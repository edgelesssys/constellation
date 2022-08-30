package role

import (
	"encoding/json"
	"strings"
)

//go:generate stringer -type=Role

// Role is a peer's role.
type Role uint

const (
	Unknown Role = iota
	ControlPlane
	Worker
	Admin
)

// MarshalJSON marshals the Role to JSON string.
func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON unmarshals the Role from JSON string.
func (r *Role) UnmarshalJSON(b []byte) error {
	var roleString string
	if err := json.Unmarshal(b, &roleString); err != nil {
		return err
	}
	switch strings.ToLower(roleString) {
	case "controlplane":
		*r = ControlPlane
	case "worker":
		*r = Worker
	case "admin":
		*r = Admin
	default:
		*r = Unknown
	}
	return nil
}
