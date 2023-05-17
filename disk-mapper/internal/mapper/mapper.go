/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package mapper uses libcryptsetup to format and map crypt devices.

This is used by the disk-mapper to set up a node's state disk.

All interaction with libcryptsetup should be done here.

Warning: This package is not thread safe, since libcryptsetup is not thread safe.
*/
package mapper
