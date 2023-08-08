/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cloudcmd provides executable commands for the CLI.

This package focuses on the interaction with the cloud provider.
It separates the cloud provider specific code from the rest of the CLI, and
provides a common interface for all cloud providers.

Exported functions must not be cloud provider specific, but rather take a
cloudprovider.Provider as an argument, perform CSP specific logic, and return a universally usable result.

It is used by the "cmd" to handle creation of cloud resources and other CSP specific interactions.
User interaction happens in the "cmd" package, and should not happen or pass through
this package.

The backend to this package is currently provided by the terraform package.
*/
package cloudcmd
