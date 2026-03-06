package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gpuai/gpuctl/internal/provider"
	"github.com/gpuai/gpuctl/internal/provision"
)

func runProvision(args []string) {
	fs := flag.NewFlagSet("provision", flag.ExitOnError)

	gpuType := fs.String("gpu-type", "h100_sxm", "GPU type (e.g. h100_sxm, a100_80gb)")
	gpuCount := fs.Int("gpu-count", 1, "Number of GPUs")
	tier := fs.String("tier", "on_demand", "Instance tier: on_demand, spot")
	region := fs.String("region", "", "Region filter (optional)")
	maxPrice := fs.Float64("max-price", 0, "Maximum price per hour (0 = no limit)")
	orgID := fs.String("org-id", "debug-org", "Organization ID")
	userID := fs.String("user-id", "debug-user", "User ID")
	sshKeyIDs := fs.String("ssh-key-ids", "", "Comma-separated SSH key UUIDs")
	dryRun := fs.Bool("dry-run", false, "Show candidates without provisioning")
	yes := fs.Bool("yes", false, "Skip approval prompt")
	fs.Parse(args)

	// Always use text logger at DEBUG level for CLI output.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Debug("initializing infrastructure")
	deps := setupCommonDeps(ctx, logger)
	defer deps.DB.Close()
	defer deps.Redis.Close()

	// Build provision request.
	req := provision.ProvisionRequest{
		OrgID:    *orgID,
		UserID:   *userID,
		GPUType:  provider.GPUType(*gpuType),
		GPUCount: *gpuCount,
		Tier:     provider.InstanceTier(*tier),
		Region:   *region,
	}
	if *maxPrice > 0 {
		req.MaxPricePerHour = maxPrice
	}
	if *sshKeyIDs != "" {
		req.SSHKeyIDs = strings.Split(*sshKeyIDs, ",")
	}

	// Query provider candidates.
	fmt.Println()
	slog.Debug("querying provider candidates",
		"gpu_type", *gpuType,
		"gpu_count", *gpuCount,
		"tier", *tier,
		"region", *region,
	)

	candidates, err := deps.Engine.SelectProviderCandidates(ctx, req)
	if err != nil {
		slog.Error("no candidates found", "error", err)
		os.Exit(1)
	}

	// Display candidates.
	fmt.Printf("\n=== Provider Candidates (%d found) ===\n\n", len(candidates))
	fmt.Printf("%-4s %-12s %-12s %-10s %-8s %-10s %s\n",
		"#", "Provider", "GPU", "Count", "Tier", "Region", "$/hr")
	fmt.Println(strings.Repeat("-", 70))

	for i, c := range candidates {
		fmt.Printf("%-4d %-12s %-12s %-10d %-8s %-10s $%.4f\n",
			i+1,
			c.Prov.Name(),
			c.Offering.GPUType,
			c.Offering.GPUCount,
			c.Offering.Tier,
			c.Offering.Region,
			c.Offering.PricePerHour,
		)
	}
	fmt.Println()

	if *dryRun {
		slog.Info("dry run complete, exiting")
		return
	}

	// Approval prompt.
	best := candidates[0]
	fmt.Printf("Selected: %s @ $%.4f/hr from %s (%s)\n",
		best.Offering.GPUType,
		best.Offering.PricePerHour,
		best.Prov.Name(),
		best.Offering.Region,
	)

	if !*yes {
		fmt.Print("\nProceed with provisioning? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return
		}
	}

	// Provision.
	fmt.Println()
	slog.Info("starting provisioning")
	start := time.Now()

	resp, err := deps.Engine.Provision(ctx, req)
	if err != nil {
		slog.Error("provisioning failed", "error", err)
		os.Exit(1)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n=== Instance Provisioned ===\n")
	fmt.Printf("  ID:       %s\n", resp.InstanceID)
	fmt.Printf("  Hostname: %s\n", resp.Hostname)
	fmt.Printf("  Price:    $%.4f/hr\n", resp.PricePerHour)
	fmt.Printf("  Status:   %s\n", resp.Status)
	fmt.Printf("  Elapsed:  %s\n", elapsed.Round(time.Millisecond))

	// Poll status until running or error (5 min timeout).
	fmt.Println("\nPolling status...")
	pollCtx, pollCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer pollCancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			slog.Warn("status polling timed out")
			return
		case <-ticker.C:
			inst, err := deps.DB.GetInstance(ctx, resp.InstanceID)
			if err != nil {
				slog.Warn("failed to check status", "error", err)
				continue
			}
			slog.Debug("instance status", "status", inst.Status)

			switch inst.Status {
			case "running":
				fmt.Printf("\nInstance %s is RUNNING\n", resp.InstanceID)
				fmt.Printf("  SSH: ssh root@%s\n", resp.Hostname)
				return
			case "error", "terminated":
				slog.Error("instance entered terminal state", "status", inst.Status)
				os.Exit(1)
			}
		}
	}
}
