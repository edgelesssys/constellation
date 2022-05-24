// Clidocgen generates a Markdown page describing all CLI commands.
package main

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/edgelesssys/constellation/cli/cmd"
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
	for _, c := range rootCmd.Commands() {
		name := c.Name()
		fmt.Fprintf(cmdList, "* [%v](#constellation-%v): %v\n", name, name, c.Short)
		if err := doc.GenMarkdown(c, body); err != nil {
			panic(err)
		}
	}

	// Remove "see also" sections. They list parent and child commands, which is not interesting for us.
	cleanedBody := seeAlsoRegexp.ReplaceAll(body.Bytes(), nil)

	fmt.Printf("Commands:\n\n%s\n%s", cmdList, cleanedBody)
}
