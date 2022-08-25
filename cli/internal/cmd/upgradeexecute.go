package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newUpgradeExecuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute an upgrade of a Constellation cluster",
		Long:  "Execute an upgrade of a Constellation cluster by applying the chosen configuration.",
		Args:  cobra.NoArgs,
		RunE:  runUpgradeExecute,
	}

	return cmd
}

func runUpgradeExecute(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	upgrader, err := cloudcmd.NewUpgrader(cmd.OutOrStdout())
	if err != nil {
		return err
	}

	return upgradeExecute(cmd, upgrader, fileHandler)
}

func upgradeExecute(cmd *cobra.Command, upgrader cloudUpgrader, fileHandler file.Handler) error {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	config, err := config.FromFile(fileHandler, configPath)
	if err != nil {
		return err
	}

	// TODO: validate upgrade config? Should be basic things like checking image is not an empty string
	// More sophisticated validation, like making sure we don't downgrade the cluster, should be done by `constellation upgrade plan`

	return upgrader.Upgrade(cmd.Context(), config.Upgrade.Image, config.Upgrade.Measurements)
}

type cloudUpgrader interface {
	Upgrade(ctx context.Context, image string, measurements map[uint32][]byte) error
}
