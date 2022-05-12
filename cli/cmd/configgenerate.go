package cmd

import (
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a default configuration file",
		Long:  "Generate a default configuration file for your selected cloud provider.",
		Args:  cobra.ExactArgs(0),
		RunE:  runConfigGenerate,
	}
	cmd.Flags().StringP("file", "f", constants.ConfigFilename, "output file")

	return cmd
}

type generateFlags struct {
	file string
}

func runConfigGenerate(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	return configGenerate(cmd, fileHandler)
}

func configGenerate(cmd *cobra.Command, fileHandler file.Handler) error {
	flags, err := parseGenerateFlags(cmd)
	if err != nil {
		return err
	}

	if flags.file == "-" {
		return yaml.NewEncoder(cmd.OutOrStdout()).Encode(config.Default())
	}

	return fileHandler.WriteYAML(flags.file, config.Default(), 0o644)
}

func parseGenerateFlags(cmd *cobra.Command) (generateFlags, error) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return generateFlags{}, err
	}
	return generateFlags{
		file: file,
	}, nil
}
