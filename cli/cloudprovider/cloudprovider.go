package cloudprovider

//go:generate stringer -type=CloudProvider

// CloudProvider is cloud provider used by the CLI.
type CloudProvider uint32

const (
	Unknown CloudProvider = iota
	AWS
	Azure
	GCP
)
