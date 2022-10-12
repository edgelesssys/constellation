/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

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

// FromString converts a string to a Role.
func FromString(s string) Role {
	switch strings.ToLower(s) {
	case "controlplane":
		return ControlPlane
	case "worker":
		return Worker
	case "admin":
		return Admin
	default:
		return Unknown
	}
}
