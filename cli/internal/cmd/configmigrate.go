/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/config/migration"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newConfigMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate a configuration file to a new version",
		Long:  "Migrate a configuration file to a new version.",
		Args:  cobra.NoArgs,
		RunE:  runConfigMigrate,
	}
	cmd.Flags().StringP("file", "f", constants.ConfigFilename, "path to output file, or '-' for stdout")

	return cmd
}

func runConfigMigrate(cmd *cobra.Command, _ []string) error {
	handler := file.NewHandler(afero.NewOsFs())
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("parsing config path flag: %w", err)
	}
	return configMigrate(cmd, configPath, handler)
}

func configMigrate(cmd *cobra.Command, configPath string, handler file.Handler) error {
	// Make sure we are reading a v2 config
	var cfgVersion struct {
		Version string `yaml:"version"`
	}
	if err := handler.ReadYAML(configPath, &cfgVersion); err != nil {
		return err
	}

	switch cfgVersion.Version {
	case config.Version3:
		cmd.Printf("Config already at version %s, nothing to do", config.Version3)
		return nil
	case migration.Version2:
		if err := migration.V2ToV3(configPath, handler); err != nil {
			return fmt.Errorf("migrating config: %w", err)
		}
		cmd.Printf("Successfully migrated config to %s\n", config.Version3)
		return nil
	default:
		return fmt.Errorf("cannot convert config version %s to %s", cfgVersion.Version, config.Version3)
	}
}
