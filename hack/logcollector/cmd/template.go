/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/hack/logcollector/internal"
	"github.com/spf13/cobra"
)

type infoMap map[string]string

var defaultInfoMap = infoMap{
	"is-debug-cluster": "false",
}

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
	templateCmd.Flags().Int("port", 5045, "Logstash port")
	templateCmd.Flags().StringToString("info", nil, "Additional fields for the Logstash pipeline in the format --info key1=value1,key2=value2,...")

	return templateCmd
}

func runTemplate(cmd *cobra.Command, _ []string) error {
	flags, err := parseTemplateFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse template flags: %w", err)
	}

	logstashPreparer := internal.NewLogstashPreparer(
		defaultInfoMap.Extend(flags.extraInfo),
		flags.username,
		flags.password,
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

	return nil
}

func (m infoMap) Extend(other infoMap) infoMap {
	for k, v := range other {
		m[k] = v
	}
	return m
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

	extraInfo, err := cmd.Flags().GetStringToString("info")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse info map: %w", err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return templateFlags{}, fmt.Errorf("parse port int: %w", err)
	}

	return templateFlags{
		dir:       dir,
		username:  username,
		password:  password,
		extraInfo: infoMap(extraInfo),
		port:      port,
	}, nil
}

type templateFlags struct {
	dir       string
	username  string
	password  string
	extraInfo infoMap
	port      int
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
