// Package v1beta1 contains API Schema definitions for the nodemaintenance v1beta1 API group
package v1beta1

// This file contains a go generate directive to download the API definition from the source (https://github.com/medik8s/node-maintenance-operator).
// We vendor the API definition to avoid a dependency on the rest of the node-maintenance-operator codebase.

//go:generate bash ./download.sh https://github.com/medik8s/node-maintenance-operator/archive/refs/tags/v0.14.0.tar.gz node-maintenance-operator-0.14.0/api/v1beta1 048323ffdb55787df9b93d85be93e4730f4495fba81b440dc6fe195408ec2533
