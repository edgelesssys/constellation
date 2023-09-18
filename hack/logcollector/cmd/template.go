/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/hack/logcollector/fields"
	"github.com/edgelesssys/constellation/v2/hack/logcollector/internal"
	"github.com/spf13/cobra"
)

func newTemplateCmd() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Templates filebeat and logstash configurations and prepares them for deployment",
		Long:  `Templates filebeat and logstash configurations and prepares them for deployment by placing them in the specified directory.`,
		RunE:  runTemplate,
	}

	templateCmd.Flags().String("dir", "", "Directory to place the templated configurations in (required)")
	must(templateCmd.MarkFlagRequired("dir"))
	must(templateCmd.MarkFlagDirname("dir"))
	templateCmd.Flags().String("username", "", "OpenSearch username (required)")
	must(templateCmd.MarkFlagRequired("username"))
	templateCmd.Flags().String("password", "", "OpenSearch password (required)")
	must(templateCmd.MarkFlagRequired("password"))
	templateCmd.Flags().String("index-prefix", "systemd-logs", "Prefix for logging index (e.g. systemd-logs)")
	templateCmd.Flags().Int("port", 5045, "Logstash port")
	templateCmd.Flags().StringToString("fields", nil, "Additional fields for the Logstash pipeline")

	return templateCmd
}

func runTemplate(cmd *cobra.Command, _ []string) error {
	flags, err := parseTemplateFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse template flags: %w", err)
	}

	if err := flags.extraFields.Check(); err != nil {
		return fmt.Errorf("validating extra fields: %w", err)
	}

	logstashPreparer := internal.NewLogstashPreparer(
		flags.extraFields,
		flags.username,
		flags.password,
		flags.indexPrefix,
		flags.port,
	)
	if err := logstashPreparer.Prepare(flags.dir); err != nil {
		return fmt.Errorf("prepare logstash: %w", err)
	}

	filebeatPreparer := internal.NewFilebeatPreparer(
		flags.port,
	)
	if err := filebeatPreparer.Prepare(flags.dir); err != nil {
		return fmt.Errorf("prepare filebeat: %w", err)
	}

	metricbeatPreparer := internal.NewMetricbeatPreparer(
		flags.port,
	)
	if err := metricbeatPreparer.Prepare(flags.dir); err != nil {
		return fmt.Errorf("prepare metricbeat: %w", err)
	}

	return nil
}

func parseTemplateFlags(cmd *cobra.Command) (templateFlags, error) {
	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse dir string: %w", err)
	}

	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse username string: %w", err)
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse password string: %w", err)
	}

	indexPrefix, err := cmd.Flags().GetString("index-prefix")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse index-prefix string: %w", err)
	}

	extraFields, err := cmd.Flags().GetStringToString("fields")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse fields map: %w", err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse port int: %w", err)
	}

	return templateFlags{
		dir:         dir,
		username:    username,
		password:    password,
		indexPrefix: indexPrefix,
		extraFields: extraFields,
		port:        port,
	}, nil
}

type templateFlags struct {
	dir         string
	username    string
	password    string
	indexPrefix string
	extraFields fields.Fields
	port        int
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
