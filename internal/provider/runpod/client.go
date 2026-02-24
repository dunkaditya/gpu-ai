// Package runpod implements the Provider interface for RunPod GPU Cloud.
package runpod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gpuai/gpuctl/internal/provider"
)

const defaultBaseURL = "https://api.runpod.io/graphql"

// Client is a GraphQL HTTP client for the RunPod API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures the RunPod client.
type ClientOption func(*Client)

// WithBaseURL overrides the default RunPod GraphQL API endpoint.
// Used for testing with httptest servers.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient creates a new RunPod GraphQL client.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// graphQLRequest is the standard GraphQL request payload.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLResponse is the standard GraphQL response envelope.
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

// graphQLError represents a single GraphQL error.
type graphQLError struct {
	Message string `json:"message"`
}

// do sends a GraphQL request and unmarshals the response data into result.
// It checks for GraphQL errors in the response body (RunPod returns HTTP 200 even on errors).
func (c *Client) do(ctx context.Context, req graphQLRequest, result any) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("read response: %w", err)
	}

	// Return the HTTP response for status code inspection by retry logic.
	if resp.StatusCode >= 400 {
		return resp, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return resp, fmt.Errorf("unmarshal response: %w", err)
	}

	// CRITICAL: Check GraphQL errors first -- RunPod returns HTTP 200 even on errors.
	if len(gqlResp.Errors) > 0 {
		errMsg := gqlResp.Errors[0].Message
		// Check for capacity-related errors.
		lower := strings.ToLower(errMsg)
		if strings.Contains(lower, "capacity") ||
			strings.Contains(lower, "insufficient") ||
			strings.Contains(lower, "no available") {
			return resp, fmt.Errorf("%s: %w", errMsg, provider.ErrNoCapacity)
		}
		return resp, fmt.Errorf("graphql error: %s", errMsg)
	}

	if result != nil && len(gqlResp.Data) > 0 {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return resp, fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return resp, nil
}

// doWithRetry executes a GraphQL request with retry logic for transient errors.
// It retries up to maxAttempts times with exponential backoff for 5xx and network errors,
// respects Retry-After headers on 429, and returns immediately for non-retryable errors.
func (c *Client) doWithRetry(ctx context.Context, op string, maxAttempts int, req graphQLRequest, result any) error {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			slog.Info("retrying RunPod API call",
				"op", op,
				"attempt", attempt+1,
				"max_attempts", maxAttempts,
			)
		}

		resp, err := c.do(ctx, req, result)
		if err == nil {
			return nil
		}
		lastErr = err

		// Don't retry capacity errors -- capacity won't appear in seconds.
		if isCapacityError(err) {
			return err
		}

		// Don't retry context cancellation.
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Determine if retryable based on HTTP status code.
		if resp != nil {
			switch {
			case resp.StatusCode == http.StatusTooManyRequests:
				// Respect Retry-After header.
				delay := parseRetryAfter(resp.Header.Get("Retry-After"))
				slog.Warn("RunPod rate limited",
					"op", op,
					"retry_after_seconds", delay.Seconds(),
				)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
				continue
			case resp.StatusCode >= 500:
				// Server error -- retryable with exponential backoff.
			case resp.StatusCode >= 400 && resp.StatusCode < 500:
				// Client error (not 429) -- not retryable.
				return err
			}
		}

		// Check if the error is a non-retryable GraphQL error (not a server/network issue).
		if resp != nil && resp.StatusCode == http.StatusOK && !isRetryableGraphQLError(err) {
			return err
		}

		// Exponential backoff: 1s, 2s, 4s.
		delay := time.Duration(1<<uint(attempt)) * time.Second
		slog.Warn("RunPod API call failed, backing off",
			"op", op,
			"attempt", attempt+1,
			"delay", delay,
			"error", err,
		)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", op, maxAttempts, lastErr)
}

// isCapacityError checks if the error is a capacity-related error.
func isCapacityError(err error) bool {
	return err != nil && strings.Contains(err.Error(), provider.ErrNoCapacity.Error())
}

// isRetryableGraphQLError checks if a GraphQL error is transient and worth retrying.
func isRetryableGraphQLError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// Server-side transient errors that may resolve on retry.
	return strings.Contains(msg, "internal server error") ||
		strings.Contains(msg, "service unavailable") ||
		strings.Contains(msg, "timeout")
}

// parseRetryAfter parses a Retry-After header value.
// Returns a default of 5 seconds if the header is missing or unparseable.
func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return 5 * time.Second
	}
	seconds, err := strconv.Atoi(val)
	if err != nil {
		return 5 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// --- GraphQL Queries and Mutations ---

const queryGPUTypes = `
query GpuTypes {
  gpuTypes {
    id
    displayName
    memoryInGb
    secureCloud
    communityCloud
    securePrice
    communityPrice
    secureSpotPrice
    communitySpotPrice
    lowestPrice(input: { gpuCount: 1 }) {
      minimumBidPrice
      uninterruptablePrice
      minVcpu
      minMemory
      stockStatus
      maxUnreservedGpuCount
    }
  }
}
`

const mutationCreateOnDemandPod = `
mutation createPod($input: PodFindAndDeployOnDemandInput!) {
  podFindAndDeployOnDemand(input: $input) {
    id
    costPerHr
    desiredStatus
    lastStatusChange
    machine {
      gpuDisplayName
      location
    }
  }
}
`

const mutationCreateSpotPod = `
mutation createSpotPod($input: PodRentInterruptableInput!) {
  podRentInterruptable(input: $input) {
    id
    costPerHr
    desiredStatus
    lastStatusChange
    machine {
      gpuDisplayName
      location
    }
  }
}
`

const queryGetPod = `
query Pod($input: PodFilter!) {
  pod(input: $input) {
    id
    desiredStatus
    lastStatusChange
    costPerHr
    runtime {
      uptimeInSeconds
      ports {
        ip
        isIpPublic
        privatePort
        publicPort
        type
      }
    }
    machine {
      gpuDisplayName
      location
    }
  }
}
`

const mutationTerminatePod = `
mutation terminatePod($input: PodTerminateInput!) {
  podTerminate(input: $input)
}
`
