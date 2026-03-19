// Package competitor scrapes GPU pricing from Lambda, CoreWeave, and AWS.
package competitor

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Prices maps canonical GPU model (e.g. "h100_sxm") to per-GPU-per-hour price.
type Prices map[string]float64

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var httpClient = &http.Client{Timeout: 30 * time.Second}

var priceRe = regexp.MustCompile(`\$(\d{1,3}(?:,\d{3})*(?:\.\d+)?)`)

// parsePrice extracts the first dollar amount from a string.
func parsePrice(s string) float64 {
	m := priceRe.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	cleaned := strings.ReplaceAll(m[1], ",", "")
	v, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return v
}

// parsePriceRaw parses a numeric string (no $ sign) into a float.
func parsePriceRaw(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, "$")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// matchLambdaGPU maps a Lambda GPU name to our canonical model.
func matchLambdaGPU(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(lower, "b200"):
		return "b200"
	case strings.Contains(lower, "h200") && !strings.Contains(lower, "gh200"):
		return "h200_sxm"
	case strings.Contains(lower, "gh200"):
		return "" // GH200 is a different product, skip
	case strings.Contains(lower, "h100 sxm"), strings.Contains(lower, "h100sxm"):
		return "h100_sxm"
	case strings.Contains(lower, "h100"):
		return "h100_pcie"
	case strings.Contains(lower, "a100"):
		// Lambda's A100 is 40GB — not the same as our a100_80gb
		return "a100_40gb"
	case strings.Contains(lower, "a6000"):
		return "rtx_a6000"
	case strings.Contains(lower, "l40s"):
		return "l40s"
	case strings.Contains(lower, "a10") && !strings.Contains(lower, "a100"):
		return "a10"
	}
	return ""
}

// matchCoreWeaveGPU maps a CoreWeave GPU name to our canonical model.
func matchCoreWeaveGPU(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(lower, "h200"):
		return "h200_sxm"
	case strings.Contains(lower, "h100"):
		return "h100_sxm"
	case strings.Contains(lower, "gb200"):
		// GB200 NVL72 is a different product, skip
		return ""
	case strings.Contains(lower, "b200"):
		return "b200"
	case strings.Contains(lower, "a100"):
		return "a100_80gb" // CoreWeave A100 is 80GB
	case strings.Contains(lower, "l40s"):
		return "l40s"
	case strings.Contains(lower, "rtx 4090"):
		return "rtx_4090"
	}
	return ""
}

// mapAWSInstance maps an AWS instance type + GPU model to our canonical model.
func mapAWSInstance(instanceType, gpuModel string) string {
	it := strings.ToLower(instanceType)
	switch {
	case strings.HasPrefix(it, "p5en"):
		return "h200_sxm"
	case strings.HasPrefix(it, "p5."):
		return "h100_sxm"
	case strings.HasPrefix(it, "p6") && strings.Contains(strings.ToLower(gpuModel), "b200"):
		return "b200"
	case strings.HasPrefix(it, "p4de"):
		return "a100_80gb"
	case strings.HasPrefix(it, "p4d."):
		return "a100_40gb"
	case strings.HasPrefix(it, "g6e"):
		return "l40s"
	}
	return ""
}
