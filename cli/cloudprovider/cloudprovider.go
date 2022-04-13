package cloudprovider

import "strings"

//go:generate stringer -type=CloudProvider

// CloudProvider is cloud provider used by the CLI.
type CloudProvider uint32

const (
	Unknown CloudProvider = iota
	AWS
	Azure
	GCP
)

// FromString returns cloud provider from string.
func FromString(s string) CloudProvider {
	s = strings.ToLower(s)
	switch s {
	case "aws":
		return AWS
	case "azure":
		return Azure
	case "gcp":
		return GCP
	default:
		return Unknown
	}
}
