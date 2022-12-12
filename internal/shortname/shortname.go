/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package shortname

import (
	"fmt"
	"strings"
)

// ToParts splits an config shortname into its parts.
// The shortname is expected to be in the format of one of:
// ref/<ref>/stream/<stream>/<version>
// stream/<stream>/<version>
// version/<version>.
func ToParts(shortname string) (string, string, string, error) {
	parts := strings.Split(shortname, "/")
	ref := "-"
	stream := "stable"
	var version string
	switch len(parts) {
	case 1:
		version = parts[0]
	case 3:
		if parts[0] != "stream" {
			return "", "", "", fmt.Errorf("invalid shortname: expected \"stream/<stream>/<version>\", got %q", shortname)
		}
		stream = parts[1]
		version = parts[2]
	case 5:
		if parts[0] != "ref" || parts[2] != "stream" {
			return "", "", "", fmt.Errorf("invalid shortname: expected \"ref/<ref>/stream/<stream>/<version>\", got %q", shortname)
		}
		ref = parts[1]
		stream = parts[3]
		version = parts[4]
	default:
		return "", "", "", fmt.Errorf("invalid shortname reference %q", shortname)
	}
	return ref, stream, version, nil
}

// FromParts joins the parts of a config shortname into a shortname.
func FromParts(ref, stream, version string) string {
	switch {
	case ref == "-" && stream == "stable":
		return version
	case ref == "-":
		return "stream/" + stream + "/" + version
	}
	return "ref/" + ref + "/stream/" + stream + "/" + version
}
