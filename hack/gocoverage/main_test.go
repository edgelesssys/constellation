/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const goTestCoverOutput = `
ok  	github.com/edgelesssys/constellation/v2/bazel/release/artifacts	(cached)	coverage: [no statements]
?   	github.com/edgelesssys/constellation/v2/bootstrapper/cmd/bootstrapper	[no test files]
?   	github.com/edgelesssys/constellation/v2/bootstrapper/initproto	[no test files]
?   	github.com/edgelesssys/constellation/v2/bootstrapper/internal/certificate	[no test files]
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/clean	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/diskencryption	(cached)	coverage: 76.9% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/initserver	(cached)	coverage: 73.7% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/joinclient	(cached)	coverage: 89.3% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/journald	(cached)	coverage: 42.1% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes	(cached)	coverage: 70.7% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi	(cached)	coverage: 8.9% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi/resources	(cached)	coverage: 22.2% of statements
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubewaiter	(cached)	coverage: 100.0% of statements
?   	github.com/edgelesssys/constellation/v2/build/metad-analyst	[no test files]
?   	github.com/edgelesssys/constellation/v2/bootstrapper/internal/logging	[no test files]
ok  	github.com/edgelesssys/constellation/v2/bootstrapper/internal/nodelock	(cached)	coverage: 75.0% of statements
?   	github.com/edgelesssys/constellation/v2/cli	[no test files]
?   	github.com/edgelesssys/constellation/v2/cli/cmd	[no test files]
ok  	github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd	(cached)	coverage: 69.1% of statements
ok  	github.com/edgelesssys/constellation/v2/cli/internal/clusterid	(cached)	coverage: 56.2% of statements
?   	github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix	[no test files]
ok  	github.com/edgelesssys/constellation/v2/cli/internal/cmd	(cached)	coverage: 54.3% of statements
?   	github.com/edgelesssys/constellation/v2/internal/constellation/featureset	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/constellation/helm/imageversion	[no test files]
?   	github.com/edgelesssys/constellation/v2/cli/internal/libvirt	[no test files]
?   	github.com/edgelesssys/constellation/v2/debugd/cmd/cdbg	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/constellation/helm	(cached)	coverage: 36.0% of statements
ok  	github.com/edgelesssys/constellation/v2/cli/internal/kubernetes	(cached)	coverage: 40.4% of statements
ok  	github.com/edgelesssys/constellation/v2/cli/internal/terraform	(cached)	coverage: 70.8% of statements
ok  	github.com/edgelesssys/constellation/v2/cli/internal/upgrade	(cached)	coverage: 66.7% of statements
ok  	github.com/edgelesssys/constellation/v2/csi/cryptmapper	(cached)	coverage: 79.2% of statements
ok  	github.com/edgelesssys/constellation/v2/csi/kms	(cached)	coverage: 70.0% of statements
?   	github.com/edgelesssys/constellation/v2/debugd/cmd/debugd	[no test files]
?   	github.com/edgelesssys/constellation/v2/debugd/internal/cdbg/cmd	[no test files]
?   	github.com/edgelesssys/constellation/v2/debugd/internal/debugd	[no test files]
?   	github.com/edgelesssys/constellation/v2/debugd/service	[no test files]
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/debugd/deploy	(cached)	coverage: 83.6% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/debugd/info	(cached)	coverage: 95.5% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/debugd/logcollector	(cached)	coverage: 15.0% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata	(cached)	coverage: 92.9% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata/cloudprovider	(cached)	coverage: 75.9% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata/fallback	(cached)	coverage: 80.0% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/debugd/server	(cached)	coverage: 71.7% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer/streamer	(cached)	coverage: 90.9% of statements
?   	github.com/edgelesssys/constellation/v2/disk-mapper/cmd	[no test files]
?   	github.com/edgelesssys/constellation/v2/disk-mapper/internal/diskencryption	[no test files]
ok  	github.com/edgelesssys/constellation/v2/disk-mapper/internal/recoveryserver	(cached)	coverage: 89.1% of statements
ok  	github.com/edgelesssys/constellation/v2/disk-mapper/internal/rejoinclient	(cached)	coverage: 91.8% of statements
ok  	github.com/edgelesssys/constellation/v2/disk-mapper/internal/setup	(cached)	coverage: 68.9% of statements
?   	github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto	[no test files]
ok  	github.com/edgelesssys/constellation/v2/disk-mapper/internal/systemd	(cached)	coverage: 25.8% of statements
?   	github.com/edgelesssys/constellation/v2/e2e	[no test files]
?   	github.com/edgelesssys/constellation/v2/e2e/internal/kubectl	[no test files]
?   	github.com/edgelesssys/constellation/v2/e2e/internal/lb	[no test files]
?   	github.com/edgelesssys/constellation/v2/e2e/internal/upgrade	[no test files]
?   	github.com/edgelesssys/constellation/v2/image/upload	[no test files]
?   	github.com/edgelesssys/constellation/v2/image/upload/internal/cmd	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/api/client	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/api/attestationconfig	(cached)	coverage: 59.2% of statements
?   	github.com/edgelesssys/constellation/v2/internal/api/fetcher	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/api/versionsapi/cli	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/api/versionsapi	(cached)	coverage: 69.8% of statements
?   	github.com/edgelesssys/constellation/v2/internal/attestation/aws	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/attestation/azure	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/atls	(cached)	coverage: 78.6% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation	(cached)	coverage: 66.7% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/aws/nitrotpm	(cached)	coverage: 43.2% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/aws/snp	(cached)	coverage: 43.2% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp	(cached)	coverage: 10.6% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch	(cached)	coverage: 4.3% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/choose	(cached)	coverage: 85.0% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/gcp	(cached)	coverage: 76.1% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest	(cached)	coverage: 75.0% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/initialize	(cached)	coverage: 10.7% of statements
?   	github.com/edgelesssys/constellation/v2/internal/attestation/qemu	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/attestation/simulator	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/measurements	(cached)	coverage: 82.8% of statements
?   	github.com/edgelesssys/constellation/v2/internal/attestation/tdx	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/attestation/variant	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/cloud	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/measurements/measurement-generator	(cached)	coverage: 0.0% of statements [no tests to run]
ok  	github.com/edgelesssys/constellation/v2/internal/attestation/vtpm	(cached)	coverage: 22.4% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/cloud/aws	(cached)	coverage: 82.6% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/cloud/azure	(cached)	coverage: 71.4% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/cloud/azureshared	(cached)	coverage: 95.6% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider	(cached)	coverage: 92.6% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/cloud/gcp	(cached)	coverage: 72.9% of statements
?   	github.com/edgelesssys/constellation/v2/internal/cloud/metadata	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/cloud/openstack	(cached)	coverage: 92.5% of statements
?   	github.com/edgelesssys/constellation/v2/internal/cloud/qemu	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/config/disktypes	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/config/imageversion	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/config/instancetypes	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/config/migration	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/constants	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/containerimage	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/compatibility	(cached)	coverage: 83.1% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/config	(cached)	coverage: 79.7% of statements
?   	github.com/edgelesssys/constellation/v2/internal/crypto/testvector	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/cryptsetup	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/crypto	(cached)	coverage: 50.0% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/file	(cached)	coverage: 88.2% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials	(cached)	coverage: 76.9% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/grpc/dialer	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/grpc/grpclog	(cached)	coverage: 73.3% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/grpc/retry	(cached)	coverage: 90.9% of statements
?   	github.com/edgelesssys/constellation/v2/internal/grpc/testdialer	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/kms/config	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/imagefetcher	(cached)	coverage: 84.4% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/installer	(cached)	coverage: 86.4% of statements
?   	github.com/edgelesssys/constellation/v2/internal/kms/kms	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/kms/kms/aws	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/kms/kms/azure	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/kms/kms/gcp	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/kms/kms/cluster	(cached)	coverage: 75.0% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/kms/kms/internal	(cached)	coverage: 86.4% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/kms/setup	(cached)	coverage: 36.2% of statements
?   	github.com/edgelesssys/constellation/v2/internal/kms/storage	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/kms/storage/awss3	(cached)	coverage: 61.3% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/kms/storage/azureblob	(cached)	coverage: 51.9% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/kms/storage/gcs	(cached)	coverage: 68.3% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/kms/storage/memfs	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/kms/uri	(cached)	coverage: 68.5% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/kubernetes	(cached)	coverage: 85.5% of statements
?   	github.com/edgelesssys/constellation/v2/internal/logger	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl	(cached)	coverage: 7.8% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/license	(cached)	coverage: 83.3% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/nodestate	(cached)	coverage: 100.0% of statements
?   	github.com/edgelesssys/constellation/v2/internal/osimage	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/osimage/archive	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/osimage/aws	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/osimage/azure	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/osimage/gcp	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/osimage/imageinfo	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/osimage/measurementsuploader	[no test files]
?   	github.com/edgelesssys/constellation/v2/internal/osimage/nop	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/osimage/secureboot	(cached)	coverage: 79.2% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/retry	(cached)	coverage: 64.3% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/role	(cached)	coverage: 70.6% of statements
?   	github.com/edgelesssys/constellation/v2/internal/sigstore/keyselect	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/semver	(cached)	coverage: 68.2% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/sigstore	(cached)	coverage: 41.6% of statements
?   	github.com/edgelesssys/constellation/v2/internal/versions/components	[no test files]
ok  	github.com/edgelesssys/constellation/v2/internal/staticupload	(cached)	coverage: 78.3% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/versions	(cached)	coverage: 13.9% of statements
ok  	github.com/edgelesssys/constellation/v2/internal/versions/hash-generator	(cached)	coverage: 0.0% of statements [no tests to run]
?   	github.com/edgelesssys/constellation/v2/joinservice/cmd	[no test files]
ok  	github.com/edgelesssys/constellation/v2/joinservice/internal/kms	(cached)	coverage: 85.7% of statements
ok  	github.com/edgelesssys/constellation/v2/joinservice/internal/kubeadm	(cached)	coverage: 76.1% of statements
ok  	github.com/edgelesssys/constellation/v2/joinservice/internal/kubernetes	(cached)	coverage: 8.5% of statements
ok  	github.com/edgelesssys/constellation/v2/joinservice/internal/kubernetesca	(cached)	coverage: 81.6% of statements
?   	github.com/edgelesssys/constellation/v2/joinservice/joinproto	[no test files]
?   	github.com/edgelesssys/constellation/v2/keyservice/cmd	[no test files]
?   	github.com/edgelesssys/constellation/v2/keyservice/keyserviceproto	[no test files]
?   	github.com/edgelesssys/constellation/v2/measurement-reader/cmd	[no test files]
?   	github.com/edgelesssys/constellation/v2/measurement-reader/internal/tdx	[no test files]
?   	github.com/edgelesssys/constellation/v2/measurement-reader/internal/tpm	[no test files]
ok  	github.com/edgelesssys/constellation/v2/joinservice/internal/server	(cached)	coverage: 76.2% of statements
ok  	github.com/edgelesssys/constellation/v2/joinservice/internal/watcher	(cached)	coverage: 80.4% of statements
ok  	github.com/edgelesssys/constellation/v2/keyservice/internal/server	(cached)	coverage: 61.9% of statements
ok  	github.com/edgelesssys/constellation/v2/measurement-reader/internal/sorted	(cached)	coverage: 94.7% of statements
?   	github.com/edgelesssys/constellation/v2/upgrade-agent/cmd	[no test files]
?   	github.com/edgelesssys/constellation/v2/upgrade-agent/upgradeproto	[no test files]
?   	github.com/edgelesssys/constellation/v2/verify/cmd	[no test files]
ok  	github.com/edgelesssys/constellation/v2/upgrade-agent/internal/server	(cached)	coverage: 14.9% of statements
ok  	github.com/edgelesssys/constellation/v2/verify/server	(cached)	coverage: 95.4% of statements
?   	github.com/edgelesssys/constellation/v2/verify/verifyproto	[no test files]
?   	github.com/edgelesssys/constellation/v2/3rdparty/node-maintenance-operator/api/v1beta1	[no test files]
?   	github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror	[no test files]
ok  	github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/bazelfiles	(cached)	coverage: 75.0% of statements
ok  	github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/issues	(cached)	coverage: 88.9% of statements
ok  	github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/mirror	(cached)	coverage: 80.2% of statements
ok  	github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/rules	(cached)	coverage: 82.2% of statements
?   	github.com/edgelesssys/constellation/v2/hack/cli-k8s-compatibility	[no test files]
?   	github.com/edgelesssys/constellation/v2/hack/clidocgen	[no test files]
ok  	github.com/edgelesssys/constellation/v2/hack/configapi	(cached)	coverage: 19.5% of statements
?   	github.com/edgelesssys/constellation/v2/hack/oci-pin	[no test files]
?   	github.com/edgelesssys/constellation/v2/hack/pseudo-version	[no test files]
?   	github.com/edgelesssys/constellation/v2/hack/pseudo-version/internal/git	[no test files]
?   	github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api	[no test files]
?   	github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper	[no test files]
ok  	github.com/edgelesssys/constellation/v2/hack/gocoverage	0.001s	coverage: [no statements] [no tests to run]
ok  	github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/extract	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/inject	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/sums	(cached)	coverage: 98.9% of statements
ok  	github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/server	(cached)	coverage: 60.7% of statements
?   	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator	[no test files]
?   	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/api	[no test files]
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/controllers	(cached)	coverage: 30.6% of statements
?   	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/fake/client	[no test files]
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/aws/client	(cached)	coverage: 78.4% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/azure/client	(cached)	coverage: 89.3% of statements
?   	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/constants	[no test files]
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/gcp/client	(cached)	coverage: 80.8% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/controlplane	(cached)	coverage: 100.0% of statements
?   	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/upgrade	[no test files]
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/deploy	(cached)	coverage: 76.6% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/etcd	(cached)	coverage: 65.8% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/executor	(cached)	coverage: 93.2% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/node	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/patch	(cached)	coverage: 100.0% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/poller	(cached)	coverage: 91.4% of statements
ok  	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/sgreconciler	(cached)	coverage: 34.9% of statements
?   	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api	[no test files]
?   	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1	[no test files]
`

