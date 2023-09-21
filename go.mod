module github.com/edgelesssys/constellation/v2

go 1.20

replace (
	k8s.io/api v0.0.0 => k8s.io/api v0.27.2
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.27.2
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.27.2
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.27.2
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.27.2
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.27.2
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.27.2
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.27.2
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.27.2
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.27.2
	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.27.2
	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.27.2
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.27.2
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.27.2
	k8s.io/dynamic-resource-allocation v0.0.0 => k8s.io/dynamic-resource-allocation v0.27.2
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.27.2
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.27.2
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.27.2
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.27.2
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.27.2
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.27.2
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.27.2
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.27.2
	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.27.2
	k8s.io/pod-security-admission v0.0.0 => k8s.io/pod-security-admission v0.27.2
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.27.2
)

replace (
	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api => ./operators/constellation-node-operator/api
	github.com/google/go-tpm => github.com/thomasten/go-tpm v0.0.0-20230629092004-f43f8e2a59eb
	github.com/martinjungblut/go-cryptsetup => github.com/daniel-weisse/go-cryptsetup v0.0.0-20230705150314-d8c07bd1723c
)

require (
	cloud.google.com/go/compute v1.20.1
	cloud.google.com/go/compute/metadata v0.2.3
	cloud.google.com/go/kms v1.12.1
	cloud.google.com/go/logging v1.7.0
	cloud.google.com/go/secretmanager v1.11.1
	cloud.google.com/go/storage v1.31.0
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.6.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.3.0
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets v0.12.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 v5.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 v4.0.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.0.0
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/aws/aws-sdk-go v1.44.297
	github.com/aws/aws-sdk-go-v2 v1.18.1
	github.com/aws/aws-sdk-go-v2/config v1.18.27
	github.com/aws/aws-sdk-go-v2/credentials v1.13.26
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.4
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.71
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.26.8
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.21.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.102.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.19.13
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.14.14
	github.com/aws/aws-sdk-go-v2/service/s3 v1.36.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.19.10
	github.com/aws/smithy-go v1.13.5
	github.com/bazelbuild/rules_go v0.41.0
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/docker/docker v23.0.6+incompatible
	github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api v0.0.0
	github.com/edgelesssys/go-azguestattestation v0.0.0-20230707101700-a683be600fcf
	github.com/edgelesssys/go-tdx-qpl v0.0.0-20230530085549-fd2878a4dead
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.14.1
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/google/go-sev-guest v0.8.0
	github.com/google/go-tpm v0.9.0
	github.com/google/go-tpm-tools v0.4.0
	github.com/google/uuid v1.3.1
	github.com/googleapis/gax-go/v2 v2.12.0
	github.com/gophercloud/gophercloud v1.5.0
	github.com/gophercloud/utils v0.0.0-20230523080330-de873b9cf00d
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0
	github.com/hashicorp/go-kms-wrapping/v2 v2.0.10
	github.com/hashicorp/go-kms-wrapping/wrappers/awskms/v2 v2.0.7
	github.com/hashicorp/go-kms-wrapping/wrappers/azurekeyvault/v2 v2.0.7
	github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2 v2.0.8
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hc-install v0.5.2
	github.com/hashicorp/hcl/v2 v2.17.0
	github.com/hashicorp/terraform-exec v0.18.1
	github.com/hashicorp/terraform-json v0.15.0
	github.com/martinjungblut/go-cryptsetup v0.0.0-20220520180014-fd0874fd07a6
	github.com/mattn/go-isatty v0.0.19
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/go-internal v1.11.0
	github.com/schollz/progressbar/v3 v3.13.1
	github.com/siderolabs/talos/pkg/machinery v1.4.6
	github.com/sigstore/rekor v1.2.2
	github.com/sigstore/sigstore v1.7.1
	github.com/spf13/afero v1.9.5
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.4
	github.com/theupdateframework/go-tuf v0.5.2
	go.uber.org/goleak v1.2.1
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.13.0
	golang.org/x/mod v0.12.0
	golang.org/x/sys v0.12.0
	golang.org/x/tools v0.10.0
	google.golang.org/api v0.130.0
	google.golang.org/grpc v1.56.2
	google.golang.org/protobuf v1.31.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm v2.17.0+incompatible
	helm.sh/helm/v3 v3.12.1
	k8s.io/api v0.27.3
	k8s.io/apiextensions-apiserver v0.27.3
	k8s.io/apimachinery v0.27.3
	k8s.io/apiserver v0.27.3
	k8s.io/client-go v0.27.3
	k8s.io/cluster-bootstrap v0.27.3
	k8s.io/kubelet v0.27.3
	k8s.io/kubernetes v1.27.3
	k8s.io/mount-utils v0.27.3
	k8s.io/utils v0.0.0-20230505201702-9f6742963106
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go v0.110.2 // indirect
	cloud.google.com/go/iam v1.1.0 // indirect
	cloud.google.com/go/longrunning v0.4.2 // indirect
	code.cloudfoundry.org/clock v0.0.0-20180518195852-02e53af36e6c // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230106234847-43070de90fa1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.7.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.22 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.12 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.0.0 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8 // indirect
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.2 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/containerd/containerd v1.7.0 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20220623050100-57a0ce2678a7 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/docker/cli v23.0.5+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/foxboron/go-uefi v0.0.0-20230808201820-18b9ba9cd4c3
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-chi/chi v4.1.2+incompatible // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-gorp/gorp/v3 v3.0.5 // indirect
	github.com/go-jose/go-jose/v3 v3.0.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.4 // indirect
	github.com/go-openapi/errors v0.20.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/loads v0.21.2 // indirect
	github.com/go-openapi/runtime v0.26.0 // indirect
	github.com/go-openapi/spec v0.20.9 // indirect
	github.com/go-openapi/strfmt v0.21.7 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-openapi/validate v0.22.1 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/certificate-transparency-go v1.1.4 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-attestation v0.5.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-containerregistry v0.15.2 // indirect
	github.com/google/go-tspi v0.3.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/logger v1.1.1 // indirect
	github.com/google/pprof v0.0.0-20221103000818-d260c55eee4c // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.4 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.2.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jedisct1/go-minisign v0.0.0-20211028175153-1c139d1cc84b // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/letsencrypt/boulder v0.0.0-20221109233200-85aa52084eaf // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc4 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rubenv/sql-migrate v1.3.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sassoftware/relic v7.2.1+incompatible // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.6.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/transparency-dev/merkle v0.0.2 // indirect
	github.com/vtolstov/go-ioctl v0.0.0-20151206205506-6be9cced4810 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	github.com/zclconf/go-cty v1.13.2 // indirect
	go.mongodb.org/mongo-driver v1.11.3 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/otel v1.14.0 // indirect
	go.opentelemetry.io/otel/trace v1.14.0 // indirect
	go.starlark.net v0.0.0-20220223235035-243c74974e97 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/oauth2 v0.9.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/term v0.12.0 // indirect
	golang.org/x/text v0.13.0
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230629202037-9506855d4529 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/cli-runtime v0.27.2 // indirect
	k8s.io/component-base v0.27.3 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/kubectl v0.27.2 // indirect
	oras.land/oras-go v1.2.3 // indirect
	sigs.k8s.io/controller-runtime v0.15.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.13.2 // indirect
	sigs.k8s.io/kustomize/kyaml v0.14.1 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
