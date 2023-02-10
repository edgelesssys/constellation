/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
The api module is used to bundle the Constellation operator API into a separate module.
This allows us to use the types from the API in other places than the operators themself, without having modules mutually depend on each other.
We can also publish this API more easily if we decide to do so.

Model for this approach is the Kubernetes API itself: https://pkg.go.dev/k8s.io/api#section-readme
*/

package api
