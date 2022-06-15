package attestationtypes

// ID holds the identifiers of a node.
type ID struct {
	Cluster []byte `json:"cluster"`
	Owner   []byte `json:"owner"`
}
