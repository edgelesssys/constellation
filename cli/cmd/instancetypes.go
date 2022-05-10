package cmd

import "strings"

func formatInstanceTypes(types []string) string {
	return "  " + strings.Join(types, "\n  ")
}