const (
	exampleReportCLI    = `{"Metadate":{"Created":"2023-08-24T16:09:02Z"},"Coverage":{"github.com/edgelesssys/constellation/v2/cli":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/cmd":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd":{"Coverage":65.5,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/clusterid":{"Coverage":56.2,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/cmd":{"Coverage":53.5,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/internal/constellation/featureset":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/internal/constellation/helm":{"Coverage":47.7,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/internal/constellation/helm/imageversion":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/internal/constellation/kubecmd":{"Coverage":54.1,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/libvirt":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/terraform":{"Coverage":71.3,"Notest":false,"Nostmt":false}}}`
	exampleReportCLIOld = `{"Metadate":{"Created":"2023-08-24T16:48:39Z"},"Coverage":{"github.com/edgelesssys/constellation/v2/cli":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/cmd":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd":{"Coverage":73.1,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/clusterid":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/cmd":{"Coverage":61.6,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/internal/constellation/featureset":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/internal/constellation/helm":{"Coverage":51.7,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/internal/constellation/helm/imageversion":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/iamid":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes":{"Coverage":49.8,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/libvirt":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/terraform":{"Coverage":66.7,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/cli/internal/upgrade":{"Coverage":83,"Notest":false,"Nostmt":false}}}`
	exampleReportDisk   = `{"Metadate":{"Created":"2023-08-24T16:40:25Z"},"Coverage":{"github.com/edgelesssys/constellation/v2/disk-mapper/cmd":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/disk-mapper/internal/diskencryption":{"Coverage":0,"Notest":true,"Nostmt":false},"github.com/edgelesssys/constellation/v2/disk-mapper/internal/recoveryserver":{"Coverage":89.1,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/disk-mapper/internal/rejoinclient":{"Coverage":91.8,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/disk-mapper/internal/setup":{"Coverage":68.9,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/disk-mapper/internal/systemd":{"Coverage":25.8,"Notest":false,"Nostmt":false},"github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto":{"Coverage":0,"Notest":true,"Nostmt":false}}}`
)

