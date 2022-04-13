package cloudprovider

import "strings"

//go:generate stringer -type=CloudProvider

// Provider is cloud provider used by the CLI.
type Provider uint32

const (
	Unknown Provider = iota
	AWS
	Azure
	GCP
)

// FromString returns cloud provider from string.
func FromString(s string) Provider {
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
