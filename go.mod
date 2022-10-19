module github.com/edgelesssys/constellation/v2

go 1.18

replace (
	k8s.io/api v0.0.0 => k8s.io/api v0.24.6
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.24.6
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.24.6
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.24.6
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.24.6
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.24.6
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.24.6
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.24.6
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.24.6
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.24.6
	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.24.6
	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.24.6
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.24.6
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.24.6
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.24.6
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.24.6
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.24.6
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.24.6
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.24.6
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.24.6
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.24.6
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.24.6
	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.24.6
	k8s.io/pod-security-admission v0.0.0 => k8s.io/pod-security-admission v0.24.6
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.24.6
)

replace github.com/google/go-tpm-tools => github.com/daniel-weisse/go-tpm-tools v0.0.0-20220929072523-22862750ed86

require (
	cloud.google.com/go/compute v1.7.0
	cloud.google.com/go/kms v1.4.0
	cloud.google.com/go/logging v1.4.2
	cloud.google.com/go/storage v1.23.0
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.1.4
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys v0.8.1
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets v0.10.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2 v2.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.4.1
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/aws/aws-sdk-go-v2 v1.16.16
	github.com/aws/aws-sdk-go-v2/config v1.17.8
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.17
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.34
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.15.20
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.32.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.18.12
	github.com/aws/aws-sdk-go-v2/service/s3 v1.27.11
	github.com/coreos/go-systemd/v22 v22.4.0
	github.com/docker/docker v20.10.19+incompatible
	github.com/fsnotify/fsnotify v1.5.4
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator/v10 v10.11.1
	github.com/google/go-tpm v0.3.3
	github.com/google/go-tpm-tools v0.3.9
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/tink/go v1.6.1
	github.com/googleapis/gax-go/v2 v2.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hc-install v0.4.0
	github.com/hashicorp/terraform-exec v0.17.3
	github.com/hashicorp/terraform-json v0.14.0
	github.com/icholy/replace v0.5.0
	github.com/manifoldco/promptui v0.9.0
	github.com/martinjungblut/go-cryptsetup v0.0.0-20220520180014-fd0874fd07a6
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/operator-framework/api v0.15.0
	github.com/pkg/errors v0.9.1
	github.com/schollz/progressbar/v3 v3.8.6
	github.com/sigstore/rekor v0.12.2
	github.com/sigstore/sigstore v1.4.2
	github.com/spf13/afero v1.9.2
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	github.com/talos-systems/talos/pkg/machinery v1.0.4
	go.uber.org/goleak v1.2.0
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.23.0
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4
	golang.org/x/sys v0.0.0-20221013171732-95e765b1cc43
	golang.org/x/text v0.3.8
	google.golang.org/api v0.96.0
	google.golang.org/genproto v0.0.0-20221014213838-99cd37c6964a
	google.golang.org/grpc v1.50.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm v2.17.0+incompatible
	helm.sh/helm/v3 v3.9.4
	k8s.io/api v0.24.6
	k8s.io/apiextensions-apiserver v0.24.6
	k8s.io/apimachinery v0.24.6
	k8s.io/apiserver v0.24.6
	k8s.io/cli-runtime v0.24.6
	k8s.io/client-go v0.24.6
	k8s.io/cluster-bootstrap v0.24.6
	k8s.io/kubelet v0.24.6
	k8s.io/kubernetes v1.24.6
	k8s.io/mount-utils v0.24.6
	k8s.io/utils v0.0.0-20220812165043-ad590609e2e5
)

require (
	cloud.google.com/go v0.103.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	code.cloudfoundry.org/clock v0.0.0-20180518195852-02e53af36e6c // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.0.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.7.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.5.1 // indirect
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.8 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.19 // indirect
	github.com/aws/smithy-go v1.13.3 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chai2010/gettext-go v0.0.0-20160711120539-c6fed771bfd5 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/containerd/containerd v1.6.6 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20210303052042-6bc126869bf4 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dnaeon/go-vcr v1.2.0 // indirect
	github.com/docker/cli v20.10.17+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-chi/chi v4.1.2+incompatible // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/analysis v0.21.4 // indirect
	github.com/go-openapi/errors v0.20.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/loads v0.21.2 // indirect
	github.com/go-openapi/runtime v0.24.1 // indirect
	github.com/go-openapi/spec v0.20.7 // indirect
	github.com/go-openapi/strfmt v0.21.3 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-openapi/validate v0.22.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/certificate-transparency-go v1.1.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-attestation v0.4.4-0.20221011162210-17f9c05652a9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-containerregistry v0.11.0 // indirect
	github.com/google/go-tspi v0.3.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/trillian v1.5.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/go-type-adapters v1.0.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jedisct1/go-minisign v0.0.0-20211028175153-1c139d1cc84b // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.8 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/letsencrypt/boulder v0.0.0-20220723181115-27de4befb95e // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20220114050600-8b9d41f48198 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rubenv/sql-migrate v1.1.1 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/sassoftware/relic v0.0.0-20210427151427-dfb082b79b74 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.4.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.13.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/tent/canonical-json-go v0.0.0-20130607151641-96e4ba3a7613 // indirect
	github.com/theupdateframework/go-tuf v0.5.1-0.20220920170306-f237d7ca5b42 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/transparency-dev/merkle v0.0.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	github.com/zclconf/go-cty v1.11.0 // indirect
	go.mongodb.org/mongo-driver v1.10.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.starlark.net v0.0.0-20220223235035-243c74974e97 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20220823124025-807a23277127 // indirect
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591 // indirect
	golang.org/x/oauth2 v0.0.0-20220822191816-0ebed06d0094 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	k8s.io/component-base v0.24.6 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220627174259-011e075b9cb8 // indirect
	k8s.io/kubectl v0.24.2 // indirect
	oras.land/oras-go v1.2.0 // indirect
	sigs.k8s.io/controller-runtime v0.12.1 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/kustomize/api v0.11.4 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
