package role

//go:generate stringer -type=Role

// Role is a peer's role.
type Role uint

const (
	Unknown Role = iota
	Coordinator
	Node
)
