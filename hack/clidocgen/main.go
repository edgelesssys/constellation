/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Clidocgen generates a Markdown page describing all CLI commands.
package main

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/edgelesssys/constellation/v2/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var seeAlsoRegexp = regexp.MustCompile(`(?s)### SEE ALSO\n.+?\n\n`)

func main() {
	cobra.EnableCommandSorting = false
	rootCmd := cmd.NewRootCmd()
	rootCmd.DisableAutoGenTag = true

	// Generate Markdown for all commands.
	cmdList := &bytes.Buffer{}
	body := &bytes.Buffer{}
	for _, c := range allSubCommands(rootCmd) {
		if c.Hidden {
			continue
		}
		name := c.Name()
		fullName, level := determineFullNameAndLevel(c)

		// First two arguments are used to create indentation for nested commands (2 spaces per level).
		fmt.Fprintf(cmdList, "%*s* [%v](#constellation-%v): %v\n", 2*level, "", name, fullName, c.Short)
		if err := doc.GenMarkdown(c, body); err != nil {
			panic(err)
		}
	}

	// Remove "see also" sections. They list parent and child commands, which is not interesting for us.
	cleanedBody := seeAlsoRegexp.ReplaceAll(body.Bytes(), nil)

	fmt.Printf("Commands:\n\n%s\n%s", cmdList, cleanedBody)
}

func allSubCommands(cmd *cobra.Command) []*cobra.Command {
	var all []*cobra.Command
	for _, c := range cmd.Commands() {
		all = append(all, c)
		all = append(all, allSubCommands(c)...)
	}
	return all
}

func determineFullNameAndLevel(cmd *cobra.Command) (string, int) {
	// Traverse the command tree upwards and determine the full name and level of the command.
	name := cmd.Name()
	level := 0
	for cmd.HasParent() && cmd.Parent().Name() != "constellation" {
		cmd = cmd.Parent()
		name = cmd.Name() + "-" + name // Use '-' as separator since we pipe it into a Markdown link.
		level++
	}
	return name, level
}
