module github.com/edgelesssys/constellation/v2

go 1.24.6

replace github.com/martinjungblut/go-cryptsetup => github.com/edgelesssys/go-cryptsetup v0.0.0-20250822075033-840d240dddf8

// TODO(daniel-weisse): revert after merging https://github.com/google/go-sev-guest/pull/173.
replace github.com/google/go-sev-guest => github.com/edgelesssys/go-sev-guest v0.0.0-20250811150530-d85b756e97f2

// Kubernetes replace directives are required because we depend on k8s.io/kubernetes/cmd/kubeadm
// k8s discourages usage of k8s.io/kubernetes as a dependency, but no external staging repositories for kubeadm exist.
// Our code does not actually depend on these packages, but `go mod download` breaks without this replace directive.
// See this issue: https://github.com/kubernetes/kubernetes/issues/79384
// And this README: https://github.com/kubernetes/kubernetes/blob/master/staging/README.md
replace (
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.33.3
	k8s.io/controller-manager => k8s.io/controller-manager v0.33.3
	k8s.io/cri-client => k8s.io/cri-client v0.33.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.33.3
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.33.3
	k8s.io/endpointslice => k8s.io/endpointslice v0.33.3
	k8s.io/externaljwt => k8s.io/externaljwt v0.33.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.33.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.33.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.33.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.33.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.30.14
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.33.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.33.3
)

