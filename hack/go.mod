module github.com/edgelesssys/constellation/v2/hack

go 1.18

replace (
	k8s.io/api v0.0.0 => k8s.io/api v0.24.3
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.24.3
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.24.3
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.24.3
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.24.3
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.24.3
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.24.3
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.24.3
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.24.3
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.24.3
	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.24.3
	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.24.3
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.24.3
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.24.3
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.24.3
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.24.3
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.24.3
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.24.3
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.24.3
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.24.3
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.24.3
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.24.3
	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.24.3
	k8s.io/pod-security-admission v0.0.0 => k8s.io/pod-security-admission v0.24.3
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.24.3
)

replace github.com/edgelesssys/constellation/v2 => ./..

require (
	cloud.google.com/go/compute v1.7.0
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.1.0
	github.com/edgelesssys/constellation/v2 v2.0.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-tpm-tools v0.3.8
	github.com/google/uuid v1.3.0
	github.com/spf13/afero v1.9.2
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	go.uber.org/goleak v1.1.12
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.21.0
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3
	google.golang.org/api v0.86.0
	google.golang.org/genproto v0.0.0-20220624142145-8cd45d7dbd1f
	google.golang.org/grpc v1.48.0
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v3 v3.0.1
	libvirt.org/go/libvirt v1.8004.0
	libvirt.org/go/libvirtxml v1.8007.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2 v2.0.0
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20210428141323-04723f9f07d7 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/go-type-adapters v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

require (
	cloud.google.com/go v0.102.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/kms v1.4.0 // indirect
	cloud.google.com/go/resourcemanager v1.2.0 // indirect
	cloud.google.com/go/storage v1.22.1 // indirect
	code.cloudfoundry.org/clock v0.0.0-20180518195852-02e53af36e6c // indirect
	github.com/Azure/azure-sdk-for-go v66.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.0.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys v0.6.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets v0.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.5.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights v1.0.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork v1.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.0.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.4.1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.27 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.18 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.11 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.5 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.5.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.15.11 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.6 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.6 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.32.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/kms v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.7 // indirect
	github.com/aws/smithy-go v1.11.3 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/certificate-transparency-go v1.1.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-attestation v0.4.4-0.20220404204839-8820d49b18d9 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/go-containerregistry v0.10.0 // indirect
	github.com/google/go-tpm v0.3.3 // indirect
	github.com/google/go-tspi v0.3.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/tink/go v1.6.1 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/letsencrypt/boulder v0.0.0-20220331220046-b23ab962616e // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/matryer/is v1.4.0 // indirect
	github.com/microsoft/ApplicationInsights-Go v0.4.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sigstore/sigstore v1.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/talos-systems/talos/pkg/machinery v1.0.4 // indirect
	github.com/theupdateframework/go-tuf v0.3.2 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel/trace v1.3.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	golang.org/x/oauth2 v0.0.0-20220622183110-fd043fe589d2 // indirect
	golang.org/x/sys v0.0.0-20220915200043-7b5979e65e41 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	helm.sh/helm v2.17.0+incompatible // indirect
	helm.sh/helm/v3 v3.9.4 // indirect
	k8s.io/api v0.24.3 // indirect
	k8s.io/apimachinery v0.24.3 // indirect
	k8s.io/client-go v0.24.3 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220627174259-011e075b9cb8 // indirect
	k8s.io/utils v0.0.0-20220812165043-ad590609e2e5 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
