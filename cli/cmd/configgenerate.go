package cmd

import (
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/talos-systems/talos/pkg/machinery/config/encoder"
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
		content, err := encoder.NewEncoder(config.Default()).Encode()
		if err != nil {
			return err
		}
		_, err = cmd.OutOrStdout().Write(content)
		return err
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
