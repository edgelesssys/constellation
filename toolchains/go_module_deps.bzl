"""Go module dependencies for Bazel.

Contains the equivalent of go.mod and go.sum files for Bazel.
"""

load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    """Declare Go module dependencies for Bazel."""
    go_repository(
        name = "ag_pack_amqp",
        importpath = "pack.ag/amqp",
        sum = "h1:cuNDWLUTbKRtEZwhB0WQBXf9pGbm87pUBXQhvcFxBWg=",
        version = "v0.11.2",
    )
    go_repository(
        name = "co_honnef_go_tools",
        importpath = "honnef.co/go/tools",
        sum = "h1:UoveltGrhghAA7ePc+e+QYDHXrBps2PqFZiHkGR/xK8=",
        version = "v0.0.1-2020.1.4",
    )
    go_repository(
        name = "com_github_acomagu_bufpipe",
        importpath = "github.com/acomagu/bufpipe",
        sum = "h1:fxAGrHZTgQ9w5QqVItgzwj235/uYZYgbXitB+dLupOk=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_adalogics_go_fuzz_headers",
        importpath = "github.com/AdaLogics/go-fuzz-headers",
        sum = "h1:V8krnnfGj4pV65YLUm3C0/8bl7V5Nry2Pwvy3ru/wLc=",
        version = "v0.0.0-20210715213245-6c3934b029d8",
    )
    go_repository(
        name = "com_github_adrg_xdg",
        importpath = "github.com/adrg/xdg",
        sum = "h1:RzRqFcjH4nE5C6oTAxhBtoE2IRyjBSa62SCbyPidvls=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_afex_hystrix_go",
        importpath = "github.com/afex/hystrix-go",
        sum = "h1:rFw4nCn9iMW+Vajsk51NtYIcwSTkXr+JGrMd36kTDJw=",
        version = "v0.0.0-20180502004556-fa1af6a1f4f5",
    )
    go_repository(
        name = "com_github_alcortesm_tgz",
        importpath = "github.com/alcortesm/tgz",
        sum = "h1:uSoVVbwJiQipAclBbw+8quDsfcvFjOpI5iCf4p/cqCs=",
        version = "v0.0.0-20161220082320-9c5fe88206d7",
    )
    go_repository(
        name = "com_github_alecthomas_kingpin",
        importpath = "github.com/alecthomas/kingpin",
        sum = "h1:5svnBTFgJjZvGKyYBtMB0+m5wvrbUHiqye8wRJMlnYI=",
        version = "v2.2.6+incompatible",
    )
    go_repository(
        name = "com_github_alecthomas_template",
        importpath = "github.com/alecthomas/template",
        sum = "h1:JYp7IbQjafoB+tBA3gMyHYHrpOtNuDiK/uB5uXxq5wM=",
        version = "v0.0.0-20190718012654-fb15b899a751",
    )
    go_repository(
        name = "com_github_alecthomas_units",
        importpath = "github.com/alecthomas/units",
        sum = "h1:UQZhZ2O0vMHr2cI+DC1Mbh0TJxzA3RcLoMsFw+aXw7E=",
        version = "v0.0.0-20190924025748-f65c72e2690d",
    )
    go_repository(
        name = "com_github_anmitsu_go_shlex",
        importpath = "github.com/anmitsu/go-shlex",
        sum = "h1:9AeTilPcZAjCFIImctFaOjnTIavg87rW78vTPkQqLI8=",
        version = "v0.0.0-20200514113438-38f4b401e2be",
    )
    go_repository(
        name = "com_github_antihax_optional",
        importpath = "github.com/antihax/optional",
        sum = "h1:xK2lYat7ZLaVVcIuj82J8kIro4V6kDe0AUDFboUCwcg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_antlr_antlr4_runtime_go_antlr",
        importpath = "github.com/antlr/antlr4/runtime/Go/antlr",
        sum = "h1:ue9pVfIcP+QMEjfgo/Ez4ZjNZfonGgR6NgjMaJMu1Cg=",
        version = "v0.0.0-20220418222510-f25a4f6275ed",
    )
    go_repository(
        name = "com_github_aokoli_goutils",
        importpath = "github.com/aokoli/goutils",
        sum = "h1:7fpzNGoJ3VA8qcrm++XEE1QUe0mIwNeLa02Nwq7RDkg=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_apache_beam",
        importpath = "github.com/apache/beam",
        sum = "h1:8MOeoZwBgORfaJjrZxpkqJWEIzwupRGLqUqG0/mvEtQ=",
        version = "v2.32.0+incompatible",
    )
    go_repository(
        name = "com_github_apache_beam_sdks_v2",
        importpath = "github.com/apache/beam/sdks/v2",
        sum = "h1:xWFFfgvHC9Mqhi2PaPZmnREENd58OpNv5LTtsksa7fA=",
        version = "v2.0.0-20211012030016-ef4364519c94",
    )
    go_repository(
        name = "com_github_apache_thrift",
        importpath = "github.com/apache/thrift",
        sum = "h1:5hryIiq9gtn+MiLVn0wP37kb/uTeRZgN08WoCsAhIhI=",
        version = "v0.13.0",
    )
    go_repository(
        name = "com_github_apex_log",
        importpath = "github.com/apex/log",
        sum = "h1:3Zk+boorIQAAGBrHn0JUtAau4ihMamT4WdnfdnXM1zQ=",
        version = "v1.1.4",
    )
    go_repository(
        name = "com_github_apex_logs",
        importpath = "github.com/apex/logs",
        sum = "h1:KmEBVwfDUOTFcBO8cfkJYwdQ5487UZSN+GteOGPmiro=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_aphistic_golf",
        importpath = "github.com/aphistic/golf",
        sum = "h1:2KLQMJ8msqoPHIPDufkxVcoTtcmE5+1sL9950m4R9Pk=",
        version = "v0.0.0-20180712155816-02c07f170c5a",
    )
    go_repository(
        name = "com_github_aphistic_sweet",
        importpath = "github.com/aphistic/sweet",
        sum = "h1:I4z+fAUqvKfvZV/CHi5dV0QuwbmIvYYFDjG0Ss5QpAs=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_apparentlymart_go_textseg",
        importpath = "github.com/apparentlymart/go-textseg",
        sum = "h1:rRmlIsPEEhUTIKQb7T++Nz/A5Q6C9IuX2wFoYVvnCs0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_apparentlymart_go_textseg_v13",
        importpath = "github.com/apparentlymart/go-textseg/v13",
        sum = "h1:Y+KvPE1NYz0xl601PVImeQfFyEy6iT90AvPUL1NNfNw=",
        version = "v13.0.0",
    )
    go_repository(
        name = "com_github_armon_circbuf",
        importpath = "github.com/armon/circbuf",
        sum = "h1:QEF07wC0T1rKkctt1RINW/+RMTVmiwxETico2l3gxJA=",
        version = "v0.0.0-20150827004946-bbbad097214e",
    )
    go_repository(
        name = "com_github_armon_consul_api",
        importpath = "github.com/armon/consul-api",
        sum = "h1:G1bPvciwNyF7IUmKXNt9Ak3m6u9DE1rF+RmtIkBpVdA=",
        version = "v0.0.0-20180202201655-eb2c6b5be1b6",
    )
    go_repository(
        name = "com_github_armon_go_metrics",
        importpath = "github.com/armon/go-metrics",
        sum = "h1:hR91U9KYmb6bLBYLQjyM+3j+rcd/UhE+G78SFnF8gJA=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_armon_go_radix",
        importpath = "github.com/armon/go-radix",
        sum = "h1:F4z6KzEeeQIMeLFa97iZU6vupzoecKdU5TX24SNppXI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_armon_go_socks5",
        importpath = "github.com/armon/go-socks5",
        sum = "h1:0CwZNZbxp69SHPdPJAN/hZIm0C4OItdklCFmMRWYpio=",
        version = "v0.0.0-20160902184237-e75332964ef5",
    )
    go_repository(
        name = "com_github_aryann_difflib",
        importpath = "github.com/aryann/difflib",
        sum = "h1:pv34s756C4pEXnjgPfGYgdhg/ZdajGhyOvzx8k+23nw=",
        version = "v0.0.0-20170710044230-e206f873d14a",
    )
    go_repository(
        name = "com_github_asaskevich_govalidator",
        importpath = "github.com/asaskevich/govalidator",
        sum = "h1:Byv0BzEl3/e6D5CLfI0j/7hiIEtvGVFPCZ7Ei2oq8iQ=",
        version = "v0.0.0-20210307081110-f21760c49a8d",
    )
    go_repository(
        name = "com_github_auth0_go_jwt_middleware",
        importpath = "github.com/auth0/go-jwt-middleware",
        sum = "h1:/fsQ4vRr4zod1wKReUH+0A3ySRjGiT9G34kypO/EKwI=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_aws_aws_lambda_go",
        importpath = "github.com/aws/aws-lambda-go",
        sum = "h1:SuCy7H3NLyp+1Mrfp+m80jcbi9KYWAs9/BXwppwRDzY=",
        version = "v1.13.3",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go",
        importpath = "github.com/aws/aws-sdk-go",
        sum = "h1:VLZuAHI9fa/3WME5JjpVjcPCNfpGHVMiHx8sLHWhMgI=",
        version = "v1.44.180",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2",
        importpath = "github.com/aws/aws-sdk-go-v2",
        sum = "h1:shN7NlnVzvDUgPQ+1rLMSxY8OWRNDRYtiqe0p/PgrhY=",
        version = "v1.17.3",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_aws_protocol_eventstream",
        importpath = "github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream",
        sum = "h1:dK82zF6kkPeCo8J1e+tGx4JdvDIQzj7ygIoLg8WMuGs=",
        version = "v1.4.10",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_config",
        importpath = "github.com/aws/aws-sdk-go-v2/config",
        sum = "h1:lDpy0WM8AHsywOnVrOHaSMfpaiV2igOw8D7svkFkXVA=",
        version = "v1.18.8",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_credentials",
        importpath = "github.com/aws/aws-sdk-go-v2/credentials",
        sum = "h1:vTrwTvv5qAwjWIGhZDSBH/oQHuIQjGmD232k01FUh6A=",
        version = "v1.13.8",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_ec2_imds",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/ec2/imds",
        sum = "h1:j9wi1kQ8b+e0FBVHxCqCGo4kxDU175hoDHcWAi0sauU=",
        version = "v1.12.21",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_s3_manager",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/s3/manager",
        sum = "h1:E884ndKWVGt8IhtUuGhXbEsmaCvdAAkTTUDu7uAok1g=",
        version = "v1.11.47",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_configsources",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/configsources",
        sum = "h1:I3cakv2Uy1vNmmhRQmFptYDxOvBnwCdNwyw63N0RaRU=",
        version = "v1.1.27",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_endpoints_v2",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/endpoints/v2",
        sum = "h1:5NbbMrIzmUn/TXFqAle6mgrH5m9cOvMLRGL7pnG8tRE=",
        version = "v2.4.21",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_ini",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/ini",
        sum = "h1:KeTxcGdNnQudb46oOl4d90f2I33DF/c6q3RnZAmvQdQ=",
        version = "v1.3.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_v4a",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/v4a",
        sum = "h1:H/mF2LNWwX00lD6FlYfKpLLZgUW7oIzCBkig78x4Xok=",
        version = "v1.0.18",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_cloudfront",
        importpath = "github.com/aws/aws-sdk-go-v2/service/cloudfront",
        sum = "h1:k1RJsiyqnUm5u+jrYOls9MoP9xgqsVUOrfsYy9mxCro=",
        version = "v1.24.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_cloudwatchlogs",
        importpath = "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs",
        sum = "h1:Ai+CjJw+5/s3r6k5pggzDuFhWor2U3iwtsrmvN0Hczc=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ec2",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ec2",
        sum = "h1:m6HYlpZlTWb9vHuuRHpWRieqPHWlS0mvQ90OJNrG/Nk=",
        version = "v1.77.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_elasticloadbalancingv2",
        importpath = "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2",
        sum = "h1:Fs+mQ2VSOH3YhNJcfImnl7dsKAm/gqw4Q9iqLRIiPWE=",
        version = "v1.19.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_accept_encoding",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding",
        sum = "h1:y2+VQzC6Zh2ojtV2LoC0MNwHWc6qXv/j2vrQtlftkdA=",
        version = "v1.9.11",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_checksum",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/checksum",
        sum = "h1:kv5vRAl00tozRxSnI0IszPWGXsJOyA7hmEUHFYqsyvw=",
        version = "v1.1.22",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_presigned_url",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/presigned-url",
        sum = "h1:5C6XgTViSb0bunmU57b3CT+MhxULqHH2721FVA+/kDM=",
        version = "v1.9.21",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_s3shared",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/s3shared",
        sum = "h1:vY5siRXvW5TrOKm2qKEf9tliBfdLxdfy0i02LOcmqUo=",
        version = "v1.13.21",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_kms",
        importpath = "github.com/aws/aws-sdk-go-v2/service/kms",
        sum = "h1:1mEQ1BVRfxU2KzcUUIzqDQ8p6yPkhzHrHT++sjtLJts=",
        version = "v1.20.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_resourcegroupstaggingapi",
        importpath = "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi",
        sum = "h1:7HElphc19oFfUwLCbgBqDN3CYxIsOP9YNxF25Ys/iAA=",
        version = "v1.14.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_s3",
        importpath = "github.com/aws/aws-sdk-go-v2/service/s3",
        sum = "h1:wddsyuESfviaiXk3w9N6/4iRwTg/a3gktjODY6jYQBo=",
        version = "v1.30.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_secretsmanager",
        importpath = "github.com/aws/aws-sdk-go-v2/service/secretsmanager",
        sum = "h1:UQDiRZyaHQGPXIuCYqKsz/wIVZknCiZdRmPW8buD/xc=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sso",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sso",
        sum = "h1:/2gzjhQowRLarkkBOGPXSRnb8sQ2RVsjdG1C/UliK/c=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ssooidc",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ssooidc",
        sum = "h1:Jfly6mRxk2ZOSlbCvZfKNS7TukSx1mIzhSsqZ/IGSZI=",
        version = "v1.14.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sts",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sts",
        sum = "h1:kOO++CYo50RcTFISESluhWEi5Prhg+gaSs4whWabiZU=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_github_aws_smithy_go",
        importpath = "github.com/aws/smithy-go",
        sum = "h1:hgz0X/DX0dGqTYpGALqXJoRKRj5oQ7150i5FdTePzO8=",
        version = "v1.13.5",
    )
    go_repository(
        name = "com_github_aybabtme_rgbterm",
        importpath = "github.com/aybabtme/rgbterm",
        sum = "h1:WWB576BN5zNSZc/M9d/10pqEx5VHNhaQ/yOVAkmj5Yo=",
        version = "v0.0.0-20170906152045-cc83f3b3ce59",
    )
    go_repository(
        name = "com_github_azure_azure_amqp_common_go_v2",
        importpath = "github.com/Azure/azure-amqp-common-go/v2",
        sum = "h1:+QbFgmWCnPzdaRMfsI0Yb6GrRdBj5jVL8N3EXuEUcBQ=",
        version = "v2.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_pipeline_go",
        importpath = "github.com/Azure/azure-pipeline-go",
        sum = "h1:OLBdZJ3yvOn2MezlWvbrBMTEUQC72zAftRZOMdj5HYo=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go",
        importpath = "github.com/Azure/azure-sdk-for-go",
        sum = "h1:QEvenaO+Y9ShPeCWsSAtolzVUcb0T0tPeek5TDsovuM=",
        version = "v67.3.0+incompatible",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_azcore",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/azcore",
        sum = "h1:VuHAcMq8pU1IWNT/m5yRaGqbK0BiQKHT8X4DTp9CHdI=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_azidentity",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/azidentity",
        sum = "h1:t/W5MYAuQy81cvM8VUNfRLzhtKpXhVUAN7Cd7KVbTyc=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_internal",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/internal",
        sum = "h1:+5VZ72z0Qan5Bog5C+ZkgSqUbeVUd9wgtHOrIKuc5b8=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_keyvault_azkeys",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys",
        sum = "h1:TOFrNxfjslms5nLLIMjW7N0+zSALX4KiGsptmpb16AA=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_keyvault_azsecrets",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets",
        sum = "h1:82w8tzLcOwDP/Q35j/wEBPt0n0kVC3cjtPdD62G8UAk=",
        version = "v0.11.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_keyvault_internal",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal",
        sum = "h1:Lg6BW0VPmCwcMlvOviL3ruHFO+H9tZNqscK0AeuFjGM=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_applicationinsights_armapplicationinsights",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights",
        sum = "h1:BpGGvzarSyE7kQF1x1hptUcGmNzZEE3yYI+uqBSNRxk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_compute_armcompute_v3",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v3",
        sum = "h1:H3g2mkmu105ON0c/Gqx3Bm+bzoIijLom8LmV9Gjn7X0=",
        version = "v3.0.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_compute_armcompute_v4",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4",
        sum = "h1:Vjq3Uy3JAU1DTxbA+uX6BegIhgO2pyFltbfbmDa9KdI=",
        version = "v4.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_internal",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/internal",
        sum = "h1:lMW1lD/17LUA5z1XTURo7LcVG2ICBPlyMHjIUrcFZNQ=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_network_armnetwork",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork",
        sum = "h1:QM6sE5k2ZT/vI5BEe0r7mqjsUSnhVBFbOsVkEuaEfiA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_network_armnetwork_v2",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2",
        sum = "h1:mk57wRUA8fyjFxVcPPGv4shLcWDXPFYokTJL9zJxQtE=",
        version = "v2.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_resources_armresources",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources",
        sum = "h1:ECsQtyERDVz3NP3kvDOTLvbQhqWp/x9EsGKtb4ogUr8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_storage_azblob",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob",
        sum = "h1:YvQv9Mz6T8oR5ypQOL6erY0Z5t71ak1uHV4QFokCOZk=",
        version = "v0.6.1",
    )
    go_repository(
        name = "com_github_azure_azure_service_bus_go",
        importpath = "github.com/Azure/azure-service-bus-go",
        sum = "h1:G1qBLQvHCFDv9pcpgwgFkspzvnGknJRR0PYJ9ytY/JA=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_azure_azure_storage_blob_go",
        importpath = "github.com/Azure/azure-storage-blob-go",
        sum = "h1:53qhf0Oxa0nOjgbDeeYPUeyiNmafAFEY95rZLK0Tj6o=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_github_azure_go_ansiterm",
        importpath = "github.com/Azure/go-ansiterm",
        sum = "h1:UQHMgLO+TxOElx5B5HZ4hJQsoJ/PvUvKRhJHDQXO8P8=",
        version = "v0.0.0-20210617225240-d185dfc1b5a1",
    )
    go_repository(
        name = "com_github_azure_go_autorest",
        importpath = "github.com/Azure/go-autorest",
        sum = "h1:V5VMDjClD3GiElqLWO7mz2MxNAK/vTfRHdAubSIPRgs=",
        version = "v14.2.0+incompatible",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest",
        importpath = "github.com/Azure/go-autorest/autorest",
        sum = "h1:ndAExarwr5Y+GaHE6VCaY1kyS/HwwGGyuimVhWsHOEM=",
        version = "v0.11.28",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_adal",
        importpath = "github.com/Azure/go-autorest/autorest/adal",
        sum = "h1:jjQnVFXPfekaqb8vIsv2G1lxshoW+oGv4MDlhRtnYZk=",
        version = "v0.9.21",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_azure_auth",
        importpath = "github.com/Azure/go-autorest/autorest/azure/auth",
        sum = "h1:P6bYXFoao05z5uhOQzbC3Qd8JqF3jUoocoTeIxkp2cA=",
        version = "v0.5.11",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_azure_cli",
        importpath = "github.com/Azure/go-autorest/autorest/azure/cli",
        sum = "h1:w77/uPk80ZET2F+AfQExZyEWtn+0Rk/uw17m9fv5Ajc=",
        version = "v0.4.6",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_date",
        importpath = "github.com/Azure/go-autorest/autorest/date",
        sum = "h1:7gUk1U5M/CQbp9WoqinNzJar+8KY+LPI6wiWrP/myHw=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_mocks",
        importpath = "github.com/Azure/go-autorest/autorest/mocks",
        sum = "h1:PGN4EDXnuQbojHbU0UWoNvmu9AGVwYHG9/fkDYhtAfw=",
        version = "v0.4.2",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_to",
        importpath = "github.com/Azure/go-autorest/autorest/to",
        sum = "h1:oXVqrxakqqV1UZdSazDOPOLvOIz+XA683u8EctwboHk=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_validation",
        importpath = "github.com/Azure/go-autorest/autorest/validation",
        sum = "h1:AgyqjAd94fwNAoTjl/WQXg4VvFeRFpO+UhNyRXqF1ac=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_azure_go_autorest_logger",
        importpath = "github.com/Azure/go-autorest/logger",
        sum = "h1:IG7i4p/mDa2Ce4TRyAO8IHnVhAVF3RFU+ZtXWSmf4Tg=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_azure_go_autorest_tracing",
        importpath = "github.com/Azure/go-autorest/tracing",
        sum = "h1:TYi4+3m5t6K48TGI9AUdb+IzbnSxvnvUMfuitfgcfuo=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_azuread_microsoft_authentication_library_for_go",
        importpath = "github.com/AzureAD/microsoft-authentication-library-for-go",
        sum = "h1:VgSJlZH5u0k2qxSpqyghcFQKmvYckj46uymKK5XzkBM=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_beeker1121_goque",
        importpath = "github.com/beeker1121/goque",
        sum = "h1:XbgLdZvVbWsK9HAhAYOp6rksTAdOVYDBQtGSVOLlJrw=",
        version = "v1.0.3-0.20191103205551-d618510128af",
    )
    go_repository(
        name = "com_github_beevik_etree",
        importpath = "github.com/beevik/etree",
        sum = "h1:T0xke/WvNtMoCqgzPhkX2r4rjY3GDZFi+FjpRZY2Jbs=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_benbjohnson_clock",
        importpath = "github.com/benbjohnson/clock",
        sum = "h1:ip6w0uFQkncKQ979AypyG0ER7mqUSBdKLOgAle/AT8A=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_beorn7_perks",
        importpath = "github.com/beorn7/perks",
        sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_bgentry_speakeasy",
        importpath = "github.com/bgentry/speakeasy",
        sum = "h1:ByYyxL9InA1OWqxJqqp2A5pYHUrCiAL6K3J+LKSsQkY=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_bketelsen_crypt",
        importpath = "github.com/bketelsen/crypt",
        sum = "h1:w/jqZtC9YD4DS/Vp9GhWfWcCpuAL58oTnLoI8vE9YHU=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_blakesmith_ar",
        importpath = "github.com/blakesmith/ar",
        sum = "h1:m935MPodAbYS46DG4pJSv7WO+VECIWUQ7OJYSoTrMh4=",
        version = "v0.0.0-20190502131153-809d4375e1fb",
    )
    go_repository(
        name = "com_github_blang_semver",
        importpath = "github.com/blang/semver",
        sum = "h1:cQNTCjp13qL8KC3Nbxr/y2Bqb63oX6wdnnjpJbkM4JQ=",
        version = "v3.5.1+incompatible",
    )
    go_repository(
        name = "com_github_blang_semver_v4",
        importpath = "github.com/blang/semver/v4",
        sum = "h1:1PFHFE6yCCTv8C1TeyNNarDzntLi7wMI5i/pzqYIsAM=",
        version = "v4.0.0",
    )
    go_repository(
        name = "com_github_boltdb_bolt",
        importpath = "github.com/boltdb/bolt",
        sum = "h1:JQmyP4ZBrce+ZQu0dY660FMfatumYDLun9hBCUVIkF4=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_bradfitz_gomemcache",
        importpath = "github.com/bradfitz/gomemcache",
        sum = "h1:L/QXpzIa3pOvUGt1D1lA5KjYhPBAN/3iWdP7xeFS9F0=",
        version = "v0.0.0-20190913173617-a41fca850d0b",
    )
    go_repository(
        name = "com_github_bshuster_repo_logrus_logstash_hook",
        importpath = "github.com/bshuster-repo/logrus-logstash-hook",
        sum = "h1:e+C0SB5R1pu//O4MQ3f9cFuPGoOVeF2fE4Og9otCc70=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_bugsnag_bugsnag_go",
        importpath = "github.com/bugsnag/bugsnag-go",
        sum = "h1:rFt+Y/IK1aEZkEHchZRSq9OQbsSzIT/OrI8YFFmRIng=",
        version = "v0.0.0-20141110184014-b1d153021fcd",
    )
    go_repository(
        name = "com_github_bugsnag_osext",
        importpath = "github.com/bugsnag/osext",
        sum = "h1:otBG+dV+YK+Soembjv71DPz3uX/V/6MMlSyD9JBQ6kQ=",
        version = "v0.0.0-20130617224835-0dd3f918b21b",
    )
    go_repository(
        name = "com_github_bugsnag_panicwrap",
        importpath = "github.com/bugsnag/panicwrap",
        sum = "h1:nvj0OLI3YqYXer/kZD8Ri1aaunCxIEsOst1BVJswV0o=",
        version = "v0.0.0-20151223152923-e2c28503fcd0",
    )
    go_repository(
        name = "com_github_burntsushi_toml",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:ksErzDEI1khOiGPgpwuI7x2ebx/uXQNw7xJpn9Eq1+I=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_burntsushi_xgb",
        importpath = "github.com/BurntSushi/xgb",
        sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
        version = "v0.0.0-20160522181843-27f122750802",
    )
    go_repository(
        name = "com_github_bwesterb_go_ristretto",
        importpath = "github.com/bwesterb/go-ristretto",
        sum = "h1:xxWOVbN5m8NNKiSDZXE1jtZvZnC6JSJ9cYFADiZcWtw=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_caarlos0_ctrlc",
        importpath = "github.com/caarlos0/ctrlc",
        sum = "h1:2DtF8GSIcajgffDFJzyG15vO+1PuBWOMUdFut7NnXhw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_campoy_unique",
        importpath = "github.com/campoy/unique",
        sum = "h1:V9a67dfYqPLAvzk5hMQOXYJlZ4SLIXgyKIE+ZiHzgGQ=",
        version = "v0.0.0-20180121183637-88950e537e7e",
    )
    go_repository(
        name = "com_github_casbin_casbin_v2",
        importpath = "github.com/casbin/casbin/v2",
        sum = "h1:bTwon/ECRx9dwBy2ewRVr5OiqjeXSGiTUY74sDPQi/g=",
        version = "v2.1.2",
    )
    go_repository(
        name = "com_github_cavaliercoder_badio",
        importpath = "github.com/cavaliercoder/badio",
        sum = "h1:YYUjy5BRwO5zPtfk+aa2gw255FIIoi93zMmuy19o0bc=",
        version = "v0.0.0-20160213150051-ce5280129e9e",
    )
    go_repository(
        name = "com_github_cavaliercoder_go_cpio",
        importpath = "github.com/cavaliercoder/go-cpio",
        sum = "h1:hHg27A0RSSp2Om9lubZpiMgVbvn39bsUmW9U5h0twqc=",
        version = "v0.0.0-20180626203310-925f9528c45e",
    )
    go_repository(
        name = "com_github_cavaliercoder_go_rpm",
        importpath = "github.com/cavaliercoder/go-rpm",
        sum = "h1:jP7ki8Tzx9ThnFPLDhBYAhEpI2+jOURnHQNURgsMvnY=",
        version = "v0.0.0-20200122174316-8cb9fd9c31a8",
    )
    go_repository(
        name = "com_github_cenkalti_backoff",
        importpath = "github.com/cenkalti/backoff",
        sum = "h1:tNowT99t7UNflLxfYYSlKYsBpXdEet03Pg2g16Swow4=",
        version = "v2.2.1+incompatible",
    )
    go_repository(
        name = "com_github_cenkalti_backoff_v3",
        importpath = "github.com/cenkalti/backoff/v3",
        sum = "h1:cfUAAO3yvKMYKPrvhDuHSwQnhZNk/RMHKdZqKTxfm6M=",
        version = "v3.2.2",
    )
    go_repository(
        name = "com_github_cenkalti_backoff_v4",
        importpath = "github.com/cenkalti/backoff/v4",
        sum = "h1:6Yo7N8UP2K6LWZnW94DLVSSrbobcWdVzAYOisuDPIFo=",
        version = "v4.1.2",
    )
    go_repository(
        name = "com_github_census_instrumentation_opencensus_proto",
        importpath = "github.com/census-instrumentation/opencensus-proto",
        sum = "h1:t/LhUZLVitR1Ow2YOnduCsavhwFUklBMoGVYUCqmCqk=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_certifi_gocertifi",
        importpath = "github.com/certifi/gocertifi",
        sum = "h1:uH66TXeswKn5PW5zdZ39xEwfS9an067BirqA+P4QaLI=",
        version = "v0.0.0-20200922220541-2c3bb06c6054",
    )
    go_repository(
        name = "com_github_cespare_xxhash",
        importpath = "github.com/cespare/xxhash",
        sum = "h1:a6HrQnmkObjyL+Gs60czilIUGqrzKutQD6XZog3p+ko=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_cespare_xxhash_v2",
        importpath = "github.com/cespare/xxhash/v2",
        sum = "h1:YRXhKfTDauu4ajMg1TPgFO5jnlC2HCbmLXMcTG5cbYE=",
        version = "v2.1.2",
    )
    go_repository(
        name = "com_github_chai2010_gettext_go",
        importpath = "github.com/chai2010/gettext-go",
        sum = "h1:1Lwwip6Q2QGsAdl/ZKPCwTe9fe0CjlUbqj5bFNSjIRk=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_checkpoint_restore_go_criu_v5",
        importpath = "github.com/checkpoint-restore/go-criu/v5",
        sum = "h1:wpFFOoomK3389ue2lAb0Boag6XPht5QYpipxmSNL4d8=",
        version = "v5.3.0",
    )
    go_repository(
        name = "com_github_chzyer_logex",
        importpath = "github.com/chzyer/logex",
        sum = "h1:Swpa1K6QvQznwJRcfTfQJmTE72DqScAa40E+fbHEXEE=",
        version = "v1.1.10",
    )
    go_repository(
        name = "com_github_chzyer_readline",
        importpath = "github.com/chzyer/readline",
        sum = "h1:fY5BOSpyZCqRo5OhCuC+XN+r/bBCmeuuJtjz+bCNIf8=",
        version = "v0.0.0-20180603132655-2972be24d48e",
    )
    go_repository(
        name = "com_github_chzyer_test",
        importpath = "github.com/chzyer/test",
        sum = "h1:q763qf9huN11kDQavWsoZXJNW3xEE4JJyHa5Q25/sd8=",
        version = "v0.0.0-20180213035817-a1ea475d72b1",
    )
    go_repository(
        name = "com_github_cilium_ebpf",
        importpath = "github.com/cilium/ebpf",
        sum = "h1:1k/q3ATgxSXRdrmPfH8d7YK0GfqVsEKZAX9dQZvs56k=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_clbanning_x2j",
        importpath = "github.com/clbanning/x2j",
        sum = "h1:EdRZT3IeKQmfCSrgo8SZ8V3MEnskuJP0wCYNpe+aiXo=",
        version = "v0.0.0-20191024224557-825249438eec",
    )
    go_repository(
        name = "com_github_client9_misspell",
        importpath = "github.com/client9/misspell",
        sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
        version = "v0.3.4",
    )
    go_repository(
        name = "com_github_cloudflare_circl",
        importpath = "github.com/cloudflare/circl",
        sum = "h1:bZgT/A+cikZnKIwn7xL2OBj012Bmvho/o6RpRvv3GKY=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_cncf_udpa_go",
        importpath = "github.com/cncf/udpa/go",
        sum = "h1:hzAQntlaYRkVSFEfj9OTWlVV1H155FMD8BTKktLv0QI=",
        version = "v0.0.0-20210930031921-04548b0d99d4",
    )
    go_repository(
        name = "com_github_cncf_xds_go",
        importpath = "github.com/cncf/xds/go",
        sum = "h1:zH8ljVhhq7yC0MIeUL/IviMtY8hx2mK8cN9wEYb8ggw=",
        version = "v0.0.0-20211011173535-cb28da3451f1",
    )
    go_repository(
        name = "com_github_cockroachdb_datadriven",
        importpath = "github.com/cockroachdb/datadriven",
        sum = "h1:xD/lrqdvwsc+O2bjSSi3YqY73Ke3LAiSCx49aCesA0E=",
        version = "v0.0.0-20200714090401-bf6692d28da5",
    )
    go_repository(
        name = "com_github_cockroachdb_errors",
        importpath = "github.com/cockroachdb/errors",
        sum = "h1:Lap807SXTH5tri2TivECb/4abUkMZC9zRoLarvcKDqs=",
        version = "v1.2.4",
    )
    go_repository(
        name = "com_github_cockroachdb_logtags",
        importpath = "github.com/cockroachdb/logtags",
        sum = "h1:o/kfcElHqOiXqcou5a3rIlMc7oJbMQkeLk0VQJ7zgqY=",
        version = "v0.0.0-20190617123548-eb05cc24525f",
    )
    go_repository(
        name = "com_github_codahale_hdrhistogram",
        importpath = "github.com/codahale/hdrhistogram",
        sum = "h1:qMd81Ts1T2OTKmB4acZcyKaMtRnY5Y44NuXGX2GFJ1w=",
        version = "v0.0.0-20161010025455-3a0bb77429bd",
    )
    go_repository(
        name = "com_github_codahale_rfc6979",
        importpath = "github.com/codahale/rfc6979",
        sum = "h1:EDmT6Q9Zs+SbUoc7Ik9EfrFqcylYqgPZ9ANSbTAntnE=",
        version = "v0.0.0-20141003034818-6a90f24967eb",
    )
    go_repository(
        name = "com_github_common_nighthawk_go_figure",
        importpath = "github.com/common-nighthawk/go-figure",
        sum = "h1:J5BL2kskAlV9ckgEsNQXscjIaLiOYiZ75d4e94E6dcQ=",
        version = "v0.0.0-20210622060536-734e95fb86be",
    )
    go_repository(
        name = "com_github_container_storage_interface_spec",
        importpath = "github.com/container-storage-interface/spec",
        sum = "h1:vwN9uCciKygX/a0toYryoYD5+qI9ZFeAMuhEEKO+JBA=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_containerd_aufs",
        importpath = "github.com/containerd/aufs",
        sum = "h1:2oeJiwX5HstO7shSrPZjrohJZLzK36wvpdmzDRkL/LY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_btrfs",
        importpath = "github.com/containerd/btrfs",
        sum = "h1:osn1exbzdub9L5SouXO5swW4ea/xVdJZ3wokxN5GrnA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_cgroups",
        importpath = "github.com/containerd/cgroups",
        sum = "h1:jN/mbWBEaz+T1pi5OFtnkQ+8qnmEbAr1Oo1FRm5B0dA=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_containerd_console",
        importpath = "github.com/containerd/console",
        sum = "h1:lIr7SlA5PxZyMV30bDW0MGbiOPXwc63yRuCP0ARubLw=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_containerd_containerd",
        importpath = "github.com/containerd/containerd",
        sum = "h1:kJ9b3mOFKf8yqo05Ob+tMoxvt1pbVWhnB0re9Y+k+8c=",
        version = "v1.6.12",
    )
    go_repository(
        name = "com_github_containerd_continuity",
        importpath = "github.com/containerd/continuity",
        sum = "h1:nisirsYROK15TAMVukJOUyGJjz4BNQJBVsNvAXZJ/eg=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_containerd_fifo",
        importpath = "github.com/containerd/fifo",
        sum = "h1:6PirWBr9/L7GDamKr+XM0IeUFXu5mf3M/BPpH9gaLBU=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_go_cni",
        importpath = "github.com/containerd/go-cni",
        sum = "h1:1yKpVCQAXI21BJIy8q7Nyk4CWpIgUno6ib7JIDca7D4=",
        version = "v1.1.7",
    )
    go_repository(
        name = "com_github_containerd_go_runc",
        importpath = "github.com/containerd/go-runc",
        sum = "h1:oU+lLv1ULm5taqgV/CJivypVODI4SUz1znWjv3nNYS0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_imgcrypt",
        importpath = "github.com/containerd/imgcrypt",
        sum = "h1:iKTstFebwy3Ak5UF0RHSeuCTahC5OIrPJa6vjMAM81s=",
        version = "v1.1.4",
    )
    go_repository(
        name = "com_github_containerd_nri",
        importpath = "github.com/containerd/nri",
        sum = "h1:6QioHRlThlKh2RkRTR4kIT3PKAcrLo3gIWnjkM4dQmQ=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_containerd_stargz_snapshotter_estargz",
        importpath = "github.com/containerd/stargz-snapshotter/estargz",
        sum = "h1:+7nYmHJb0tEkcRaAW+MHqoKaJYZmkikupxCqVtmPuY0=",
        version = "v0.12.1",
    )
    go_repository(
        name = "com_github_containerd_ttrpc",
        importpath = "github.com/containerd/ttrpc",
        sum = "h1:GbtyLRxb0gOLR0TYQWt3O6B0NvT8tMdorEHqIQo/lWI=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_containerd_typeurl",
        importpath = "github.com/containerd/typeurl",
        sum = "h1:Chlt8zIieDbzQFzXzAeBEF92KhExuE4p9p92/QmY7aY=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_containerd_zfs",
        importpath = "github.com/containerd/zfs",
        sum = "h1:cXLJbx+4Jj7rNsTiqVfm6i+RNLx6FFA2fMmDlEf+Wm8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containernetworking_cni",
        importpath = "github.com/containernetworking/cni",
        sum = "h1:wtRGZVv7olUHMOqouPpn3cXJWpJgM6+EUl31EQbXALQ=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_containernetworking_plugins",
        importpath = "github.com/containernetworking/plugins",
        sum = "h1:+AGfFigZ5TiQH00vhR8qPeSatj53eNGz0C1d3wVYlHE=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_containers_ocicrypt",
        importpath = "github.com/containers/ocicrypt",
        sum = "h1:uMxn2wTb4nDR7GqG3rnZSfpJXqWURfzZ7nKydzIeKpA=",
        version = "v1.1.3",
    )
    go_repository(
        name = "com_github_coredns_caddy",
        importpath = "github.com/coredns/caddy",
        sum = "h1:ezvsPrT/tA/7pYDBZxu0cT0VmWk75AfIaf6GSYCNMf0=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_coredns_corefile_migration",
        importpath = "github.com/coredns/corefile-migration",
        sum = "h1:tNwh8+4WOANV6NjSljwgW7qViJfhvPUt1kosj4rR8yg=",
        version = "v1.0.17",
    )
    go_repository(
        name = "com_github_coreos_bbolt",
        importpath = "github.com/coreos/bbolt",
        sum = "h1:wZwiHHUieZCquLkDL0B8UhzreNWsPHooDAG3q34zk0s=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_coreos_etcd",
        importpath = "github.com/coreos/etcd",
        sum = "h1:8F3hqu9fGYLBifCmRCJsicFqDx/D68Rt3q1JMazcgBQ=",
        version = "v3.3.13+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_etcd",
        importpath = "github.com/coreos/go-etcd",
        sum = "h1:bXhRBIXoTm9BYHS3gE0TtQuyNZyeEMux2sDi4oo5YOo=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_oidc",
        importpath = "github.com/coreos/go-oidc",
        sum = "h1:sdJrfw8akMnCuUlaZU3tE/uYXFgfqom8DBE9so9EBsM=",
        version = "v2.1.0+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_oidc_v3",
        importpath = "github.com/coreos/go-oidc/v3",
        sum = "h1:VxKtbccHZxs8juq7RdJntSqtXFtde9YpNpGn0yqgEHw=",
        version = "v3.5.0",
    )
    go_repository(
        name = "com_github_coreos_go_semver",
        importpath = "github.com/coreos/go-semver",
        sum = "h1:wkHLiw0WNATZnSG7epLsujiMCgPAc9xhjJ4tgnAxmfM=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_coreos_go_systemd",
        importpath = "github.com/coreos/go-systemd",
        sum = "h1:Wf6HqHfScWJN9/ZjdUKyjop4mf3Qdd+1TvvltAvM3m8=",
        version = "v0.0.0-20190321100706-95778dfbb74e",
    )
    go_repository(
        name = "com_github_coreos_go_systemd_v22",
        importpath = "github.com/coreos/go-systemd/v22",
        sum = "h1:RrqgGjYQKalulkV8NGVIfkXQf6YYmOyiJKk8iXXhfZs=",
        version = "v22.5.0",
    )
    go_repository(
        name = "com_github_coreos_pkg",
        importpath = "github.com/coreos/pkg",
        sum = "h1:lBNOc5arjvs8E5mO2tbpBpLoyyu8B6e44T7hJy6potg=",
        version = "v0.0.0-20180928190104-399ea9e2e55f",
    )
    go_repository(
        name = "com_github_cosi_project_runtime",
        importpath = "github.com/cosi-project/runtime",
        sum = "h1:4jzNWoftvsHCU3pv49mzmALc6iaXfNwMpRD2U9LSfqE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man",
        importpath = "github.com/cpuguy83/go-md2man",
        sum = "h1:BSKMNlYxDvnunlTymqtgONjNnaRV1sTpcovwwjF22jk=",
        version = "v1.0.10",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man_v2",
        importpath = "github.com/cpuguy83/go-md2man/v2",
        sum = "h1:p1EgwI/C7NhT0JmVkwCD2ZBK8j4aeHQX2pMHHBfMQ6w=",
        version = "v2.0.2",
    )
    go_repository(
        name = "com_github_creack_pty",
        importpath = "github.com/creack/pty",
        sum = "h1:07n33Z8lZxZ2qwegKbObQohDhXDQxiMMz1NOUGYlesw=",
        version = "v1.1.11",
    )
    go_repository(
        name = "com_github_cyberphone_json_canonicalization",
        importpath = "github.com/cyberphone/json-canonicalization",
        sum = "h1:7AjYfmq7AmviXsuZjV5DcE7PuhJ4dWMi8gLllpLVDQY=",
        version = "v0.0.0-20210303052042-6bc126869bf4",
    )
    go_repository(
        name = "com_github_cyphar_filepath_securejoin",
        importpath = "github.com/cyphar/filepath-securejoin",
        sum = "h1:YX6ebbZCZP7VkM3scTTokDgBL2TY741X51MTk3ycuNI=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_danieljoos_wincred",
        importpath = "github.com/danieljoos/wincred",
        sum = "h1:QLdCxFs1/Yl4zduvBdcHB8goaYk9RARS2SgLLRuAyr0=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_data_dog_go_sqlmock",
        importpath = "github.com/DATA-DOG/go-sqlmock",
        sum = "h1:Shsta01QNfFxHCfpW6YH2STWB0MudeXXEWMr20OEh60=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_daviddengcn_go_colortext",
        importpath = "github.com/daviddengcn/go-colortext",
        sum = "h1:ANqDyC0ys6qCSvuEK7l3g5RaehL/Xck9EX8ATG8oKsE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_denisenkom_go_mssqldb",
        importpath = "github.com/denisenkom/go-mssqldb",
        sum = "h1:RSohk2RsiZqLZ0zCjtfn3S4Gp4exhpBWHyQ7D0yGjAk=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_devigned_tab",
        importpath = "github.com/devigned/tab",
        sum = "h1:3mD6Kb1mUOYeLpJvTVSDwSg5ZsfSxfvxGRTxRsJsITA=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_dgrijalva_jwt_go",
        importpath = "github.com/dgrijalva/jwt-go",
        sum = "h1:7qlOGliEKZXTDg6OTjfoBKDXWrumCAMpl/TFQ4/5kLM=",
        version = "v3.2.0+incompatible",
    )
    go_repository(
        name = "com_github_dgryski_go_rendezvous",
        importpath = "github.com/dgryski/go-rendezvous",
        sum = "h1:lO4WD4F/rVNCu3HqELle0jiPLLBs70cWOduZpkS1E78=",
        version = "v0.0.0-20200823014737-9f7001d12a5f",
    )
    go_repository(
        name = "com_github_dgryski_go_sip13",
        importpath = "github.com/dgryski/go-sip13",
        sum = "h1:RMLoZVzv4GliuWafOuPuQDKSm1SJph7uCRnnS61JAn4=",
        version = "v0.0.0-20181026042036-e10d5fee7954",
    )
    go_repository(
        name = "com_github_dimchansky_utfbom",
        importpath = "github.com/dimchansky/utfbom",
        sum = "h1:vV6w1AhK4VMnhBno/TPVCoK9U/LP0PkLCS9tbxHdi/U=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_distribution_distribution_v3",
        importpath = "github.com/distribution/distribution/v3",
        sum = "h1:hbCT8ZPPMqefiAWD2ZKjn7ypokIGViTvBBg/ExLSdCk=",
        version = "v3.0.0-20220526142353-ffbd94cbe269",
    )
    go_repository(
        name = "com_github_dnaeon_go_vcr",
        importpath = "github.com/dnaeon/go-vcr",
        sum = "h1:zHCHvJYTMh1N7xnV7zf1m1GPBF9Ad0Jk/whtQ1663qI=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_docker_cli",
        importpath = "github.com/docker/cli",
        sum = "h1:lWQbHSHUFs7KraSN2jOJK7zbMS2jNCHI4mt4xUFUVQ4=",
        version = "v20.10.20+incompatible",
    )
    go_repository(
        name = "com_github_docker_distribution",
        importpath = "github.com/docker/distribution",
        sum = "h1:Q50tZOPR6T/hjNsyc9g8/syEs6bk8XXApsHjKukMl68=",
        version = "v2.8.1+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker",
        importpath = "github.com/docker/docker",
        sum = "h1:6jX4yB+NtcbldT90k7vBSaWJDB3i+zkVJT9BEK8kQkk=",
        version = "v20.10.22+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker_credential_helpers",
        importpath = "github.com/docker/docker-credential-helpers",
        sum = "h1:xtCHsjxogADNZcdv1pKUHXryefjlVRqWqIhk/uXJp0A=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_docker_go_connections",
        importpath = "github.com/docker/go-connections",
        sum = "h1:El9xVISelRB7BuFusrZozjnkIM5YnzCViNKohAFqRJQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_docker_go_events",
        importpath = "github.com/docker/go-events",
        sum = "h1:+pKlWGMw7gf6bQ+oDZB4KHQFypsfjYlq/C4rfL7D3g8=",
        version = "v0.0.0-20190806004212-e31b211e4f1c",
    )
    go_repository(
        name = "com_github_docker_go_metrics",
        importpath = "github.com/docker/go-metrics",
        sum = "h1:AgB/0SvBxihN0X8OR4SjsblXkbMvalQ8cjmtKQ2rQV8=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_docker_go_units",
        importpath = "github.com/docker/go-units",
        sum = "h1:69rxXcBk27SvSaaxTtLh/8llcHD8vYHT7WSdRZ/jvr4=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_docker_libtrust",
        importpath = "github.com/docker/libtrust",
        sum = "h1:ZClxb8laGDf5arXfYcAtECDFgAgHklGI8CxgjHnXKJ4=",
        version = "v0.0.0-20150114040149-fa567046d9b1",
    )
    go_repository(
        name = "com_github_docopt_docopt_go",
        importpath = "github.com/docopt/docopt-go",
        sum = "h1:bWDMxwH3px2JBh6AyO7hdCn/PkvCZXii8TGj7sbtEbQ=",
        version = "v0.0.0-20180111231733-ee0de3bc6815",
    )
    go_repository(
        name = "com_github_dustin_go_humanize",
        importpath = "github.com/dustin/go-humanize",
        sum = "h1:VSnTsYCnlFHaM2/igO1h6X3HA71jcobQuxemgkq4zYo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_eapache_go_resiliency",
        importpath = "github.com/eapache/go-resiliency",
        sum = "h1:1NtRmCAqadE2FN4ZcN6g90TP3uk8cg9rn9eNK2197aU=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_eapache_go_xerial_snappy",
        importpath = "github.com/eapache/go-xerial-snappy",
        sum = "h1:YEetp8/yCZMuEPMUDHG0CW/brkkEp8mzqk2+ODEitlw=",
        version = "v0.0.0-20180814174437-776d5712da21",
    )
    go_repository(
        name = "com_github_eapache_queue",
        importpath = "github.com/eapache/queue",
        sum = "h1:YOEu7KNc61ntiQlcEeUIoDTJ2o8mQznoNvUhiigpIqc=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_edsrzf_mmap_go",
        importpath = "github.com/edsrzf/mmap-go",
        sum = "h1:CEBF7HpRnUCSJgGUb5h1Gm7e3VkmVDrR8lvWVLtrOFw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_eggsampler_acme_v3",
        importpath = "github.com/eggsampler/acme/v3",
        sum = "h1:5M7vwYRy65iPpCFHZ01RyWXmYT8e8MlcWn/9BUUB7Ro=",
        version = "v3.3.0",
    )
    go_repository(
        name = "com_github_elazarl_goproxy",
        importpath = "github.com/elazarl/goproxy",
        sum = "h1:yUdfgN0XgIJw7foRItutHYUIhlcKzcSf5vDpdhQAKTc=",
        version = "v0.0.0-20180725130230-947c36da3153",
    )
    go_repository(
        name = "com_github_emicklei_go_restful",
        importpath = "github.com/emicklei/go-restful",
        sum = "h1:spTtZBk5DYEvbxMVutUuTyh1Ao2r4iyvLdACqsl/Ljk=",
        version = "v2.9.5+incompatible",
    )
    go_repository(
        name = "com_github_emicklei_go_restful_v3",
        importpath = "github.com/emicklei/go-restful/v3",
        sum = "h1:eCZ8ulSerjdAiaNpF7GxXIE7ZCMo1moN1qX+S609eVw=",
        version = "v3.8.0",
    )
    go_repository(
        name = "com_github_emirpasic_gods",
        importpath = "github.com/emirpasic/gods",
        sum = "h1:FXtiHYKDGKCW2KzwZKx0iC0PQmdlorYgdFG9jPXJ1Bc=",
        version = "v1.18.1",
    )
    go_repository(
        name = "com_github_envoyproxy_go_control_plane",
        importpath = "github.com/envoyproxy/go-control-plane",
        sum = "h1:xvqufLtNVwAhN8NMyWklVgxnWohi+wtMGQMhtxexlm0=",
        version = "v0.10.2-0.20220325020618-49ff273808a1",
    )
    go_repository(
        name = "com_github_envoyproxy_protoc_gen_validate",
        importpath = "github.com/envoyproxy/protoc-gen-validate",
        sum = "h1:bV5JGEB1ouEzZa0hgVDFFiClrUEuGWRaAc/3mxR2QK0=",
        version = "v0.3.0-java",
    )
    go_repository(
        name = "com_github_etcd_io_gofail",
        importpath = "github.com/etcd-io/gofail",
        sum = "h1:Y2I0lxOttdUKz+hNaIdG3FtjuQrTmwXun1opRV65IZc=",
        version = "v0.0.0-20190801230047-ad7f989257ca",
    )
    go_repository(
        name = "com_github_euank_go_kmsg_parser",
        importpath = "github.com/euank/go-kmsg-parser",
        sum = "h1:cHD53+PLQuuQyLZeriD1V/esuG4MuU0Pjs5y6iknohY=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_evanphx_json_patch",
        importpath = "github.com/evanphx/json-patch",
        sum = "h1:jBYDEEiFBPxA0v50tFdvOzQQTCvpL6mnFh5mB2/l16U=",
        version = "v5.6.0+incompatible",
    )
    go_repository(
        name = "com_github_evanphx_json_patch_v5",
        importpath = "github.com/evanphx/json-patch/v5",
        sum = "h1:b91NhWfaz02IuVxO9faSllyAtNXHMPkC5J8sJCLunww=",
        version = "v5.6.0",
    )
    go_repository(
        name = "com_github_exponent_io_jsonpath",
        importpath = "github.com/exponent-io/jsonpath",
        sum = "h1:105gxyaGwCFad8crR9dcMQWvV9Hvulu6hwUh4tWPJnM=",
        version = "v0.0.0-20151013193312-d6023ce2651d",
    )
    go_repository(
        name = "com_github_facebookgo_clock",
        importpath = "github.com/facebookgo/clock",
        sum = "h1:yDWHCSQ40h88yih2JAcL6Ls/kVkSE8GFACTGVnMPruw=",
        version = "v0.0.0-20150410010913-600d898af40a",
    )
    go_repository(
        name = "com_github_facebookgo_limitgroup",
        importpath = "github.com/facebookgo/limitgroup",
        sum = "h1:IeaD1VDVBPlx3viJT9Md8if8IxxJnO+x0JCGb054heg=",
        version = "v0.0.0-20150612190941-6abd8d71ec01",
    )
    go_repository(
        name = "com_github_facebookgo_muster",
        importpath = "github.com/facebookgo/muster",
        sum = "h1:a4DFiKFJiDRGFD1qIcqGLX/WlUMD9dyLSLDt+9QZgt8=",
        version = "v0.0.0-20150708232844-fd3d7953fd52",
    )
    go_repository(
        name = "com_github_fatih_camelcase",
        importpath = "github.com/fatih/camelcase",
        sum = "h1:hxNvNX/xYBp0ovncs8WyWZrOrpBNub/JfaMvbURyft8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_fatih_color",
        importpath = "github.com/fatih/color",
        sum = "h1:8LOYc1KYPPmyKMuN8QV2DNRWNbLo6LZ0iLs8+mlH53w=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_github_felixge_httpsnoop",
        importpath = "github.com/felixge/httpsnoop",
        sum = "h1:s/nj+GCswXYzN5v2DpNMuMQYe+0DDwt5WVCU6CWBdXk=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_flynn_go_docopt",
        importpath = "github.com/flynn/go-docopt",
        sum = "h1:Ss/B3/5wWRh8+emnK0++g5zQzwDTi30W10pKxKc4JXI=",
        version = "v0.0.0-20140912013429-f6dd2ebbb31e",
    )
    go_repository(
        name = "com_github_flynn_go_shlex",
        importpath = "github.com/flynn/go-shlex",
        sum = "h1:BHsljHzVlRcyQhjrss6TZTdY2VfCqZPbv5k3iBFa2ZQ=",
        version = "v0.0.0-20150515145356-3f9db97f8568",
    )
    go_repository(
        name = "com_github_form3tech_oss_jwt_go",
        importpath = "github.com/form3tech-oss/jwt-go",
        sum = "h1:/l4kBbb4/vGSsdtB5nUe8L7B9mImVMaBPw9L/0TBHU8=",
        version = "v3.2.5+incompatible",
    )
    go_repository(
        name = "com_github_fortytw2_leaktest",
        importpath = "github.com/fortytw2/leaktest",
        sum = "h1:u8491cBMTQ8ft8aeV+adlcytMZylmA5nnwwkRZjI8vw=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_franela_goblin",
        importpath = "github.com/franela/goblin",
        sum = "h1:gb2Z18BhTPJPpLQWj4T+rfKHYCHxRHCtRxhKKjRidVw=",
        version = "v0.0.0-20200105215937-c9ffbefa60db",
    )
    go_repository(
        name = "com_github_franela_goreq",
        importpath = "github.com/franela/goreq",
        sum = "h1:a9ENSRDFBUPkJ5lCgVZh26+ZbGyoVJG7yb5SSzF5H54=",
        version = "v0.0.0-20171204163338-bcd34c9993f8",
    )
    go_repository(
        name = "com_github_frankban_quicktest",
        importpath = "github.com/frankban/quicktest",
        sum = "h1:FJKSZTDHjyhriyC81FLQ0LY93eSai0ZyR/ZIkd3ZUKE=",
        version = "v1.14.3",
    )
    go_repository(
        name = "com_github_fsnotify_fsnotify",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:n+5WquG0fcWoWp6xPWfHdbskMCQaFnG6PfBrh1Ky4HY=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_fullstorydev_grpcurl",
        importpath = "github.com/fullstorydev/grpcurl",
        sum = "h1:xJWosq3BQovQ4QrdPO72OrPiWuGgEsxY8ldYsJbPrqI=",
        version = "v1.8.7",
    )
    go_repository(
        name = "com_github_fvbommel_sortorder",
        importpath = "github.com/fvbommel/sortorder",
        sum = "h1:dSnXLt4mJYH25uDDGa3biZNQsozaUWDSWeKJ0qqFfzE=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_fxamacker_cbor_v2",
        importpath = "github.com/fxamacker/cbor/v2",
        sum = "h1:ri0ArlOR+5XunOP8CRUowT0pSJOwhW098ZCUyskZD88=",
        version = "v2.4.0",
    )
    go_repository(
        name = "com_github_gertd_go_pluralize",
        importpath = "github.com/gertd/go-pluralize",
        sum = "h1:M3uASbVjMnTsPb0PNqg+E/24Vwigyo/tvyMTtAlLgiA=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_getkin_kin_openapi",
        importpath = "github.com/getkin/kin-openapi",
        sum = "h1:j77zg3Ec+k+r+GA3d8hBoXpAc6KX9TbBPrwQGBIy2sY=",
        version = "v0.76.0",
    )
    go_repository(
        name = "com_github_getsentry_raven_go",
        importpath = "github.com/getsentry/raven-go",
        sum = "h1:no+xWJRb5ZI7eE8TWgIq1jLulQiIoLG0IfYxv5JYMGs=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_ghodss_yaml",
        importpath = "github.com/ghodss/yaml",
        sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_gin_contrib_sse",
        importpath = "github.com/gin-contrib/sse",
        sum = "h1:Y/yl/+YNO8GZSjAhjMsSuLt29uWRFHdHYUb5lYOV9qE=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_gin_gonic_gin",
        importpath = "github.com/gin-gonic/gin",
        sum = "h1:fi+bqFAx/oLK54somfCtEZs9HeH1LHVoEPUgARpTqyc=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_gliderlabs_ssh",
        importpath = "github.com/gliderlabs/ssh",
        sum = "h1:OcaySEmAQJgyYcArR+gGGTHCyE7nvhEMTlYY+Dp8CpY=",
        version = "v0.3.5",
    )
    go_repository(
        name = "com_github_go_chi_chi",
        importpath = "github.com/go-chi/chi",
        sum = "h1:fGFk2Gmi/YKXk0OmGfBh0WgmN3XB8lVnEyNz34tQRec=",
        version = "v4.1.2+incompatible",
    )
    go_repository(
        name = "com_github_go_errors_errors",
        importpath = "github.com/go-errors/errors",
        sum = "h1:J6MZopCL4uSllY1OfXM374weqZFFItUbrImctkmUxIA=",
        version = "v1.4.2",
    )
    go_repository(
        name = "com_github_go_git_gcfg",
        importpath = "github.com/go-git/gcfg",
        sum = "h1:Q5ViNfGF8zFgyJWPqYwA7qGFoMTEiBmdlkcfRmpIMa4=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_go_git_go_billy_v5",
        importpath = "github.com/go-git/go-billy/v5",
        sum = "h1:Vaw7LaSTRJOUric7pe4vnzBSgyuf2KrLsu2Y4ZpQBDE=",
        version = "v5.4.0",
    )
    go_repository(
        name = "com_github_go_git_go_git_fixtures_v4",
        importpath = "github.com/go-git/go-git-fixtures/v4",
        sum = "h1:y5z6dd3qi8Hl+stezc8p3JxDkoTRqMAlKnXHuzrfjTQ=",
        version = "v4.3.1",
    )
    go_repository(
        name = "com_github_go_git_go_git_v5",
        importpath = "github.com/go-git/go-git/v5",
        sum = "h1:v8lgZa5k9ylUw+OR/roJHTxR4QItsNFI5nKtAXFuynw=",
        version = "v5.5.2",
    )
    go_repository(
        name = "com_github_go_gl_glfw",
        importpath = "github.com/go-gl/glfw",
        sum = "h1:QbL/5oDUmRBzO9/Z7Seo6zf912W/a6Sr4Eu0G/3Jho0=",
        version = "v0.0.0-20190409004039-e6da0acd62b1",
    )
    go_repository(
        name = "com_github_go_gl_glfw_v3_3_glfw",
        importpath = "github.com/go-gl/glfw/v3.3/glfw",
        sum = "h1:WtGNWLvXpe6ZudgnXrq0barxBImvnnJoMEhXAzcbM0I=",
        version = "v0.0.0-20200222043503-6f7a984d4dc4",
    )
    go_repository(
        name = "com_github_go_gorp_gorp_v3",
        importpath = "github.com/go-gorp/gorp/v3",
        sum = "h1:ULqJXIekoqMx29FI5ekXXFoH1dT2Vc8UhnRzBg+Emz4=",
        version = "v3.0.2",
    )
    go_repository(
        name = "com_github_go_ini_ini",
        importpath = "github.com/go-ini/ini",
        sum = "h1:Mujh4R/dH6YL8bxuISne3xX2+qcQ9p0IxKAP6ExWoUo=",
        version = "v1.25.4",
    )
    go_repository(
        name = "com_github_go_jose_go_jose_v3",
        importpath = "github.com/go-jose/go-jose/v3",
        sum = "h1:s6rrhirfEP/CGIoc6p+PZAeogN2SxKav6Wp7+dyMWVo=",
        version = "v3.0.0",
    )
    go_repository(
        name = "com_github_go_kit_kit",
        importpath = "github.com/go-kit/kit",
        sum = "h1:dXFJfIHVvUcpSgDOV+Ne6t7jXri8Tfv2uOLHUZ2XNuo=",
        version = "v0.10.0",
    )
    go_repository(
        name = "com_github_go_kit_log",
        importpath = "github.com/go-kit/log",
        sum = "h1:7i2K3eKTos3Vc0enKCfnVcgHh2olr/MyfboYq7cAcFw=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_go_logfmt_logfmt",
        importpath = "github.com/go-logfmt/logfmt",
        sum = "h1:otpy5pqBCBZ1ng9RQ0dPu4PN7ba75Y/aA+UpowDyNVA=",
        version = "v0.5.1",
    )
    go_repository(
        name = "com_github_go_logr_logr",
        importpath = "github.com/go-logr/logr",
        sum = "h1:2DntVwHkVopvECVRSlL5PSo9eG+cAkDCuckLubN+rq0=",
        version = "v1.2.3",
    )
    go_repository(
        name = "com_github_go_logr_stdr",
        importpath = "github.com/go-logr/stdr",
        sum = "h1:hSWxHoqTgW2S2qGc0LTAI563KZ5YKYRhT3MFKZMbjag=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_go_logr_zapr",
        importpath = "github.com/go-logr/zapr",
        sum = "h1:a9vnzlIBPQBBkeaR9IuMUfmVOrQlkoC4YfPoFkX3T7A=",
        version = "v1.2.3",
    )
    go_repository(
        name = "com_github_go_openapi_analysis",
        importpath = "github.com/go-openapi/analysis",
        sum = "h1:ZDFLvSNxpDaomuCueM0BlSXxpANBlFYiBvr+GXrvIHc=",
        version = "v0.21.4",
    )
    go_repository(
        name = "com_github_go_openapi_errors",
        importpath = "github.com/go-openapi/errors",
        sum = "h1:rz6kiC84sqNQoqrtulzaL/VERgkoCyB6WdEkc2ujzUc=",
        version = "v0.20.3",
    )
    go_repository(
        name = "com_github_go_openapi_jsonpointer",
        importpath = "github.com/go-openapi/jsonpointer",
        sum = "h1:gZr+CIYByUqjcgeLXnQu2gHYQC9o73G2XUeOFYEICuY=",
        version = "v0.19.5",
    )
    go_repository(
        name = "com_github_go_openapi_jsonreference",
        importpath = "github.com/go-openapi/jsonreference",
        sum = "h1:MYlu0sBgChmCfJxxUKZ8g1cPWFOB37YSZqewK7OKeyA=",
        version = "v0.20.0",
    )
    go_repository(
        name = "com_github_go_openapi_loads",
        importpath = "github.com/go-openapi/loads",
        sum = "h1:r2a/xFIYeZ4Qd2TnGpWDIQNcP80dIaZgf704za8enro=",
        version = "v0.21.2",
    )
    go_repository(
        name = "com_github_go_openapi_runtime",
        importpath = "github.com/go-openapi/runtime",
        sum = "h1:Sml5cgQKGYQHF+M7yYSHaH1eOjvTykrddTE/KtQVjqo=",
        version = "v0.24.1",
    )
    go_repository(
        name = "com_github_go_openapi_spec",
        importpath = "github.com/go-openapi/spec",
        sum = "h1:1Rlu/ZrOCCob0n+JKKJAWhNWMPW8bOZRg8FJaY+0SKI=",
        version = "v0.20.7",
    )
    go_repository(
        name = "com_github_go_openapi_strfmt",
        importpath = "github.com/go-openapi/strfmt",
        sum = "h1:xwhj5X6CjXEZZHMWy1zKJxvW9AfHC9pkyUjLvHtKG7o=",
        version = "v0.21.3",
    )
    go_repository(
        name = "com_github_go_openapi_swag",
        importpath = "github.com/go-openapi/swag",
        sum = "h1:yMBqmnQ0gyZvEb/+KzuWZOXgllrXT4SADYbvDaXHv/g=",
        version = "v0.22.3",
    )
    go_repository(
        name = "com_github_go_openapi_validate",
        importpath = "github.com/go-openapi/validate",
        sum = "h1:b0QecH6VslW/TxtpKgzpO1SNG7GU2FsaqKdP1E2T50Y=",
        version = "v0.22.0",
    )
    go_repository(
        name = "com_github_go_ozzo_ozzo_validation",
        importpath = "github.com/go-ozzo/ozzo-validation",
        sum = "h1:sUy/in/P6askYr16XJgTKq/0SZhiWsdg4WZGaLsGQkM=",
        version = "v3.5.0+incompatible",
    )
    go_repository(
        name = "com_github_go_playground_assert_v2",
        importpath = "github.com/go-playground/assert/v2",
        sum = "h1:MsBgLAaY856+nPRTKrp3/OZK38U/wa0CcBYNjji3q3A=",
        version = "v2.0.1",
    )
    go_repository(
        name = "com_github_go_playground_locales",
        importpath = "github.com/go-playground/locales",
        sum = "h1:EWaQ/wswjilfKLTECiXz7Rh+3BjFhfDFKv/oXslEjJA=",
        version = "v0.14.1",
    )
    go_repository(
        name = "com_github_go_playground_universal_translator",
        importpath = "github.com/go-playground/universal-translator",
        sum = "h1:82dyy6p4OuJq4/CByFNOn/jYrnRPArHwAcmLoJZxyho=",
        version = "v0.18.0",
    )
    go_repository(
        name = "com_github_go_playground_validator_v10",
        importpath = "github.com/go-playground/validator/v10",
        sum = "h1:prmOlTVv+YjZjmRmNSF3VmspqJIxJWXmqUsHwfTRRkQ=",
        version = "v10.11.1",
    )
    go_repository(
        name = "com_github_go_redis_redis",
        importpath = "github.com/go-redis/redis",
        sum = "h1:K0pv1D7EQUjfyoMql+r/jZqCLizCGKFlFgcHWWmHQjg=",
        version = "v6.15.9+incompatible",
    )
    go_repository(
        name = "com_github_go_redis_redis_v8",
        importpath = "github.com/go-redis/redis/v8",
        sum = "h1:AcZZR7igkdvfVmQTPnu9WE37LRrO/YrBH5zWyjDC0oI=",
        version = "v8.11.5",
    )
    go_repository(
        name = "com_github_go_rod_rod",
        importpath = "github.com/go-rod/rod",
        sum = "h1:xbSaA9trZ8v/+eJRGOM6exK1RCsLPwwnzA78vpES0gk=",
        version = "v0.112.3",
    )
    go_repository(
        name = "com_github_go_sql_driver_mysql",
        importpath = "github.com/go-sql-driver/mysql",
        sum = "h1:BCTh4TKNUYmOmMUcQ3IipzF5prigylS7XXjEkfCHuOE=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_go_stack_stack",
        importpath = "github.com/go-stack/stack",
        sum = "h1:ntEHSVwIt7PNXNpgPmVfMrNhLtgjlmnZha2kOpuRiDw=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_go_task_slim_sprig",
        importpath = "github.com/go-task/slim-sprig",
        sum = "h1:p104kn46Q8WdvHunIJ9dAyjPVtrBPhSr3KT2yUst43I=",
        version = "v0.0.0-20210107165309-348f09dbbbc0",
    )
    go_repository(
        name = "com_github_go_test_deep",
        importpath = "github.com/go-test/deep",
        sum = "h1:WOcxcdHcvdgThNXjw0t76K42FXTU7HpNQWHpA2HHNlg=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_gobuffalo_attrs",
        importpath = "github.com/gobuffalo/attrs",
        sum = "h1:hSkbZ9XSyjyBirMeqSqUrK+9HboWrweVlzRNqoBi2d4=",
        version = "v0.0.0-20190224210810-a9411de4debd",
    )
    go_repository(
        name = "com_github_gobuffalo_depgen",
        importpath = "github.com/gobuffalo/depgen",
        sum = "h1:31atYa/UW9V5q8vMJ+W6wd64OaaTHUrCUXER358zLM4=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_gobuffalo_envy",
        importpath = "github.com/gobuffalo/envy",
        sum = "h1:GlXgaiBkmrYMHco6t4j7SacKO4XUjvh5pwXh0f4uxXU=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_gobuffalo_flect",
        importpath = "github.com/gobuffalo/flect",
        sum = "h1:3GQ53z7E3o00C/yy7Ko8VXqQXoJGLkrTQCLTF1EjoXU=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_gobuffalo_genny",
        importpath = "github.com/gobuffalo/genny",
        sum = "h1:iQ0D6SpNXIxu52WESsD+KoQ7af2e3nCfnSBoSF/hKe0=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_gobuffalo_gitgen",
        importpath = "github.com/gobuffalo/gitgen",
        sum = "h1:mSVZ4vj4khv+oThUfS+SQU3UuFIZ5Zo6UNcvK8E8Mz8=",
        version = "v0.0.0-20190315122116-cc086187d211",
    )
    go_repository(
        name = "com_github_gobuffalo_gogen",
        importpath = "github.com/gobuffalo/gogen",
        sum = "h1:dLg+zb+uOyd/mKeQUYIbwbNmfRsr9hd/WtYWepmayhI=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_gobuffalo_logger",
        importpath = "github.com/gobuffalo/logger",
        sum = "h1:nnZNpxYo0zx+Aj9RfMPBm+x9zAU2OayFh/xrAWi34HU=",
        version = "v1.0.6",
    )
    go_repository(
        name = "com_github_gobuffalo_mapi",
        importpath = "github.com/gobuffalo/mapi",
        sum = "h1:fq9WcL1BYrm36SzK6+aAnZ8hcp+SrmnDyAxhNx8dvJk=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_gobuffalo_packd",
        importpath = "github.com/gobuffalo/packd",
        sum = "h1:U2wXfRr4E9DH8IdsDLlRFwTZTK7hLfq9qT/QHXGVe/0=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_gobuffalo_packr_v2",
        importpath = "github.com/gobuffalo/packr/v2",
        sum = "h1:xE1yzvnO56cUC0sTpKR3DIbxZgB54AftTFMhB2XEWlY=",
        version = "v2.8.3",
    )
    go_repository(
        name = "com_github_gobuffalo_syncx",
        importpath = "github.com/gobuffalo/syncx",
        sum = "h1:tpom+2CJmpzAWj5/VEHync2rJGi+epHNIeRSWjzGA+4=",
        version = "v0.0.0-20190224160051-33c29581e754",
    )
    go_repository(
        name = "com_github_gobwas_glob",
        importpath = "github.com/gobwas/glob",
        sum = "h1:A4xDbljILXROh+kObIiy5kIaPYD8e96x1tgBhUI5J+Y=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_godbus_dbus",
        importpath = "github.com/godbus/dbus",
        sum = "h1:WqqLRTsQic3apZUK9qC5sGNfXthmPXzUZ7nQPrNITa4=",
        version = "v4.1.0+incompatible",
    )
    go_repository(
        name = "com_github_godbus_dbus_v5",
        importpath = "github.com/godbus/dbus/v5",
        sum = "h1:mkgN1ofwASrYnJ5W6U/BxG15eXXXjirgZc7CLqkcaro=",
        version = "v5.0.6",
    )
    go_repository(
        name = "com_github_godror_godror",
        importpath = "github.com/godror/godror",
        sum = "h1:uxGAD7UdnNGjX5gf4NnEIGw0JAPTIFiqAyRBZTPKwXs=",
        version = "v0.24.2",
    )
    go_repository(
        name = "com_github_gofrs_flock",
        importpath = "github.com/gofrs/flock",
        sum = "h1:+gYjHKf32LDeiEEFhQaotPbLuUXjY5ZqxKgXy7n59aw=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_github_gofrs_uuid",
        importpath = "github.com/gofrs/uuid",
        sum = "h1:yyYWMnhkhrKwwr8gAOcOCYxOOscHgDS9yZgBrnJfGa0=",
        version = "v4.2.0+incompatible",
    )
    go_repository(
        name = "com_github_gogo_googleapis",
        importpath = "github.com/gogo/googleapis",
        sum = "h1:zgVt4UpGxcqVOw97aRGxT4svlcmdK35fynLNctY32zI=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_gogo_protobuf",
        importpath = "github.com/gogo/protobuf",
        sum = "h1:Ov1cvc58UF3b5XjBnZv7+opcTcQFZebYjWzi34vdm4Q=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_golang_glog",
        importpath = "github.com/golang/glog",
        sum = "h1:nfP3RFugxnNRyKgeWd4oI1nYvXpxrx8ck8ZrcizshdQ=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_golang_groupcache",
        importpath = "github.com/golang/groupcache",
        sum = "h1:oI5xCqsCo564l8iNU+DwB5epxmsaqB+rhGL0m5jtYqE=",
        version = "v0.0.0-20210331224755-41bb18bfe9da",
    )
    go_repository(
        name = "com_github_golang_jwt_jwt",
        importpath = "github.com/golang-jwt/jwt",
        sum = "h1:73Z+4BJcrTC+KczS6WvTPvRGOp1WmfEP4Q1lOd9Z/+c=",
        version = "v3.2.1+incompatible",
    )
    go_repository(
        name = "com_github_golang_jwt_jwt_v4",
        importpath = "github.com/golang-jwt/jwt/v4",
        sum = "h1:rcc4lwaZgFMCZ5jxF9ABolDcIHdBytAFgqFPbSJQAYs=",
        version = "v4.4.2",
    )
    go_repository(
        name = "com_github_golang_mock",
        importpath = "github.com/golang/mock",
        sum = "h1:ErTB+efbowRARo13NNdxyJji2egdxLGQhRaY+DUumQc=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_golang_protobuf",
        importpath = "github.com/golang/protobuf",
        sum = "h1:ROPKBNFfQgOUMifHyP+KYbvpjbdoFNs+aK7DXlji0Tw=",
        version = "v1.5.2",
    )
    go_repository(
        name = "com_github_golang_snappy",
        importpath = "github.com/golang/snappy",
        sum = "h1:yAGX7huGHXlcLOEtBnF4w7FQwA26wojNCwOYAEhLjQM=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_golang_sql_civil",
        importpath = "github.com/golang-sql/civil",
        sum = "h1:lXe2qZdvpiX5WZkZR4hgp4KJVfY3nMkvmwbVkpv1rVY=",
        version = "v0.0.0-20190719163853-cb61b32ac6fe",
    )
    go_repository(
        name = "com_github_gomodule_redigo",
        importpath = "github.com/gomodule/redigo",
        sum = "h1:H5XSIre1MB5NbPYFp+i1NBbb5qN1W8Y8YAQoAYbkm8k=",
        version = "v1.8.2",
    )
    go_repository(
        name = "com_github_google_btree",
        importpath = "github.com/google/btree",
        sum = "h1:xf4v41cLI2Z6FxbKm+8Bu+m8ifhj15JuZ9sa0jZCMUU=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_google_cadvisor",
        importpath = "github.com/google/cadvisor",
        sum = "h1:bXQze1sd8srZiQwiQ19Qaq/AoMIZS8YceBXrIaEvkX0=",
        version = "v0.45.0",
    )
    go_repository(
        name = "com_github_google_cel_go",
        importpath = "github.com/google/cel-go",
        sum = "h1:DmzaiSgoaqGCjtpPQWl26/gND+yRpim56H1jCVev6d8=",
        version = "v0.12.5",
    )
    go_repository(
        name = "com_github_google_certificate_transparency_go",
        importpath = "github.com/google/certificate-transparency-go",
        sum = "h1:4hE0GEId6NAW28dFpC+LrRGwQX5dtmXQGDbg8+/MZOM=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_google_gnostic",
        importpath = "github.com/google/gnostic",
        sum = "h1:FhTMOKj2VhjpouxvWJAV1TL304uMlb9zcDqkl6cEI54=",
        version = "v0.5.7-v3refs",
    )
    go_repository(
        name = "com_github_google_go_attestation",
        importpath = "github.com/google/go-attestation",
        sum = "h1:V3P3BqDEVzAAEw0nlx7OhVRynBUUNgL4znTlqg2Udbc=",
        version = "v0.4.4-0.20221011162210-17f9c05652a9",
    )
    go_repository(
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        sum = "h1:O2Tfq5qg4qc4AmwVlvv0oLiVAGB7enBSJ2x2DqQFi38=",
        version = "v0.5.9",
    )
    go_repository(
        name = "com_github_google_go_containerregistry",
        importpath = "github.com/google/go-containerregistry",
        sum = "h1:W1mzdNUTx4Zla4JaixCRLhORcR7G6KxE5hHl5fkPsp8=",
        version = "v0.12.1",
    )
    go_repository(
        name = "com_github_google_go_github_v28",
        importpath = "github.com/google/go-github/v28",
        sum = "h1:kORf5ekX5qwXO2mGzXXOjMe/g6ap8ahVe0sBEulhSxo=",
        version = "v28.1.1",
    )
    go_repository(
        name = "com_github_google_go_licenses",
        importpath = "github.com/google/go-licenses",
        sum = "h1:JtmsUf+m+KdwCOgLG578T0Mvd0+l+dezPrJh5KYnXZg=",
        version = "v0.0.0-20210329231322-ce1d9163b77d",
    )
    go_repository(
        name = "com_github_google_go_querystring",
        importpath = "github.com/google/go-querystring",
        sum = "h1:Xkwi/a1rcvNg1PPYe5vI8GbeBY/jrVuDX5ASuANWTrk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_google_go_replayers_grpcreplay",
        importpath = "github.com/google/go-replayers/grpcreplay",
        sum = "h1:eNb1y9rZFmY4ax45uEEECSa8fsxGRU+8Bil52ASAwic=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_go_replayers_httpreplay",
        importpath = "github.com/google/go-replayers/httpreplay",
        sum = "h1:AX7FUb4BjrrzNvblr/OlgwrmFiep6soj5K2QSDW7BGk=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_go_sev_guest",
        importpath = "github.com/google/go-sev-guest",
        sum = "h1:IjxtGAvzR+zSyAqMc1FWfYKCg1cwPkBly9+Xog3YMZc=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_google_go_tpm",
        importpath = "github.com/google/go-tpm",
        sum = "h1:P/ZFNBZYXRxc+z7i5uyd8VP7MaDteuLZInzrH2idRGo=",
        version = "v0.3.3",
    )
    go_repository(
        name = "com_github_google_go_tpm_tools",
        importpath = "github.com/google/go-tpm-tools",
        replace = "github.com/daniel-weisse/go-tpm-tools",
        sum = "h1:cJp8NluCBP5idNdtfo2nauPeepn94/jIaeSqTTac9zc=",
        version = "v0.0.0-20230105122812-f7474d459dfc",
    )
    go_repository(
        name = "com_github_google_go_tspi",
        importpath = "github.com/google/go-tspi",
        sum = "h1:ADtq8RKfP+jrTyIWIZDIYcKOMecRqNJFOew2IT0Inus=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_google_gofuzz",
        importpath = "github.com/google/gofuzz",
        sum = "h1:xRy4A+RhZaiKjJ1bPfwQ8sedCA+YS2YcCHW6ec7JMi0=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_google_licenseclassifier",
        importpath = "github.com/google/licenseclassifier",
        sum = "h1:EfzlPF5MRmoWsCGvSkPZ1Nh9uVzHf4FfGnDQ6CXd2NA=",
        version = "v0.0.0-20210325184830-bb04aff29e72",
    )
    go_repository(
        name = "com_github_google_logger",
        importpath = "github.com/google/logger",
        sum = "h1:+6Z2geNxc9G+4D4oDO9njjjn2d0wN5d7uOo0vOIW1NQ=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_google_martian",
        importpath = "github.com/google/martian",
        sum = "h1:xmapqc1AyLoB+ddYT6r04bD9lIjlOqGaREovi0SzFaE=",
        version = "v2.1.1-0.20190517191504-25dcb96d9e51+incompatible",
    )
    go_repository(
        name = "com_github_google_martian_v3",
        importpath = "github.com/google/martian/v3",
        sum = "h1:IqNFLAmvJOgVlpdEBiQbDc2EwKW77amAycfTuWKdfvw=",
        version = "v3.3.2",
    )
    go_repository(
        name = "com_github_google_pprof",
        importpath = "github.com/google/pprof",
        sum = "h1:K6RDEckDVWvDI9JAJYCmNdQXq6neHJOYx3V6jnqNEec=",
        version = "v0.0.0-20210720184732-4bb14d4b1be1",
    )
    go_repository(
        name = "com_github_google_renameio",
        importpath = "github.com/google/renameio",
        sum = "h1:GOZbcHa3HfsPKPlmyPyN2KEohoMXOhdMbHrvbpl2QaA=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_rpmpack",
        importpath = "github.com/google/rpmpack",
        sum = "h1:Fv9Ni1vIq9+Gv4Sm0Xq+NnPYcnsMbdNhJ4Cu4rkbPBM=",
        version = "v0.0.0-20210518075352-dc539ef4f2ea",
    )
    go_repository(
        name = "com_github_google_shlex",
        importpath = "github.com/google/shlex",
        sum = "h1:El6M4kTTCOh6aBiKaUGG7oYTSPP8MxqL4YI3kZKwcP4=",
        version = "v0.0.0-20191202100458-e7afc7fbc510",
    )
    go_repository(
        name = "com_github_google_subcommands",
        importpath = "github.com/google/subcommands",
        sum = "h1:/eqq+otEXm5vhfBrbREPCSVQbvofip6kIz+mX5TUH7k=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_google_tink_go",
        importpath = "github.com/google/tink/go",
        sum = "h1:6Eox8zONGebBFcCBqkVmt60LaWZa6xg1cl/DwAh/J1w=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_google_trillian",
        importpath = "github.com/google/trillian",
        sum = "h1:I5pIN18bKlXtlj1Tk919rQ3mWBU2BzNNR6JhLISGMB4=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_google_uuid",
        importpath = "github.com/google/uuid",
        sum = "h1:t6JiXgmwXMjEs8VusXIJk2BXHsn+wx8BZdTaoZ5fu7I=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_google_wire",
        importpath = "github.com/google/wire",
        sum = "h1:I7ELFeVBr3yfPIcc8+MWvrjk+3VjbcSzoXm3JVa+jD8=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_googleapis_enterprise_certificate_proxy",
        importpath = "github.com/googleapis/enterprise-certificate-proxy",
        sum = "h1:RY7tHKZcRlk788d5WSo/e83gOyyy742E8GSs771ySpg=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_googleapis_gax_go",
        importpath = "github.com/googleapis/gax-go",
        sum = "h1:silFMLAnr330+NRuag/VjIGF7TLp/LBrV2CJKFLWEww=",
        version = "v2.0.2+incompatible",
    )
    go_repository(
        name = "com_github_googleapis_gax_go_v2",
        importpath = "github.com/googleapis/gax-go/v2",
        sum = "h1:IcsPKeInNvYi7eqSaDjiZqDDKu5rsmunY0Y1YupQSSQ=",
        version = "v2.7.0",
    )
    go_repository(
        name = "com_github_googleapis_go_type_adapters",
        importpath = "github.com/googleapis/go-type-adapters",
        sum = "h1:9XdMn+d/G57qq1s8dNc5IesGCXHf6V2HZ2JwRxfA2tA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_googleapis_google_cloud_go_testing",
        importpath = "github.com/googleapis/google-cloud-go-testing",
        sum = "h1:tlyzajkF3030q6M8SvmJSemC9DTHL/xaMa18b65+JM4=",
        version = "v0.0.0-20200911160855-bcd43fbb19e8",
    )
    go_repository(
        name = "com_github_googlecloudplatform_cloudsql_proxy",
        importpath = "github.com/GoogleCloudPlatform/cloudsql-proxy",
        sum = "h1:sTOp2Ajiew5XIH92YSdwhYc+bgpUX5j5TKK/Ac8Saw8=",
        version = "v0.0.0-20191009163259-e802c2cb94ae",
    )
    go_repository(
        name = "com_github_googlecloudplatform_k8s_cloud_provider",
        importpath = "github.com/GoogleCloudPlatform/k8s-cloud-provider",
        sum = "h1:Heo1J/ttaQFgGJSVnCZquy3e5eH5j1nqxBuomztB3P0=",
        version = "v1.18.1-0.20220218231025-f11817397a1b",
    )
    go_repository(
        name = "com_github_gophercloud_gophercloud",
        importpath = "github.com/gophercloud/gophercloud",
        sum = "h1:P/nh25+rzXouhytV2pUHBb65fnds26Ghl8/391+sT5o=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_gopherjs_gopherjs",
        importpath = "github.com/gopherjs/gopherjs",
        sum = "h1:EGx4pi6eqNxGaHF6qqu48+N2wcFQ5qg5FXgOdqsJ5d8=",
        version = "v0.0.0-20181017120253-0766667cb4d1",
    )
    go_repository(
        name = "com_github_gordonklaus_ineffassign",
        importpath = "github.com/gordonklaus/ineffassign",
        sum = "h1:vc7Dmrk4JwS0ZPS6WZvWlwDflgDTA26jItmbSj83nug=",
        version = "v0.0.0-20200309095847-7953dde2c7bf",
    )
    go_repository(
        name = "com_github_goreleaser_goreleaser",
        importpath = "github.com/goreleaser/goreleaser",
        sum = "h1:3ua5fSYvc/doKd5sVTzvXbyqW6H/eyuvy46woLrUFic=",
        version = "v0.134.0",
    )
    go_repository(
        name = "com_github_goreleaser_nfpm",
        importpath = "github.com/goreleaser/nfpm",
        sum = "h1:AEnu9XVmupRDTR930Z2rAs31Mj6sLIPxFcR9ESYvgDA=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_gorilla_context",
        importpath = "github.com/gorilla/context",
        sum = "h1:AWwleXJkX/nhcU9bZSnZoi3h/qGYqQAGhq6zZe/aQW8=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_gorilla_handlers",
        importpath = "github.com/gorilla/handlers",
        sum = "h1:9lRY6j8DEeeBT10CvO9hGW0gmky0BprnvDI5vfhUHH4=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_gorilla_mux",
        importpath = "github.com/gorilla/mux",
        sum = "h1:i40aqfkR1h2SlN9hojwV5ZA91wcXFOvkdNIeFDP5koI=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_gorilla_websocket",
        importpath = "github.com/gorilla/websocket",
        sum = "h1:+/TMaTYc4QFitKJxsQ7Yye35DkWvkdLcvGKqM+x0Ufc=",
        version = "v1.4.2",
    )
    go_repository(
        name = "com_github_gosuri_uitable",
        importpath = "github.com/gosuri/uitable",
        sum = "h1:IG2xLKRvErL3uhY6e1BylFzG+aJiwQviDDTfOKeKTpY=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_gregjones_httpcache",
        importpath = "github.com/gregjones/httpcache",
        sum = "h1:+ngKgrYPPJrOjhax5N+uePQ0Fh1Z7PheYoUI/0nzkPA=",
        version = "v0.0.0-20190611155906-901d90724c79",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_middleware",
        importpath = "github.com/grpc-ecosystem/go-grpc-middleware",
        sum = "h1:+9834+KizmvFV7pXQGSXQTsaWhq2GjuNUt0aUU0YBYw=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_prometheus",
        importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
        sum = "h1:Ovs26xHkKqVztRpIrF/92BcuyuQ/YW4NSIpoGtfXNho=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway",
        importpath = "github.com/grpc-ecosystem/grpc-gateway",
        sum = "h1:gmcG1KaJ57LophUzW0Hy8NmPhnMZb4M0+kPpLofRdBo=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway_v2",
        importpath = "github.com/grpc-ecosystem/grpc-gateway/v2",
        sum = "h1:fi9bGIUJOGzzrHBbP8NWbTfNC5fKO6X7kFw40TOqGB8=",
        version = "v2.13.0",
    )
    go_repository(
        name = "com_github_hashicorp_consul_api",
        importpath = "github.com/hashicorp/consul/api",
        sum = "h1:HXNYlRkkM/t+Y/Yhxtwcy02dlYwIaoxzvxPnS+cqy78=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_hashicorp_consul_sdk",
        importpath = "github.com/hashicorp/consul/sdk",
        sum = "h1:UOxjlb4xVNF93jak1mzzoBatyFju9nrkxpVwIp/QqxQ=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_hashicorp_errwrap",
        importpath = "github.com/hashicorp/errwrap",
        sum = "h1:OxrOeh75EUXMY8TBjag2fzXGZ40LB6IKw45YeGUDY2I=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_checkpoint",
        importpath = "github.com/hashicorp/go-checkpoint",
        sum = "h1:MFYpPZCnQqQTE18jFwSII6eUQrD/oxMFp3mlgcqk5mU=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_cleanhttp",
        importpath = "github.com/hashicorp/go-cleanhttp",
        sum = "h1:035FKYIWjmULyFRBKPs8TBQoi0x6d9G4xc9neXJWAZQ=",
        version = "v0.5.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_hclog",
        importpath = "github.com/hashicorp/go-hclog",
        sum = "h1:vDwF1DFNZhntP4DAjuTpOw3uEgMUpXh1pB5fW9DqHpo=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_immutable_radix",
        importpath = "github.com/hashicorp/go-immutable-radix",
        sum = "h1:DKHmCUm2hRBK510BaiZlwvpD40f8bJFeZnpfm2KLowc=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_msgpack",
        importpath = "github.com/hashicorp/go-msgpack",
        sum = "h1:zKjpN5BK/P5lMYrLmBHdBULWbJ0XpYR+7NGzqkZzoD4=",
        version = "v0.5.3",
    )
    go_repository(
        name = "com_github_hashicorp_go_multierror",
        importpath = "github.com/hashicorp/go-multierror",
        sum = "h1:H5DkEtf6CXdFp0N0Em5UCwQpXMWke8IA0+lD48awMYo=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_net",
        importpath = "github.com/hashicorp/go.net",
        sum = "h1:sNCoNyDEvN1xa+X0baata4RdcpKwcMS6DH+xwfqPgjw=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_plugin",
        importpath = "github.com/hashicorp/go-plugin",
        sum = "h1:MDV3UrKQBM3du3G7MApDGvOsMYy3JQJ4exhSoKBAeVA=",
        version = "v1.4.6",
    )
    go_repository(
        name = "com_github_hashicorp_go_retryablehttp",
        importpath = "github.com/hashicorp/go-retryablehttp",
        sum = "h1:sUiuQAnLlbvmExtFQs72iFW/HXeUn8Z1aJLQ4LJJbTQ=",
        version = "v0.7.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_rootcerts",
        importpath = "github.com/hashicorp/go-rootcerts",
        sum = "h1:jzhAVGtqPKbwpyCPELlgNWhE1znq+qwJtW5Oi2viEzc=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_secure_stdlib_mlock",
        importpath = "github.com/hashicorp/go-secure-stdlib/mlock",
        sum = "h1:p4AKXPPS24tO8Wc8i1gLvSKdmkiSY5xuju57czJ/IJQ=",
        version = "v0.1.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_secure_stdlib_parseutil",
        importpath = "github.com/hashicorp/go-secure-stdlib/parseutil",
        sum = "h1:UpiO20jno/eV1eVZcxqWnUohyKRe1g8FPV/xH1s/2qs=",
        version = "v0.1.7",
    )
    go_repository(
        name = "com_github_hashicorp_go_secure_stdlib_strutil",
        importpath = "github.com/hashicorp/go-secure-stdlib/strutil",
        sum = "h1:kes8mmyCpxJsI7FTwtzRqEy9CdjCtrXrXGuOpxEA7Ts=",
        version = "v0.1.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_sockaddr",
        importpath = "github.com/hashicorp/go-sockaddr",
        sum = "h1:ztczhD1jLxIRjVejw8gFomI1BQZOe2WoVOu0SyteCQc=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_syslog",
        importpath = "github.com/hashicorp/go-syslog",
        sum = "h1:KaodqZuhUoZereWVIYmpUgZysurB1kBLX2j0MwMrUAE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_uuid",
        importpath = "github.com/hashicorp/go-uuid",
        sum = "h1:2gKiV6YVmrJ1i2CKKa9obLvRieoRGviZFL26PcT/Co8=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_hashicorp_go_version",
        importpath = "github.com/hashicorp/go-version",
        sum = "h1:feTTfFNnjP967rlCxM/I9g701jU+RN74YKx2mOkIeek=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_hashicorp_golang_lru",
        importpath = "github.com/hashicorp/golang-lru",
        sum = "h1:YDjusn29QI/Das2iO9M0BHnIbxPeyuCHsjMW+lJfyTc=",
        version = "v0.5.4",
    )
    go_repository(
        name = "com_github_hashicorp_hc_install",
        importpath = "github.com/hashicorp/hc-install",
        sum = "h1:cZkRFr1WVa0Ty6x5fTvL1TuO1flul231rWkGH92oYYk=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_hashicorp_hcl",
        importpath = "github.com/hashicorp/hcl",
        sum = "h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_logutils",
        importpath = "github.com/hashicorp/logutils",
        sum = "h1:dLEQVugN8vlakKOUE3ihGLTZJRB4j+M2cdTm/ORI65Y=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_mdns",
        importpath = "github.com/hashicorp/mdns",
        sum = "h1:WhIgCr5a7AaVH6jPUwjtRuuE7/RDufnUvzIr48smyxs=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_memberlist",
        importpath = "github.com/hashicorp/memberlist",
        sum = "h1:EmmoJme1matNzb+hMpDuR/0sbJSUisxyqBGG676r31M=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_hashicorp_serf",
        importpath = "github.com/hashicorp/serf",
        sum = "h1:YZ7UKsJv+hKjqGVUUbtE3HNj79Eln2oQ75tniF6iPt0=",
        version = "v0.8.2",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_exec",
        importpath = "github.com/hashicorp/terraform-exec",
        sum = "h1:MX14Kvnka/oWGmIkyuyvL6POx25ZmKrjlaclkx3eErU=",
        version = "v0.17.3",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_json",
        importpath = "github.com/hashicorp/terraform-json",
        sum = "h1:sh9iZ1Y8IFJLx+xQiKHGud6/TSUCM0N8e17dKDpqV7s=",
        version = "v0.14.0",
    )
    go_repository(
        name = "com_github_hashicorp_vault_api",
        importpath = "github.com/hashicorp/vault/api",
        sum = "h1:C7OL9YtOtwQbTKI9ogB0A1wffRbCN+rH/LLCHO3d8HM=",
        version = "v1.8.2",
    )
    go_repository(
        name = "com_github_hashicorp_vault_sdk",
        importpath = "github.com/hashicorp/vault/sdk",
        sum = "h1:sjZC1z4j5Rh2GXYbkxn5BLK05S1p7+MhW4AgdUmgRUA=",
        version = "v0.6.1",
    )
    go_repository(
        name = "com_github_hashicorp_yamux",
        importpath = "github.com/hashicorp/yamux",
        sum = "h1:yrQxtgseBDrq9Y652vSRDvsKCJKOUD+GzTS4Y0Y8pvE=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_heketi_heketi",
        importpath = "github.com/heketi/heketi",
        sum = "h1:X4DBFPzcyWZWhia32d94UhDECQJHH0M5kpRb1gxxUHk=",
        version = "v10.3.0+incompatible",
    )
    go_repository(
        name = "com_github_heketi_tests",
        importpath = "github.com/heketi/tests",
        sum = "h1:oJ/NLadJn5HoxvonA6VxG31lg0d6XOURNA09BTtM4fY=",
        version = "v0.0.0-20151005000721-f3775cbcefd6",
    )
    go_repository(
        name = "com_github_honeycombio_beeline_go",
        importpath = "github.com/honeycombio/beeline-go",
        sum = "h1:cUDe555oqvw8oD76BQJ8alk7FP0JZ/M/zXpNvOEDLDc=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_honeycombio_libhoney_go",
        importpath = "github.com/honeycombio/libhoney-go",
        sum = "h1:kPpqoz6vbOzgp7jC6SR7SkNj7rua7rgxvznI6M3KdHc=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_github_howeyc_gopass",
        importpath = "github.com/howeyc/gopass",
        sum = "h1:aY2hhxLhjEAbfXOx2nRJxCXezC6CO2V/yN+OCr1srtk=",
        version = "v0.0.0-20190910152052-7cb4b85ec19c",
    )
    go_repository(
        name = "com_github_hpcloud_tail",
        importpath = "github.com/hpcloud/tail",
        sum = "h1:nfCOvKYfkgYP8hkirhJocXT2+zOD8yUNjXaWfTlyFKI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_huandu_xstrings",
        importpath = "github.com/huandu/xstrings",
        sum = "h1:L18LIDzqlW6xN2rEkpdV8+oL/IXWJ1APd+vsdYy4Wdw=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_hudl_fargo",
        importpath = "github.com/hudl/fargo",
        sum = "h1:0U6+BtN6LhaYuTnIJq4Wyq5cpn6O2kWrxAtcqBmYY6w=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_ianlancetaylor_demangle",
        importpath = "github.com/ianlancetaylor/demangle",
        sum = "h1:mV02weKRL81bEnm8A0HT1/CAelMQDBuQIfLw8n+d6xI=",
        version = "v0.0.0-20200824232613-28f6c0f3b639",
    )
    go_repository(
        name = "com_github_imdario_mergo",
        importpath = "github.com/imdario/mergo",
        sum = "h1:lFzP57bqS/wsqKssCGmtLAb8A0wKjLGrve2q3PPVcBk=",
        version = "v0.3.13",
    )
    go_repository(
        name = "com_github_in_toto_in_toto_golang",
        importpath = "github.com/in-toto/in-toto-golang",
        sum = "h1:FU8tuL4IWx/Hq55AO4+13AZn3Kd6uk3Z44OCIZ9coTw=",
        version = "v0.3.4-0.20211211042327-af1f9fb822bf",
    )
    go_repository(
        name = "com_github_inconshreveable_mousetrap",
        importpath = "github.com/inconshreveable/mousetrap",
        sum = "h1:U3uMjPSQEBMNp1lFxmllqCPM6P5u/Xq7Pgzkat/bFNc=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_influxdata_influxdb1_client",
        importpath = "github.com/influxdata/influxdb1-client",
        sum = "h1:/WZQPMZNsjZ7IlCpsLGdQBINg5bxKQ1K1sh6awxLtkA=",
        version = "v0.0.0-20191209144304-8bf82d3c094d",
    )
    go_repository(
        name = "com_github_intel_goresctrl",
        importpath = "github.com/intel/goresctrl",
        sum = "h1:JyZjdMQu9Kl/wLXe9xA6s1X+tF6BWsQPFGJMEeCfWzE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_ishidawataru_sctp",
        importpath = "github.com/ishidawataru/sctp",
        sum = "h1:qPmlgoeRS18y2dT+iAH5vEKZgIqgiPi2Y8UCu/b7Aq8=",
        version = "v0.0.0-20190723014705-7c296d48a2b5",
    )
    go_repository(
        name = "com_github_jarcoal_httpmock",
        importpath = "github.com/jarcoal/httpmock",
        sum = "h1:cHtVEcTxRSX4J0je7mWPfc9BpDpqzXSJ5HbymZmyHck=",
        version = "v1.0.5",
    )
    go_repository(
        name = "com_github_jbenet_go_context",
        importpath = "github.com/jbenet/go-context",
        sum = "h1:BQSFePA1RWJOlocH6Fxy8MmwDt+yVQYULKfN0RoTN8A=",
        version = "v0.0.0-20150711004518-d14ea06fba99",
    )
    go_repository(
        name = "com_github_jedisct1_go_minisign",
        importpath = "github.com/jedisct1/go-minisign",
        sum = "h1:ZGiXF8sz7PDk6RgkP+A/SFfUD0ZR/AgG6SpRNEDKZy8=",
        version = "v0.0.0-20211028175153-1c139d1cc84b",
    )
    go_repository(
        name = "com_github_jeffashton_win_pdh",
        importpath = "github.com/JeffAshton/win_pdh",
        sum = "h1:UKkYhof1njT1/xq4SEg5z+VpTgjmNeHwPGRQl7takDI=",
        version = "v0.0.0-20161109143554-76bb4ee9f0ab",
    )
    go_repository(
        name = "com_github_jellydator_ttlcache_v2",
        importpath = "github.com/jellydator/ttlcache/v2",
        sum = "h1:AZGME43Eh2Vv3giG6GeqeLeFXxwxn1/qHItqWZl6U64=",
        version = "v2.11.1",
    )
    go_repository(
        name = "com_github_jessevdk_go_flags",
        importpath = "github.com/jessevdk/go-flags",
        sum = "h1:1jKYvbxEjfUl0fmqTCOfonvskHHXMjBySTLW4y9LFvc=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_jhump_protoreflect",
        importpath = "github.com/jhump/protoreflect",
        sum = "h1:MBbQK392K3u8NTLbKOCIi3XdI+y+c6yt5oMq0X3xviw=",
        version = "v1.14.0",
    )
    go_repository(
        name = "com_github_jmespath_go_jmespath",
        importpath = "github.com/jmespath/go-jmespath",
        sum = "h1:BEgLn5cpjn8UN1mAw4NjwDrS35OdebyEtFe+9YPoQUg=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_jmespath_go_jmespath_internal_testify",
        importpath = "github.com/jmespath/go-jmespath/internal/testify",
        sum = "h1:shLQSRRSCCPj3f2gpwzGwWFoC7ycTf1rcQZHOlsJ6N8=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_jmhodges_clock",
        importpath = "github.com/jmhodges/clock",
        sum = "h1:dYTbLf4m0a5u0KLmPfB6mgxbcV7588bOCx79hxa5Sr4=",
        version = "v0.0.0-20160418191101-880ee4c33548",
    )
    go_repository(
        name = "com_github_jmoiron_sqlx",
        importpath = "github.com/jmoiron/sqlx",
        sum = "h1:vFFPA71p1o5gAeqtEAwLU4dnX2napprKtHr7PYIcN3g=",
        version = "v1.3.5",
    )
    go_repository(
        name = "com_github_joho_godotenv",
        importpath = "github.com/joho/godotenv",
        sum = "h1:Zjp+RcGpHhGlrMbJzXTrZZPrWj+1vfm90La1wgB6Bhc=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_jonboulle_clockwork",
        importpath = "github.com/jonboulle/clockwork",
        sum = "h1:9BSCMi8C+0qdApAp4auwX0RkLGUjs956h0EkuQymUhg=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_josharian_intern",
        importpath = "github.com/josharian/intern",
        sum = "h1:vlS4z54oSdjm0bgjRigI+G1HpF+tI+9rE5LLzOg8HmY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_josharian_native",
        importpath = "github.com/josharian/native",
        sum = "h1:Ts/E8zCSEsG17dUqv7joXJFybuMLjQfWE04tsBODTxk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jpillora_backoff",
        importpath = "github.com/jpillora/backoff",
        sum = "h1:uvFg412JmmHBHw7iwprIxkPMI+sGQ4kzOWsMeHnm2EA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jsimonetti_rtnetlink",
        importpath = "github.com/jsimonetti/rtnetlink",
        sum = "h1:lScjubfLwewsD1F+YaDLiq1HDDq7IGADIhGATPwlKHg=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_json_iterator_go",
        importpath = "github.com/json-iterator/go",
        sum = "h1:PV8peI4a0ysnczrg+LtxykD8LfKY9ML6u2jnxaEnrnM=",
        version = "v1.1.12",
    )
    go_repository(
        name = "com_github_jstemmer_go_junit_report",
        importpath = "github.com/jstemmer/go-junit-report",
        sum = "h1:6QPYqodiu3GuPL+7mfx+NwDdp2eTkp9IfEUpgAwUN0o=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_jtolds_gls",
        importpath = "github.com/jtolds/gls",
        sum = "h1:xdiiI2gbIgH/gLH7ADydsJ1uDOEzR8yvV7C0MuV77Wo=",
        version = "v4.20.0+incompatible",
    )
    go_repository(
        name = "com_github_juju_ratelimit",
        importpath = "github.com/juju/ratelimit",
        sum = "h1:+7AIFJVQ0EQgq/K9+0Krm7m530Du7tIz0METWzN0RgY=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_julienschmidt_httprouter",
        importpath = "github.com/julienschmidt/httprouter",
        sum = "h1:U0609e9tgbseu3rBINet9P48AI/D3oJs4dN7jwJOQ1U=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_k0kubun_go_ansi",
        importpath = "github.com/k0kubun/go-ansi",
        sum = "h1:qGQQKEcAR99REcMpsXCp3lJ03zYT1PkRd3kQGPn9GVg=",
        version = "v0.0.0-20180517002512-3bf9e2903213",
    )
    go_repository(
        name = "com_github_karrick_godirwalk",
        importpath = "github.com/karrick/godirwalk",
        sum = "h1:DynhcF+bztK8gooS0+NDJFrdNZjJ3gzVzC545UNA9iw=",
        version = "v1.16.1",
    )
    go_repository(
        name = "com_github_kevinburke_ssh_config",
        importpath = "github.com/kevinburke/ssh_config",
        sum = "h1:x584FjTGwHzMwvHx18PXxbBVzfnxogHaAReU4gf13a4=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_kisielk_errcheck",
        importpath = "github.com/kisielk/errcheck",
        sum = "h1:e8esj/e4R+SAOwFwN+n3zr0nYeCyeweozKfO23MvHzY=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_kisielk_gotool",
        importpath = "github.com/kisielk/gotool",
        sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_klauspost_compress",
        importpath = "github.com/klauspost/compress",
        sum = "h1:YClS/PImqYbn+UILDnqxQCZ3RehC9N318SU3kElDUEM=",
        version = "v1.15.12",
    )
    go_repository(
        name = "com_github_knetic_govaluate",
        importpath = "github.com/Knetic/govaluate",
        sum = "h1:1G1pk05UrOh0NlF1oeaaix1x8XzrfjIDK47TY0Zehcw=",
        version = "v3.0.1-0.20171022003610-9aa49832a739+incompatible",
    )
    go_repository(
        name = "com_github_konsorten_go_windows_terminal_sequences",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:CE8S1cTafDpPvMhIxNJKvHsGVBgn1xWYf1NbHQhywc8=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_kortschak_utter",
        importpath = "github.com/kortschak/utter",
        sum = "h1:AJVccwLrdrikvkH0aI5JKlzZIORLpfMeGBQ5tHfIXis=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_kr_fs",
        importpath = "github.com/kr/fs",
        sum = "h1:Jskdu9ieNAYnjxsi0LbQp1ulIKZV1LAFgK1tWhpZgl8=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_kr_logfmt",
        importpath = "github.com/kr/logfmt",
        sum = "h1:T+h1c/A9Gawja4Y9mFVWj2vyii2bbUNDw3kt9VxK2EY=",
        version = "v0.0.0-20140226030751-b84e30acd515",
    )
    go_repository(
        name = "com_github_kr_pretty",
        importpath = "github.com/kr/pretty",
        sum = "h1:WgNl7dwNpEZ6jJ9k1snq4pZsg7DOEN8hP9Xw0Tsjwk0=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_kr_pty",
        importpath = "github.com/kr/pty",
        sum = "h1:AkaSdXYQOWeaO3neb8EM634ahkXXe3jYbVh/F9lq+GI=",
        version = "v1.1.8",
    )
    go_repository(
        name = "com_github_kr_text",
        importpath = "github.com/kr/text",
        sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_kylelemons_godebug",
        importpath = "github.com/kylelemons/godebug",
        sum = "h1:RPNrshWIDI6G2gRW9EHilWtl7Z6Sb1BR0xunSBf0SNc=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_lann_builder",
        importpath = "github.com/lann/builder",
        sum = "h1:SOEGU9fKiNWd/HOJuq6+3iTQz8KNCLtVX6idSoTLdUw=",
        version = "v0.0.0-20180802200727-47ae307949d0",
    )
    go_repository(
        name = "com_github_lann_ps",
        importpath = "github.com/lann/ps",
        sum = "h1:P6pPBnrTSX3DEVR4fDembhRWSsG5rVo6hYhAB/ADZrk=",
        version = "v0.0.0-20150810152359-62de8c46ede0",
    )
    go_repository(
        name = "com_github_leodido_go_urn",
        importpath = "github.com/leodido/go-urn",
        sum = "h1:BqpAaACuzVSgi/VLzGZIobT2z4v53pjosyNd9Yv6n/w=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_letsencrypt_boulder",
        importpath = "github.com/letsencrypt/boulder",
        sum = "h1:ndns1qx/5dL43g16EQkPV/i8+b3l5bYQwLeoSBe7tS8=",
        version = "v0.0.0-20221109233200-85aa52084eaf",
    )
    go_repository(
        name = "com_github_letsencrypt_challtestsrv",
        importpath = "github.com/letsencrypt/challtestsrv",
        sum = "h1:Lzv4jM+wSgVMCeO5a/F/IzSanhClstFMnX6SfrAJXjI=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_letsencrypt_pkcs11key_v4",
        importpath = "github.com/letsencrypt/pkcs11key/v4",
        sum = "h1:qLc/OznH7xMr5ARJgkZCCWk+EomQkiNTOoOF5LAgagc=",
        version = "v4.0.0",
    )
    go_repository(
        name = "com_github_lib_pq",
        importpath = "github.com/lib/pq",
        sum = "h1:jbk+ZieJ0D7EVGJYpL9QTz7/YW6UHbmdnZWYyK5cdBs=",
        version = "v1.10.6",
    )
    go_repository(
        name = "com_github_libopenstorage_openstorage",
        importpath = "github.com/libopenstorage/openstorage",
        sum = "h1:GLPam7/0mpdP8ZZtKjbfcXJBTIA/T1O6CBErVEFEyIM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_liggitt_tabwriter",
        importpath = "github.com/liggitt/tabwriter",
        sum = "h1:9TO3cAIGXtEhnIaL+V+BEER86oLrvS+kWobKpbJuye0=",
        version = "v0.0.0-20181228230101-89fcab3d43de",
    )
    go_repository(
        name = "com_github_lightstep_lightstep_tracer_common_golang_gogo",
        importpath = "github.com/lightstep/lightstep-tracer-common/golang/gogo",
        sum = "h1:143Bb8f8DuGWck/xpNUOckBVYfFbBTnLevfRZ1aVVqo=",
        version = "v0.0.0-20190605223551-bc2310a04743",
    )
    go_repository(
        name = "com_github_lightstep_lightstep_tracer_go",
        importpath = "github.com/lightstep/lightstep-tracer-go",
        sum = "h1:vi1F1IQ8N7hNWytK9DpJsUfQhGuNSc19z330K6vl4zk=",
        version = "v0.18.1",
    )
    go_repository(
        name = "com_github_lithammer_dedent",
        importpath = "github.com/lithammer/dedent",
        sum = "h1:VNzHMVCBNG1j0fh3OrsFRkVUwStdDArbgBWoPAffktY=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_lpabon_godbc",
        importpath = "github.com/lpabon/godbc",
        sum = "h1:ilqjArN1UOENJJdM34I2YHKmF/B0gGq4VLoSGy9iAao=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_lyft_protoc_gen_validate",
        importpath = "github.com/lyft/protoc-gen-validate",
        sum = "h1:KNt/RhmQTOLr7Aj8PsJ7mTronaFyx80mRTT9qF261dA=",
        version = "v0.0.13",
    )
    go_repository(
        name = "com_github_magiconair_properties",
        importpath = "github.com/magiconair/properties",
        sum = "h1:5ibWZ6iY0NctNGWo87LalDlEZ6R41TqbbDamhfG/Qzo=",
        version = "v1.8.6",
    )
    go_repository(
        name = "com_github_mailru_easyjson",
        importpath = "github.com/mailru/easyjson",
        sum = "h1:UGYAvKxe3sBsEDzO8ZeWOSlIQfWFlxbzLZe7hwFURr0=",
        version = "v0.7.7",
    )
    go_repository(
        name = "com_github_makenowjust_heredoc",
        importpath = "github.com/MakeNowJust/heredoc",
        sum = "h1:cXCdzVdstXyiTqTvfqk9SDHpKNjxuom+DOlyEeQ4pzQ=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_manifoldco_promptui",
        importpath = "github.com/manifoldco/promptui",
        sum = "h1:3V4HzJk1TtXW1MTZMP7mdlwbBpIinw3HztaIlYthEiA=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_markbates_errx",
        importpath = "github.com/markbates/errx",
        sum = "h1:QDFeR+UP95dO12JgW+tgi2UVfo0V8YBHiUIOaeBPiEI=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_markbates_oncer",
        importpath = "github.com/markbates/oncer",
        sum = "h1:E83IaVAHygyndzPimgUYJjbshhDTALZyXxvk9FOlQRY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_markbates_safe",
        importpath = "github.com/markbates/safe",
        sum = "h1:yjZkbvRM6IzKj9tlu/zMJLS0n/V351OZWRnF3QfaUxI=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_martinjungblut_go_cryptsetup",
        importpath = "github.com/martinjungblut/go-cryptsetup",
        sum = "h1:YDjLk3wsL5ZLhLC4TIwIvT2NkSCAdAV6pzzZaRfj4jk=",
        version = "v0.0.0-20220520180014-fd0874fd07a6",
    )
    go_repository(
        name = "com_github_masterminds_goutils",
        importpath = "github.com/Masterminds/goutils",
        sum = "h1:5nUrii3FMTL5diU80unEVvNevw1nH4+ZV4DSLVJLSYI=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_masterminds_semver",
        importpath = "github.com/Masterminds/semver",
        sum = "h1:H65muMkzWKEuNDnfl9d70GUjFniHKHRbFPGBuZ3QEww=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_masterminds_semver_v3",
        importpath = "github.com/Masterminds/semver/v3",
        sum = "h1:hLg3sBzpNErnxhQtUy/mmLR2I9foDujNK030IGemrRc=",
        version = "v3.1.1",
    )
    go_repository(
        name = "com_github_masterminds_sprig",
        importpath = "github.com/Masterminds/sprig",
        sum = "h1:z4yfnGrZ7netVz+0EDJ0Wi+5VZCSYp4Z0m2dk6cEM60=",
        version = "v2.22.0+incompatible",
    )
    go_repository(
        name = "com_github_masterminds_sprig_v3",
        importpath = "github.com/Masterminds/sprig/v3",
        sum = "h1:17jRggJu518dr3QaafizSXOjKYp94wKfABxUmyxvxX8=",
        version = "v3.2.2",
    )
    go_repository(
        name = "com_github_masterminds_squirrel",
        importpath = "github.com/Masterminds/squirrel",
        sum = "h1:YPpoceAcxuzIljlr5iWpNKaql7hLeG1KLSrhvdHpkZc=",
        version = "v1.5.3",
    )
    go_repository(
        name = "com_github_masterminds_vcs",
        importpath = "github.com/Masterminds/vcs",
        sum = "h1:IIA2aBdXvfbIM+yl/eTnL4hb1XwdpvuQLglAix1gweE=",
        version = "v1.13.3",
    )
    go_repository(
        name = "com_github_matryer_is",
        importpath = "github.com/matryer/is",
        sum = "h1:92UTHpy8CDwaJ08GqLDzhhuixiBUUD1p3AU6PHddz4A=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_mattn_go_colorable",
        importpath = "github.com/mattn/go-colorable",
        sum = "h1:fFA4WZxdEF4tXPZVKMLwD8oUnCTTo08duU7wxecdEvA=",
        version = "v0.1.13",
    )
    go_repository(
        name = "com_github_mattn_go_ieproxy",
        importpath = "github.com/mattn/go-ieproxy",
        sum = "h1:HfxbT6/JcvIljmERptWhwa8XzP7H3T+Z2N26gTsaDaA=",
        version = "v0.0.0-20190610004146-91bb50d98149",
    )
    go_repository(
        name = "com_github_mattn_go_isatty",
        importpath = "github.com/mattn/go-isatty",
        sum = "h1:BTarxUcIeDqL27Mc+vyvdWYSL28zpIhv3RoTdsLMPng=",
        version = "v0.0.17",
    )
    go_repository(
        name = "com_github_mattn_go_oci8",
        importpath = "github.com/mattn/go-oci8",
        sum = "h1:aEUDxNAyDG0tv8CA3TArnDQNyc4EhnWlsfxRgDHABHM=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_mattn_go_runewidth",
        importpath = "github.com/mattn/go-runewidth",
        sum = "h1:+xnbZSEeDbOIg5/mE6JF0w6n9duR1l3/WmbinWVwUuU=",
        version = "v0.0.14",
    )
    go_repository(
        name = "com_github_mattn_go_shellwords",
        importpath = "github.com/mattn/go-shellwords",
        sum = "h1:M2zGm7EW6UQJvDeQxo4T51eKPurbeFbe8WtebGE2xrk=",
        version = "v1.0.12",
    )
    go_repository(
        name = "com_github_mattn_go_sqlite3",
        importpath = "github.com/mattn/go-sqlite3",
        sum = "h1:qZgc/Rwetq+MtyE18WhzjokPD93dNqLGNT3QJuLvBGw=",
        version = "v1.14.14",
    )
    go_repository(
        name = "com_github_mattn_go_zglob",
        importpath = "github.com/mattn/go-zglob",
        sum = "h1:xsEx/XUoVlI6yXjqBK062zYhRTZltCNmYPx6v+8DNaY=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_matttproud_golang_protobuf_extensions",
        importpath = "github.com/matttproud/golang_protobuf_extensions",
        sum = "h1:mmDVorXM7PCGKw94cs5zkfA9PSy5pEvNWRP0ET0TIVo=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_mdlayher_ethtool",
        importpath = "github.com/mdlayher/ethtool",
        sum = "h1:oxZoVGGvhhgrFsGWHGesfa4WjHJYx2dt/CKQ+EDJE8E=",
        version = "v0.0.0-20220830195143-0e16326d06d1",
    )
    go_repository(
        name = "com_github_mdlayher_genetlink",
        importpath = "github.com/mdlayher/genetlink",
        sum = "h1:4yrIkRV5Wfk1WfpWTcoOlGmsWgQj3OtQN9ZsbrE+XtU=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_mdlayher_netlink",
        importpath = "github.com/mdlayher/netlink",
        sum = "h1:ZNGI4V7i1fJ94DPYtWhI/R85i/Q7ZxnuhUJQcJMoodI=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_mdlayher_socket",
        importpath = "github.com/mdlayher/socket",
        sum = "h1:280wsy40IC9M9q1uPGcLBwXpcTQDtoGwVt+BNoITxIw=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_medik8s_node_maintenance_operator",
        importpath = "github.com/medik8s/node-maintenance-operator",
        sum = "h1:Vt9u4C12vpQe2T1/K11fZW7kVEGkUgAj40W3YIqoxhg=",
        version = "v0.14.0",
    )
    go_repository(
        name = "com_github_mediocregopher_radix_v4",
        importpath = "github.com/mediocregopher/radix/v4",
        sum = "h1:JkZBEp0y8pWGNZkmO3RR5oEO5huwd4zKKt4rh1C+P8s=",
        version = "v4.1.1",
    )
    go_repository(
        name = "com_github_mgutz_ansi",
        importpath = "github.com/mgutz/ansi",
        sum = "h1:j7+1HpAFS1zy5+Q4qx1fWh90gTKwiN4QCGoY9TWyyO4=",
        version = "v0.0.0-20170206155736-9520e82c474b",
    )
    go_repository(
        name = "com_github_microsoft_applicationinsights_go",
        importpath = "github.com/microsoft/ApplicationInsights-Go",
        sum = "h1:G4+H9WNs6ygSCe6sUyxRc2U81TI5Es90b2t/MwX5KqY=",
        version = "v0.4.4",
    )
    go_repository(
        name = "com_github_microsoft_go_winio",
        importpath = "github.com/Microsoft/go-winio",
        sum = "h1:slsWYD/zyx7lCXoZVlvQrj0hPTM1HI4+v1sIda2yDvg=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_microsoft_hcsshim",
        importpath = "github.com/Microsoft/hcsshim",
        sum = "h1:AbV+VPfTrIVffukazHcpxmz/sRiE6YaMDzHWR9BXZHo=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_github_miekg_dns",
        importpath = "github.com/miekg/dns",
        sum = "h1:DQUfb9uc6smULcREF09Uc+/Gd46YWqJd5DbpPE9xkcA=",
        version = "v1.1.50",
    )
    go_repository(
        name = "com_github_miekg_pkcs11",
        importpath = "github.com/miekg/pkcs11",
        sum = "h1:Ugu9pdy6vAYku5DEpVWVFPYnzV+bxB+iRdbuFSu7TvU=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_mindprince_gonvml",
        importpath = "github.com/mindprince/gonvml",
        sum = "h1:PS1dLCGtD8bb9RPKJrc8bS7qHL6JnW1CZvwzH9dPoUs=",
        version = "v0.0.0-20190828220739-9ebdce4bb989",
    )
    go_repository(
        name = "com_github_mistifyio_go_zfs",
        importpath = "github.com/mistifyio/go-zfs",
        sum = "h1:aKW/4cBs+yK6gpqU3K/oIwk9Q/XICqd3zOX/UFuvqmk=",
        version = "v2.1.2-0.20190413222219-f784269be439+incompatible",
    )
    go_repository(
        name = "com_github_mitchellh_cli",
        importpath = "github.com/mitchellh/cli",
        sum = "h1:PvH+lL2B7IQ101xQL63Of8yFS2y+aDlsFcsqNc+u/Kw=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_mitchellh_colorstring",
        importpath = "github.com/mitchellh/colorstring",
        sum = "h1:62I3jR2EmQ4l5rM/4FEfDWcRD+abF5XlKShorW5LRoQ=",
        version = "v0.0.0-20190213212951-d06e56a500db",
    )
    go_repository(
        name = "com_github_mitchellh_copystructure",
        importpath = "github.com/mitchellh/copystructure",
        sum = "h1:vpKXTN4ewci03Vljg/q9QvCGUDttBOGBIa15WveJJGw=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_homedir",
        importpath = "github.com/mitchellh/go-homedir",
        sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_testing_interface",
        importpath = "github.com/mitchellh/go-testing-interface",
        sum = "h1:jrgshOhYAUVNMAJiKbEu7EqAwgJJ2JqpQmpLJOu07cU=",
        version = "v1.14.1",
    )
    go_repository(
        name = "com_github_mitchellh_go_wordwrap",
        importpath = "github.com/mitchellh/go-wordwrap",
        sum = "h1:6GlHJ/LTGMrIJbwgdqdl2eEH8o+Exx/0m8ir9Gns0u4=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_gox",
        importpath = "github.com/mitchellh/gox",
        sum = "h1:lfGJxY7ToLJQjHHwi0EX6uYBdK78egf954SQl13PQJc=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_mitchellh_iochan",
        importpath = "github.com/mitchellh/iochan",
        sum = "h1:C+X3KsSTLFVBr/tK1eYN/vs4rJcvsiLU338UhYPJWeY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_mapstructure",
        importpath = "github.com/mitchellh/mapstructure",
        sum = "h1:jeMsZIYE/09sWLaz43PL7Gy6RuMjD2eJVyuac5Z2hdY=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_mitchellh_reflectwalk",
        importpath = "github.com/mitchellh/reflectwalk",
        sum = "h1:G2LzWKi524PWgd3mLHV8Y5k7s6XUvT0Gef6zxSIeXaQ=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_moby_ipvs",
        importpath = "github.com/moby/ipvs",
        sum = "h1:aoZ7fhLTXgDbzVrAnvV+XbKOU8kOET7B3+xULDF/1o0=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_moby_locker",
        importpath = "github.com/moby/locker",
        sum = "h1:fOXqR41zeveg4fFODix+1Ch4mj/gT0NE1XJbp/epuBg=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_moby_spdystream",
        importpath = "github.com/moby/spdystream",
        sum = "h1:cjW1zVyyoiM0T7b6UoySUFqzXMoqRckQtXwGPiBhOM8=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_moby_sys_mountinfo",
        importpath = "github.com/moby/sys/mountinfo",
        sum = "h1:BzJjoreD5BMFNmD9Rus6gdd1pLuecOFPt8wC+Vygl78=",
        version = "v0.6.2",
    )
    go_repository(
        name = "com_github_moby_sys_signal",
        importpath = "github.com/moby/sys/signal",
        sum = "h1:aDpY94H8VlhTGa9sNYUFCFsMZIUh5wm0B6XkIoJj/iY=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_moby_sys_symlink",
        importpath = "github.com/moby/sys/symlink",
        sum = "h1:tk1rOM+Ljp0nFmfOIBtlV3rTDlWOwFRhjEeAhZB0nZc=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_moby_term",
        importpath = "github.com/moby/term",
        sum = "h1:dcztxKSvZ4Id8iPpHERQBbIJfabdt4wUm5qy3wOL2Zc=",
        version = "v0.0.0-20210619224110-3f7ff695adc6",
    )
    go_repository(
        name = "com_github_modern_go_concurrent",
        importpath = "github.com/modern-go/concurrent",
        sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
        version = "v0.0.0-20180306012644-bacd9c7ef1dd",
    )
    go_repository(
        name = "com_github_modern_go_reflect2",
        importpath = "github.com/modern-go/reflect2",
        sum = "h1:xBagoLtFs94CBntxluKeaWgTMpvLxC4ur3nMaC9Gz0M=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_modocache_gover",
        importpath = "github.com/modocache/gover",
        sum = "h1:8Q0qkMVC/MmWkpIdlvZgcv2o2jrlF6zqVOh7W5YHdMA=",
        version = "v0.0.0-20171022184752-b58185e213c5",
    )
    go_repository(
        name = "com_github_mohae_deepcopy",
        importpath = "github.com/mohae/deepcopy",
        sum = "h1:RWengNIwukTxcDr9M+97sNutRR1RKhG96O6jWumTTnw=",
        version = "v0.0.0-20170929034955-c48cc78d4826",
    )
    go_repository(
        name = "com_github_monochromegane_go_gitignore",
        importpath = "github.com/monochromegane/go-gitignore",
        sum = "h1:n6/2gBQ3RWajuToeY6ZtZTIKv2v7ThUy5KKusIT0yc0=",
        version = "v0.0.0-20200626010858-205db1a8cc00",
    )
    go_repository(
        name = "com_github_montanaflynn_stats",
        importpath = "github.com/montanaflynn/stats",
        sum = "h1:Duep6KMIDpY4Yo11iFsvyqJDyfzLF9+sndUKT+v64GQ=",
        version = "v0.6.6",
    )
    go_repository(
        name = "com_github_morikuni_aec",
        importpath = "github.com/morikuni/aec",
        sum = "h1:nP9CBfwrvYnBRgY6qfDQkygYDmYwOilePFkwzv4dU8A=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mrunalp_fileutils",
        importpath = "github.com/mrunalp/fileutils",
        sum = "h1:NKzVxiH7eSk+OQ4M+ZYW1K6h27RUV3MI6NUTsHhU6Z4=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_munnerz_goautoneg",
        importpath = "github.com/munnerz/goautoneg",
        sum = "h1:C3w9PqII01/Oq1c1nUAm88MOHcQC9l5mIlSMApZMrHA=",
        version = "v0.0.0-20191010083416-a7dc8b61c822",
    )
    go_repository(
        name = "com_github_mvdan_xurls",
        importpath = "github.com/mvdan/xurls",
        sum = "h1:OpuDelGQ1R1ueQ6sSryzi6P+1RtBpfQHM8fJwlE45ww=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_mwitkow_go_conntrack",
        importpath = "github.com/mwitkow/go-conntrack",
        sum = "h1:KUppIJq7/+SVif2QVs3tOP0zanoHgBEVAwHxUSIzRqU=",
        version = "v0.0.0-20190716064945-2f068394615f",
    )
    go_repository(
        name = "com_github_mwitkow_go_proto_validators",
        importpath = "github.com/mwitkow/go-proto-validators",
        sum = "h1:F6LFfmgVnfULfaRsQWBbe7F7ocuHCr9+7m+GAeDzNbQ=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_mxk_go_flowrate",
        importpath = "github.com/mxk/go-flowrate",
        sum = "h1:y5//uYreIhSUg3J1GEMiLbxo1LJaP8RfCpH6pymGZus=",
        version = "v0.0.0-20140419014527-cca7078d478f",
    )
    go_repository(
        name = "com_github_nats_io_jwt",
        importpath = "github.com/nats-io/jwt",
        sum = "h1:+RB5hMpXUUA2dfxuhBTEkMOrYmM+gKIZYS1KjSostMI=",
        version = "v0.3.2",
    )
    go_repository(
        name = "com_github_nats_io_nats_go",
        importpath = "github.com/nats-io/nats.go",
        sum = "h1:ik3HbLhZ0YABLto7iX80pZLPw/6dx3T+++MZJwLnMrQ=",
        version = "v1.9.1",
    )
    go_repository(
        name = "com_github_nats_io_nats_server_v2",
        importpath = "github.com/nats-io/nats-server/v2",
        sum = "h1:i2Ly0B+1+rzNZHHWtD4ZwKi+OU5l+uQo1iDHZ2PmiIc=",
        version = "v2.1.2",
    )
    go_repository(
        name = "com_github_nats_io_nkeys",
        importpath = "github.com/nats-io/nkeys",
        sum = "h1:6JrEfig+HzTH85yxzhSVbjHRJv9cn0p6n3IngIcM5/k=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_nats_io_nuid",
        importpath = "github.com/nats-io/nuid",
        sum = "h1:5iA8DT8V7q8WK2EScv2padNa/rTESc1KdnPw4TC2paw=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_niemeyer_pretty",
        importpath = "github.com/niemeyer/pretty",
        sum = "h1:fD57ERR4JtEqsWbfPhv4DMiApHyliiK5xCTNVSPiaAs=",
        version = "v0.0.0-20200227124842-a10e7caefd8e",
    )
    go_repository(
        name = "com_github_nishanths_predeclared",
        importpath = "github.com/nishanths/predeclared",
        sum = "h1:3f0nxAmdj/VoCGN/ijdMy7bj6SBagaqYg1B0hu8clMA=",
        version = "v0.0.0-20200524104333-86fad755b4d3",
    )
    go_repository(
        name = "com_github_nxadm_tail",
        importpath = "github.com/nxadm/tail",
        sum = "h1:nPr65rt6Y5JFSKQO7qToXr7pePgD6Gwiw05lkbyAQTE=",
        version = "v1.4.8",
    )
    go_repository(
        name = "com_github_nytimes_gziphandler",
        importpath = "github.com/NYTimes/gziphandler",
        sum = "h1:ZUDjpQae29j0ryrS0u/B8HZfJBtBQHjqw2rQ2cqUQ3I=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_oklog_oklog",
        importpath = "github.com/oklog/oklog",
        sum = "h1:wVfs8F+in6nTBMkA7CbRw+zZMIB7nNM825cM1wuzoTk=",
        version = "v0.3.2",
    )
    go_repository(
        name = "com_github_oklog_run",
        importpath = "github.com/oklog/run",
        sum = "h1:GEenZ1cK0+q0+wsJew9qUg/DyD8k3JzYsZAi5gYi2mA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_oklog_ulid",
        importpath = "github.com/oklog/ulid",
        sum = "h1:EGfNDEx6MqHz8B3uNV6QAib1UR2Lm97sHi3ocA6ESJ4=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_olekukonko_tablewriter",
        importpath = "github.com/olekukonko/tablewriter",
        sum = "h1:P2Ga83D34wi1o9J6Wh1mRuqd4mF/x/lgBS7N7AbDhec=",
        version = "v0.0.5",
    )
    go_repository(
        name = "com_github_oneofone_xxhash",
        importpath = "github.com/OneOfOne/xxhash",
        sum = "h1:KMrpdQIwFcEqXDklaen+P1axHaj9BSKzvpUUfnHldSE=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_onsi_ginkgo",
        importpath = "github.com/onsi/ginkgo",
        sum = "h1:8xi0RTUf59SOSfEtZMvwTvXYMzG4gV23XVHOZiXNtnE=",
        version = "v1.16.5",
    )
    go_repository(
        name = "com_github_onsi_ginkgo_v2",
        importpath = "github.com/onsi/ginkgo/v2",
        sum = "h1:/XxtEV3I3Eif/HobnVx9YmJgk8ENdRsuUmM+fLCFNow=",
        version = "v2.7.0",
    )
    go_repository(
        name = "com_github_onsi_gomega",
        importpath = "github.com/onsi/gomega",
        sum = "h1:J/tulyYK6JwBldPViHJReihxxZ+22FHs0piGjQAvoUE=",
        version = "v1.24.2",
    )
    go_repository(
        name = "com_github_op_go_logging",
        importpath = "github.com/op/go-logging",
        sum = "h1:lDH9UUVJtmYCjyT0CI4q8xvlXPxeZ0gYCVvWbmPlp88=",
        version = "v0.0.0-20160315200505-970db520ece7",
    )
    go_repository(
        name = "com_github_opencontainers_go_digest",
        importpath = "github.com/opencontainers/go-digest",
        sum = "h1:apOUWs51W5PlhuyGyz9FCeeBIOUDA/6nW8Oi/yOhh5U=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_opencontainers_image_spec",
        importpath = "github.com/opencontainers/image-spec",
        sum = "h1:2zx/Stx4Wc5pIPDvIxHXvXtQFW/7XWJGmnM7r3wg034=",
        version = "v1.1.0-rc2",
    )
    go_repository(
        name = "com_github_opencontainers_runc",
        importpath = "github.com/opencontainers/runc",
        sum = "h1:vIXrkId+0/J2Ymu2m7VjGvbSlAId9XNRPhn2p4b+d8w=",
        version = "v1.1.3",
    )
    go_repository(
        name = "com_github_opencontainers_runtime_spec",
        importpath = "github.com/opencontainers/runtime-spec",
        sum = "h1:3snG66yBm59tKhhSPQrQ/0bCrv1LQbKt40LnUPiUxdc=",
        version = "v1.0.3-0.20210326190908-1c3f411f0417",
    )
    go_repository(
        name = "com_github_opencontainers_selinux",
        importpath = "github.com/opencontainers/selinux",
        sum = "h1:09LIPVRP3uuZGQvgR+SgMSNBd1Eb3vlRbGqQpoHsF8w=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_github_opentracing_basictracer_go",
        importpath = "github.com/opentracing/basictracer-go",
        sum = "h1:YyUAhaEfjoWXclZVJ9sGoNct7j4TVk7lZWlQw5UXuoo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_opentracing_contrib_go_observer",
        importpath = "github.com/opentracing-contrib/go-observer",
        sum = "h1:lM6RxxfUMrYL/f8bWEUqdXrANWtrL7Nndbm9iFN0DlU=",
        version = "v0.0.0-20170622124052-a52f23424492",
    )
    go_repository(
        name = "com_github_opentracing_opentracing_go",
        importpath = "github.com/opentracing/opentracing-go",
        sum = "h1:uEJPy/1a5RIPAJ0Ov+OIO8OxWu77jEv+1B0VhjKrZUs=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_openzipkin_contrib_zipkin_go_opentracing",
        importpath = "github.com/openzipkin-contrib/zipkin-go-opentracing",
        sum = "h1:ZCnq+JUrvXcDVhX/xRolRBZifmabN1HcS1wrPSvxhrU=",
        version = "v0.4.5",
    )
    go_repository(
        name = "com_github_openzipkin_zipkin_go",
        importpath = "github.com/openzipkin/zipkin-go",
        sum = "h1:nY8Hti+WKaP0cRsSeQ026wU03QsM762XBeCXBb9NAWI=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_otiai10_copy",
        importpath = "github.com/otiai10/copy",
        sum = "h1:HvG945u96iNadPoG2/Ja2+AUJeW5YuFQMixq9yirC+k=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_otiai10_curr",
        importpath = "github.com/otiai10/curr",
        sum = "h1:TJIWdbX0B+kpNagQrjgq8bCMrbhiuX73M2XwgtDMoOI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_otiai10_mint",
        importpath = "github.com/otiai10/mint",
        sum = "h1:BCmzIS3n71sGfHB5NMNDB3lHYPz8fWSkCAErHed//qc=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_pact_foundation_pact_go",
        importpath = "github.com/pact-foundation/pact-go",
        sum = "h1:OYkFijGHoZAYbOIb1LWXrwKQbMMRUv1oQ89blD2Mh2Q=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_pascaldekloe_goe",
        importpath = "github.com/pascaldekloe/goe",
        sum = "h1:Lgl0gzECD8GnQ5QCWA8o6BtfL6mDH5rQgM4/fX3avOs=",
        version = "v0.0.0-20180627143212-57f6aae5913c",
    )
    go_repository(
        name = "com_github_pborman_uuid",
        importpath = "github.com/pborman/uuid",
        sum = "h1:J7Q5mO4ysT1dv8hyrUGHb9+ooztCXu1D8MY8DZYsu3g=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_pelletier_go_buffruneio",
        importpath = "github.com/pelletier/go-buffruneio",
        sum = "h1:U4t4R6YkofJ5xHm3dJzuRpPZ0mr5MMCoAWooScCR7aA=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_pelletier_go_toml",
        importpath = "github.com/pelletier/go-toml",
        sum = "h1:4yBQzkHv+7BHq2PQUZF3Mx0IYxG7LsP222s7Agd3ve8=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_github_pelletier_go_toml_v2",
        importpath = "github.com/pelletier/go-toml/v2",
        sum = "h1:ipoSadvV8oGUjnUbMub59IDPPwfxF694nG/jwbMiyQg=",
        version = "v2.0.5",
    )
    go_repository(
        name = "com_github_performancecopilot_speed",
        importpath = "github.com/performancecopilot/speed",
        sum = "h1:2WnRzIquHa5QxaJKShDkLM+sc0JPuwhXzK8OYOyt3Vg=",
        version = "v3.0.0+incompatible",
    )
    go_repository(
        name = "com_github_peterbourgon_diskv",
        importpath = "github.com/peterbourgon/diskv",
        sum = "h1:UBdAOUP5p4RWqPBg048CAvpKN+vxiaj6gdUUzhl4XmI=",
        version = "v2.0.1+incompatible",
    )
    go_repository(
        name = "com_github_phayes_freeport",
        importpath = "github.com/phayes/freeport",
        sum = "h1:Ii+DKncOVM8Cu1Hc+ETb5K+23HdAMvESYE3ZJ5b5cMI=",
        version = "v0.0.0-20220201140144-74d24b5ae9f5",
    )
    go_repository(
        name = "com_github_pierrec_lz4",
        importpath = "github.com/pierrec/lz4",
        sum = "h1:9UY3+iC23yxF0UfGaYrGplQ+79Rg+h/q9FV9ix19jjM=",
        version = "v2.6.1+incompatible",
    )
    go_repository(
        name = "com_github_pjbgf_sha1cd",
        importpath = "github.com/pjbgf/sha1cd",
        sum = "h1:uKQP/7QOzNtKYH7UTohZLcjF5/55EnTw0jO/Ru4jZwI=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_pkg_browser",
        importpath = "github.com/pkg/browser",
        sum = "h1:KoWmjvw+nsYOo29YJK9vDA65RGE3NrOnUtO7a+RF9HU=",
        version = "v0.0.0-20210911075715-681adbf594b8",
    )
    go_repository(
        name = "com_github_pkg_diff",
        importpath = "github.com/pkg/diff",
        sum = "h1:aoZm08cpOy4WuID//EZDgcC4zIxODThtZNPirFr42+A=",
        version = "v0.0.0-20210226163009-20ebb0f2a09e",
    )
    go_repository(
        name = "com_github_pkg_errors",
        importpath = "github.com/pkg/errors",
        sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_pkg_profile",
        importpath = "github.com/pkg/profile",
        sum = "h1:F++O52m40owAmADcojzM+9gyjmMOY/T4oYJkgFDH8RE=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_pkg_sftp",
        importpath = "github.com/pkg/sftp",
        sum = "h1:I2qBYMChEhIjOgazfJmV3/mZM256btk6wkCDRmW7JYs=",
        version = "v1.13.1",
    )
    go_repository(
        name = "com_github_pmezard_go_difflib",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_posener_complete",
        importpath = "github.com/posener/complete",
        sum = "h1:ccV59UEOTzVDnDUEFdT95ZzHVZ+5+158q8+SJb2QV5w=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_poy_onpar",
        importpath = "github.com/poy/onpar",
        sum = "h1:oL4IBbcqwhhNWh31bjOX8C/OCy0zs9906d/VUru+bqg=",
        version = "v0.0.0-20190519213022-ee068f8ea4d1",
    )
    go_repository(
        name = "com_github_pquerna_cachecontrol",
        importpath = "github.com/pquerna/cachecontrol",
        sum = "h1:yJMy84ti9h/+OEWa752kBTKv4XC30OtVVHYv/8cTqKc=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_prometheus_client_golang",
        importpath = "github.com/prometheus/client_golang",
        sum = "h1:b71QUfeo5M8gq2+evJdTPfZhYMAU0uKPkyPJ7TPsloU=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_github_prometheus_client_model",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:UBgGFHqYdG/TPFD1B1ogZywDqEkwp3fBMvqdiQ7Xew4=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_prometheus_common",
        importpath = "github.com/prometheus/common",
        sum = "h1:ccBbHCgIiT9uSoFY0vX8H3zsNR5eLt17/RQLUvn8pXE=",
        version = "v0.37.0",
    )
    go_repository(
        name = "com_github_prometheus_procfs",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:ODq8ZFEaYeCaZOJlZZdJA2AbQR98dSHSM1KW/You5mo=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_github_prometheus_prometheus",
        importpath = "github.com/prometheus/prometheus",
        sum = "h1:7QPitgO2kOFG8ecuRn9O/4L9+10He72rVRJvMXrE9Hg=",
        version = "v2.5.0+incompatible",
    )
    go_repository(
        name = "com_github_prometheus_tsdb",
        importpath = "github.com/prometheus/tsdb",
        sum = "h1:YZcsG11NqnK4czYLrWd9mpEuAJIHVQLwdrleYfszMAA=",
        version = "v0.7.1",
    )
    go_repository(
        name = "com_github_protonmail_go_crypto",
        importpath = "github.com/ProtonMail/go-crypto",
        sum = "h1:ra2OtmuW0AE5csawV4YXMNGNQQXvLRps3z2Z59OPO+I=",
        version = "v0.0.0-20221026131551-cf6655e29de4",
    )
    go_repository(
        name = "com_github_protonmail_go_mime",
        importpath = "github.com/ProtonMail/go-mime",
        sum = "h1:CGq7OieOz3wyQJ1fO8S0eO9TCW1JyvLrf8fhzz1i8ko=",
        version = "v0.0.0-20220302105931-303f85f7fe0f",
    )
    go_repository(
        name = "com_github_protonmail_gopenpgp_v2",
        importpath = "github.com/ProtonMail/gopenpgp/v2",
        sum = "h1:EYgkxzwmQvsa6kxxkgP1AwzkFqKHscF2UINxaSn6rdI=",
        version = "v2.4.10",
    )
    go_repository(
        name = "com_github_pseudomuto_protoc_gen_doc",
        importpath = "github.com/pseudomuto/protoc-gen-doc",
        sum = "h1:Ah259kcrio7Ix1Rhb6u8FCaOkzf9qRBqXnvAufg061w=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_pseudomuto_protokit",
        importpath = "github.com/pseudomuto/protokit",
        sum = "h1:hlnBDcy3YEDXH7kc9gV+NLaN0cDzhDvD1s7Y6FZ8RpM=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_puerkitobio_purell",
        importpath = "github.com/PuerkitoBio/purell",
        sum = "h1:WEQqlqaGbrPkxLJWfBwQmfEAE1Z7ONdDLqrN38tNFfI=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_puerkitobio_urlesc",
        importpath = "github.com/PuerkitoBio/urlesc",
        sum = "h1:d+Bc7a5rLufV/sSk/8dngufqelfh6jnri85riMAaF/M=",
        version = "v0.0.0-20170810143723-de5bf2ad4578",
    )
    go_repository(
        name = "com_github_qur_ar",
        importpath = "github.com/qur/ar",
        sum = "h1:A6sXY4zAECrW5Obx41PVMGr4kOw1rd1kmwcHa5M0dTg=",
        version = "v0.0.0-20130629153254-282534b91770",
    )
    go_repository(
        name = "com_github_rcrowley_go_metrics",
        importpath = "github.com/rcrowley/go-metrics",
        sum = "h1:9ZKAASQSHhDYGoxY8uLVpewe1GDZ2vu2Tr/vTdVAkFQ=",
        version = "v0.0.0-20181016184325-3113b8401b8a",
    )
    go_repository(
        name = "com_github_rivo_uniseg",
        importpath = "github.com/rivo/uniseg",
        sum = "h1:utMvzDsuh3suAEnhH0RdHmoPbU648o6CvXxTx4SBMOw=",
        version = "v0.4.3",
    )
    go_repository(
        name = "com_github_robfig_cron_v3",
        importpath = "github.com/robfig/cron/v3",
        sum = "h1:WdRxkvbJztn8LMz/QEvLN5sBU+xKpSqwwUO1Pjr4qDs=",
        version = "v3.0.1",
    )
    go_repository(
        name = "com_github_rogpeppe_fastuuid",
        importpath = "github.com/rogpeppe/fastuuid",
        sum = "h1:Ppwyp6VYCF1nvBTXL3trRso7mXMlRrw9ooo375wvi2s=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_rogpeppe_go_internal",
        importpath = "github.com/rogpeppe/go-internal",
        sum = "h1:geMPLpDpQOgVyCg5z5GoRwLHepNdb71NXb67XFkP+Eg=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_rs_cors",
        importpath = "github.com/rs/cors",
        sum = "h1:KCooALfAYGs415Cwu5ABvv9n9509fSiG5SQJn/AQo4U=",
        version = "v1.8.2",
    )
    go_repository(
        name = "com_github_rubenv_sql_migrate",
        importpath = "github.com/rubenv/sql-migrate",
        sum = "h1:9M6oj4e//owVVHYrFISmY9LBRw6gzkCNmD9MV36tZeQ=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_rubiojr_go_vhd",
        importpath = "github.com/rubiojr/go-vhd",
        sum = "h1:if3/24+h9Sq6eDx8UUz1SO9cT9tizyIsATfB7b4D3tc=",
        version = "v0.0.0-20200706105327-02e210299021",
    )
    go_repository(
        name = "com_github_russross_blackfriday",
        importpath = "github.com/russross/blackfriday",
        sum = "h1:KqfZb0pUVN2lYqZUYRddxF4OR8ZMURnJIG5Y3VRLtww=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_russross_blackfriday_v2",
        importpath = "github.com/russross/blackfriday/v2",
        sum = "h1:JIOH55/0cWyOuilr9/qlrm0BSXldqnqwMsf35Ld67mk=",
        version = "v2.1.0",
    )
    go_repository(
        name = "com_github_ryanuber_columnize",
        importpath = "github.com/ryanuber/columnize",
        sum = "h1:UFr9zpz4xgTnIE5yIMtWAMngCdZ9p/+q6lTbgelo80M=",
        version = "v0.0.0-20160712163229-9b3edd62028f",
    )
    go_repository(
        name = "com_github_ryanuber_go_glob",
        importpath = "github.com/ryanuber/go-glob",
        sum = "h1:iQh3xXAumdQ+4Ufa5b25cRpC5TYKlno6hsv6Cb3pkBk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_samuel_go_zookeeper",
        importpath = "github.com/samuel/go-zookeeper",
        sum = "h1:p3Vo3i64TCLY7gIfzeQaUJ+kppEO5WQG3cL8iE8tGHU=",
        version = "v0.0.0-20190923202752-2cc03de413da",
    )
    go_repository(
        name = "com_github_sassoftware_go_rpmutils",
        importpath = "github.com/sassoftware/go-rpmutils",
        sum = "h1:ZHMXpGoMHL/cpQJ1byyhEBW73TDtpNq4inYa9M4FxyY=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_sassoftware_relic",
        importpath = "github.com/sassoftware/relic",
        sum = "h1:sUNzanSKA9z/h8xXl+ZJoxIYZL0Qx306MmxqRrvUgr0=",
        version = "v0.0.0-20210427151427-dfb082b79b74",
    )
    go_repository(
        name = "com_github_satori_go_uuid",
        importpath = "github.com/satori/go.uuid",
        sum = "h1:0uYX9dsZ2yD7q2RtLRtPSdGDWzjeM3TbMJP9utgA0ww=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_schollz_progressbar_v3",
        importpath = "github.com/schollz/progressbar/v3",
        sum = "h1:9TeeWRcjW2qd05I8Kf9knPkW4vLM/hYoa6z9ABvxje8=",
        version = "v3.13.0",
    )
    go_repository(
        name = "com_github_sean_seed",
        importpath = "github.com/sean-/seed",
        sum = "h1:nn5Wsu0esKSJiIVhscUtVbo7ada43DJhG55ua/hjS5I=",
        version = "v0.0.0-20170313163322-e2103e2c3529",
    )
    go_repository(
        name = "com_github_sebdah_goldie",
        importpath = "github.com/sebdah/goldie",
        sum = "h1:9GNhIat69MSlz/ndaBg48vl9dF5fI+NBB6kfOxgfkMc=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_seccomp_libseccomp_golang",
        importpath = "github.com/seccomp/libseccomp-golang",
        sum = "h1:RpforrEYXWkmGwJHIGnLZ3tTWStkjVVstwzNGqxX2Ds=",
        version = "v0.9.2-0.20220502022130-f33da4d89646",
    )
    go_repository(
        name = "com_github_secure_systems_lab_go_securesystemslib",
        importpath = "github.com/secure-systems-lab/go-securesystemslib",
        sum = "h1:b23VGrQhTA8cN2CbBw7/FulN9fTtqYUdS5+Oxzt+DUE=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_segmentio_ksuid",
        importpath = "github.com/segmentio/ksuid",
        sum = "h1:sBo2BdShXjmcugAMwjugoGUdUV0pcxY5mW4xKRn3v4c=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_sergi_go_diff",
        importpath = "github.com/sergi/go-diff",
        sum = "h1:XU+rvMAioB0UC3q1MFrIQy4Vo5/4VsRDQQXHsEya6xQ=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_shibumi_go_pathspec",
        importpath = "github.com/shibumi/go-pathspec",
        sum = "h1:QUyMZhFo0Md5B8zV8x2tesohbb5kfbpTi9rBnKh5dkI=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_shopify_logrus_bugsnag",
        importpath = "github.com/Shopify/logrus-bugsnag",
        sum = "h1:UrqY+r/OJnIp5u0s1SbQ8dVfLCZJsnvazdBP5hS4iRs=",
        version = "v0.0.0-20171204204709-577dee27f20d",
    )
    go_repository(
        name = "com_github_shopify_sarama",
        importpath = "github.com/Shopify/sarama",
        sum = "h1:9oksLxC6uxVPHPVYUmq6xhr1BOF/hHobWH2UzO67z1s=",
        version = "v1.19.0",
    )
    go_repository(
        name = "com_github_shopify_toxiproxy",
        importpath = "github.com/Shopify/toxiproxy",
        sum = "h1:TKdv8HiTLgE5wdJuEML90aBgNWsokNbMijUGhmcoBJc=",
        version = "v2.1.4+incompatible",
    )
    go_repository(
        name = "com_github_shopspring_decimal",
        importpath = "github.com/shopspring/decimal",
        sum = "h1:abSATXmQEYyShuxI4/vyW3tV1MrKAJzCZ/0zLUXYbsQ=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_shurcool_sanitized_anchor_name",
        importpath = "github.com/shurcooL/sanitized_anchor_name",
        sum = "h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_siderolabs_crypto",
        importpath = "github.com/siderolabs/crypto",
        sum = "h1:o1KIR1KyevUcY9nbJlSyQAj7+p+rveGGF8LjAAFMtjc=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_siderolabs_gen",
        importpath = "github.com/siderolabs/gen",
        sum = "h1:NoPmxjLFwOGRFaKypk2hoT5/XIzoM6AhF0eqy61bQLk=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_siderolabs_go_api_signature",
        importpath = "github.com/siderolabs/go-api-signature",
        sum = "h1:C5tUzuFsJYidpYyVfJGYpgQwETglA8B62ET4obkLDGE=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_siderolabs_go_blockdevice",
        importpath = "github.com/siderolabs/go-blockdevice",
        sum = "h1:4vY4B6amYRLhohqJkmVB4FIKJnFy7N21r7Lm3huag9w=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_siderolabs_go_debug",
        importpath = "github.com/siderolabs/go-debug",
        sum = "h1:c8styCvp+MO0oPO8q4N1CKSF3fVuAT0qnuUIeZ/BiW0=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_siderolabs_go_pointer",
        importpath = "github.com/siderolabs/go-pointer",
        sum = "h1:6TshPKep2doDQJAAtHUuHWXbca8ZfyRySjSBT/4GsMU=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_siderolabs_net",
        importpath = "github.com/siderolabs/net",
        sum = "h1:1bOgVay/ijPkJz4qct98nHsiB/ysLQU0KLoBC4qLm7I=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_siderolabs_protoenc",
        importpath = "github.com/siderolabs/protoenc",
        sum = "h1:QFxWIAo//12+/bm27GNYoK/TpQGTYsRrrZCu9jSghvU=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_siderolabs_talos_pkg_machinery",
        importpath = "github.com/siderolabs/talos/pkg/machinery",
        sum = "h1:UqaPWMzxu85SjM09c2jn6LPbOABKCevYvau4I/fxI0Y=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_sigstore_rekor",
        importpath = "github.com/sigstore/rekor",
        sum = "h1:rcESXSNkAPRWFYZel9rarspdvneET60F2ngNkadi89c=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore",
        importpath = "github.com/sigstore/sigstore",
        sum = "h1:iUou0QJW8eQKMUkTXbFyof9ZOblDtfaW2Sn2+QI8Tcs=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_sirupsen_logrus",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:trlNQbNUG3OdDrDil03MCb1H2o9nJ1x4/5LYw7byDE0=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_github_skeema_knownhosts",
        importpath = "github.com/skeema/knownhosts",
        sum = "h1:Wvr9V0MxhjRbl3f9nMnKnFfiWTJmtECJ9Njkea3ysW0=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_skratchdot_open_golang",
        importpath = "github.com/skratchdot/open-golang",
        sum = "h1:JIAuq3EEf9cgbU6AtGPK4CTG3Zf6CKMNqf0MHTggAUA=",
        version = "v0.0.0-20200116055534-eef842397966",
    )
    go_repository(
        name = "com_github_smartystreets_assertions",
        importpath = "github.com/smartystreets/assertions",
        sum = "h1:UVQPSSmc3qtTi+zPPkCXvZX9VvW/xT/NsRvKfwY81a8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_smartystreets_go_aws_auth",
        importpath = "github.com/smartystreets/go-aws-auth",
        sum = "h1:hp2CYQUINdZMHdvTdXtPOY2ainKl4IoMcpAXEf2xj3Q=",
        version = "v0.0.0-20180515143844-0c1422d1fdb9",
    )
    go_repository(
        name = "com_github_smartystreets_goconvey",
        importpath = "github.com/smartystreets/goconvey",
        sum = "h1:fv0U8FUIMPNf1L9lnHLvLhgicrIVChEkdzIKYqbNC9s=",
        version = "v1.6.4",
    )
    go_repository(
        name = "com_github_smartystreets_gunit",
        importpath = "github.com/smartystreets/gunit",
        sum = "h1:RyPDUFcJbvtXlhJPk7v+wnxZRY2EUokhEYl2EJOPToI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_soheilhy_cmux",
        importpath = "github.com/soheilhy/cmux",
        sum = "h1:jjzc5WVemNEDTLwv9tlmemhC73tI08BNOIGwBOo10Js=",
        version = "v0.1.5",
    )
    go_repository(
        name = "com_github_sony_gobreaker",
        importpath = "github.com/sony/gobreaker",
        sum = "h1:oMnRNZXX5j85zso6xCPRNPtmAycat+WcoKbklScLDgQ=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_spaolacci_murmur3",
        importpath = "github.com/spaolacci/murmur3",
        sum = "h1:qLC7fQah7D6K1B0ujays3HV9gkFtllcxhzImRR7ArPQ=",
        version = "v0.0.0-20180118202830-f09979ecbc72",
    )
    go_repository(
        name = "com_github_spf13_afero",
        importpath = "github.com/spf13/afero",
        sum = "h1:41FoI0fD7OR7mGcKE/aOiLkGreyf8ifIOQmJANWogMk=",
        version = "v1.9.3",
    )
    go_repository(
        name = "com_github_spf13_cast",
        importpath = "github.com/spf13/cast",
        sum = "h1:rj3WzYc11XZaIZMPKmwP96zkFEnnAmV8s6XbB2aY32w=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_spf13_cobra",
        importpath = "github.com/spf13/cobra",
        sum = "h1:o94oiPyS4KD1mPy2fmcYYHHfCxLqYjJOhGsCHFZtEzA=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_github_spf13_jwalterweatherman",
        importpath = "github.com/spf13/jwalterweatherman",
        sum = "h1:ue6voC5bR5F8YxI5S67j9i582FU4Qvo2bmqnqMYADFk=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_spf13_pflag",
        importpath = "github.com/spf13/pflag",
        sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
        version = "v1.0.5",
    )
    go_repository(
        name = "com_github_spf13_viper",
        importpath = "github.com/spf13/viper",
        sum = "h1:BWSJ/M+f+3nmdz9bxB+bWX28kkALN2ok11D0rSo8EJU=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_github_src_d_gcfg",
        importpath = "github.com/src-d/gcfg",
        sum = "h1:xXbNR5AlLSA315x2UO+fTSSAXCDf+Ar38/6oyGbDKQ4=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_stefanberger_go_pkcs11uri",
        importpath = "github.com/stefanberger/go-pkcs11uri",
        sum = "h1:lIOOHPEbXzO3vnmx2gok1Tfs31Q8GQqKLc8vVqyQq/I=",
        version = "v0.0.0-20201008174630-78d3cae3a980",
    )
    go_repository(
        name = "com_github_stoewer_go_strcase",
        importpath = "github.com/stoewer/go-strcase",
        sum = "h1:Z2iHWqGXH00XYgqDmNgQbIBxf3wrNq0F3feEy0ainaU=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_streadway_amqp",
        importpath = "github.com/streadway/amqp",
        sum = "h1:kuuDrUJFZL1QYL9hUNuCxNObNzB0bV/ZG5jV3RWAQgo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_streadway_handy",
        importpath = "github.com/streadway/handy",
        sum = "h1:AhmOdSHeswKHBjhsLs/7+1voOxT+LLrSk/Nxvk35fug=",
        version = "v0.0.0-20190108123426-d5acb3125c2a",
    )
    go_repository(
        name = "com_github_stretchr_objx",
        importpath = "github.com/stretchr/objx",
        sum = "h1:1zr/of2m5FGMsad5YfcqgdqdWrIhu+EBEJRhR1U7z/c=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_stretchr_testify",
        importpath = "github.com/stretchr/testify",
        sum = "h1:w7B6lhMri9wdJUVmEZPGGhZzrYTPvgJArz7wNPgYKsk=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_subosito_gotenv",
        importpath = "github.com/subosito/gotenv",
        sum = "h1:jyEFiXpy21Wm81FBN71l9VoMMV8H8jG+qIK3GCpY6Qs=",
        version = "v1.4.1",
    )
    go_repository(
        name = "com_github_syndtr_gocapability",
        importpath = "github.com/syndtr/gocapability",
        sum = "h1:kdXcSzyDtseVEc4yCz2qF8ZrQvIDBJLl4S1c3GCXmoI=",
        version = "v0.0.0-20200815063812-42c35b437635",
    )
    go_repository(
        name = "com_github_syndtr_goleveldb",
        importpath = "github.com/syndtr/goleveldb",
        sum = "h1:vfofYNRScrDdvS342BElfbETmL1Aiz3i2t0zfRj16Hs=",
        version = "v1.0.1-0.20220721030215-126854af5e6d",
    )
    go_repository(
        name = "com_github_tchap_go_patricia",
        importpath = "github.com/tchap/go-patricia",
        sum = "h1:JvoDL7JSoIP2HDE8AbDH3zC8QBPxmzYe32HHy5yQ+Ck=",
        version = "v2.2.6+incompatible",
    )
    go_repository(
        name = "com_github_tedsuo_ifrit",
        importpath = "github.com/tedsuo/ifrit",
        sum = "h1:LUUe4cdABGrIJAhl1P1ZpWY76AwukVszFdwkVFVLwIk=",
        version = "v0.0.0-20180802180643-bea94bb476cc",
    )
    go_repository(
        name = "com_github_tent_canonical_json_go",
        importpath = "github.com/tent/canonical-json-go",
        sum = "h1:iGnD/q9160NWqKZZ5vY4p0dMiYMRknzctfSkqA4nBDw=",
        version = "v0.0.0-20130607151641-96e4ba3a7613",
    )
    go_repository(
        name = "com_github_theupdateframework_go_tuf",
        importpath = "github.com/theupdateframework/go-tuf",
        sum = "h1:1i/Afw3rmaR1gF3sfVkG2X6ldkikQwA9zY380LrR5YI=",
        version = "v0.5.2-0.20220930112810-3890c1e7ace4",
    )
    go_repository(
        name = "com_github_tidwall_pretty",
        importpath = "github.com/tidwall/pretty",
        sum = "h1:RWIZEg2iJ8/g6fDDYzMpobmaoGh5OLl4AXtGUGPcqCs=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_tilinna_clock",
        importpath = "github.com/tilinna/clock",
        sum = "h1:6IQQQCo6KoBxVudv6gwtY8o4eDfhHo8ojA5dP0MfhSs=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_titanous_rocacheck",
        importpath = "github.com/titanous/rocacheck",
        sum = "h1:e/5i7d4oYZ+C1wj2THlRK+oAhjeS/TRQwMfkIuet3w0=",
        version = "v0.0.0-20171023193734-afe73141d399",
    )
    go_repository(
        name = "com_github_tj_assert",
        importpath = "github.com/tj/assert",
        sum = "h1:Rw8kxzWo1mr6FSaYXjQELRe88y2KdfynXdnK72rdjtA=",
        version = "v0.0.0-20171129193455-018094318fb0",
    )
    go_repository(
        name = "com_github_tj_go_elastic",
        importpath = "github.com/tj/go-elastic",
        sum = "h1:eGaGNxrtoZf/mBURsnNQKDR7u50Klgcf2eFDQEnc8Bc=",
        version = "v0.0.0-20171221160941-36157cbbebc2",
    )
    go_repository(
        name = "com_github_tj_go_kinesis",
        importpath = "github.com/tj/go-kinesis",
        sum = "h1:m74UWYy+HBs+jMFR9mdZU6shPewugMyH5+GV6LNgW8w=",
        version = "v0.0.0-20171128231115-08b17f58cb1b",
    )
    go_repository(
        name = "com_github_tj_go_spin",
        importpath = "github.com/tj/go-spin",
        sum = "h1:lhdWZsvImxvZ3q1C5OIB7d72DuOwP4O2NdBg9PyzNds=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_tmc_grpc_websocket_proxy",
        importpath = "github.com/tmc/grpc-websocket-proxy",
        sum = "h1:uruHq4dN7GR16kFc5fp3d1RIYzJW5onx8Ybykw2YQFA=",
        version = "v0.0.0-20201229170055-e5319fda7802",
    )
    go_repository(
        name = "com_github_tomasen_realip",
        importpath = "github.com/tomasen/realip",
        sum = "h1:fb190+cK2Xz/dvi9Hv8eCYJYvIGUTN2/KLq1pT6CjEc=",
        version = "v0.0.0-20180522021738-f0c99a92ddce",
    )
    go_repository(
        name = "com_github_transparency_dev_merkle",
        importpath = "github.com/transparency-dev/merkle",
        sum = "h1:T9/9gYB8uZl7VOJIhdwjALeRWlxUxSfDEysjfmx+L9E=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_ugorji_go",
        importpath = "github.com/ugorji/go",
        sum = "h1:/68gy2h+1mWMrwZFeD1kQialdSzAb432dtpeJ42ovdo=",
        version = "v1.1.7",
    )
    go_repository(
        name = "com_github_ugorji_go_codec",
        importpath = "github.com/ugorji/go/codec",
        sum = "h1:2SvQaVZ1ouYrrKKwoSk2pzd4A9evlKJb9oTL+OaLUSs=",
        version = "v1.1.7",
    )
    go_repository(
        name = "com_github_ulikunitz_xz",
        importpath = "github.com/ulikunitz/xz",
        sum = "h1:t92gobL9l3HE202wg3rlk19F6X+JOxl9BBrCCMYEYd8=",
        version = "v0.5.10",
    )
    go_repository(
        name = "com_github_urfave_cli",
        importpath = "github.com/urfave/cli",
        sum = "h1:aXiFAgRugfJ27UFDsGJ9DB2FvTC73hlVXFSqq5bo9eU=",
        version = "v1.22.7",
    )
    go_repository(
        name = "com_github_urfave_negroni",
        importpath = "github.com/urfave/negroni",
        sum = "h1:kIimOitoypq34K7TG7DUaJ9kq/N4Ofuwi1sjz0KipXc=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_vbatts_tar_split",
        importpath = "github.com/vbatts/tar-split",
        sum = "h1:Via6XqJr0hceW4wff3QRzD5gAk/tatMw/4ZA7cTlIME=",
        version = "v0.11.2",
    )
    go_repository(
        name = "com_github_veraison_go_cose",
        importpath = "github.com/veraison/go-cose",
        sum = "h1:4qA7dbFJGvt7gcqv5MCIyCQvN+NpHFPkW7do3EeDLb8=",
        version = "v1.0.0-rc.1",
    )
    go_repository(
        name = "com_github_vishvananda_netlink",
        importpath = "github.com/vishvananda/netlink",
        sum = "h1:+UB2BJA852UkGH42H+Oee69djmxS3ANzl2b/JtT1YiA=",
        version = "v1.1.1-0.20210330154013-f5de75959ad5",
    )
    go_repository(
        name = "com_github_vishvananda_netns",
        importpath = "github.com/vishvananda/netns",
        sum = "h1:p4VB7kIXpOQvVn1ZaTIVp+3vuYAXFe3OJEvjbUYJLaA=",
        version = "v0.0.0-20210104183010-2eb08e3e575f",
    )
    go_repository(
        name = "com_github_vividcortex_gohistogram",
        importpath = "github.com/VividCortex/gohistogram",
        sum = "h1:6+hBz+qvs0JOrrNhhmR7lFxo5sINxBCGXrdtl/UvroE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_vmihailenco_msgpack",
        importpath = "github.com/vmihailenco/msgpack",
        sum = "h1:wapg9xDUZDzGCNFlwc5SqI1rvcciqcxEHac4CYj89xI=",
        version = "v3.3.3+incompatible",
    )
    go_repository(
        name = "com_github_vmihailenco_msgpack_v4",
        importpath = "github.com/vmihailenco/msgpack/v4",
        sum = "h1:07s4sz9IReOgdikxLTKNbBdqDMLsjPKXwvCazn8G65U=",
        version = "v4.3.12",
    )
    go_repository(
        name = "com_github_vmihailenco_msgpack_v5",
        importpath = "github.com/vmihailenco/msgpack/v5",
        sum = "h1:5gO0H1iULLWGhs2H5tbAHIZTV8/cYafcFOr9znI5mJU=",
        version = "v5.3.5",
    )
    go_repository(
        name = "com_github_vmihailenco_tagparser",
        importpath = "github.com/vmihailenco/tagparser",
        sum = "h1:quXMXlA39OCbd2wAdTsGDlK9RkOk6Wuw+x37wVyIuWY=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_vmihailenco_tagparser_v2",
        importpath = "github.com/vmihailenco/tagparser/v2",
        sum = "h1:y09buUbR+b5aycVFQs/g70pqKVZNBmxwAhO7/IwNM9g=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_vmware_govmomi",
        importpath = "github.com/vmware/govmomi",
        sum = "h1:gpw/0Ku+6RgF3jsi7fnCLmlcikBHfKBCUcu1qgc16OU=",
        version = "v0.20.3",
    )
    go_repository(
        name = "com_github_weppos_publicsuffix_go",
        importpath = "github.com/weppos/publicsuffix-go",
        sum = "h1:eR9jm8DVMdrDUuVji4eOxPK4r/dANDlDBdISSUUV96s=",
        version = "v0.20.1-0.20221031080346-e4081aa8a6de",
    )
    go_repository(
        name = "com_github_x448_float16",
        importpath = "github.com/x448/float16",
        sum = "h1:qLwI1I70+NjRFUR3zs1JPUCgaCXSh3SW62uAKT1mSBM=",
        version = "v0.8.4",
    )
    go_repository(
        name = "com_github_xanzy_go_gitlab",
        importpath = "github.com/xanzy/go-gitlab",
        sum = "h1:+nHztQuCXGSMluKe5Q9IRaPdz6tO8O0gMkQ0vqGpiBk=",
        version = "v0.31.0",
    )
    go_repository(
        name = "com_github_xanzy_ssh_agent",
        importpath = "github.com/xanzy/ssh-agent",
        sum = "h1:+/15pJfg/RsTxqYcX6fHqOXZwwMP+2VyYWJeWM2qQFM=",
        version = "v0.3.3",
    )
    go_repository(
        name = "com_github_xdg_go_pbkdf2",
        importpath = "github.com/xdg-go/pbkdf2",
        sum = "h1:Su7DPu48wXMwC3bs7MCNG+z4FhcyEuz5dlvchbq0B0c=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_xdg_go_scram",
        importpath = "github.com/xdg-go/scram",
        sum = "h1:VOMT+81stJgXW3CpHyqHN3AXDYIMsx56mEFrB37Mb/E=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_xdg_go_stringprep",
        importpath = "github.com/xdg-go/stringprep",
        sum = "h1:kdwGpVNwPFtjs98xCGkHjQtGKh86rDcRZN17QEMCOIs=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonpointer",
        importpath = "github.com/xeipuuv/gojsonpointer",
        sum = "h1:J9EGpcZtP0E/raorCMxlFGSTBrsSlaDGf3jU/qvAE2c=",
        version = "v0.0.0-20180127040702-4e3ac2762d5f",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonreference",
        importpath = "github.com/xeipuuv/gojsonreference",
        sum = "h1:EzJWgHovont7NscjpAxXsDA8S8BMYve8Y5+7cuRE7R0=",
        version = "v0.0.0-20180127040603-bd5ef7bd5415",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonschema",
        importpath = "github.com/xeipuuv/gojsonschema",
        sum = "h1:LhYJRs+L4fBtjZUfuSZIKGeVu0QRy8e5Xi7D17UxZ74=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_xi2_xz",
        importpath = "github.com/xi2/xz",
        sum = "h1:nIPpBwaJSVYIxUFsDv3M8ofmx9yWTog9BfvIu0q41lo=",
        version = "v0.0.0-20171230120015-48954b6210f8",
    )
    go_repository(
        name = "com_github_xiang90_probing",
        importpath = "github.com/xiang90/probing",
        sum = "h1:eY9dn8+vbi4tKz5Qo6v2eYzo7kUS51QINcR5jNpbZS8=",
        version = "v0.0.0-20190116061207-43a291ad63a2",
    )
    go_repository(
        name = "com_github_xlab_treeprint",
        importpath = "github.com/xlab/treeprint",
        sum = "h1:G/1DjNkPpfZCFt9CSh6b5/nY4VimlbHF3Rh4obvtzDk=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_xordataexchange_crypt",
        importpath = "github.com/xordataexchange/crypt",
        sum = "h1:ESFSdwYZvkeru3RtdrYueztKhOBCSAAzS4Gf+k0tEow=",
        version = "v0.0.3-0.20170626215501-b2862e3d0a77",
    )
    go_repository(
        name = "com_github_youmark_pkcs8",
        importpath = "github.com/youmark/pkcs8",
        sum = "h1:splanxYIlg+5LfHAM6xpdFEAYOk8iySO56hMFq6uLyA=",
        version = "v0.0.0-20181117223130-1be2e3e5546d",
    )
    go_repository(
        name = "com_github_ysmood_goob",
        importpath = "github.com/ysmood/goob",
        sum = "h1:HsxXhyLBeGzWXnqVKtmT9qM7EuVs/XOgkX7T6r1o1AQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_ysmood_gson",
        importpath = "github.com/ysmood/gson",
        sum = "h1:QFkWbTH8MxyUTKPkVWAENJhxqdBa4lYTQWqZCiLG6kE=",
        version = "v0.7.3",
    )
    go_repository(
        name = "com_github_ysmood_leakless",
        importpath = "github.com/ysmood/leakless",
        sum = "h1:BzLrVoiwxikpgEQR0Lk8NyBN5Cit2b1z+u0mgL4ZJak=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_github_yuin_goldmark",
        importpath = "github.com/yuin/goldmark",
        sum = "h1:fVcFKWvrslecOb/tg+Cc05dkeYx540o0FuFt3nUVDoE=",
        version = "v1.4.13",
    )
    go_repository(
        name = "com_github_yvasiyarov_go_metrics",
        importpath = "github.com/yvasiyarov/go-metrics",
        sum = "h1:+lm10QQTNSBd8DVTNGHx7o/IKu9HYDvLMffDhbyLccI=",
        version = "v0.0.0-20140926110328-57bccd1ccd43",
    )
    go_repository(
        name = "com_github_yvasiyarov_gorelic",
        importpath = "github.com/yvasiyarov/gorelic",
        sum = "h1:hlE8//ciYMztlGpl/VA+Zm1AcTPHYkHJPbHqE6WJUXE=",
        version = "v0.0.0-20141212073537-a9bba5b9ab50",
    )
    go_repository(
        name = "com_github_yvasiyarov_newrelic_platform_go",
        importpath = "github.com/yvasiyarov/newrelic_platform_go",
        sum = "h1:ERexzlUfuTvpE74urLSbIQW0Z/6hF9t8U4NsJLaioAY=",
        version = "v0.0.0-20140908184405-b21fdbd4370f",
    )
    go_repository(
        name = "com_github_zalando_go_keyring",
        importpath = "github.com/zalando/go-keyring",
        sum = "h1:w2V9lcx/Uj4l+dzAf1m9s+DJ1O8ROkEHnynonHjTcYE=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_zclconf_go_cty",
        importpath = "github.com/zclconf/go-cty",
        sum = "h1:726SxLdi2SDnjY+BStqB9J1hNp4+2WlzyXLuimibIe0=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_github_zclconf_go_cty_debug",
        importpath = "github.com/zclconf/go-cty-debug",
        sum = "h1:FosyBZYxY34Wul7O/MSKey3txpPYyCqVO5ZyceuQJEI=",
        version = "v0.0.0-20191215020915-b22d67c1ba0b",
    )
    go_repository(
        name = "com_github_ziutek_mymysql",
        importpath = "github.com/ziutek/mymysql",
        sum = "h1:GB0qdRGsTwQSBVYuVShFBKaXSnSnYYC2d9knnE1LHFs=",
        version = "v1.5.4",
    )
    go_repository(
        name = "com_github_zmap_zcrypto",
        importpath = "github.com/zmap/zcrypto",
        sum = "h1:+nr36qrZEH0RIYNjcUEnOrCUdcSG3om2ANaFA6iSVWA=",
        version = "v0.0.0-20220402174210-599ec18ecbac",
    )
    go_repository(
        name = "com_github_zmap_zlint_v3",
        importpath = "github.com/zmap/zlint/v3",
        sum = "h1:Xs/lrMJY74MpJx/jSx2oVvZBrqlyUyFaLLBRyf68cqg=",
        version = "v3.4.0",
    )
    go_repository(
        name = "com_google_cloud_go",
        importpath = "cloud.google.com/go",
        sum = "h1:qkj22L7bgkl6vIeZDlOY2po43Mx/TIa2Wsa7VR+PEww=",
        version = "v0.107.0",
    )
    go_repository(
        name = "com_google_cloud_go_accessapproval",
        importpath = "cloud.google.com/go/accessapproval",
        sum = "h1:/nTivgnV/n1CaAeo+ekGexTYUsKEU9jUVkoY5359+3Q=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_accesscontextmanager",
        importpath = "cloud.google.com/go/accesscontextmanager",
        sum = "h1:CFhNhU7pcD11cuDkQdrE6PQJgv0EXNKNv06jIzbLlCU=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_aiplatform",
        importpath = "cloud.google.com/go/aiplatform",
        sum = "h1:DBi3Jk9XjCJ4pkkLM4NqKgj3ozUL1wq4l+d3/jTGXAI=",
        version = "v1.27.0",
    )
    go_repository(
        name = "com_google_cloud_go_analytics",
        importpath = "cloud.google.com/go/analytics",
        sum = "h1:NKw6PpQi6V1O+KsjuTd+bhip9d0REYu4NevC45vtGp8=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_google_cloud_go_apigateway",
        importpath = "cloud.google.com/go/apigateway",
        sum = "h1:IIoXKR7FKrEAQhMTz5hK2wiDz2WNFHS7eVr/L1lE/rM=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeconnect",
        importpath = "cloud.google.com/go/apigeeconnect",
        sum = "h1:AONoTYJviyv1vS4IkvWzq69gEVdvHx35wKXc+e6wjZQ=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_appengine",
        importpath = "cloud.google.com/go/appengine",
        sum = "h1:lmG+O5oaR9xNwaRBwE2XoMhwQHsHql5IoiGr1ptdDwU=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_area120",
        importpath = "cloud.google.com/go/area120",
        sum = "h1:TCMhwWEWhCn8d44/Zs7UCICTWje9j3HuV6nVGMjdpYw=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_artifactregistry",
        importpath = "cloud.google.com/go/artifactregistry",
        sum = "h1:3d0LRAU1K6vfqCahhl9fx2oGHcq+s5gftdix4v8Ibrc=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_asset",
        importpath = "cloud.google.com/go/asset",
        sum = "h1:aCrlaLGJWTODJX4G56ZYzJefITKEWNfbjjtHSzWpxW0=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_assuredworkloads",
        importpath = "cloud.google.com/go/assuredworkloads",
        sum = "h1:hhIdCOowsT1GG5eMCIA0OwK6USRuYTou/1ZeNxCSRtA=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_automl",
        importpath = "cloud.google.com/go/automl",
        sum = "h1:BMioyXSbg7d7xLibn47cs0elW6RT780IUWr42W8rp2Q=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_baremetalsolution",
        importpath = "cloud.google.com/go/baremetalsolution",
        sum = "h1:g9KO6SkakcYPcc/XjAzeuUrEOXlYPnMpuiaywYaGrmQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_batch",
        importpath = "cloud.google.com/go/batch",
        sum = "h1:1jvEBY55OH4Sd2FxEXQfxGExFWov1A/IaRe+Z5Z71Fw=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_beyondcorp",
        importpath = "cloud.google.com/go/beyondcorp",
        sum = "h1:w+4kThysgl0JiKshi2MKDCg2NZgOyqOI0wq2eBZyrzA=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_bigquery",
        importpath = "cloud.google.com/go/bigquery",
        sum = "h1:Wi4dITi+cf9VYp4VH2T9O41w0kCW0uQTELq2Z6tukN0=",
        version = "v1.44.0",
    )
    go_repository(
        name = "com_google_cloud_go_billing",
        importpath = "cloud.google.com/go/billing",
        sum = "h1:Xkii76HWELHwBtkQVZvqmSo9GTr0O+tIbRNnMcGdlg4=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_binaryauthorization",
        importpath = "cloud.google.com/go/binaryauthorization",
        sum = "h1:pL70vXWn9TitQYXBWTK2abHl2JHLwkFRjYw6VflRqEA=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_certificatemanager",
        importpath = "cloud.google.com/go/certificatemanager",
        sum = "h1:tzbR4UHBbgsewMWUD93JHi8EBi/gHBoSAcY1/sThFGk=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_channel",
        importpath = "cloud.google.com/go/channel",
        sum = "h1:pNuUlZx0Jb0Ts9P312bmNMuH5IiFWIR4RUtLb70Ke5s=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_cloudbuild",
        importpath = "cloud.google.com/go/cloudbuild",
        sum = "h1:TAAmCmAlOJ4uNBu6zwAjwhyl/7fLHHxIEazVhr3QBbQ=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_clouddms",
        importpath = "cloud.google.com/go/clouddms",
        sum = "h1:UhzHIlgFfMr6luVYVNydw/pl9/U5kgtjCMJHnSvoVws=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_cloudtasks",
        importpath = "cloud.google.com/go/cloudtasks",
        sum = "h1:faUiUgXjW8yVZ7XMnKHKm1WE4OldPBUWWfIRN/3z1dc=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_compute",
        importpath = "cloud.google.com/go/compute",
        sum = "h1:7UGq3QknM33pw5xATlpzeoomNxsacIVvTqTTvbfajmE=",
        version = "v1.15.1",
    )
    go_repository(
        name = "com_google_cloud_go_compute_metadata",
        importpath = "cloud.google.com/go/compute/metadata",
        sum = "h1:mg4jlk7mCAj6xXp9UJ4fjI9VUI5rubuGBW5aJ7UnBMY=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_google_cloud_go_contactcenterinsights",
        importpath = "cloud.google.com/go/contactcenterinsights",
        sum = "h1:tTQLI/ZvguUf9Hv+36BkG2+/PeC8Ol1q4pBW+tgCx0A=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_container",
        importpath = "cloud.google.com/go/container",
        sum = "h1:nbEK/59GyDRKKlo1SqpohY1TK8LmJ2XNcvS9Gyom2A0=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_containeranalysis",
        importpath = "cloud.google.com/go/containeranalysis",
        sum = "h1:2824iym832ljKdVpCBnpqm5K94YT/uHTVhNF+dRTXPI=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_datacatalog",
        importpath = "cloud.google.com/go/datacatalog",
        sum = "h1:6kZ4RIOW/uT7QWC5SfPfq/G8sYzr/v+UOmOAxy4Z1TE=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataflow",
        importpath = "cloud.google.com/go/dataflow",
        sum = "h1:CW3541Fm7KPTyZjJdnX6NtaGXYFn5XbFC5UcjgALKvU=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataform",
        importpath = "cloud.google.com/go/dataform",
        sum = "h1:vLwowLF2ZB5J5gqiZCzv076lDI/Rd7zYQQFu5XO1PSg=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_datafusion",
        importpath = "cloud.google.com/go/datafusion",
        sum = "h1:j5m2hjWovTZDTQak4MJeXAR9yN7O+zMfULnjGw/OOLg=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_datalabeling",
        importpath = "cloud.google.com/go/datalabeling",
        sum = "h1:dp8jOF21n/7jwgo/uuA0RN8hvLcKO4q6s/yvwevs2ZM=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataplex",
        importpath = "cloud.google.com/go/dataplex",
        sum = "h1:cNxeA2DiWliQGi21kPRqnVeQ5xFhNoEjPRt1400Pm8Y=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataproc",
        importpath = "cloud.google.com/go/dataproc",
        sum = "h1:gVOqNmElfa6n/ccG/QDlfurMWwrK3ezvy2b2eDoCmS0=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataqna",
        importpath = "cloud.google.com/go/dataqna",
        sum = "h1:gx9jr41ytcA3dXkbbd409euEaWtofCVXYBvJz3iYm18=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_datastore",
        importpath = "cloud.google.com/go/datastore",
        sum = "h1:4siQRf4zTiAVt/oeH4GureGkApgb2vtPQAtOmhpqQwE=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_datastream",
        importpath = "cloud.google.com/go/datastream",
        sum = "h1:PgIgbhedBtYBU6POGXFMn2uSl9vpqubc3ewTNdcU8Mk=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_deploy",
        importpath = "cloud.google.com/go/deploy",
        sum = "h1:kI6dxt8Ml0is/x7YZjLveTvR7YPzXAUD/8wQZ2nH5zA=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_dialogflow",
        importpath = "cloud.google.com/go/dialogflow",
        sum = "h1:HYHVOkoxQ9bSfNIelSZYNAtUi4CeSrCnROyOsbOqPq8=",
        version = "v1.19.0",
    )
    go_repository(
        name = "com_google_cloud_go_dlp",
        importpath = "cloud.google.com/go/dlp",
        sum = "h1:9I4BYeJSVKoSKgjr70fLdRDumqcUeVmHV4fd5f9LR6Y=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_documentai",
        importpath = "cloud.google.com/go/documentai",
        sum = "h1:jfq09Fdjtnpnmt/MLyf6A3DM3ynb8B2na0K+vSXvpFM=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_domains",
        importpath = "cloud.google.com/go/domains",
        sum = "h1:pu3JIgC1rswIqi5romW0JgNO6CTUydLYX8zyjiAvO1c=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_edgecontainer",
        importpath = "cloud.google.com/go/edgecontainer",
        sum = "h1:hd6J2n5dBBRuAqnNUEsKWrp6XNPKsaxwwIyzOPZTokk=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_errorreporting",
        importpath = "cloud.google.com/go/errorreporting",
        sum = "h1:kj1XEWMu8P0qlLhm3FwcaFsUvXChV/OraZwA70trRR0=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_essentialcontacts",
        importpath = "cloud.google.com/go/essentialcontacts",
        sum = "h1:b6csrQXCHKQmfo9h3dG/pHyoEh+fQG1Yg78a53LAviY=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_eventarc",
        importpath = "cloud.google.com/go/eventarc",
        sum = "h1:AgCqrmMMIcel5WWKkzz5EkCUKC3Rl5LNMMYsS+LvsI0=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_filestore",
        importpath = "cloud.google.com/go/filestore",
        sum = "h1:yjKOpzvqtDmL5AXbKttLc8j0hL20kuC1qPdy5HPcxp0=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_firestore",
        importpath = "cloud.google.com/go/firestore",
        sum = "h1:IBlRyxgGySXu5VuW0RgGFlTtLukSnNkpDiEOMkQkmpA=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_functions",
        importpath = "cloud.google.com/go/functions",
        sum = "h1:35tgv1fQOtvKqH/uxJMzX3w6usneJ0zXpsFr9KAVhNE=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_gaming",
        importpath = "cloud.google.com/go/gaming",
        sum = "h1:97OAEQtDazAJD7yh/kvQdSCQuTKdR0O+qWAJBZJ4xiA=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkebackup",
        importpath = "cloud.google.com/go/gkebackup",
        sum = "h1:4K+jiv4ocqt1niN8q5Imd8imRoXBHTrdnJVt/uFFxF4=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkeconnect",
        importpath = "cloud.google.com/go/gkeconnect",
        sum = "h1:zAcvDa04tTnGdu6TEZewaLN2tdMtUOJJ7fEceULjguA=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkehub",
        importpath = "cloud.google.com/go/gkehub",
        sum = "h1:JTcTaYQRGsVm+qkah7WzHb6e9sf1C0laYdRPn9aN+vg=",
        version = "v0.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkemulticloud",
        importpath = "cloud.google.com/go/gkemulticloud",
        sum = "h1:8F1NhJj8ucNj7lK51UZMtAjSWTgP1zO18XF6vkfiPPU=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_grafeas",
        importpath = "cloud.google.com/go/grafeas",
        sum = "h1:CYjC+xzdPvbV65gi6Dr4YowKcmLo045pm18L0DhdELM=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_gsuiteaddons",
        importpath = "cloud.google.com/go/gsuiteaddons",
        sum = "h1:TGT2oGmO5q3VH6SjcrlgPUWI0njhYv4kywLm6jag0to=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_iam",
        importpath = "cloud.google.com/go/iam",
        sum = "h1:E2osAkZzxI/+8pZcxVLcDtAQx/u+hZXVryUaYQ5O0Kk=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_iap",
        importpath = "cloud.google.com/go/iap",
        sum = "h1:BGEXovwejOCt1zDk8hXq0bOhhRu9haXKWXXXp2B4wBM=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_ids",
        importpath = "cloud.google.com/go/ids",
        sum = "h1:LncHK4HHucb5Du310X8XH9/ICtMwZ2PCfK0ScjWiJoY=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_iot",
        importpath = "cloud.google.com/go/iot",
        sum = "h1:Y9+oZT9jD4GUZzORXTU45XsnQrhxmDT+TFbPil6pRVQ=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_kms",
        importpath = "cloud.google.com/go/kms",
        sum = "h1:VrJLOsMRzW7IqTTYn+OYupqF3iKSE060Nrn+PECrYjg=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_language",
        importpath = "cloud.google.com/go/language",
        sum = "h1:3Wa+IUMamL4JH3Zd3cDZUHpwyqplTACt6UZKRD2eCL4=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_lifesciences",
        importpath = "cloud.google.com/go/lifesciences",
        sum = "h1:tIqhivE2LMVYkX0BLgG7xL64oNpDaFFI7teunglt1tI=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_logging",
        importpath = "cloud.google.com/go/logging",
        sum = "h1:ZBsZK+JG+oCDT+vaxwqF2egKNRjz8soXiS6Xv79benI=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_longrunning",
        importpath = "cloud.google.com/go/longrunning",
        sum = "h1:NjljC+FYPV3uh5/OwWT6pVU+doBqMg2x/rZlE+CamDs=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_managedidentities",
        importpath = "cloud.google.com/go/managedidentities",
        sum = "h1:3Kdajn6X25yWQFhFCErmKSYTSvkEd3chJROny//F1A0=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_maps",
        importpath = "cloud.google.com/go/maps",
        sum = "h1:kLReRbclTgJefw2fcCbdLPLhPj0U6UUWN10ldG8sdOU=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_google_cloud_go_mediatranslation",
        importpath = "cloud.google.com/go/mediatranslation",
        sum = "h1:qAJzpxmEX+SeND10Y/4868L5wfZpo4Y3BIEnIieP4dk=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_memcache",
        importpath = "cloud.google.com/go/memcache",
        sum = "h1:yLxUzJkZVSH2kPaHut7k+7sbIBFpvSh1LW9qjM2JDjA=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_metastore",
        importpath = "cloud.google.com/go/metastore",
        sum = "h1:3KcShzqWdqxrDEXIBWpYJpOOrgpDj+HlBi07Grot49Y=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_monitoring",
        importpath = "cloud.google.com/go/monitoring",
        sum = "h1:c9riaGSPQ4dUKWB+M1Fl0N+iLxstMbCktdEwYSPGDvA=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkconnectivity",
        importpath = "cloud.google.com/go/networkconnectivity",
        sum = "h1:BVdIKaI68bihnXGdCVL89Jsg9kq2kg+II30fjVqo62E=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkmanagement",
        importpath = "cloud.google.com/go/networkmanagement",
        sum = "h1:mDHA3CDW00imTvC5RW6aMGsD1bH+FtKwZm/52BxaiMg=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_networksecurity",
        importpath = "cloud.google.com/go/networksecurity",
        sum = "h1:qDEX/3sipg9dS5JYsAY+YvgTjPR63cozzAWop8oZS94=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_notebooks",
        importpath = "cloud.google.com/go/notebooks",
        sum = "h1:AC8RPjNvel3ExgXjO1YOAz+teg9+j+89TNxa7pIZfww=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_optimization",
        importpath = "cloud.google.com/go/optimization",
        sum = "h1:7PxOq9VTT7TMib/6dMoWpMvWS2E4dJEvtYzjvBreaec=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_orchestration",
        importpath = "cloud.google.com/go/orchestration",
        sum = "h1:39d6tqvNjd/wsSub1Bn4cEmrYcet5Ur6xpaN+SxOxtY=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_orgpolicy",
        importpath = "cloud.google.com/go/orgpolicy",
        sum = "h1:erF5PHqDZb6FeFrUHiYj2JK2BMhsk8CyAg4V4amJ3rE=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_osconfig",
        importpath = "cloud.google.com/go/osconfig",
        sum = "h1:NO0RouqCOM7M2S85Eal6urMSSipWwHU8evzwS+siqUI=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_oslogin",
        importpath = "cloud.google.com/go/oslogin",
        sum = "h1:pKGDPfeZHDybtw48WsnVLjoIPMi9Kw62kUE5TXCLCN4=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_phishingprotection",
        importpath = "cloud.google.com/go/phishingprotection",
        sum = "h1:OrwHLSRSZyaiOt3tnY33dsKSedxbMzsXvqB21okItNQ=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_policytroubleshooter",
        importpath = "cloud.google.com/go/policytroubleshooter",
        sum = "h1:NQklJuOUoz1BPP+Epjw81COx7IISWslkZubz/1i0UN8=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_privatecatalog",
        importpath = "cloud.google.com/go/privatecatalog",
        sum = "h1:Vz86uiHCtNGm1DeC32HeG2VXmOq5JRYA3VRPf8ZEcSg=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_pubsub",
        importpath = "cloud.google.com/go/pubsub",
        sum = "h1:q+J/Nfr6Qx4RQeu3rJcnN48SNC0qzlYzSeqkPq93VHs=",
        version = "v1.27.1",
    )
    go_repository(
        name = "com_google_cloud_go_pubsublite",
        importpath = "cloud.google.com/go/pubsublite",
        sum = "h1:iqrD8vp3giTb7hI1q4TQQGj77cj8zzgmMPsTZtLnprM=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_recaptchaenterprise",
        importpath = "cloud.google.com/go/recaptchaenterprise",
        sum = "h1:u6EznTGzIdsyOsvm+Xkw0aSuKFXQlyjGE9a4exk6iNQ=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_google_cloud_go_recaptchaenterprise_v2",
        importpath = "cloud.google.com/go/recaptchaenterprise/v2",
        sum = "h1:UqzFfb/WvhwXGDF1eQtdHLrmni+iByZXY4h3w9Kdyv8=",
        version = "v2.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_recommendationengine",
        importpath = "cloud.google.com/go/recommendationengine",
        sum = "h1:6w+WxPf2LmUEqX0YyvfCoYb8aBYOcbIV25Vg6R0FLGw=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_recommender",
        importpath = "cloud.google.com/go/recommender",
        sum = "h1:9kMZQGeYfcOD/RtZfcNKGKtoex3DdoB4zRgYU/WaIwE=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_redis",
        importpath = "cloud.google.com/go/redis",
        sum = "h1:/zTwwBKIAD2DEWTrXZp8WD9yD/gntReF/HkPssVYd0U=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_resourcemanager",
        importpath = "cloud.google.com/go/resourcemanager",
        sum = "h1:NDao6CHMwEZIaNsdWy+tuvHaavNeGP06o1tgrR0kLvU=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_resourcesettings",
        importpath = "cloud.google.com/go/resourcesettings",
        sum = "h1:eTzOwB13WrfF0kuzG2ZXCfB3TLunSHBur4s+HFU6uSM=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_retail",
        importpath = "cloud.google.com/go/retail",
        sum = "h1:N9fa//ecFUOEPsW/6mJHfcapPV0wBSwIUwpVZB7MQ3o=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_run",
        importpath = "cloud.google.com/go/run",
        sum = "h1:AWPuzU7Xtaj3Jf+QarDWIs6AJ5hM1VFQ+F6Q+VZ6OT4=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_scheduler",
        importpath = "cloud.google.com/go/scheduler",
        sum = "h1:K/mxOewgHGeKuATUJNGylT75Mhtjmx1TOkKukATqMT8=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_secretmanager",
        importpath = "cloud.google.com/go/secretmanager",
        sum = "h1:pu03bha7ukxF8otyPKTFdDz+rr9sE3YauS5PliDXK60=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_security",
        importpath = "cloud.google.com/go/security",
        sum = "h1:KSKzzJMyUoMRQzcz7azIgqAUqxo7rmQ5rYvimMhikqg=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_securitycenter",
        importpath = "cloud.google.com/go/securitycenter",
        sum = "h1:QTVtk/Reqnx2bVIZtJKm1+mpfmwRwymmNvlaFez7fQY=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_google_cloud_go_servicecontrol",
        importpath = "cloud.google.com/go/servicecontrol",
        sum = "h1:ImIzbOu6y4jL6ob65I++QzvqgFaoAKgHOG+RU9/c4y8=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_servicedirectory",
        importpath = "cloud.google.com/go/servicedirectory",
        sum = "h1:f7M8IMcVzO3T425AqlZbP3yLzeipsBHtRza8vVFYMhQ=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_servicemanagement",
        importpath = "cloud.google.com/go/servicemanagement",
        sum = "h1:TpkCO5M7dhKSy1bKUD9o/sSEW/U1Gtx7opA1fsiMx0c=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_serviceusage",
        importpath = "cloud.google.com/go/serviceusage",
        sum = "h1:b0EwJxPJLpavSljMQh0RcdHsUrr5DQ+Nelt/3BAs5ro=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_shell",
        importpath = "cloud.google.com/go/shell",
        sum = "h1:b1LFhFBgKsG252inyhtmsUUZwchqSz3WTvAIf3JFo4g=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_spanner",
        importpath = "cloud.google.com/go/spanner",
        sum = "h1:NvdTpRwf7DTegbfFdPjAWyD7bOVu0VeMqcvR9aCQCAc=",
        version = "v1.41.0",
    )
    go_repository(
        name = "com_google_cloud_go_speech",
        importpath = "cloud.google.com/go/speech",
        sum = "h1:yK0ocnFH4Wsf0cMdUyndJQ/hPv02oTJOxzi6AgpBy4s=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_storage",
        importpath = "cloud.google.com/go/storage",
        sum = "h1:F5QDG5ChchaAVQhINh24U99OWHURqrW8OmQcGKXcbgI=",
        version = "v1.28.1",
    )
    go_repository(
        name = "com_google_cloud_go_storagetransfer",
        importpath = "cloud.google.com/go/storagetransfer",
        sum = "h1:fUe3OydbbvHcAYp07xY+2UpH4AermGbmnm7qdEj3tGE=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_talent",
        importpath = "cloud.google.com/go/talent",
        sum = "h1:MrekAGxLqAeAol4Sc0allOVqUGO8j+Iim8NMvpiD7tM=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_texttospeech",
        importpath = "cloud.google.com/go/texttospeech",
        sum = "h1:ccPiHgTewxgyAeCWgQWvZvrLmbfQSFABTMAfrSPLPyY=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_tpu",
        importpath = "cloud.google.com/go/tpu",
        sum = "h1:ztIdKoma1Xob2qm6QwNh4Xi9/e7N3IfvtwG5AcNsj1g=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_trace",
        importpath = "cloud.google.com/go/trace",
        sum = "h1:qO9eLn2esajC9sxpqp1YKX37nXC3L4BfGnPS0Cx9dYo=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_translate",
        importpath = "cloud.google.com/go/translate",
        sum = "h1:AOYOH3MspzJ/bH1YXzB+xTE8fMpn3mwhLjugwGXvMPI=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_video",
        importpath = "cloud.google.com/go/video",
        sum = "h1:ttlvO4J5c1VGq6FkHqWPD/aH6PfdxujHt+muTJlW1Zk=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_videointelligence",
        importpath = "cloud.google.com/go/videointelligence",
        sum = "h1:RPFgVVXbI2b5vnrciZjtsUgpNKVtHO/WIyXUhEfuMhA=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_vision",
        importpath = "cloud.google.com/go/vision",
        sum = "h1:/CsSTkbmO9HC8iQpxbK8ATms3OQaX3YQUeTMGCxlaK4=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_vision_v2",
        importpath = "cloud.google.com/go/vision/v2",
        sum = "h1:TQHxRqvLMi19azwm3qYuDbEzZWmiKJNTpGbkNsfRCik=",
        version = "v2.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_vmmigration",
        importpath = "cloud.google.com/go/vmmigration",
        sum = "h1:A2Tl2ZmwMRpvEmhV2ibISY85fmQR+Y5w9a0PlRz5P3s=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_vmwareengine",
        importpath = "cloud.google.com/go/vmwareengine",
        sum = "h1:JMPZaOT/gIUxVlTqSl/QQ32Y2k+r0stNeM1NSqhVP9o=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_google_cloud_go_vpcaccess",
        importpath = "cloud.google.com/go/vpcaccess",
        sum = "h1:woHXXtnW8b9gLFdWO9HLPalAddBQ9V4LT+1vjKwR3W8=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_webrisk",
        importpath = "cloud.google.com/go/webrisk",
        sum = "h1:ypSnpGlJnZSXbN9a13PDmAYvVekBLnGKxQ3Q9SMwnYY=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_websecurityscanner",
        importpath = "cloud.google.com/go/websecurityscanner",
        sum = "h1:y7yIFg/h/mO+5Y5aCOtVAnpGUOgqCH5rXQ2Oc8Oq2+g=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_workflows",
        importpath = "cloud.google.com/go/workflows",
        sum = "h1:7Chpin9p50NTU8Tb7qk+I11U/IwVXmDhEoSsdccvInE=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_shuralyov_dmitri_gpu_mtl",
        importpath = "dmitri.shuralyov.com/gpu/mtl",
        sum = "h1:VpgP7xuJadIUuKccphEpTJnWhS2jkQyMt6Y7pJCD7fY=",
        version = "v0.0.0-20190408044501-666a987793e9",
    )
    go_repository(
        name = "com_sourcegraph_sourcegraph_appdash",
        importpath = "sourcegraph.com/sourcegraph/appdash",
        sum = "h1:ucqkfpjg9WzSUubAO62csmucvxl4/JeW3F4I4909XkM=",
        version = "v0.0.0-20190731080439-ebfcffb1b5c0",
    )
    go_repository(
        name = "dev_gocloud",
        importpath = "gocloud.dev",
        sum = "h1:TAqxeJMp07tXRPbBRBhwBtAs2Z8dtHq7xIuMxrnrvJ8=",
        version = "v0.24.1-0.20211119014450-028788aaaa4c",
    )
    go_repository(
        name = "in_gopkg_alecthomas_kingpin_v2",
        importpath = "gopkg.in/alecthomas/kingpin.v2",
        sum = "h1:jMFz6MfLP0/4fUyZle81rXUoxOBFi19VUFKVDOQfozc=",
        version = "v2.2.6",
    )
    go_repository(
        name = "in_gopkg_alexcesaro_statsd_v2",
        importpath = "gopkg.in/alexcesaro/statsd.v2",
        sum = "h1:FXkZSCZIH17vLCO5sO2UucTHsH9pc+17F6pl3JVCwMc=",
        version = "v2.0.0",
    )
    go_repository(
        name = "in_gopkg_check_v1",
        importpath = "gopkg.in/check.v1",
        sum = "h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=",
        version = "v1.0.0-20201130134442-10cb98267c6c",
    )
    go_repository(
        name = "in_gopkg_cheggaaa_pb_v1",
        importpath = "gopkg.in/cheggaaa/pb.v1",
        sum = "h1:n1tBJnnK2r7g9OW2btFH91V92STTUevLXYFb8gy9EMk=",
        version = "v1.0.28",
    )
    go_repository(
        name = "in_gopkg_errgo_v2",
        importpath = "gopkg.in/errgo.v2",
        sum = "h1:0vLT13EuvQ0hNvakwLuFZ/jYrLp5F3kcWHXdRggjCE8=",
        version = "v2.1.0",
    )
    go_repository(
        name = "in_gopkg_fsnotify_v1",
        importpath = "gopkg.in/fsnotify.v1",
        sum = "h1:xOHLXZwVvI9hhs+cLKq5+I5onOuwQLhQwiu63xxlHs4=",
        version = "v1.4.7",
    )
    go_repository(
        name = "in_gopkg_gcfg_v1",
        importpath = "gopkg.in/gcfg.v1",
        sum = "h1:m8OOJ4ccYHnx2f4gQwpno8nAX5OGOh7RLaaz0pj3Ogs=",
        version = "v1.2.3",
    )
    go_repository(
        name = "in_gopkg_go_playground_assert_v1",
        importpath = "gopkg.in/go-playground/assert.v1",
        sum = "h1:xoYuJVE7KT85PYWrN730RguIQO0ePzVRfFMXadIrXTM=",
        version = "v1.2.1",
    )
    go_repository(
        name = "in_gopkg_go_playground_validator_v9",
        importpath = "gopkg.in/go-playground/validator.v9",
        sum = "h1:SvGtYmN60a5CVKTOzMSyfzWDeZRxRuGvRQyEAKbw1xc=",
        version = "v9.29.1",
    )
    go_repository(
        name = "in_gopkg_inf_v0",
        importpath = "gopkg.in/inf.v0",
        sum = "h1:73M5CoZyi3ZLMOyDlQh031Cx6N9NDJ2Vvfl76EDAgDc=",
        version = "v0.9.1",
    )
    go_repository(
        name = "in_gopkg_ini_v1",
        importpath = "gopkg.in/ini.v1",
        sum = "h1:Dgnx+6+nfE+IfzjUEISNeydPJh9AXNNsWbGP9KzCsOA=",
        version = "v1.67.0",
    )
    go_repository(
        name = "in_gopkg_natefinch_lumberjack_v2",
        importpath = "gopkg.in/natefinch/lumberjack.v2",
        sum = "h1:1Lc07Kr7qY4U2YPouBjpCLxpiyxIVoxqXgkXLknAOE8=",
        version = "v2.0.0",
    )
    go_repository(
        name = "in_gopkg_resty_v1",
        importpath = "gopkg.in/resty.v1",
        sum = "h1:CuXP0Pjfw9rOuY6EP+UvtNvt5DSqHpIxILZKT/quCZI=",
        version = "v1.12.0",
    )
    go_repository(
        name = "in_gopkg_square_go_jose_v2",
        importpath = "gopkg.in/square/go-jose.v2",
        sum = "h1:NGk74WTnPKBNUhNzQX7PYcTLUjoq7mzKk2OKbvwk2iI=",
        version = "v2.6.0",
    )
    go_repository(
        name = "in_gopkg_src_d_go_billy_v4",
        importpath = "gopkg.in/src-d/go-billy.v4",
        sum = "h1:0SQA1pRztfTFx2miS8sA97XvooFeNOmvUenF4o0EcVg=",
        version = "v4.3.2",
    )
    go_repository(
        name = "in_gopkg_src_d_go_git_fixtures_v3",
        importpath = "gopkg.in/src-d/go-git-fixtures.v3",
        sum = "h1:ivZFOIltbce2Mo8IjzUHAFoq/IylO9WHhNOAJK+LsJg=",
        version = "v3.5.0",
    )
    go_repository(
        name = "in_gopkg_src_d_go_git_v4",
        importpath = "gopkg.in/src-d/go-git.v4",
        sum = "h1:SRtFyV8Kxc0UP7aCHcijOMQGPxHSmMOPrzulQWolkYE=",
        version = "v4.13.1",
    )
    go_repository(
        name = "in_gopkg_tomb_v1",
        importpath = "gopkg.in/tomb.v1",
        sum = "h1:uRGJdciOHaEIrze2W8Q3AKkepLTh2hOroT7a+7czfdQ=",
        version = "v1.0.0-20141024135613-dd632973f1e7",
    )
    go_repository(
        name = "in_gopkg_warnings_v0",
        importpath = "gopkg.in/warnings.v0",
        sum = "h1:wFXVbFY8DY5/xOe1ECiWdKCzZlxgshcYVNkBHstARME=",
        version = "v0.1.2",
    )
    go_repository(
        name = "in_gopkg_yaml_v2",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=",
        version = "v2.4.0",
    )
    go_repository(
        name = "in_gopkg_yaml_v3",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=",
        version = "v3.0.1",
    )
    go_repository(
        name = "io_etcd_go_bbolt",
        importpath = "go.etcd.io/bbolt",
        sum = "h1:/ecaJf0sk1l4l6V4awd65v2C3ILy7MSj+s/x1ADCIMU=",
        version = "v1.3.6",
    )
    go_repository(
        name = "io_etcd_go_etcd",
        importpath = "go.etcd.io/etcd",
        sum = "h1:VcrIfasaLFkyjk6KNlXQSzO+B0fZcnECiDrKJsfxka0=",
        version = "v0.0.0-20191023171146-3cf2f69b5738",
    )
    go_repository(
        name = "io_etcd_go_etcd_api_v3",
        importpath = "go.etcd.io/etcd/api/v3",
        sum = "h1:sbcmosSVesNrWOJ58ZQFitHMdncusIifYcrBfwrlJSY=",
        version = "v3.5.7",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_pkg_v3",
        importpath = "go.etcd.io/etcd/client/pkg/v3",
        sum = "h1:y3kf5Gbp4e4q7egZdn5T7W9TSHUvkClN6u+Rq9mEOmg=",
        version = "v3.5.7",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v2",
        importpath = "go.etcd.io/etcd/client/v2",
        sum = "h1:Dcx3/MYyfKcPNLpR4VVQUP5KgYrBeJtktBwEKkw08Ao=",
        version = "v2.305.4",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v3",
        importpath = "go.etcd.io/etcd/client/v3",
        sum = "h1:u/OhpiuCgYY8awOHlhIhmGIGpxfBU/GZBUP3m/3/Iz4=",
        version = "v3.5.7",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdctl_v3",
        importpath = "go.etcd.io/etcd/etcdctl/v3",
        sum = "h1:LVFzhocId7Vb8SqK3YanpW0rKjlvtkN80ShJpcBDDZk=",
        version = "v3.5.4",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdutl_v3",
        importpath = "go.etcd.io/etcd/etcdutl/v3",
        sum = "h1:TeQGkpXMGnQ+Tgn/dB5yuADyeSZatehBBy6XXSxnO7U=",
        version = "v3.5.4",
    )
    go_repository(
        name = "io_etcd_go_etcd_pkg_v3",
        importpath = "go.etcd.io/etcd/pkg/v3",
        sum = "h1:V5Dvl7S39ZDwjkKqJG2BfXgxZ3QREqqKifWQgIw5IM0=",
        version = "v3.5.4",
    )
    go_repository(
        name = "io_etcd_go_etcd_raft_v3",
        importpath = "go.etcd.io/etcd/raft/v3",
        sum = "h1:YGrnAgRfgXloBNuqa+oBI/aRZMcK/1GS6trJePJ/Gqc=",
        version = "v3.5.4",
    )
    go_repository(
        name = "io_etcd_go_etcd_server_v3",
        importpath = "go.etcd.io/etcd/server/v3",
        sum = "h1:CMAZd0g8Bn5NRhynW6pKhc4FRg41/0QYy3d7aNm9874=",
        version = "v3.5.4",
    )
    go_repository(
        name = "io_etcd_go_etcd_tests_v3",
        importpath = "go.etcd.io/etcd/tests/v3",
        sum = "h1:wiYG8vbDwZO2UatQE9Z3GIv2z52jGg5DvEkTDXm090c=",
        version = "v3.5.4",
    )
    go_repository(
        name = "io_etcd_go_etcd_v3",
        importpath = "go.etcd.io/etcd/v3",
        sum = "h1:IWyDYI27KTWKGv1OS0Hzysr6514E6e7qfRUVpzr4YFQ=",
        version = "v3.5.4",
    )
    go_repository(
        name = "io_filippo_edwards25519",
        importpath = "filippo.io/edwards25519",
        sum = "h1:0wAIcmJUqRdI8IJ/3eGi5/HwXZWPujYXXlkrQogz0Ek=",
        version = "v1.0.0",
    )
    go_repository(
        name = "io_gitea_code_sdk_gitea",
        importpath = "code.gitea.io/sdk/gitea",
        sum = "h1:CdI3J82Mqn0mElyEKa5DUSr3Wi2R+qm/6uVtCkSSqSM=",
        version = "v0.11.3",
    )
    go_repository(
        name = "io_k8s_api",
        importpath = "k8s.io/api",
        sum = "h1:3YO8J4RtmG7elEgaWMb4HgmpS2CfY1QlaOz9nwB+ZSs=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_apiextensions_apiserver",
        importpath = "k8s.io/apiextensions-apiserver",
        sum = "h1:7hu9pF+xikxQuQZ7/30z/qxIPZc2J1lFElPtr7f+B6U=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_apimachinery",
        importpath = "k8s.io/apimachinery",
        sum = "h1:CtXsuaitMESSu339tfhVXhQrPET+EiWnIY1rcurKnAc=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_apiserver",
        importpath = "k8s.io/apiserver",
        sum = "h1:/3TwZcgLqX7wUxq7TtXOUqXeBTwXIblVMQdhR5XZ7yo=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_cli_runtime",
        importpath = "k8s.io/cli-runtime",
        sum = "h1:GTSBN7aKBrc2LqpdO30CmHQqJtRmotxV7XsMSP+QZIk=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_client_go",
        importpath = "k8s.io/client-go",
        sum = "h1:3RNRDffAkNU56M/a7gUfXaEzdhZlYhoW8dgViGy5fn8=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_cloud_provider",
        importpath = "k8s.io/cloud-provider",
        replace = "k8s.io/cloud-provider",
        sum = "h1:1X1BKXm0fp8/ZkaQKNDyWqgh6t7m9O5MDSbO9OA4muk=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_cluster_bootstrap",
        importpath = "k8s.io/cluster-bootstrap",
        sum = "h1:m50ICwMsEW13N7Z/tdTmLwELGHt4SJEJaeriPdQRxs0=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_code_generator",
        importpath = "k8s.io/code-generator",
        sum = "h1:tjQ7/+9eN7UOiU2DP+0v4ntTI4JZLi2c1N0WllpFhTc=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_component_base",
        importpath = "k8s.io/component-base",
        sum = "h1:n1bjg9Yt+G1C0WnIDJmg2fo6wbEU1UGMRiQSjmj7hNQ=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_component_helpers",
        importpath = "k8s.io/component-helpers",
        sum = "h1:A4xQEFq7tbnhB3CTwZTLcQtyEhFFZN2TyQjNgziuSEI=",
        version = "v0.25.2",
    )
    go_repository(
        name = "io_k8s_controller_manager",
        importpath = "k8s.io/controller-manager",
        replace = "k8s.io/controller-manager",
        sum = "h1:MaPr9kgLB8DNNJ5fPGGpecGQ1gqQSLX1DsVbiZHfftM=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_cri_api",
        importpath = "k8s.io/cri-api",
        sum = "h1:INwdXsCDSA/0hGNdPxdE2dQD6ft/5K1EaKXZixvSQxg=",
        version = "v0.25.0",
    )
    go_repository(
        name = "io_k8s_csi_translation_lib",
        importpath = "k8s.io/csi-translation-lib",
        replace = "k8s.io/csi-translation-lib",
        sum = "h1:gNHocpHaD4O+QZT31mHFvwC8ZFl5q0Tp7BvhV900HPo=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_gengo",
        importpath = "k8s.io/gengo",
        sum = "h1:TT1WdmqqXareKxZ/oNXEUSwKlLiHzPMyB0t8BaFeBYI=",
        version = "v0.0.0-20211129171323-c02415ce4185",
    )
    go_repository(
        name = "io_k8s_klog",
        importpath = "k8s.io/klog",
        sum = "h1:Pt+yjF5aB1xDSVbau4VsWe+dQNzA0qv1LlXdC2dF6Q8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "io_k8s_klog_v2",
        importpath = "k8s.io/klog/v2",
        sum = "h1:atnLQ121W371wYYFawwYx1aEY2eUfs4l3J72wtgAwV4=",
        version = "v2.80.1",
    )
    go_repository(
        name = "io_k8s_kube_aggregator",
        importpath = "k8s.io/kube-aggregator",
        replace = "k8s.io/kube-aggregator",
        sum = "h1:eOG9S4GPICAXWIFeQDHjnhqYaYPpgLIC1NunJu9pZCs=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_kube_controller_manager",
        importpath = "k8s.io/kube-controller-manager",
        replace = "k8s.io/kube-controller-manager",
        sum = "h1:Ouh627GVtSq2dM8+ATrM/W6EJNVXv8sZpOE6ANw10lM=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_kube_openapi",
        importpath = "k8s.io/kube-openapi",
        sum = "h1:MQ8BAZPZlWk3S9K4a9NCkIFQtZShWqoha7snGixVgEA=",
        version = "v0.0.0-20220803162953-67bda5d908f1",
    )
    go_repository(
        name = "io_k8s_kube_proxy",
        importpath = "k8s.io/kube-proxy",
        replace = "k8s.io/kube-proxy",
        sum = "h1:3iaWXdYyHhuPe+BvEUeO6LEwteiYMUm4lk96O7VTg04=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_kube_scheduler",
        importpath = "k8s.io/kube-scheduler",
        replace = "k8s.io/kube-scheduler",
        sum = "h1:ezMPpdHc2u8kB6CyU/N3/AjGNZ2unXeITGp3FgoWbgA=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_kubectl",
        importpath = "k8s.io/kubectl",
        sum = "h1:2993lTeVimxKSWx/7z2PiJxUILygRa3tmC4QhFaeioA=",
        version = "v0.25.2",
    )
    go_repository(
        name = "io_k8s_kubelet",
        importpath = "k8s.io/kubelet",
        sum = "h1:24MmTTQGBHr08UkMYFC/RaLjuiMREM53HfRgJKWRquI=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_kubernetes",
        importpath = "k8s.io/kubernetes",
        sum = "h1:M1+MR8IxE64zHhSSDn30twChLaOI+p0Kt77pvyQMKwU=",
        version = "v1.25.4",
    )
    go_repository(
        name = "io_k8s_legacy_cloud_providers",
        importpath = "k8s.io/legacy-cloud-providers",
        replace = "k8s.io/legacy-cloud-providers",
        sum = "h1:v7M/U8jE6kJn2gEBVofDApCOkgUlsFGBTXosTKVTtAo=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_metrics",
        importpath = "k8s.io/metrics",
        sum = "h1:105TuPaIFfr4EHzN56WwZJO7r1UesuDytNTzeMqGySo=",
        version = "v0.25.2",
    )
    go_repository(
        name = "io_k8s_mount_utils",
        importpath = "k8s.io/mount-utils",
        sum = "h1:+j1GBo6rH6sM1GvOI4jcu1IpjA5WssuwW7UEkQevaTU=",
        version = "v0.25.4",
    )
    go_repository(
        name = "io_k8s_pod_security_admission",
        importpath = "k8s.io/pod-security-admission",
        replace = "k8s.io/pod-security-admission",
        sum = "h1:2HnXWKUIDSez2sWtvxeGgGVUFvYnJJHutL4AI1MIuwk=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_sample_apiserver",
        importpath = "k8s.io/sample-apiserver",
        replace = "k8s.io/sample-apiserver",
        sum = "h1:LyeMlp4dvLMCMldhzEcPXh73osVY+uVGWrlyOrT444s=",
        version = "v0.25.3",
    )
    go_repository(
        name = "io_k8s_sigs_apiserver_network_proxy_konnectivity_client",
        importpath = "sigs.k8s.io/apiserver-network-proxy/konnectivity-client",
        sum = "h1:LYqFq+6Cj2D0gFfrJvL7iElD4ET6ir3VDdhDdTK7rgc=",
        version = "v0.0.33",
    )
    go_repository(
        name = "io_k8s_sigs_controller_runtime",
        importpath = "sigs.k8s.io/controller-runtime",
        sum = "h1:tUsRCSJVM1QQOOeViGeX3GMT3dQF1eePPw6sEE3xSlg=",
        version = "v0.13.1",
    )
    go_repository(
        name = "io_k8s_sigs_json",
        importpath = "sigs.k8s.io/json",
        sum = "h1:iXTIw73aPyC+oRdyqqvVJuloN1p0AC/kzH07hu3NE+k=",
        version = "v0.0.0-20220713155537-f223a00ba0e2",
    )
    go_repository(
        name = "io_k8s_sigs_kustomize_api",
        importpath = "sigs.k8s.io/kustomize/api",
        sum = "h1:7YM7gW3kYBwtKvoY216ZzY+8hM+lV53LUayghNRJ0vM=",
        version = "v0.12.1",
    )
    go_repository(
        name = "io_k8s_sigs_kustomize_kustomize_v4",
        importpath = "sigs.k8s.io/kustomize/kustomize/v4",
        sum = "h1:cDW6AVMl6t/SLuQaezMET8hgnadZGIAr8tUrxFVOrpg=",
        version = "v4.5.7",
    )
    go_repository(
        name = "io_k8s_sigs_kustomize_kyaml",
        importpath = "sigs.k8s.io/kustomize/kyaml",
        sum = "h1:Qz53EAaFFANyNgyOEJbT/yoIHygK40/ZcvU3rgry2Tk=",
        version = "v0.13.9",
    )
    go_repository(
        name = "io_k8s_sigs_release_utils",
        importpath = "sigs.k8s.io/release-utils",
        sum = "h1:6pS8x6c5RmdUgR9qcg1LO6hjUzuE4Yo9TGZ3DemrZdM=",
        version = "v0.7.3",
    )
    go_repository(
        name = "io_k8s_sigs_structured_merge_diff_v4",
        importpath = "sigs.k8s.io/structured-merge-diff/v4",
        sum = "h1:PRbqxJClWWYMNV1dhaG4NsibJbArud9kFxnAMREiWFE=",
        version = "v4.2.3",
    )
    go_repository(
        name = "io_k8s_sigs_yaml",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:a2VclLzOGrwOHDiV8EfBGhvjHvP46CtW5j6POvhYGGo=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_k8s_system_validators",
        importpath = "k8s.io/system-validators",
        sum = "h1:tYD3hojdYc58CKtiEsh7BWW6Pg0x6xbclUiwjaIiiYo=",
        version = "v1.7.0",
    )
    go_repository(
        name = "io_k8s_utils",
        importpath = "k8s.io/utils",
        sum = "h1:KTgPnR10d5zhztWptI952TNtt/4u5h3IzDXkdIMuo2Y=",
        version = "v0.0.0-20221128185143-99ec85e7a448",
    )
    go_repository(
        name = "io_opencensus_go",
        importpath = "go.opencensus.io",
        sum = "h1:y73uSU6J157QMP2kn2r30vwW1A2W2WFwSCGnAVxeaD0=",
        version = "v0.24.0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_aws",
        importpath = "contrib.go.opencensus.io/exporter/aws",
        sum = "h1:YsbWYxDZkC7x2OxlsDEYvvEXZ3cBI3qBgUK5BqkZvRw=",
        version = "v0.0.0-20181029163544-2befc13012d0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_ocagent",
        importpath = "contrib.go.opencensus.io/exporter/ocagent",
        sum = "h1:TKXjQSRS0/cCDrP7KvkgU6SmILtF/yV2TOs/02K/WZQ=",
        version = "v0.5.0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_stackdriver",
        importpath = "contrib.go.opencensus.io/exporter/stackdriver",
        sum = "h1:bjBKzIf7/TAkxd7L2utGaLM78bmUWlCval5K9UeElbY=",
        version = "v0.13.12",
    )
    go_repository(
        name = "io_opencensus_go_contrib_integrations_ocsql",
        importpath = "contrib.go.opencensus.io/integrations/ocsql",
        sum = "h1:kfg5Yyy1nYUrqzyfW5XX+dzMASky8IJXhtHe0KTYNS4=",
        version = "v0.1.4",
    )
    go_repository(
        name = "io_opencensus_go_contrib_resource",
        importpath = "contrib.go.opencensus.io/resource",
        sum = "h1:4r2CANuYhKGmYWP02+5E94rLRcS/YeD+KlxSrOsMxk0=",
        version = "v0.1.1",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib",
        importpath = "go.opentelemetry.io/contrib",
        sum = "h1:xJawAzMuR3s4Au5p/ABHqYFychHjK2AHB9JvkBuBbTA=",
        version = "v1.6.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_github_com_emicklei_go_restful_otelrestful",
        importpath = "go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful",
        sum = "h1:8YW+SL62UmcwRQJFZVfnyOlIUUtmlR13NaMKi+Fa6Fo=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_google_golang_org_grpc_otelgrpc",
        importpath = "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
        sum = "h1:Ky1MObd188aGbgb5OgNnwGuEEwI9MVIcc7rBW6zk5Ak=",
        version = "v0.28.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_net_http_otelhttp",
        importpath = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
        sum = "h1:Q3C9yzW6I9jqEc8sawxzxZmY48fs9u220KXq6d5s3XU=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel",
        importpath = "go.opentelemetry.io/otel",
        sum = "h1:APxLf0eiBwLl+SOXiJJCVYzA1OOJNyAoV8C5RNRyy7Y=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp",
        importpath = "go.opentelemetry.io/otel/exporters/otlp",
        sum = "h1:PTNgq9MRmQqqJY0REVbZFvwkYOA85vbdQU/nVfxDyqg=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_internal_retry",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/internal/retry",
        sum = "h1:R/OBkMoGgfy2fLhs2QhkCI1w4HLEQX92GCcJB6SSdNk=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace",
        sum = "h1:giGm8w67Ja7amYNfYMdme7xSp2pIxThWopw8+QP51Yk=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace_otlptracegrpc",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc",
        sum = "h1:VQbUHoJqytHHSJ1OZodPH9tvZZSVzUHjPHpkO85sT6k=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace_otlptracehttp",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp",
        sum = "h1:Ydage/P0fRrSPpZeCVxzjqGcI6iVmG2xb43+IR8cjqM=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_metric",
        importpath = "go.opentelemetry.io/otel/metric",
        sum = "h1:4kzhXFP+btKm4jwxpjIqjs41A7MakRFUS86bqLHTIw8=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_oteltest",
        importpath = "go.opentelemetry.io/otel/oteltest",
        sum = "h1:HiITxCawalo5vQzdHfKeZurV8x7ljcqAgiWzF6Vaeaw=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk",
        importpath = "go.opentelemetry.io/otel/sdk",
        sum = "h1:3278edCoH89MEJ0Ky8WQXVmDQv3FX4ZJ3Pp+9fJreAI=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk_export_metric",
        importpath = "go.opentelemetry.io/otel/sdk/export/metric",
        sum = "h1:c5VRjxCXdQlx1HjzwGdQHzZaVI82b5EbBgOu2ljD92g=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk_metric",
        importpath = "go.opentelemetry.io/otel/sdk/metric",
        sum = "h1:7ao1wpzHRVKf0OQ7GIxiQJA6X7DLX9o14gmVon7mMK8=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_trace",
        importpath = "go.opentelemetry.io/otel/trace",
        sum = "h1:doy8Hzb1RJ+I3yFhtDmwNc7tIyw1tNMOIsyPzp1NOGY=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_opentelemetry_go_proto_otlp",
        importpath = "go.opentelemetry.io/proto/otlp",
        sum = "h1:cLDgIBTf4lLOlztkhzAEdQsJ4Lj+i5Wc9k6Nn0K1VyU=",
        version = "v0.11.0",
    )
    go_repository(
        name = "io_rsc_binaryregexp",
        importpath = "rsc.io/binaryregexp",
        sum = "h1:HfqmD5MEmC0zvwBuF187nq9mdnXjXsSivRiXN7SmRkE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "io_rsc_quote_v3",
        importpath = "rsc.io/quote/v3",
        sum = "h1:9JKUTTIUgS6kzR9mK1YuGKv6Nl+DijDNIc0ghT58FaY=",
        version = "v3.1.0",
    )
    go_repository(
        name = "io_rsc_sampler",
        importpath = "rsc.io/sampler",
        sum = "h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/QiW4=",
        version = "v1.3.0",
    )
    go_repository(
        name = "land_oras_oras_go",
        importpath = "oras.land/oras-go",
        sum = "h1:yoKosVIbsPoFMqAIFHTnrmOuafHal+J/r+I5bdbVWu4=",
        version = "v1.2.0",
    )
    go_repository(
        name = "net_starlark_go",
        importpath = "go.starlark.net",
        sum = "h1:ghIB+2LQvihWROIGpcAVPq/ce5O2uMQersgxXiOeTS4=",
        version = "v0.0.0-20220223235035-243c74974e97",
    )
    go_repository(
        name = "org_bazil_fuse",
        importpath = "bazil.org/fuse",
        sum = "h1:FNCRpXiquG1aoyqcIWVFmpTSKVcx2bQD38uZZeGtdlw=",
        version = "v0.0.0-20180421153158-65cc252bf669",
    )
    go_repository(
        name = "org_bitbucket_bertimus9_systemstat",
        importpath = "bitbucket.org/bertimus9/systemstat",
        sum = "h1:n0aLnh2Jo4nBUBym9cE5PJDG8GT6g+4VuS2Ya2jYYpA=",
        version = "v0.5.0",
    )
    go_repository(
        name = "org_bitbucket_creachadair_shell",
        importpath = "bitbucket.org/creachadair/shell",
        sum = "h1:Z96pB6DkSb7F3Y3BBnJeOZH2gazyMTWlvecSD4vDqfk=",
        version = "v0.0.7",
    )
    go_repository(
        name = "org_cloudfoundry_code_clock",
        importpath = "code.cloudfoundry.org/clock",
        sum = "h1:5eeuG0BHx1+DHeT3AP+ISKZ2ht1UjGhm581ljqYpVeQ=",
        version = "v0.0.0-20180518195852-02e53af36e6c",
    )
    go_repository(
        name = "org_golang_google_api",
        importpath = "google.golang.org/api",
        sum = "h1:I2SlFjD8ZWabaIFOfeEDg3pf0BHJDh6iYQ1ic3Yu/UU=",
        version = "v0.107.0",
    )
    go_repository(
        name = "org_golang_google_appengine",
        importpath = "google.golang.org/appengine",
        sum = "h1:FZR1q0exgwxzPzp/aF+VccGrSfxfPpkBqjIIEq3ru6c=",
        version = "v1.6.7",
    )
    go_repository(
        name = "org_golang_google_genproto",
        importpath = "google.golang.org/genproto",
        sum = "h1:uQ2vjV/sHTsWSqdKeLqmwitzgvjMl7o4IdtHwUDXSJY=",
        version = "v0.0.0-20221227171554-f9683d7f8bef",
    )
    go_repository(
        name = "org_golang_google_grpc",
        importpath = "google.golang.org/grpc",
        sum = "h1:E1eGv1FTqoLIdnBCZufiSHgKjlqG6fKFf6pPWtMTh8U=",
        version = "v1.51.0",
    )
    go_repository(
        name = "org_golang_google_grpc_cmd_protoc_gen_go_grpc",
        importpath = "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
        sum = "h1:TLkBREm4nIsEcexnCjgQd5GQWaHcqMzwQV0TX9pq8S0=",
        version = "v1.2.0",
    )
    go_repository(
        name = "org_golang_google_protobuf",
        importpath = "google.golang.org/protobuf",
        sum = "h1:d0NfwRgPtno5B1Wa6L2DAG+KivqkdutMf1UhdNx175w=",
        version = "v1.28.1",
    )
    go_repository(
        name = "org_golang_x_crypto",
        importpath = "golang.org/x/crypto",
        sum = "h1:U/0M97KRkSFvyD/3FSmdP5W5swImpNgle/EHFhOsQPE=",
        version = "v0.5.0",
    )
    go_repository(
        name = "org_golang_x_exp",
        importpath = "golang.org/x/exp",
        sum = "h1:S4NrSKDfihhl3+4jSTgwoIevKxX9p7Iv9x++OEIptDo=",
        version = "v0.0.0-20220823124025-807a23277127",
    )
    go_repository(
        name = "org_golang_x_image",
        importpath = "golang.org/x/image",
        sum = "h1:+qEpEAPhDZ1o0x3tHzZTQDArnOixOzGD9HUJfcg0mb4=",
        version = "v0.0.0-20190802002840-cff245a6509b",
    )
    go_repository(
        name = "org_golang_x_lint",
        importpath = "golang.org/x/lint",
        sum = "h1:VLliZ0d+/avPrXXH+OakdXhpJuEoBZuwh1m2j7U6Iug=",
        version = "v0.0.0-20210508222113-6edffad5e616",
    )
    go_repository(
        name = "org_golang_x_mobile",
        importpath = "golang.org/x/mobile",
        sum = "h1:4+4C/Iv2U4fMZBiMCc98MG1In4gJY5YRhtpDNeDeHWs=",
        version = "v0.0.0-20190719004257-d2bd2a29d028",
    )
    go_repository(
        name = "org_golang_x_mod",
        importpath = "golang.org/x/mod",
        sum = "h1:LapD9S96VoQRhi/GrNTqeBJFrUjs5UHCAtTlgwA5oZA=",
        version = "v0.7.0",
    )
    go_repository(
        name = "org_golang_x_net",
        importpath = "golang.org/x/net",
        sum = "h1:GyT4nK/YDHSqa1c4753ouYCDajOYKTja9Xb/OHtgvSw=",
        version = "v0.5.0",
    )
    go_repository(
        name = "org_golang_x_oauth2",
        importpath = "golang.org/x/oauth2",
        sum = "h1:NF0gk8LVPg1Ml7SSbGyySuoxdsXitj7TvgvuRxIMc/M=",
        version = "v0.4.0",
    )
    go_repository(
        name = "org_golang_x_sync",
        importpath = "golang.org/x/sync",
        sum = "h1:wsuoTGHzEhffawBOhz5CYhcrV4IdKZbEyZjBMuTp12o=",
        version = "v0.1.0",
    )
    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:Zr2JFtRQNX3BCZ8YtxRE9hNJYC8J6I1MVbMg6owUp18=",
        version = "v0.4.0",
    )
    go_repository(
        name = "org_golang_x_term",
        importpath = "golang.org/x/term",
        sum = "h1:O7UWfv5+A2qiuulQk30kVinPoMtoIPeVaKLEgLpVkvg=",
        version = "v0.4.0",
    )
    go_repository(
        name = "org_golang_x_text",
        importpath = "golang.org/x/text",
        sum = "h1:3XmdazWV+ubf7QgHSTWeykHOci5oeekaGJBLkrkaw4k=",
        version = "v0.6.0",
    )
    go_repository(
        name = "org_golang_x_time",
        importpath = "golang.org/x/time",
        sum = "h1:52I/1L54xyEQAYdtcSuxtiT84KGYTBGXwayxmIpNJhE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "org_golang_x_tools",
        importpath = "golang.org/x/tools",
        sum = "h1:+bSpV5HIeWkuvgaMfI3UmKRThoTA5ODJTUd8T17NO+4=",
        version = "v0.5.0",
    )
    go_repository(
        name = "org_golang_x_xerrors",
        importpath = "golang.org/x/xerrors",
        sum = "h1:H2TDz8ibqkAF6YGhCdN3jS9O0/s90v0rJh3X/OLHEUk=",
        version = "v0.0.0-20220907171357-04be3eba64a2",
    )
    go_repository(
        name = "org_gonum_v1_gonum",
        importpath = "gonum.org/v1/gonum",
        sum = "h1:4r+yNT0+8SWcOkXP+63H2zQbN+USnC73cjGUxnDF94Q=",
        version = "v0.6.2",
    )
    go_repository(
        name = "org_gonum_v1_netlib",
        importpath = "gonum.org/v1/netlib",
        sum = "h1:jRyg0XfpwWlhEV8mDfdNGBeSJM2fuyh9Yjrnd8kF2Ts=",
        version = "v0.0.0-20190331212654-76723241ea4e",
    )
    go_repository(
        name = "org_libvirt_go_libvirt",
        importpath = "libvirt.org/go/libvirt",
        sum = "h1:q4UeuyzSp7GiYUB5T3ytzeXh41hp9JqNQu150NkO+7A=",
        version = "v1.8010.0",
    )
    go_repository(
        name = "org_libvirt_go_libvirtxml",
        importpath = "libvirt.org/go/libvirtxml",
        sum = "h1:29wyVe4m08S3JBSJnMJ/0WCWWcq8Xl5zF5NIa2RtXdI=",
        version = "v1.8009.0",
    )
    go_repository(
        name = "org_mongodb_go_mongo_driver",
        importpath = "go.mongodb.org/mongo-driver",
        sum = "h1:UtV6N5k14upNp4LTduX0QCufG124fSu25Wz9tu94GLg=",
        version = "v1.10.0",
    )
    go_repository(
        name = "org_mozilla_go_pkcs7",
        importpath = "go.mozilla.org/pkcs7",
        sum = "h1:A/5uWzF44DlIgdm/PQFwfMkW0JX+cIcQi/SwLAmZP5M=",
        version = "v0.0.0-20200128120323-432b2356ecb1",
    )
    go_repository(
        name = "org_uber_go_atomic",
        importpath = "go.uber.org/atomic",
        sum = "h1:9qC72Qh0+3MqyJbAn8YU5xVq1frD8bn3JtD2oXtafVQ=",
        version = "v1.10.0",
    )
    go_repository(
        name = "org_uber_go_goleak",
        importpath = "go.uber.org/goleak",
        sum = "h1:xqgm/S+aQvhWFTtR0XK3Jvg7z8kGV8P4X14IzwN3Eqk=",
        version = "v1.2.0",
    )
    go_repository(
        name = "org_uber_go_multierr",
        importpath = "go.uber.org/multierr",
        sum = "h1:7fIwc/ZtS0q++VgcfqFDxSBZVv/Xo49/SYnDFupUwlI=",
        version = "v1.9.0",
    )
    go_repository(
        name = "org_uber_go_tools",
        importpath = "go.uber.org/tools",
        sum = "h1:0mgffUl7nfd+FpvXMVz4IDEaUSmT1ysygQC7qYo7sG4=",
        version = "v0.0.0-20190618225709-2cfd321de3ee",
    )
    go_repository(
        name = "org_uber_go_zap",
        importpath = "go.uber.org/zap",
        sum = "h1:FiJd5l1UOLj0wCgbSE0rwwXHzEdAZS6hiiSnxJN/D60=",
        version = "v1.24.0",
    )
    go_repository(
        name = "sh_helm_helm",
        importpath = "helm.sh/helm",
        sum = "h1:cSe3FaQOpRWLDXvTObQNj0P7WI98IG5yloU6tQVls2k=",
        version = "v2.17.0+incompatible",
    )
    go_repository(
        name = "sh_helm_helm_v3",
        importpath = "helm.sh/helm/v3",
        sum = "h1:wL7IUZ7Zyukm5Kz0OUmIFZgKHuAgByCrUcJBtY0kDyw=",
        version = "v3.10.3",
    )
    go_repository(
        name = "sm_step_go_crypto",
        importpath = "go.step.sm/crypto",
        sum = "h1:Qwxk5JrqG0Q1t8tOfDM3zKTNECG6J5J24qgWZCRM0Ic=",
        version = "v0.21.0",
    )
    go_repository(
        name = "tools_gotest_v3",
        importpath = "gotest.tools/v3",
        sum = "h1:4AuOwCGf4lLR9u3YOe2awrHygurzhO/HeQ6laiA6Sx0=",
        version = "v3.0.3",
    )
    go_repository(
        name = "xyz_gomodules_jsonpatch_v2",
        importpath = "gomodules.xyz/jsonpatch/v2",
        sum = "h1:4pT439QV83L+G9FkcCriY6EkpcK6r6bK+A5FBUMI7qY=",
        version = "v2.2.0",
    )
