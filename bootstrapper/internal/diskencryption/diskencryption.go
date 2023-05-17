/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package diskencryption handles interaction with a node's state disk.

This package is not thread safe, since libcryptsetup is not thread safe.
There should only be one instance using this package per process.
*/
package diskencryption