require (
	cloud.google.com/go/compute v1.41.0
	cloud.google.com/go/compute/metadata v0.7.0
	cloud.google.com/go/kms v1.22.0
	cloud.google.com/go/secretmanager v1.15.0
	cloud.google.com/go/storage v1.56.0
	dario.cat/mergo v1.0.2
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.10.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7 v7.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7 v7.0.0
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets v1.4.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.2
	github.com/BurntSushi/toml v1.5.0
	github.com/aws/aws-sdk-go v1.55.7
	github.com/aws/aws-sdk-go-v2 v1.37.1
	github.com/aws/aws-sdk-go-v2/config v1.30.2
	github.com/aws/aws-sdk-go-v2/credentials v1.18.2
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.1
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.18.2
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.55.1
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.49.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.238.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.47.1
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.27.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.85.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.36.1
	github.com/aws/smithy-go v1.22.5
	github.com/bazelbuild/buildtools v0.0.0-20250715102656-62b9413b08bb
	github.com/bazelbuild/rules_go v0.55.1
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/docker/docker v28.3.3+incompatible
	github.com/edgelesssys/go-azguestattestation v0.0.0-20250408071817-8c4457b235ff
	github.com/edgelesssys/go-tdx-qpl v0.0.0-20250129202750-607ac61e2377
	github.com/foxboron/go-uefi v0.0.0-20250625111927-a3183a1bfc84
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.27.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/go-sev-guest v0.13.0
	github.com/google/go-tdx-guest v0.3.2-0.20250505161510-9efd53b4a100
	github.com/google/go-tpm v0.9.5
	github.com/google/go-tpm-tools v0.4.4
	github.com/google/uuid v1.6.0
	github.com/googleapis/gax-go/v2 v2.15.0
	github.com/gophercloud/gophercloud/v2 v2.7.0
	github.com/gophercloud/utils/v2 v2.0.0-20250711132455-9770683b100a
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.2
	github.com/hashicorp/go-kms-wrapping/v2 v2.0.18
	github.com/hashicorp/go-kms-wrapping/wrappers/awskms/v2 v2.0.11
	github.com/hashicorp/go-kms-wrapping/wrappers/azurekeyvault/v2 v2.0.14
	github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2 v2.0.13
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/hc-install v0.9.2
	github.com/hashicorp/hcl/v2 v2.24.0
	github.com/hashicorp/terraform-exec v0.23.0
	github.com/hashicorp/terraform-json v0.25.0
	github.com/hashicorp/terraform-plugin-framework v1.15.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.18.0
	github.com/hashicorp/terraform-plugin-go v0.28.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-testing v1.13.2
	github.com/hexops/gotextdiff v1.0.3
	github.com/martinjungblut/go-cryptsetup v0.0.0-20220520180014-fd0874fd07a6
	github.com/mattn/go-isatty v0.0.20
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.38.0
	github.com/pkg/errors v0.9.1
	github.com/regclient/regclient v0.9.0
	github.com/rogpeppe/go-internal v1.14.1
	github.com/samber/slog-multi v1.4.1
	github.com/schollz/progressbar/v3 v3.18.0
	github.com/secure-systems-lab/go-securesystemslib v0.9.0
	github.com/siderolabs/talos/pkg/machinery v1.10.5
	github.com/sigstore/rekor v1.3.10
	github.com/sigstore/sigstore v1.9.5
	github.com/spf13/afero v1.14.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.7
	github.com/stretchr/testify v1.10.0
	github.com/tink-crypto/tink-go/v2 v2.4.0
	github.com/vincent-petithory/dataurl v1.0.0
	go.etcd.io/etcd/api/v3 v3.6.4
	go.etcd.io/etcd/client/pkg/v3 v3.6.4
	go.etcd.io/etcd/client/v3 v3.6.4
	go.uber.org/goleak v1.3.0
	golang.org/x/crypto v0.40.0
	golang.org/x/exp v0.0.0-20250718183923-645b1fa84792
	golang.org/x/mod v0.26.0
	golang.org/x/sys v0.34.0
	golang.org/x/text v0.27.0
	golang.org/x/tools v0.35.0
	google.golang.org/api v0.244.0
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.18.5
	k8s.io/api v0.33.3
	k8s.io/apiextensions-apiserver v0.33.3
	k8s.io/apimachinery v0.33.3
	k8s.io/apiserver v0.33.3
	k8s.io/client-go v0.33.3
	k8s.io/cluster-bootstrap v0.33.3
	k8s.io/kubelet v0.33.3
	k8s.io/kubernetes v1.33.4
	k8s.io/mount-utils v0.33.3
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397
	libvirt.org/go/libvirt v1.11005.0
	sigs.k8s.io/controller-runtime v0.21.0
	sigs.k8s.io/yaml v1.6.0
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.4 // indirect
	cloud.google.com/go/auth v0.16.3 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/longrunning v0.6.7 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys v0.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.7.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.27.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.53.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.53.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.26.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.31.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.35.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/containerd/containerd v1.7.27 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/coredns/caddy v1.1.1 // indirect
	github.com/coredns/corefile-migration v1.0.25 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20220623050100-57a0ce2678a7 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/danieljoos/wincred v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-chi/chi v4.1.2+incompatible // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/errors v0.22.1 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/runtime v0.28.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/strfmt v0.23.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-openapi/validate v0.24.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/uuid/v5 v5.3.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/certificate-transparency-go v1.1.8 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-attestation v0.5.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-configfs-tsm v0.3.3-0.20240919001351-b4b5b84fdcbc // indirect
	github.com/google/go-containerregistry v0.20.3 // indirect
	github.com/google/go-tspi v0.3.0 // indirect
	github.com/google/logger v1.1.1 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty v1.5.0 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.6.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.9 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.6 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.37.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.5 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jedisct1/go-minisign v0.0.0-20211028175153-1c139d1cc84b // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/letsencrypt/boulder v0.0.0-20240620165639-de9c06129bec // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20241121165744-79df5c4772f2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rubenv/sql-migrate v1.8.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/samber/slog-common v0.19.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/sassoftware/relic v7.2.1+incompatible // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sigstore/protobuf-specs v0.4.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/theupdateframework/go-tuf v0.7.0 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/transparency-dev/merkle v0.0.2 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/vtolstov/go-ioctl v0.0.0-20151206205506-6be9cced4810 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/zclconf/go-cty v1.16.3 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	go.mongodb.org/mongo-driver v1.14.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.36.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250721164621-a45f3dfb1074 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250728155136-f173205681a0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gotest.tools/v3 v3.4.0 // indirect
	k8s.io/cli-runtime v0.33.3 // indirect
	k8s.io/component-base v0.33.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	k8s.io/kubectl v0.33.3 // indirect
	oras.land/oras-go/v2 v2.6.0 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/kustomize/api v0.19.0 // indirect
	sigs.k8s.io/kustomize/kyaml v0.19.0 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
)
