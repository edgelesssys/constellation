/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package upgrade tests that the CLI's apply command works as expected and
// the operators eventually upgrade all nodes inside the cluster.
// The test is written as a go test because:
// 1. the helm cli does not directly provide the chart version of a release
//
// 2. the image patch needs to be parsed from the image-api's info.json
//
// 3. there is some logic required to setup the test correctly:
//
//   - set microservice, k8s version correctly depending on input
//
//   - set or fetch measurements depending on target image
package upgrade
