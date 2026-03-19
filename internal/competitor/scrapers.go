package competitor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// fetchDoc fetches a URL and returns a goquery document.
func fetchDoc(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

// ScrapeLambda scrapes lambda.ai/pricing for GPU per-hour prices.
// Lambda uses HTML tables with data-plan attribute on <tr> and
// data-label="PRICE/GPU/HR*" on the price <td>.
func ScrapeLambda(ctx context.Context) (Prices, error) {
	doc, err := fetchDoc(ctx, "https://lambda.ai/pricing")
	if err != nil {
		return nil, fmt.Errorf("lambda: %w", err)
	}

	prices := make(Prices)

	// Scan all table rows that have a data-plan attribute (GPU name).
	doc.Find("tr[data-plan]").Each(func(i int, row *goquery.Selection) {
		plan, _ := row.Attr("data-plan")
		model := matchLambdaGPU(plan)
		if model == "" {
			return
		}

		// Find price cell by data-label containing "PRICE".
		var priceText string
		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			label, exists := cell.Attr("data-label")
			if exists && strings.Contains(strings.ToUpper(label), "PRICE") {
				priceText = strings.TrimSpace(cell.Text())
			}
		})

		if priceText == "" {
			return
		}

		price := parsePrice(priceText)
		if price <= 0 || price > 100 {
			return
		}

		// Keep lowest price per model (clusters may be cheaper than single instances).
		if existing, ok := prices[model]; !ok || price < existing {
			prices[model] = price
		}
	})

	// Also scan for cluster pricing sections that may not use data-plan.
	// Look for any table row containing a known GPU name and a price.
	doc.Find("table tr").Each(func(i int, row *goquery.Selection) {
		text := row.Text()
		for _, keyword := range []string{"H100", "H200", "B200", "A100", "L40S"} {
			if !strings.Contains(text, keyword) {
				continue
			}
			// Already captured via data-plan? Skip if so.
			if _, exists := row.Attr("data-plan"); exists {
				break
			}
			// Try to extract GPU name and price from row text.
			model := matchLambdaGPU(text)
			if model == "" {
				break
			}
			price := parsePrice(text)
			if price <= 0 || price > 100 {
				break
			}
			if existing, ok := prices[model]; !ok || price < existing {
				prices[model] = price
			}
			break
		}
	})

	if len(prices) == 0 {
		return nil, fmt.Errorf("lambda: no prices found (page structure may have changed)")
	}

	slog.Debug("scraped lambda prices", "count", len(prices))
	return prices, nil
}

// ScrapeCoreWeave scrapes coreweave.com/pricing for GPU per-hour prices.
// CoreWeave uses a Webflow div grid with .table-row class.
// The instance-price cell contains the total node price; we divide by GPU count.
func ScrapeCoreWeave(ctx context.Context) (Prices, error) {
	doc, err := fetchDoc(ctx, "https://www.coreweave.com/pricing")
	if err != nil {
		return nil, fmt.Errorf("coreweave: %w", err)
	}

	prices := make(Prices)

	doc.Find(".table-row-v2").Each(func(i int, row *goquery.Selection) {
		// GPU name from the --name cell.
		gpuName := strings.TrimSpace(row.Find(".table-v2-cell--name").Text())
		model := matchCoreWeaveGPU(gpuName)
		if model == "" {
			return
		}

		// GPU count: first plain .table-v2-cell (not --name, not --vram) that contains a small number.
		gpuCount := 8 // default
		row.Find(".table-v2-cell").Each(func(j int, cell *goquery.Selection) {
			t := strings.TrimSpace(cell.Text())
			// Clean superscript markers like "8^1"
			if idx := strings.Index(t, "^"); idx > 0 {
				t = t[:idx]
			}
			if n, err := strconv.Atoi(t); err == nil && n >= 1 && n <= 16 {
				gpuCount = n
			}
		})

		// Instance (on-demand) price.
		priceText := strings.TrimSpace(row.Find(".instance-price .item-value").Text())
		if priceText == "" || strings.EqualFold(priceText, "N/A") {
			return
		}

		totalPrice := parsePrice(priceText)
		if totalPrice <= 0 {
			return
		}

		perGPU := totalPrice / float64(gpuCount)
		if perGPU > 50 {
			return // sanity check
		}

		if existing, ok := prices[model]; !ok || perGPU < existing {
			prices[model] = perGPU
		}
	})

	if len(prices) == 0 {
		return nil, fmt.Errorf("coreweave: no prices found (page structure may have changed)")
	}

	slog.Debug("scraped coreweave prices", "count", len(prices))
	return prices, nil
}

// ScrapeAWS scrapes aws-pricing.com/gpu.html for GPU per-hour prices.
// The page has a standard HTML table with columns:
// [instance_type, vCPUs, clock, RAM, GPU_count, GPU_vendor, GPU_model, GPU_mem, ..., on_demand_price, ...]
func ScrapeAWS(ctx context.Context) (Prices, error) {
	doc, err := fetchDoc(ctx, "https://aws-pricing.com/gpu.html")
	if err != nil {
		return nil, fmt.Errorf("aws: %w", err)
	}

	prices := make(Prices)

	doc.Find("table tbody tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 10 {
			return
		}

		instanceType := strings.TrimSpace(cells.Eq(0).Text())
		gpuCountStr := strings.TrimSpace(cells.Eq(4).Text())
		gpuModel := strings.TrimSpace(cells.Eq(6).Text())
		onDemandStr := strings.TrimSpace(cells.Eq(9).Text())

		model := mapAWSInstance(instanceType, gpuModel)
		if model == "" {
			return
		}

		gpuCount, err := strconv.Atoi(gpuCountStr)
		if err != nil || gpuCount <= 0 {
			return
		}

		totalPrice := parsePriceRaw(onDemandStr)
		if totalPrice <= 0 {
			return
		}

		perGPU := totalPrice / float64(gpuCount)
		if perGPU > 50 {
			return
		}

		// Keep lowest per-GPU price (smaller instances may have different per-GPU economics).
		if existing, ok := prices[model]; !ok || perGPU < existing {
			prices[model] = perGPU
		}
	})

	if len(prices) == 0 {
		return nil, fmt.Errorf("aws: no prices found (page structure may have changed)")
	}

	slog.Debug("scraped aws prices", "count", len(prices))
	return prices, nil
}
