/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package externalcontrollers contains a controllers that reconcile on cloud provider infrastructure.
// It is used to create, delete and update the spec of infrastructure-related k8s resources based on the
// actual state of the infrastructure.
// It uses polling (with additional triggers) to check the state of the infrastructure.
package externalcontrollers
