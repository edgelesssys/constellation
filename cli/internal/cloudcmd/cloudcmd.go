/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cloudcmd provides executable command for the CLI.

This package focuses on the interaction with the cloud provider.
It separates the cloud provider specific code from the rest of the CLI, and
provides a common interface for all cloud providers.

Exported functions must not be cloud provider specific, but rather take a
cloudprovider.Provider as an argument.

User interaction happens in the cmd package, and should not happen or pass through
this package.

The backend to this package is currently provided by the terraform package.
*/
package cloudcmd