func TestParseStreaming(t *testing.T) {
	assert := assert.New(t)
	in := bytes.NewBufferString(goTestCoverOutput)
	out := bytes.NewBuffer(nil)
	err := parseStreaming(in, out)
	assert.NoError(err)
}

func TestParseTestOutput(t *testing.T) {
	assert := assert.New(t)
	report, err := parseTestOutput(bytes.NewBufferString(goTestCoverOutput))
	assert.NoError(err)
	assert.Len(report.Coverage, 208)
}

func TestDiffCoverage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var oldreport, newreport report
	err := json.Unmarshal([]byte(exampleReportCLI), &oldreport)
	require.NoError(err)
	newreport = oldreport
	diff, err := diffCoverage(oldreport, newreport)
	assert.NoError(err)
	assert.Len(diff, 12)

	err = json.Unmarshal([]byte(exampleReportDisk), &newreport)
	require.NoError(err)
	diff, err = diffCoverage(oldreport, newreport)
	assert.NoError(err)
	assert.Len(diff, 19)
}

func TestDiffsToMd(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	var oldreport, newreport report
	err := json.Unmarshal([]byte(exampleReportCLI), &oldreport)
	require.NoError(err)
	err = json.Unmarshal([]byte(exampleReportCLIOld), &newreport)
	require.NoError(err)
	diff, err := diffCoverage(oldreport, newreport)
	require.NoError(err)

	out := new(bytes.Buffer)
	err = diffsToMd(diff, out, []string{})
	assert.NoError(err)
	assert.NotEmpty(out)
	lines := strings.Split(out.String(), "\n")
	assert.Len(lines, 17)

	out = new(bytes.Buffer)
	err = diffsToMd(diff, out, []string{})
	assert.NoError(err)
	assert.NotEmpty(out)
	lines = strings.Split(out.String(), "\n")
	require.Len(lines, 17)
}
