/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# AttestationConfig API

The AttestationConfig API provides values for the attestation key in the Constellation config.

This package defines API types that represents objects of the AttestationConfig API.
The types provide helper methods for validation and commonly used operations on the
information contained in the objects. Especially the paths used for the API are defined
in these helper methods.

Regarding the decision to implement new types over using the existing types from internal/config:
AttestationCfg objects for AttestationCfg API need to hold some version information (for sorting, recognizing latest).
Thus, existing config types (AWSNitroTPM, AzureSEVSNP, ...) can not be extended to implement apiObject interface.
Instead, we need a separate type that wraps _all_ attestation types. In the codebase this is done using the AttestationCfg interface.
The new type AttestationCfgGet needs to be located inside internal/config in order to implement UnmarshalJSON.
*/
package attestationconfigapi
