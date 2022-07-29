package client

import (
	"errors"
	"regexp"

	"github.com/spf13/afero"
)

var projectIDRegex = regexp.MustCompile(`(?m)^project-id = (.*)$`)

// loadProjectID loads the project id from the gce config file.
func loadProjectID(fs afero.Fs, path string) (string, error) {
	rawConfig, err := afero.ReadFile(fs, path)
	if err != nil {
		return "", err
	}
	// find project-id line
	matches := projectIDRegex.FindStringSubmatch(string(rawConfig))
	if len(matches) != 2 {
		return "", errors.New("invalid config: project-id not found")
	}
	return matches[1], nil
}
