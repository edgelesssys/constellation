/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/logcollection"
	"github.com/spf13/cobra"
)

func newPrepareCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepares deployment files",
		Long:  `Prepares deployment files.`,
	}

	deployCmd.PersistentFlags().String("csp", "", "(required) CSP to retrieve OpenSearch credentials from (e.g. 'azure')")
	must(cobra.MarkFlagRequired(deployCmd.PersistentFlags(), "csp"))

	deployCmd.AddCommand(newPrepareLogstashCmd())

	return deployCmd
}

func newPrepareLogstashCmd() *cobra.Command {
	prepareLogstashCmd := &cobra.Command{
		Use:   "logstash",
		Short: "Prepares Logstash deployment files",
		Long:  `Prepares Logstash deployment files.`,
		RunE:  runPrepareLogstash,
	}

	return prepareLogstashCmd
}

func runPrepareLogstash(cmd *cobra.Command, args []string) error {
	cspString := cmd.Flag("csp").Value.String()
	csp := cloudprovider.FromString(cspString)

	ls, err := logcollection.NewLogstash()
	if err != nil {
		return fmt.Errorf("create Logstash: %w", err)
	}

	// get credentials
	credGetter, err := logcollection.NewCloudCredentialGetter(cmd.Context(), csp, "")
	if err != nil {
		return fmt.Errorf("create cloud credential getter: %w", err)
	}
	creds, err := credGetter.GetOpensearchCredentials(cmd.Context())
	if err != nil {
		return fmt.Errorf("get credentials: %w", err)
	}

	// write templates
	pipelineConfInput := logcollection.NewLogstashPipelineConfInput(
		creds,
		map[string]string{},
		logcollection.LogMetadata{},
	)
	if err := ls.WritePipelineConf(pipelineConfInput, "pipeline.conf"); err != nil {
		return fmt.Errorf("write pipeline config: %w", err)
	}
	if err := ls.WriteHelmValues("pipeline.conf", "values.yml"); err != nil {
		return fmt.Errorf("write helm values: %w", err)
	}

	cmd.Println("Successfully prepared Logstash deployment files.\n`make deploy` to deploy.")
	return nil
}
