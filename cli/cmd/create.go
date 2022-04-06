package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create aws|gcp|azure",
		Short: "Create instances on a cloud platform for your Constellation.",
		Long:  "Create instances on a cloud platform for your Constellation.",
	}
	cmd.PersistentFlags().String("name", "constell", "Set this flag to create the Constellation with the specified name.")
	cmd.PersistentFlags().BoolP("yes", "y", false, "Set this flag to create the Constellation without further confirmation.")

	cmd.AddCommand(newCreateAWSCmd())
	cmd.AddCommand(newCreateGCPCmd())
	cmd.AddCommand(newCreateAzureCmd())
	return cmd
}

// checkDirClean checks if files of a previous Constellation are left in the current working dir.
func checkDirClean(fileHandler file.Handler) error {
	if _, err := fileHandler.Stat(constants.StateFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory, run 'constellation terminate' before creating a new one", constants.StateFilename)
	}
	if _, err := fileHandler.Stat(constants.AdminConfFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory, run 'constellation terminate' before creating a new one", constants.AdminConfFilename)
	}
	if _, err := fileHandler.Stat(constants.MasterSecretFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory, clean it up first", constants.MasterSecretFilename)
	}

	return nil
}
