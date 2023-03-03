/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package helm provides a higher level interface to the Helm GO SDK.

It is used by the CLI to:
- load embedded charts
- install charts
- update helm releases
- get versions for installed helm releases
- create local backups before running service upgrades
*/
package helm
