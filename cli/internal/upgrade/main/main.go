package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

func main() {
	ctx := context.Background()
	tfClient, err := terraform.New(ctx, filepath.Join(constants.UpgradeDir, "test", constants.TerraformUpgradeWorkingDir))
	if err != nil {
		panic(fmt.Errorf("setting up terraform client: %w", err))
	}
	// give me a writer
	outWriter := bytes.NewBuffer(nil)
	tfUpgrader, err := upgrade.NewTerraformUpgrader(tfClient, outWriter)
	if err != nil {
		panic(fmt.Errorf("setting up terraform upgrader: %w", err))
	}
	diff, err := tfUpgrader.PlanIAMMigration(ctx, upgrade.TerraformUpgradeOptions{
		CSP:      cloudprovider.AWS,
		LogLevel: terraform.LogLevelDebug,
	}, "test")
	if err != nil {
		panic(fmt.Errorf("planning terraform migrations: %w", err))
	}
	fmt.Println(diff)
}
