package client

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

var (
	numberedNameRegex       = regexp.MustCompile(`^(.+)-(\d+)$`)
	instanceTemplateIDRegex = regexp.MustCompile(`projects/([^/]+)/global/instanceTemplates/([^/]+)`)
)

// generateInstanceTemplateName generates a unique name for an instance template by incrementing a counter.
// The name is in the format <prefix>-<counter>.
func generateInstanceTemplateName(last string) (string, error) {
	if len(last) > 0 && last[len(last)-1] == '-' {
		return last + "1", nil
	}
	matches := numberedNameRegex.FindStringSubmatch(last)
	if len(matches) != 3 {
		return last + "-1", nil
	}
	n, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", err
	}
	if n < 1 || n == math.MaxInt {
		return "", fmt.Errorf("invalid counter: %v", n)
	}
	return matches[1] + "-" + strconv.Itoa(n+1), nil
}

// splitInstanceTemplateID splits an instance template ID into its project and name components.
func splitInstanceTemplateID(instanceTemplateID string) (project, templateName string, err error) {
	matches := instanceTemplateIDRegex.FindStringSubmatch(instanceTemplateID)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("error splitting instanceTemplateID: %v", instanceTemplateID)
	}
	return matches[1], matches[2], nil
}

// joinInstanceTemplateURI joins a project and template name into an instance template URI.
func joinInstanceTemplateURI(project, templateName string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%v/global/instanceTemplates/%v", project, templateName)
}
