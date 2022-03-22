package cmd

import (
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version of this CLI",
		Long:  `Display version of this CLI`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("CLI Version: v%s \n", config.Version)
		},
	}
	return cmd
}
