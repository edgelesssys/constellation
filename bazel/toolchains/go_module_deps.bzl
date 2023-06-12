"""Go module dependencies for Bazel.

Contains the equivalent of go.mod and go.sum files for Bazel.
"""

load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    """Declare Go module dependencies for Bazel."""
    go_repository(
        name = "ag_pack_amqp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "pack.ag/amqp",
        sum = "h1:cuNDWLUTbKRtEZwhB0WQBXf9pGbm87pUBXQhvcFxBWg=",
        version = "v0.11.2",
    )
    go_repository(
        name = "cc_mvdan_editorconfig",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "mvdan.cc/editorconfig",
        sum = "h1:XL+7ys6ls/RKrkUNFQvEwIvNHh+JKx8Mj1pUV5wQxQE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "cc_mvdan_unparam",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "mvdan.cc/unparam",
        sum = "h1:VuJo4Mt0EVPychre4fNlDWDuE5AjXtPJpRUWqZDQhaI=",
        version = "v0.0.0-20230312165513-e84e2d14e3b8",
    )
    go_repository(
        name = "co_honnef_go_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "honnef.co/go/tools",
        sum = "h1:o/n5/K5gXqk8Gozvs2cnL0F2S1/g1vcGCAx2vETjITw=",
        version = "v0.4.3",
    )
    go_repository(
        name = "com_github_a8m_expect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/a8m/expect",
        sum = "h1:o0PXeXn7zLB77ajwOyT1s1HcPJ4hbV6jAvCWUwvFBUM=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_acomagu_bufpipe",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/acomagu/bufpipe",
        sum = "h1:e3H4WUzM3npvo5uv95QuJM3cQspFNtFBzvJ2oNjKIDQ=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_adalogics_go_fuzz_headers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/AdaLogics/go-fuzz-headers",
        sum = "h1:EKPd1INOIyr5hWOWhvpmQpY6tKjeG0hT1s3AMC/9fic=",
        version = "v0.0.0-20230106234847-43070de90fa1",
    )
    go_repository(
        name = "com_github_adamkorcz_go_118_fuzz_build",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/AdamKorcz/go-118-fuzz-build",
        sum = "h1:+vTEFqeoeur6XSq06bs+roX3YiT49gUniJK7Zky7Xjg=",
        version = "v0.0.0-20221215162035-5330a85ea652",
    )
    go_repository(
        name = "com_github_adamkorcz_go_fuzz_headers_1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/AdamKorcz/go-fuzz-headers-1",
        sum = "h1:1tbEqR4NyQLgiod7vLXSswHteGetAVZrMGCqrJxLKRs=",
        version = "v0.0.0-20230329111138-12e09aba5ebd",
    )

    go_repository(
        name = "com_github_adrg_xdg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/adrg/xdg",
        sum = "h1:RzRqFcjH4nE5C6oTAxhBtoE2IRyjBSa62SCbyPidvls=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_afex_hystrix_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/afex/hystrix-go",
        sum = "h1:rFw4nCn9iMW+Vajsk51NtYIcwSTkXr+JGrMd36kTDJw=",
        version = "v0.0.0-20180502004556-fa1af6a1f4f5",
    )

    go_repository(
        name = "com_github_alcortesm_tgz",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alcortesm/tgz",
        sum = "h1:uSoVVbwJiQipAclBbw+8quDsfcvFjOpI5iCf4p/cqCs=",
        version = "v0.0.0-20161220082320-9c5fe88206d7",
    )

    go_repository(
        name = "com_github_alecthomas_kingpin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/kingpin",
        sum = "h1:5svnBTFgJjZvGKyYBtMB0+m5wvrbUHiqye8wRJMlnYI=",
        version = "v2.2.6+incompatible",
    )
    go_repository(
        name = "com_github_alecthomas_kingpin_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/kingpin/v2",
        sum = "h1:ANLJcKmQm4nIaog7xdr/id6FM6zm5hHnfZrvtKPxqGg=",
        version = "v2.3.1",
    )

    go_repository(
        name = "com_github_alecthomas_template",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/template",
        sum = "h1:JYp7IbQjafoB+tBA3gMyHYHrpOtNuDiK/uB5uXxq5wM=",
        version = "v0.0.0-20190718012654-fb15b899a751",
    )
    go_repository(
        name = "com_github_alecthomas_units",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/units",
        sum = "h1:s6gZFSlWYmbqAuRjVTiNNhvNRfY2Wxp9nhfyel4rklc=",
        version = "v0.0.0-20211218093645-b94a6e3cc137",
    )

    go_repository(
        name = "com_github_andybalholm_brotli",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/andybalholm/brotli",
        sum = "h1:V7DdXeJtZscaqfNuAdSRuRFzuiKlHSC/Zh3zl9qY3JY=",
        version = "v1.0.4",
    )

    go_repository(
        name = "com_github_anmitsu_go_shlex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/anmitsu/go-shlex",
        sum = "h1:9AeTilPcZAjCFIImctFaOjnTIavg87rW78vTPkQqLI8=",
        version = "v0.0.0-20200514113438-38f4b401e2be",
    )
    go_repository(
        name = "com_github_antihax_optional",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/antihax/optional",
        sum = "h1:xK2lYat7ZLaVVcIuj82J8kIro4V6kDe0AUDFboUCwcg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_antlr_antlr4_runtime_go_antlr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/antlr/antlr4/runtime/Go/antlr",
        sum = "h1:yL7+Jz0jTC6yykIK/Wh74gnTJnrGr5AyrNMXuA0gves=",
        version = "v1.4.10",
    )
    go_repository(
        name = "com_github_aokoli_goutils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aokoli/goutils",
        sum = "h1:7fpzNGoJ3VA8qcrm++XEE1QUe0mIwNeLa02Nwq7RDkg=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_apache_arrow_go_v11",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apache/arrow/go/v11",
        sum = "h1:hqauxvFQxww+0mEU/2XHG6LT7eZternCZq+A5Yly2uM=",
        version = "v11.0.0",
    )

    go_repository(
        name = "com_github_apache_beam",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apache/beam",
        sum = "h1:8MOeoZwBgORfaJjrZxpkqJWEIzwupRGLqUqG0/mvEtQ=",
        version = "v2.32.0+incompatible",
    )

    go_repository(
        name = "com_github_apache_beam_sdks_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apache/beam/sdks/v2",
        sum = "h1:Y/FmwSp2FfVwbfM81Ix+LPENDurLIbIwkz35tXu8CIQ=",
        version = "v2.47.0-RC3",
    )
    go_repository(
        name = "com_github_apache_thrift",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apache/thrift",
        sum = "h1:qEy6UW60iVOlUy+b9ZR0d5WzUWYGOo4HfopoyBaNmoY=",
        version = "v0.16.0",
    )

    go_repository(
        name = "com_github_apex_log",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apex/log",
        sum = "h1:3Zk+boorIQAAGBrHn0JUtAau4ihMamT4WdnfdnXM1zQ=",
        version = "v1.1.4",
    )
    go_repository(
        name = "com_github_apex_logs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apex/logs",
        sum = "h1:KmEBVwfDUOTFcBO8cfkJYwdQ5487UZSN+GteOGPmiro=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_aphistic_golf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aphistic/golf",
        sum = "h1:2KLQMJ8msqoPHIPDufkxVcoTtcmE5+1sL9950m4R9Pk=",
        version = "v0.0.0-20180712155816-02c07f170c5a",
    )
    go_repository(
        name = "com_github_aphistic_sweet",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aphistic/sweet",
        sum = "h1:I4z+fAUqvKfvZV/CHi5dV0QuwbmIvYYFDjG0Ss5QpAs=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_apparentlymart_go_textseg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apparentlymart/go-textseg",
        sum = "h1:rRmlIsPEEhUTIKQb7T++Nz/A5Q6C9IuX2wFoYVvnCs0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_apparentlymart_go_textseg_v13",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apparentlymart/go-textseg/v13",
        sum = "h1:Y+KvPE1NYz0xl601PVImeQfFyEy6iT90AvPUL1NNfNw=",
        version = "v13.0.0",
    )

    go_repository(
        name = "com_github_armon_circbuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/armon/circbuf",
        sum = "h1:QEF07wC0T1rKkctt1RINW/+RMTVmiwxETico2l3gxJA=",
        version = "v0.0.0-20150827004946-bbbad097214e",
    )
    go_repository(
        name = "com_github_armon_consul_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/armon/consul-api",
        sum = "h1:G1bPvciwNyF7IUmKXNt9Ak3m6u9DE1rF+RmtIkBpVdA=",
        version = "v0.0.0-20180202201655-eb2c6b5be1b6",
    )
    go_repository(
        name = "com_github_armon_go_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/armon/go-metrics",
        sum = "h1:8GUt8eRujhVEGZFFEjBj46YV4rDjvGrNxb0KMWYkL2I=",
        version = "v0.0.0-20180917152333-f0300d1749da",
    )
    go_repository(
        name = "com_github_armon_go_radix",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/armon/go-radix",
        sum = "h1:F4z6KzEeeQIMeLFa97iZU6vupzoecKdU5TX24SNppXI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_armon_go_socks5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/armon/go-socks5",
        sum = "h1:0CwZNZbxp69SHPdPJAN/hZIm0C4OItdklCFmMRWYpio=",
        version = "v0.0.0-20160902184237-e75332964ef5",
    )
    go_repository(
        name = "com_github_aryann_difflib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aryann/difflib",
        sum = "h1:pv34s756C4pEXnjgPfGYgdhg/ZdajGhyOvzx8k+23nw=",
        version = "v0.0.0-20170710044230-e206f873d14a",
    )

    go_repository(
        name = "com_github_asaskevich_govalidator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/asaskevich/govalidator",
        sum = "h1:DklsrG3dyBCFEj5IhUbnKptjxatkF07cF2ak3yi77so=",
        version = "v0.0.0-20230301143203-a9d515a09cc2",
    )

    go_repository(
        name = "com_github_aws_aws_lambda_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-lambda-go",
        sum = "h1:SuCy7H3NLyp+1Mrfp+m80jcbi9KYWAs9/BXwppwRDzY=",
        version = "v1.13.3",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go",
        sum = "h1:HwelXYZZ8c34uFFhgVw3ybu2gB5fkk8KLj2idTvzZb8=",
        version = "v1.44.257",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2",
        sum = "h1:882kkTpSFhdgYRKVZ/VCgf7sd0ru57p2JCxz4/oN5RY=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_aws_protocol_eventstream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream",
        sum = "h1:dK82zF6kkPeCo8J1e+tGx4JdvDIQzj7ygIoLg8WMuGs=",
        version = "v1.4.10",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_config",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/config",
        sum = "h1:gc3lPsAnZpwfi2exupmgHfva0JiAY2BWDg5JWYlmA28=",
        version = "v1.18.23",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_credentials",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/credentials",
        sum = "h1:Hp9rwJS4giQ48xqonRV/s7QcDf/wxF6UY7osRmBabvI=",
        version = "v1.13.22",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_ec2_imds",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/ec2/imds",
        sum = "h1:jJPgroehGvjrde3XufFIJUZVK5A2L9a3KwSFgKy9n8w=",
        version = "v1.13.3",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_s3_manager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/s3/manager",
        sum = "h1:BRLcU36boYxw6BPOEvwJbDPuCtP7FqMhXMFk2NM6poM=",
        version = "v1.11.60",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_configsources",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/configsources",
        sum = "h1:kG5eQilShqmJbv11XL1VpyDbaEJzWxd4zRiCG30GSn4=",
        version = "v1.1.33",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_endpoints_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/endpoints/v2",
        sum = "h1:vFQlirhuM8lLlpI7imKOMsjdQLuN9CPi+k44F/OFVsk=",
        version = "v2.4.27",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_ini",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/ini",
        sum = "h1:gGLG7yKaXG02/jBlg210R7VgQIotiQntNhsCFejawx8=",
        version = "v1.3.34",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_v4a",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/v4a",
        sum = "h1:DWYZIsyqagnWL00f8M/SOr9fN063OEQWn9LLTbdYXsk=",
        version = "v1.0.23",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_autoscaling",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/autoscaling",
        sum = "h1:OpzahvFZn/B+TNWLZf0ARovZoQB0Q2MvM+y13gdL+WY=",
        version = "v1.28.6",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_cloudfront",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/cloudfront",
        sum = "h1:I+LH+gwRMD/xUt13G03O0xIBq/+R1rN1/vXQQqBVrtA=",
        version = "v1.26.2",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_cloudwatchlogs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs",
        sum = "h1:Sv9ixBhjrihZUZih+SJfyo892LXutFspfqPt5XQGc9Q=",
        version = "v1.20.7",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ec2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ec2",
        sum = "h1:xn5CI639mnWvdiweqoRx/H221Ia9Asx9XxfIRhe0MPo=",
        version = "v1.92.1",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_elasticloadbalancingv2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2",
        sum = "h1:XpIms0tmerNg/t6IiGrbKU6Au25CHyXqs8Yc3zOET5o=",
        version = "v1.19.7",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_accept_encoding",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding",
        sum = "h1:y2+VQzC6Zh2ojtV2LoC0MNwHWc6qXv/j2vrQtlftkdA=",
        version = "v1.9.11",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_checksum",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/checksum",
        sum = "h1:CeuSeq/8FnYpPtnuIeLQEEvDv9zUjneuYi8EghMBdwQ=",
        version = "v1.1.26",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_presigned_url",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/presigned-url",
        sum = "h1:0iKliEXAcCa2qVtRs7Ot5hItA2MsufrphbRFlz1Owxo=",
        version = "v1.9.27",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_s3shared",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/s3shared",
        sum = "h1:e2ooMhpYGhDnBfSvIyusvAwX7KexuZaHbQY2Dyei7VU=",
        version = "v1.14.0",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/kms",
        sum = "h1:Q03Jqh1enA8keCiGZpLetpk58Ll9iGejE5bOErxyGAU=",
        version = "v1.21.1",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_resourcegroupstaggingapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi",
        sum = "h1:68kjp2WO8gv2tqBxVmffdmMUGuk7SXEinCwrPSPcXXo=",
        version = "v1.14.7",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_s3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/s3",
        sum = "h1:PJH4I+qYjPXclKRbVCW47iYUvtXEh1u6YmDhn5J8VQE=",
        version = "v1.31.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_secretsmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/secretsmanager",
        sum = "h1:+rANS0SbrDUqF3VJeil1HJHhNK8vdUu1VGqnkr4o6kw=",
        version = "v1.19.1",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sso",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sso",
        sum = "h1:UBQjaMTCKwyUYwiVnUt6toEJwGXsLBI6al083tpjJzY=",
        version = "v1.12.10",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ssooidc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ssooidc",
        sum = "h1:PkHIIJs8qvq0e5QybnZoG1K/9QTrLr9OsqCIo59jOBA=",
        version = "v1.14.10",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sts",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sts",
        sum = "h1:uBE+Zj478pfxV98L6SEpvxYiADNjTlMNY714PJLE7uo=",
        version = "v1.18.11",
    )

    go_repository(
        name = "com_github_aws_smithy_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/smithy-go",
        sum = "h1:hgz0X/DX0dGqTYpGALqXJoRKRj5oQ7150i5FdTePzO8=",
        version = "v1.13.5",
    )

    go_repository(
        name = "com_github_aybabtme_rgbterm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aybabtme/rgbterm",
        sum = "h1:WWB576BN5zNSZc/M9d/10pqEx5VHNhaQ/yOVAkmj5Yo=",
        version = "v0.0.0-20170906152045-cc83f3b3ce59",
    )

    go_repository(
        name = "com_github_azure_azure_amqp_common_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-amqp-common-go/v2",
        sum = "h1:+QbFgmWCnPzdaRMfsI0Yb6GrRdBj5jVL8N3EXuEUcBQ=",
        version = "v2.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_pipeline_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-pipeline-go",
        sum = "h1:OLBdZJ3yvOn2MezlWvbrBMTEUQC72zAftRZOMdj5HYo=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go",
        sum = "h1:fcYLmCpyNYRnvJbPerq7U0hS+6+I79yEDJBqVNcqUzU=",
        version = "v68.0.0+incompatible",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_azcore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/azcore",
        sum = "h1:8kDqDngH+DmVBiCtIjCFTGa7MBnsIOkF9IccInFEbjk=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_azidentity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/azidentity",
        sum = "h1:uqM+VoHjVH6zdlkLF2b6O0ZANcHoj3rO0PoQ3jglUJA=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/internal",
        sum = "h1:sXr+ck84g/ZlZUOZiNELInmMgOsuGwdjjVkEIde0OtY=",
        version = "v1.3.0",
    )

    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_keyvault_azsecrets",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets",
        sum = "h1:xnO4sFyG8UH2fElBkcqLTOZsAajvKfnSlgBBW8dXYjw=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_keyvault_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal",
        sum = "h1:FbH3BbSb4bvGluTesZZ+ttN/MDsnMmQP36OSnDuSXqw=",
        version = "v0.7.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_applicationinsights_armapplicationinsights",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights",
        sum = "h1:hBrFatNIiVAwDb5GzMLjpkQ6l2/waFSvBWMBWZRH8WI=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_compute_armcompute_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4",
        sum = "h1:UPeCRD+XY7QlaGQte2EVI2iOcWvUYA2XY8w5T/8v0NQ=",
        version = "v4.2.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/internal",
        sum = "h1:mLY+pNLjCUeKhgnAJWAKhEUQM+RJQo2H1fuGSw1Ky1E=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_network_armnetwork",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork",
        sum = "h1:QM6sE5k2ZT/vI5BEe0r7mqjsUSnhVBFbOsVkEuaEfiA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_network_armnetwork_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2",
        sum = "h1:bWh0Z2rOEDfB/ywv/l0iHN1JgyazE6kW/aIA89+CEK0=",
        version = "v2.2.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_resources_armresources",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources",
        sum = "h1:ECsQtyERDVz3NP3kvDOTLvbQhqWp/x9EsGKtb4ogUr8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_storage_azblob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob",
        sum = "h1:u/LLAOFgsMv7HmNL4Qufg58y+qElGOt5qv0z1mURkRY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_azure_azure_service_bus_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-service-bus-go",
        sum = "h1:G1qBLQvHCFDv9pcpgwgFkspzvnGknJRR0PYJ9ytY/JA=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_azure_azure_storage_blob_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-storage-blob-go",
        sum = "h1:53qhf0Oxa0nOjgbDeeYPUeyiNmafAFEY95rZLK0Tj6o=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_github_azure_go_ansiterm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-ansiterm",
        sum = "h1:UQHMgLO+TxOElx5B5HZ4hJQsoJ/PvUvKRhJHDQXO8P8=",
        version = "v0.0.0-20210617225240-d185dfc1b5a1",
    )
    go_repository(
        name = "com_github_azure_go_autorest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest",
        sum = "h1:V5VMDjClD3GiElqLWO7mz2MxNAK/vTfRHdAubSIPRgs=",
        version = "v14.2.0+incompatible",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest",
        sum = "h1:I4+HL/JDvErx2LjyzaVxllw2lRDB5/BT2Bm4g20iqYw=",
        version = "v0.11.29",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_adal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest/adal",
        sum = "h1:/GblQdIudfEM3AWWZ0mrYJQSd7JS4S/Mbzh6F0ov0Xc=",
        version = "v0.9.22",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_azure_auth",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest/azure/auth",
        sum = "h1:wkAZRgT/pn8HhFyzfe9UnqOjJYqlembgCTi72Bm/xKk=",
        version = "v0.5.12",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_azure_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest/azure/cli",
        sum = "h1:w77/uPk80ZET2F+AfQExZyEWtn+0Rk/uw17m9fv5Ajc=",
        version = "v0.4.6",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_date",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest/date",
        sum = "h1:7gUk1U5M/CQbp9WoqinNzJar+8KY+LPI6wiWrP/myHw=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_mocks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest/mocks",
        sum = "h1:PGN4EDXnuQbojHbU0UWoNvmu9AGVwYHG9/fkDYhtAfw=",
        version = "v0.4.2",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_to",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest/to",
        sum = "h1:oXVqrxakqqV1UZdSazDOPOLvOIz+XA683u8EctwboHk=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_validation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/autorest/validation",
        sum = "h1:AgyqjAd94fwNAoTjl/WQXg4VvFeRFpO+UhNyRXqF1ac=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_azure_go_autorest_logger",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/logger",
        sum = "h1:IG7i4p/mDa2Ce4TRyAO8IHnVhAVF3RFU+ZtXWSmf4Tg=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_azure_go_autorest_tracing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest/tracing",
        sum = "h1:TYi4+3m5t6K48TGI9AUdb+IzbnSxvnvUMfuitfgcfuo=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_azuread_microsoft_authentication_library_for_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/AzureAD/microsoft-authentication-library-for-go",
        sum = "h1:UE9n9rkJF62ArLb1F3DEjRt8O3jLwMWdSoypKV4f3MU=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_bazelbuild_buildtools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bazelbuild/buildtools",
        sum = "h1:XmPu4mXICgdGnC5dXGjUGbwUD/kUmS0l5Aop3LaevBM=",
        version = "v0.0.0-20230317132445-9c3c1fc0106e",
    )
    go_repository(
        name = "com_github_bazelbuild_rules_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bazelbuild/rules_go",
        sum = "h1:wkJLUDx59dntWMghuL8++GteoU1To6sRoKJXuyFtmf8=",
        version = "v0.39.1",
    )

    go_repository(
        name = "com_github_beeker1121_goque",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/beeker1121/goque",
        sum = "h1:XbgLdZvVbWsK9HAhAYOp6rksTAdOVYDBQtGSVOLlJrw=",
        version = "v1.0.3-0.20191103205551-d618510128af",
    )

    go_repository(
        name = "com_github_benbjohnson_clock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/benbjohnson/clock",
        sum = "h1:ip6w0uFQkncKQ979AypyG0ER7mqUSBdKLOgAle/AT8A=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_beorn7_perks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/beorn7/perks",
        sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_bgentry_speakeasy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bgentry/speakeasy",
        sum = "h1:ByYyxL9InA1OWqxJqqp2A5pYHUrCiAL6K3J+LKSsQkY=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_bketelsen_crypt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bketelsen/crypt",
        sum = "h1:w/jqZtC9YD4DS/Vp9GhWfWcCpuAL58oTnLoI8vE9YHU=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_blakesmith_ar",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/blakesmith/ar",
        sum = "h1:m935MPodAbYS46DG4pJSv7WO+VECIWUQ7OJYSoTrMh4=",
        version = "v0.0.0-20190502131153-809d4375e1fb",
    )
    go_repository(
        name = "com_github_blang_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/blang/semver",
        sum = "h1:cQNTCjp13qL8KC3Nbxr/y2Bqb63oX6wdnnjpJbkM4JQ=",
        version = "v3.5.1+incompatible",
    )
    go_repository(
        name = "com_github_blang_semver_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/blang/semver/v4",
        sum = "h1:1PFHFE6yCCTv8C1TeyNNarDzntLi7wMI5i/pzqYIsAM=",
        version = "v4.0.0",
    )

    go_repository(
        name = "com_github_bshuster_repo_logrus_logstash_hook",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bshuster-repo/logrus-logstash-hook",
        sum = "h1:e+C0SB5R1pu//O4MQ3f9cFuPGoOVeF2fE4Og9otCc70=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_bugsnag_bugsnag_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bugsnag/bugsnag-go",
        sum = "h1:rFt+Y/IK1aEZkEHchZRSq9OQbsSzIT/OrI8YFFmRIng=",
        version = "v0.0.0-20141110184014-b1d153021fcd",
    )
    go_repository(
        name = "com_github_bugsnag_osext",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bugsnag/osext",
        sum = "h1:otBG+dV+YK+Soembjv71DPz3uX/V/6MMlSyD9JBQ6kQ=",
        version = "v0.0.0-20130617224835-0dd3f918b21b",
    )
    go_repository(
        name = "com_github_bugsnag_panicwrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bugsnag/panicwrap",
        sum = "h1:nvj0OLI3YqYXer/kZD8Ri1aaunCxIEsOst1BVJswV0o=",
        version = "v0.0.0-20151223152923-e2c28503fcd0",
    )

    go_repository(
        name = "com_github_burntsushi_toml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:9F2/+DoOYIOksmaJFPw1tGFy1eDnIJXg+UHjuD8lTak=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_burntsushi_xgb",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/BurntSushi/xgb",
        sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
        version = "v0.0.0-20160522181843-27f122750802",
    )
    go_repository(
        name = "com_github_bwesterb_go_ristretto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bwesterb/go-ristretto",
        sum = "h1:1w53tCkGhCQ5djbat3+MH0BAQ5Kfgbt56UZQ/JMzngw=",
        version = "v1.2.3",
    )

    go_repository(
        name = "com_github_caarlos0_ctrlc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/caarlos0/ctrlc",
        sum = "h1:2DtF8GSIcajgffDFJzyG15vO+1PuBWOMUdFut7NnXhw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_campoy_unique",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/campoy/unique",
        sum = "h1:V9a67dfYqPLAvzk5hMQOXYJlZ4SLIXgyKIE+ZiHzgGQ=",
        version = "v0.0.0-20180121183637-88950e537e7e",
    )
    go_repository(
        name = "com_github_casbin_casbin_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/casbin/casbin/v2",
        sum = "h1:bTwon/ECRx9dwBy2ewRVr5OiqjeXSGiTUY74sDPQi/g=",
        version = "v2.1.2",
    )

    go_repository(
        name = "com_github_cavaliercoder_badio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cavaliercoder/badio",
        sum = "h1:YYUjy5BRwO5zPtfk+aa2gw255FIIoi93zMmuy19o0bc=",
        version = "v0.0.0-20160213150051-ce5280129e9e",
    )
    go_repository(
        name = "com_github_cavaliercoder_go_cpio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cavaliercoder/go-cpio",
        sum = "h1:hHg27A0RSSp2Om9lubZpiMgVbvn39bsUmW9U5h0twqc=",
        version = "v0.0.0-20180626203310-925f9528c45e",
    )
    go_repository(
        name = "com_github_cavaliercoder_go_rpm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cavaliercoder/go-rpm",
        sum = "h1:jP7ki8Tzx9ThnFPLDhBYAhEpI2+jOURnHQNURgsMvnY=",
        version = "v0.0.0-20200122174316-8cb9fd9c31a8",
    )
    go_repository(
        name = "com_github_cenkalti_backoff",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cenkalti/backoff",
        sum = "h1:tNowT99t7UNflLxfYYSlKYsBpXdEet03Pg2g16Swow4=",
        version = "v2.2.1+incompatible",
    )

    go_repository(
        name = "com_github_cenkalti_backoff_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cenkalti/backoff/v3",
        sum = "h1:cfUAAO3yvKMYKPrvhDuHSwQnhZNk/RMHKdZqKTxfm6M=",
        version = "v3.2.2",
    )
    go_repository(
        name = "com_github_cenkalti_backoff_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cenkalti/backoff/v4",
        sum = "h1:HN5dHm3WBOgndBH6E8V0q2jIYIR3s9yglV8k/+MN3u4=",
        version = "v4.2.0",
    )
    go_repository(
        name = "com_github_census_instrumentation_opencensus_proto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/census-instrumentation/opencensus-proto",
        sum = "h1:iKLQ0xPNFxR/2hzXZMrBo8f1j86j5WHzznCCQxV/b8g=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_certifi_gocertifi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/certifi/gocertifi",
        sum = "h1:uH66TXeswKn5PW5zdZ39xEwfS9an067BirqA+P4QaLI=",
        version = "v0.0.0-20200922220541-2c3bb06c6054",
    )
    go_repository(
        name = "com_github_cespare_xxhash",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cespare/xxhash",
        sum = "h1:a6HrQnmkObjyL+Gs60czilIUGqrzKutQD6XZog3p+ko=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_cespare_xxhash_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cespare/xxhash/v2",
        sum = "h1:DC2CZ1Ep5Y4k3ZQ899DldepgrayRUGE6BBZ/cd9Cj44=",
        version = "v2.2.0",
    )
    go_repository(
        name = "com_github_chai2010_gettext_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chai2010/gettext-go",
        sum = "h1:1Lwwip6Q2QGsAdl/ZKPCwTe9fe0CjlUbqj5bFNSjIRk=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_checkpoint_restore_go_criu_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/checkpoint-restore/go-criu/v5",
        sum = "h1:wpFFOoomK3389ue2lAb0Boag6XPht5QYpipxmSNL4d8=",
        version = "v5.3.0",
    )

    go_repository(
        name = "com_github_chzyer_logex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chzyer/logex",
        sum = "h1:+eqR0HfOetur4tgnC8ftU5imRnhi4te+BadWS95c5AM=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_chzyer_readline",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chzyer/readline",
        sum = "h1:lSwwFrbNviGePhkewF1az4oLmcwqCZijQ2/Wi3BGHAI=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_chzyer_test",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chzyer/test",
        sum = "h1:dZ0/VyGgQdVGAss6Ju0dt5P0QltE0SFY5Woh6hbIfiQ=",
        version = "v0.0.0-20210722231415-061457976a23",
    )
    go_repository(
        name = "com_github_cilium_ebpf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cilium/ebpf",
        sum = "h1:64sn2K3UKw8NbP/blsixRpF3nXuyhz/VjRlRzvlBRu4=",
        version = "v0.9.1",
    )

    go_repository(
        name = "com_github_clbanning_x2j",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/clbanning/x2j",
        sum = "h1:EdRZT3IeKQmfCSrgo8SZ8V3MEnskuJP0wCYNpe+aiXo=",
        version = "v0.0.0-20191024224557-825249438eec",
    )

    go_repository(
        name = "com_github_client9_misspell",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/client9/misspell",
        sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
        version = "v0.3.4",
    )
    go_repository(
        name = "com_github_cloudflare_circl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cloudflare/circl",
        # keep
        patches = [
            "//3rdparty/bazel/com_github_cloudflare_circl:math_fp448_BUILD_bazel.patch",
            "//3rdparty/bazel/com_github_cloudflare_circl:math_fp25519_BUILD_bazel.patch",
            "//3rdparty/bazel/com_github_cloudflare_circl:dh_x448_BUILD_bazel.patch",
            "//3rdparty/bazel/com_github_cloudflare_circl:dh_x25519_BUILD_bazel.patch",
        ],
        sum = "h1:fE/Qz0QdIGqeWfnwq0RE0R7MI51s0M2E4Ga9kq5AEMs=",
        version = "v1.3.3",
    )

    go_repository(
        name = "com_github_cncf_udpa_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cncf/udpa/go",
        sum = "h1:QQ3GSy+MqSHxm/d8nCtnAiZdYFd45cYZPs8vOOIYKfk=",
        version = "v0.0.0-20220112060539-c52dc94e7fbe",
    )
    go_repository(
        name = "com_github_cncf_xds_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cncf/xds/go",
        sum = "h1:58f1tJ1ra+zFINPlwLWvQsR9CzAKt2e+EWV2yX9oXQ4=",
        version = "v0.0.0-20230310173818-32f1caf87195",
    )

    go_repository(
        name = "com_github_cockroachdb_cockroach_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cockroachdb/cockroach-go/v2",
        sum = "h1:fNmtG6XhoA1DhdDCIu66YyGSsNb1szj4CaAsbDxRmy4=",
        version = "v2.3.3",
    )

    go_repository(
        name = "com_github_cockroachdb_datadriven",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cockroachdb/datadriven",
        sum = "h1:xD/lrqdvwsc+O2bjSSi3YqY73Ke3LAiSCx49aCesA0E=",
        version = "v0.0.0-20200714090401-bf6692d28da5",
    )
    go_repository(
        name = "com_github_cockroachdb_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cockroachdb/errors",
        sum = "h1:Lap807SXTH5tri2TivECb/4abUkMZC9zRoLarvcKDqs=",
        version = "v1.2.4",
    )
    go_repository(
        name = "com_github_cockroachdb_logtags",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cockroachdb/logtags",
        sum = "h1:o/kfcElHqOiXqcou5a3rIlMc7oJbMQkeLk0VQJ7zgqY=",
        version = "v0.0.0-20190617123548-eb05cc24525f",
    )

    go_repository(
        name = "com_github_codahale_hdrhistogram",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/codahale/hdrhistogram",
        sum = "h1:qMd81Ts1T2OTKmB4acZcyKaMtRnY5Y44NuXGX2GFJ1w=",
        version = "v0.0.0-20161010025455-3a0bb77429bd",
    )

    go_repository(
        name = "com_github_codahale_rfc6979",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/codahale/rfc6979",
        sum = "h1:EDmT6Q9Zs+SbUoc7Ik9EfrFqcylYqgPZ9ANSbTAntnE=",
        version = "v0.0.0-20141003034818-6a90f24967eb",
    )

    go_repository(
        name = "com_github_common_nighthawk_go_figure",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/common-nighthawk/go-figure",
        sum = "h1:J5BL2kskAlV9ckgEsNQXscjIaLiOYiZ75d4e94E6dcQ=",
        version = "v0.0.0-20210622060536-734e95fb86be",
    )
    go_repository(
        name = "com_github_container_orchestrated_devices_container_device_interface",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/container-orchestrated-devices/container-device-interface",
        sum = "h1:PqQGqJqQttMP5oJ/qNGEg8JttlHqGY3xDbbcKb5T9E8=",
        version = "v0.5.4",
    )

    go_repository(
        name = "com_github_container_storage_interface_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/container-storage-interface/spec",
        sum = "h1:gW8eyFQUZWWrMWa8p1seJ28gwDoN5CVJ4uAbQ+Hdycw=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_containerd_aufs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/aufs",
        sum = "h1:2oeJiwX5HstO7shSrPZjrohJZLzK36wvpdmzDRkL/LY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_btrfs_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/btrfs/v2",
        sum = "h1:FN4wsx7KQrYoLXN7uLP0vBV4oVWHOIKDRQ1G2Z0oL5M=",
        version = "v2.0.0",
    )

    go_repository(
        name = "com_github_containerd_cgroups",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/cgroups",
        sum = "h1:v8rEWFl6EoqHB+swVNjVoCJE8o3jX7e8nqBGPLaDFBM=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_containerd_cgroups_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/cgroups/v3",
        sum = "h1:4hfGvu8rfGIwVIDd+nLzn/B9ZXx4BcCjzt5ToenJRaE=",
        version = "v3.0.1",
    )

    go_repository(
        name = "com_github_containerd_console",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/console",
        sum = "h1:lIr7SlA5PxZyMV30bDW0MGbiOPXwc63yRuCP0ARubLw=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_containerd_containerd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/containerd",
        sum = "h1:G/ZQr3gMZs6ZT0qPUZ15znx5QSdQdASW11nXTLTM2Pg=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_containerd_continuity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/continuity",
        sum = "h1:nisirsYROK15TAMVukJOUyGJjz4BNQJBVsNvAXZJ/eg=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_containerd_fifo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/fifo",
        sum = "h1:4I2mbh5stb1u6ycIABlBw9zgtlK8viPI9QkQNRQEEmY=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_containerd_go_cni",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/go-cni",
        sum = "h1:ORi7P1dYzCwVM6XPN4n3CbkuOx/NZ2DOqy+SHRdo9rU=",
        version = "v1.1.9",
    )
    go_repository(
        name = "com_github_containerd_go_runc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/go-runc",
        sum = "h1:oU+lLv1ULm5taqgV/CJivypVODI4SUz1znWjv3nNYS0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_imgcrypt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/imgcrypt",
        sum = "h1:WSf9o9EQ0KGHiUx2ESFZ+PKf4nxK9BcvV/nJDX8RkB4=",
        version = "v1.1.7",
    )
    go_repository(
        name = "com_github_containerd_nri",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/nri",
        sum = "h1:2ZM4WImye1ypSnE7COjOvPAiLv84kaPILBDvb1tbDK8=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_containerd_stargz_snapshotter_estargz",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/stargz-snapshotter/estargz",
        sum = "h1:OqlDCK3ZVUO6C3B/5FSkDwbkEETK84kQgEeFwDC+62k=",
        version = "v0.14.3",
    )
    go_repository(
        name = "com_github_containerd_ttrpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/ttrpc",
        sum = "h1:VWv/Rzx023TBLv4WQ+9WPXlBG/s3rsRjY3i9AJ2BJdE=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_containerd_typeurl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/typeurl",
        sum = "h1:Chlt8zIieDbzQFzXzAeBEF92KhExuE4p9p92/QmY7aY=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_containerd_typeurl_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/typeurl/v2",
        sum = "h1:yNAhJvbNEANt7ck48IlEGOxP7YAp6LLpGn5jZACDNIE=",
        version = "v2.1.0",
    )

    go_repository(
        name = "com_github_containerd_zfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/zfs",
        sum = "h1:cXLJbx+4Jj7rNsTiqVfm6i+RNLx6FFA2fMmDlEf+Wm8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containernetworking_cni",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containernetworking/cni",
        sum = "h1:wtRGZVv7olUHMOqouPpn3cXJWpJgM6+EUl31EQbXALQ=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_containernetworking_plugins",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containernetworking/plugins",
        sum = "h1:SWgg3dQG1yzUo4d9iD8cwSVh1VqI+bP7mkPDoSfP9VU=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_containers_ocicrypt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containers/ocicrypt",
        sum = "h1:uoG52u2e91RE4UqmBICZY8dNshgfvkdl3BW6jnxiFaI=",
        version = "v1.1.6",
    )
    go_repository(
        name = "com_github_coredns_caddy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coredns/caddy",
        sum = "h1:ezvsPrT/tA/7pYDBZxu0cT0VmWk75AfIaf6GSYCNMf0=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_coredns_corefile_migration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coredns/corefile-migration",
        sum = "h1:MdOkT6F3ehju/n9tgxlGct8XAajOX2vN+wG7To4BWSI=",
        version = "v1.0.20",
    )
    go_repository(
        name = "com_github_coreos_bbolt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/bbolt",
        sum = "h1:wZwiHHUieZCquLkDL0B8UhzreNWsPHooDAG3q34zk0s=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_coreos_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/etcd",
        sum = "h1:8F3hqu9fGYLBifCmRCJsicFqDx/D68Rt3q1JMazcgBQ=",
        version = "v3.3.13+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-etcd",
        sum = "h1:bXhRBIXoTm9BYHS3gE0TtQuyNZyeEMux2sDi4oo5YOo=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_oidc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-oidc",
        sum = "h1:sdJrfw8akMnCuUlaZU3tE/uYXFgfqom8DBE9so9EBsM=",
        version = "v2.1.0+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_oidc_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-oidc/v3",
        sum = "h1:VxKtbccHZxs8juq7RdJntSqtXFtde9YpNpGn0yqgEHw=",
        version = "v3.5.0",
    )
    go_repository(
        name = "com_github_coreos_go_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-semver",
        sum = "h1:yi21YpKnrx1gt5R+la8n5WgS0kCrsPp33dmEyHReZr4=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_coreos_go_systemd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-systemd",
        sum = "h1:Wf6HqHfScWJN9/ZjdUKyjop4mf3Qdd+1TvvltAvM3m8=",
        version = "v0.0.0-20190321100706-95778dfbb74e",
    )
    go_repository(
        name = "com_github_coreos_go_systemd_v22",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-systemd/v22",
        sum = "h1:RrqgGjYQKalulkV8NGVIfkXQf6YYmOyiJKk8iXXhfZs=",
        version = "v22.5.0",
    )
    go_repository(
        name = "com_github_coreos_pkg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/pkg",
        sum = "h1:lBNOc5arjvs8E5mO2tbpBpLoyyu8B6e44T7hJy6potg=",
        version = "v0.0.0-20180928190104-399ea9e2e55f",
    )
    go_repository(
        name = "com_github_cosi_project_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cosi-project/runtime",
        sum = "h1:f8++A7HUu7pQv9G3IhQworfA4TFLdzGWl3W+jLQF3Oo=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cpuguy83/go-md2man",
        sum = "h1:BSKMNlYxDvnunlTymqtgONjNnaRV1sTpcovwwjF22jk=",
        version = "v1.0.10",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cpuguy83/go-md2man/v2",
        sum = "h1:p1EgwI/C7NhT0JmVkwCD2ZBK8j4aeHQX2pMHHBfMQ6w=",
        version = "v2.0.2",
    )
    go_repository(
        name = "com_github_creack_pty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/creack/pty",
        sum = "h1:n56/Zwd5o6whRC5PMGretI4IdRLlmBXYNjScPaBgsbY=",
        version = "v1.1.18",
    )
    go_repository(
        name = "com_github_cyberphone_json_canonicalization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cyberphone/json-canonicalization",
        sum = "h1:vU+EP9ZuFUCYE0NYLwTSob+3LNEJATzNfP/DC7SWGWI=",
        version = "v0.0.0-20220623050100-57a0ce2678a7",
    )
    go_repository(
        name = "com_github_cyphar_filepath_securejoin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cyphar/filepath-securejoin",
        sum = "h1:YX6ebbZCZP7VkM3scTTokDgBL2TY741X51MTk3ycuNI=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_danieljoos_wincred",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/danieljoos/wincred",
        sum = "h1:QLdCxFs1/Yl4zduvBdcHB8goaYk9RARS2SgLLRuAyr0=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_data_dog_go_sqlmock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/DATA-DOG/go-sqlmock",
        sum = "h1:Shsta01QNfFxHCfpW6YH2STWB0MudeXXEWMr20OEh60=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_daviddengcn_go_colortext",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/daviddengcn/go-colortext",
        sum = "h1:ANqDyC0ys6qCSvuEK7l3g5RaehL/Xck9EX8ATG8oKsE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_denisenkom_go_mssqldb",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/denisenkom/go-mssqldb",
        sum = "h1:RSohk2RsiZqLZ0zCjtfn3S4Gp4exhpBWHyQ7D0yGjAk=",
        version = "v0.9.0",
    )

    go_repository(
        name = "com_github_devigned_tab",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/devigned/tab",
        sum = "h1:3mD6Kb1mUOYeLpJvTVSDwSg5ZsfSxfvxGRTxRsJsITA=",
        version = "v0.1.1",
    )

    go_repository(
        name = "com_github_dgrijalva_jwt_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dgrijalva/jwt-go",
        sum = "h1:7qlOGliEKZXTDg6OTjfoBKDXWrumCAMpl/TFQ4/5kLM=",
        version = "v3.2.0+incompatible",
    )

    go_repository(
        name = "com_github_dgryski_go_rendezvous",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dgryski/go-rendezvous",
        sum = "h1:lO4WD4F/rVNCu3HqELle0jiPLLBs70cWOduZpkS1E78=",
        version = "v0.0.0-20200823014737-9f7001d12a5f",
    )
    go_repository(
        name = "com_github_dgryski_go_sip13",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dgryski/go-sip13",
        sum = "h1:RMLoZVzv4GliuWafOuPuQDKSm1SJph7uCRnnS61JAn4=",
        version = "v0.0.0-20181026042036-e10d5fee7954",
    )

    go_repository(
        name = "com_github_dimchansky_utfbom",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dimchansky/utfbom",
        sum = "h1:vV6w1AhK4VMnhBno/TPVCoK9U/LP0PkLCS9tbxHdi/U=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_distribution_distribution_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/distribution/distribution/v3",
        sum = "h1:aBfCb7iqHmDEIp6fBvC/hQUddQfg+3qdYjwzaiP9Hnc=",
        version = "v3.0.0-20221208165359-362910506bc2",
    )

    go_repository(
        name = "com_github_dnaeon_go_vcr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dnaeon/go-vcr",
        sum = "h1:zHCHvJYTMh1N7xnV7zf1m1GPBF9Ad0Jk/whtQ1663qI=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_docker_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/cli",
        sum = "h1:LRyWITpGzl2C9e9uGxzisptnxAn1zfZKXy13Ul2Q5oM=",
        version = "v23.0.1+incompatible",
    )
    go_repository(
        name = "com_github_docker_distribution",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/distribution",
        sum = "h1:Q50tZOPR6T/hjNsyc9g8/syEs6bk8XXApsHjKukMl68=",
        version = "v2.8.1+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/docker",
        sum = "h1:9GhVsShNWz1hO//9BNg/dpMnZW25KydO4wtVxWAIbho=",
        version = "v23.0.3+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker_credential_helpers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/docker-credential-helpers",
        sum = "h1:xtCHsjxogADNZcdv1pKUHXryefjlVRqWqIhk/uXJp0A=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_docker_go_connections",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/go-connections",
        sum = "h1:El9xVISelRB7BuFusrZozjnkIM5YnzCViNKohAFqRJQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_docker_go_events",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/go-events",
        sum = "h1:+pKlWGMw7gf6bQ+oDZB4KHQFypsfjYlq/C4rfL7D3g8=",
        version = "v0.0.0-20190806004212-e31b211e4f1c",
    )
    go_repository(
        name = "com_github_docker_go_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/go-metrics",
        sum = "h1:AgB/0SvBxihN0X8OR4SjsblXkbMvalQ8cjmtKQ2rQV8=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_docker_go_units",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/go-units",
        sum = "h1:69rxXcBk27SvSaaxTtLh/8llcHD8vYHT7WSdRZ/jvr4=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_docker_libtrust",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/libtrust",
        sum = "h1:ZClxb8laGDf5arXfYcAtECDFgAgHklGI8CxgjHnXKJ4=",
        version = "v0.0.0-20150114040149-fa567046d9b1",
    )
    go_repository(
        name = "com_github_docopt_docopt_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docopt/docopt-go",
        sum = "h1:bWDMxwH3px2JBh6AyO7hdCn/PkvCZXii8TGj7sbtEbQ=",
        version = "v0.0.0-20180111231733-ee0de3bc6815",
    )
    go_repository(
        name = "com_github_dustin_go_humanize",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dustin/go-humanize",
        sum = "h1:GzkhY7T5VNhEkwH0PVJgjz+fX1rhBrR7pRT3mDkpeCY=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_eapache_go_resiliency",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/eapache/go-resiliency",
        sum = "h1:1NtRmCAqadE2FN4ZcN6g90TP3uk8cg9rn9eNK2197aU=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_eapache_go_xerial_snappy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/eapache/go-xerial-snappy",
        sum = "h1:YEetp8/yCZMuEPMUDHG0CW/brkkEp8mzqk2+ODEitlw=",
        version = "v0.0.0-20180814174437-776d5712da21",
    )
    go_repository(
        name = "com_github_eapache_queue",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/eapache/queue",
        sum = "h1:YOEu7KNc61ntiQlcEeUIoDTJ2o8mQznoNvUhiigpIqc=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_edgelesssys_go_azguestattestation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/edgelesssys/go-azguestattestation",
        sum = "h1:J9k1gV8YA5beC6jANKQy5O7UtaKS3ueuanxUan5Y5NU=",
        version = "v0.0.0-20230303085714-62ede861d33f",
    )
    go_repository(
        name = "com_github_edgelesssys_go_tdx_qpl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/edgelesssys/go-tdx-qpl",
        sum = "h1:Q2TI34V/NCLGQQkdc0/KmPx/7ix9YnGDUQDT+gqvDw0=",
        version = "v0.0.0-20230530085549-fd2878a4dead",
    )

    go_repository(
        name = "com_github_edsrzf_mmap_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/edsrzf/mmap-go",
        sum = "h1:CEBF7HpRnUCSJgGUb5h1Gm7e3VkmVDrR8lvWVLtrOFw=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_eggsampler_acme_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/eggsampler/acme/v3",
        sum = "h1:5M7vwYRy65iPpCFHZ01RyWXmYT8e8MlcWn/9BUUB7Ro=",
        version = "v3.3.0",
    )

    go_repository(
        name = "com_github_emicklei_go_restful_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/emicklei/go-restful/v3",
        sum = "h1:rc42Y5YTp7Am7CS630D7JmhRjq4UlEUuEKfrDac4bSQ=",
        version = "v3.10.1",
    )

    go_repository(
        name = "com_github_emirpasic_gods",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/emirpasic/gods",
        sum = "h1:FXtiHYKDGKCW2KzwZKx0iC0PQmdlorYgdFG9jPXJ1Bc=",
        version = "v1.18.1",
    )
    go_repository(
        name = "com_github_envoyproxy_go_control_plane",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/envoyproxy/go-control-plane",
        sum = "h1:jtLewhRR2vMRNnq2ZZUoCjUlgut+Y0+sDDWPOfwOi1o=",
        version = "v0.11.0",
    )
    go_repository(
        name = "com_github_envoyproxy_protoc_gen_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/envoyproxy/protoc-gen-validate",
        sum = "h1:oIfnZFdC0YhpNNEX+SuIqko4cqqVZeN9IGTrhZje83Y=",
        version = "v0.10.0",
    )

    go_repository(
        name = "com_github_etcd_io_gofail",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/etcd-io/gofail",
        sum = "h1:Y2I0lxOttdUKz+hNaIdG3FtjuQrTmwXun1opRV65IZc=",
        version = "v0.0.0-20190801230047-ad7f989257ca",
    )
    go_repository(
        name = "com_github_euank_go_kmsg_parser",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/euank/go-kmsg-parser",
        sum = "h1:cHD53+PLQuuQyLZeriD1V/esuG4MuU0Pjs5y6iknohY=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_evanphx_json_patch",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/evanphx/json-patch",
        sum = "h1:jBYDEEiFBPxA0v50tFdvOzQQTCvpL6mnFh5mB2/l16U=",
        version = "v5.6.0+incompatible",
    )
    go_repository(
        name = "com_github_evanphx_json_patch_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/evanphx/json-patch/v5",
        sum = "h1:b91NhWfaz02IuVxO9faSllyAtNXHMPkC5J8sJCLunww=",
        version = "v5.6.0",
    )
    go_repository(
        name = "com_github_exponent_io_jsonpath",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/exponent-io/jsonpath",
        sum = "h1:105gxyaGwCFad8crR9dcMQWvV9Hvulu6hwUh4tWPJnM=",
        version = "v0.0.0-20151013193312-d6023ce2651d",
    )
    go_repository(
        name = "com_github_facebookgo_clock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/facebookgo/clock",
        sum = "h1:yDWHCSQ40h88yih2JAcL6Ls/kVkSE8GFACTGVnMPruw=",
        version = "v0.0.0-20150410010913-600d898af40a",
    )
    go_repository(
        name = "com_github_facebookgo_limitgroup",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/facebookgo/limitgroup",
        sum = "h1:IeaD1VDVBPlx3viJT9Md8if8IxxJnO+x0JCGb054heg=",
        version = "v0.0.0-20150612190941-6abd8d71ec01",
    )
    go_repository(
        name = "com_github_facebookgo_muster",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/facebookgo/muster",
        sum = "h1:a4DFiKFJiDRGFD1qIcqGLX/WlUMD9dyLSLDt+9QZgt8=",
        version = "v0.0.0-20150708232844-fd3d7953fd52",
    )

    go_repository(
        name = "com_github_fatih_camelcase",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fatih/camelcase",
        sum = "h1:hxNvNX/xYBp0ovncs8WyWZrOrpBNub/JfaMvbURyft8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_fatih_color",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fatih/color",
        sum = "h1:qfhVLaG5s+nCROl1zJsZRxFeYrHLqWroPOQ8BWiNb4w=",
        version = "v1.14.1",
    )

    go_repository(
        name = "com_github_felixge_httpsnoop",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/felixge/httpsnoop",
        sum = "h1:s/nj+GCswXYzN5v2DpNMuMQYe+0DDwt5WVCU6CWBdXk=",
        version = "v1.0.3",
    )

    go_repository(
        name = "com_github_flynn_go_docopt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/flynn/go-docopt",
        sum = "h1:Ss/B3/5wWRh8+emnK0++g5zQzwDTi30W10pKxKc4JXI=",
        version = "v0.0.0-20140912013429-f6dd2ebbb31e",
    )
    go_repository(
        name = "com_github_flynn_go_shlex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/flynn/go-shlex",
        sum = "h1:BHsljHzVlRcyQhjrss6TZTdY2VfCqZPbv5k3iBFa2ZQ=",
        version = "v0.0.0-20150515145356-3f9db97f8568",
    )
    go_repository(
        name = "com_github_form3tech_oss_jwt_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/form3tech-oss/jwt-go",
        sum = "h1:/l4kBbb4/vGSsdtB5nUe8L7B9mImVMaBPw9L/0TBHU8=",
        version = "v3.2.5+incompatible",
    )
    go_repository(
        name = "com_github_fortytw2_leaktest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fortytw2/leaktest",
        sum = "h1:u8491cBMTQ8ft8aeV+adlcytMZylmA5nnwwkRZjI8vw=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_foxcpp_go_mockdns",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/foxcpp/go-mockdns",
        sum = "h1:7jBqxd3WDWwi/6WhDvacvH1XsN3rOLXyHM1uhvIx6FI=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_franela_goblin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/franela/goblin",
        sum = "h1:gb2Z18BhTPJPpLQWj4T+rfKHYCHxRHCtRxhKKjRidVw=",
        version = "v0.0.0-20200105215937-c9ffbefa60db",
    )
    go_repository(
        name = "com_github_franela_goreq",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/franela/goreq",
        sum = "h1:a9ENSRDFBUPkJ5lCgVZh26+ZbGyoVJG7yb5SSzF5H54=",
        version = "v0.0.0-20171204163338-bcd34c9993f8",
    )

    go_repository(
        name = "com_github_frankban_quicktest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/frankban/quicktest",
        sum = "h1:g2rn0vABPOOXmZUj+vbmUp0lPoXEMuhTpIluN0XL9UY=",
        version = "v1.14.4",
    )
    go_repository(
        name = "com_github_fsnotify_fsnotify",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:n+5WquG0fcWoWp6xPWfHdbskMCQaFnG6PfBrh1Ky4HY=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_fullstorydev_grpcurl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fullstorydev/grpcurl",
        sum = "h1:xJWosq3BQovQ4QrdPO72OrPiWuGgEsxY8ldYsJbPrqI=",
        version = "v1.8.7",
    )
    go_repository(
        name = "com_github_fvbommel_sortorder",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fvbommel/sortorder",
        sum = "h1:dSnXLt4mJYH25uDDGa3biZNQsozaUWDSWeKJ0qqFfzE=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_fxamacker_cbor_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fxamacker/cbor/v2",
        sum = "h1:ri0ArlOR+5XunOP8CRUowT0pSJOwhW098ZCUyskZD88=",
        version = "v2.4.0",
    )
    go_repository(
        name = "com_github_gabriel_vasile_mimetype",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gabriel-vasile/mimetype",
        sum = "h1:w5qFW6JKBz9Y393Y4q372O9A7cUSequkh1Q7OhCmWKU=",
        version = "v1.4.2",
    )

    go_repository(
        name = "com_github_gertd_go_pluralize",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gertd/go-pluralize",
        sum = "h1:M3uASbVjMnTsPb0PNqg+E/24Vwigyo/tvyMTtAlLgiA=",
        version = "v0.2.1",
    )

    go_repository(
        name = "com_github_getsentry_raven_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/getsentry/raven-go",
        sum = "h1:no+xWJRb5ZI7eE8TWgIq1jLulQiIoLG0IfYxv5JYMGs=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_ghodss_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ghodss/yaml",
        sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_gin_contrib_sse",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gin-contrib/sse",
        sum = "h1:Y/yl/+YNO8GZSjAhjMsSuLt29uWRFHdHYUb5lYOV9qE=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_gin_gonic_gin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gin-gonic/gin",
        sum = "h1:fi+bqFAx/oLK54somfCtEZs9HeH1LHVoEPUgARpTqyc=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_gliderlabs_ssh",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gliderlabs/ssh",
        sum = "h1:OcaySEmAQJgyYcArR+gGGTHCyE7nvhEMTlYY+Dp8CpY=",
        version = "v0.3.5",
    )

    go_repository(
        name = "com_github_go_chi_chi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-chi/chi",
        sum = "h1:fGFk2Gmi/YKXk0OmGfBh0WgmN3XB8lVnEyNz34tQRec=",
        version = "v4.1.2+incompatible",
    )
    go_repository(
        name = "com_github_go_errors_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-errors/errors",
        sum = "h1:J6MZopCL4uSllY1OfXM374weqZFFItUbrImctkmUxIA=",
        version = "v1.4.2",
    )
    go_repository(
        name = "com_github_go_git_gcfg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-git/gcfg",
        sum = "h1:Q5ViNfGF8zFgyJWPqYwA7qGFoMTEiBmdlkcfRmpIMa4=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_go_git_go_billy_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-git/go-billy/v5",
        sum = "h1:Uwp5tDRkPr+l/TnbHOQzp+tmJfLceOlbVucgpTz8ix4=",
        version = "v5.4.1",
    )
    go_repository(
        name = "com_github_go_git_go_git_fixtures_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-git/go-git-fixtures/v4",
        sum = "h1:y5z6dd3qi8Hl+stezc8p3JxDkoTRqMAlKnXHuzrfjTQ=",
        version = "v4.3.1",
    )
    go_repository(
        name = "com_github_go_git_go_git_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-git/go-git/v5",
        sum = "h1:q4ZRqQl4pR/ZJHc1L5CFjGA1a10u76aV1iC+nh+bHsk=",
        version = "v5.6.1",
    )
    go_repository(
        name = "com_github_go_gl_glfw",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-gl/glfw",
        sum = "h1:QbL/5oDUmRBzO9/Z7Seo6zf912W/a6Sr4Eu0G/3Jho0=",
        version = "v0.0.0-20190409004039-e6da0acd62b1",
    )
    go_repository(
        name = "com_github_go_gl_glfw_v3_3_glfw",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-gl/glfw/v3.3/glfw",
        sum = "h1:WtGNWLvXpe6ZudgnXrq0barxBImvnnJoMEhXAzcbM0I=",
        version = "v0.0.0-20200222043503-6f7a984d4dc4",
    )
    go_repository(
        name = "com_github_go_gorp_gorp_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-gorp/gorp/v3",
        sum = "h1:PUjzYdYu3HBOh8LE+UUmRG2P0IRDak9XMeGNvaeq4Ow=",
        version = "v3.0.5",
    )
    go_repository(
        name = "com_github_go_ini_ini",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-ini/ini",
        sum = "h1:Mujh4R/dH6YL8bxuISne3xX2+qcQ9p0IxKAP6ExWoUo=",
        version = "v1.25.4",
    )
    go_repository(
        name = "com_github_go_jose_go_jose_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-jose/go-jose/v3",
        sum = "h1:s6rrhirfEP/CGIoc6p+PZAeogN2SxKav6Wp7+dyMWVo=",
        version = "v3.0.0",
    )
    go_repository(
        name = "com_github_go_kit_kit",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-kit/kit",
        sum = "h1:dXFJfIHVvUcpSgDOV+Ne6t7jXri8Tfv2uOLHUZ2XNuo=",
        version = "v0.10.0",
    )
    go_repository(
        name = "com_github_go_kit_log",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-kit/log",
        sum = "h1:MRVx0/zhvdseW+Gza6N9rVzU/IVzaeE1SFI4raAhmBU=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_go_logfmt_logfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logfmt/logfmt",
        sum = "h1:otpy5pqBCBZ1ng9RQ0dPu4PN7ba75Y/aA+UpowDyNVA=",
        version = "v0.5.1",
    )
    go_repository(
        name = "com_github_go_logr_logr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logr/logr",
        sum = "h1:g01GSCwiDw2xSZfjJ2/T9M+S6pFdcNtFYsp+Y43HYDQ=",
        version = "v1.2.4",
    )
    go_repository(
        name = "com_github_go_logr_stdr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logr/stdr",
        sum = "h1:hSWxHoqTgW2S2qGc0LTAI563KZ5YKYRhT3MFKZMbjag=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_go_logr_zapr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logr/zapr",
        sum = "h1:QHVo+6stLbfJmYGkQ7uGHUCu5hnAFAj6mDe6Ea0SeOo=",
        version = "v1.2.4",
    )

    go_repository(
        name = "com_github_go_openapi_analysis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/analysis",
        sum = "h1:ZDFLvSNxpDaomuCueM0BlSXxpANBlFYiBvr+GXrvIHc=",
        version = "v0.21.4",
    )
    go_repository(
        name = "com_github_go_openapi_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/errors",
        sum = "h1:rz6kiC84sqNQoqrtulzaL/VERgkoCyB6WdEkc2ujzUc=",
        version = "v0.20.3",
    )
    go_repository(
        name = "com_github_go_openapi_jsonpointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/jsonpointer",
        sum = "h1:eCs3fxoIi3Wh6vtgmLTOjdhSpiqphQ+DaPn38N2ZdrE=",
        version = "v0.19.6",
    )
    go_repository(
        name = "com_github_go_openapi_jsonreference",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/jsonreference",
        sum = "h1:FBLnyygC4/IZZr893oiomc9XaghoveYTrLC1F86HID8=",
        version = "v0.20.1",
    )
    go_repository(
        name = "com_github_go_openapi_loads",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/loads",
        sum = "h1:r2a/xFIYeZ4Qd2TnGpWDIQNcP80dIaZgf704za8enro=",
        version = "v0.21.2",
    )
    go_repository(
        name = "com_github_go_openapi_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/runtime",
        sum = "h1:HYOFtG00FM1UvqrcxbEJg/SwvDRvYLQKGhw2zaQjTcc=",
        version = "v0.26.0",
    )
    go_repository(
        name = "com_github_go_openapi_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/spec",
        sum = "h1:xnlYNQAwKd2VQRRfwTEI0DcK+2cbuvI/0c7jx3gA8/8=",
        version = "v0.20.9",
    )
    go_repository(
        name = "com_github_go_openapi_strfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/strfmt",
        sum = "h1:rspiXgNWgeUzhjo1YU01do6qsahtJNByjLVbPLNHb8k=",
        version = "v0.21.7",
    )
    go_repository(
        name = "com_github_go_openapi_swag",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/swag",
        sum = "h1:yMBqmnQ0gyZvEb/+KzuWZOXgllrXT4SADYbvDaXHv/g=",
        version = "v0.22.3",
    )
    go_repository(
        name = "com_github_go_openapi_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/validate",
        sum = "h1:G+c2ub6q47kfX1sOBLwIQwzBVt8qmOAARyo/9Fqs9NU=",
        version = "v0.22.1",
    )

    go_repository(
        name = "com_github_go_playground_assert_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-playground/assert/v2",
        sum = "h1:JvknZsQTYeFEAhQwI4qEt9cyV5ONwRHC+lYKSsYSR8s=",
        version = "v2.2.0",
    )
    go_repository(
        name = "com_github_go_playground_locales",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-playground/locales",
        sum = "h1:EWaQ/wswjilfKLTECiXz7Rh+3BjFhfDFKv/oXslEjJA=",
        version = "v0.14.1",
    )
    go_repository(
        name = "com_github_go_playground_universal_translator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-playground/universal-translator",
        sum = "h1:Bcnm0ZwsGyWbCzImXv+pAJnYK9S473LQFuzCbDbfSFY=",
        version = "v0.18.1",
    )
    go_repository(
        name = "com_github_go_playground_validator_v10",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-playground/validator/v10",
        sum = "h1:vgvQWe3XCz3gIeFDm/HnTIbj6UGmg/+t63MyGU2n5js=",
        version = "v10.14.0",
    )
    go_repository(
        name = "com_github_go_redis_redis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-redis/redis",
        sum = "h1:K0pv1D7EQUjfyoMql+r/jZqCLizCGKFlFgcHWWmHQjg=",
        version = "v6.15.9+incompatible",
    )
    go_repository(
        name = "com_github_go_redis_redis_v8",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-redis/redis/v8",
        sum = "h1:AcZZR7igkdvfVmQTPnu9WE37LRrO/YrBH5zWyjDC0oI=",
        version = "v8.11.5",
    )
    go_repository(
        name = "com_github_go_redis_redismock_v9",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-redis/redismock/v9",
        sum = "h1:mtHQi2l51lCmXIbTRTqb1EiHYe9tL5Yk5oorlSJJqR0=",
        version = "v9.0.3",
    )

    go_repository(
        name = "com_github_go_rod_rod",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-rod/rod",
        sum = "h1:uA/yLbB+t0UlqJcLJtK2pZrCNPzd15dOKRUEOnmnt9k=",
        version = "v0.112.9",
    )
    go_repository(
        name = "com_github_go_sql_driver_mysql",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-sql-driver/mysql",
        sum = "h1:lUIinVbN1DY0xBg0eMOzmmtGoHwWBbvnWubQUrtU8EI=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_github_go_stack_stack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-stack/stack",
        sum = "h1:5SgMzNM5HxrEjV0ww2lTmX6E2Izsfxas4+YHWRs3Lsk=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_go_task_slim_sprig",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-task/slim-sprig",
        sum = "h1:tfuBGBXKqDEevZMzYi5KSi8KkcZtzBcTgAUUtapy0OI=",
        version = "v0.0.0-20230315185526-52ccab3ef572",
    )
    go_repository(
        name = "com_github_go_test_deep",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-test/deep",
        sum = "h1:WOcxcdHcvdgThNXjw0t76K42FXTU7HpNQWHpA2HHNlg=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_gobuffalo_attrs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/attrs",
        sum = "h1:hSkbZ9XSyjyBirMeqSqUrK+9HboWrweVlzRNqoBi2d4=",
        version = "v0.0.0-20190224210810-a9411de4debd",
    )
    go_repository(
        name = "com_github_gobuffalo_depgen",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/depgen",
        sum = "h1:31atYa/UW9V5q8vMJ+W6wd64OaaTHUrCUXER358zLM4=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_gobuffalo_envy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/envy",
        sum = "h1:GlXgaiBkmrYMHco6t4j7SacKO4XUjvh5pwXh0f4uxXU=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_gobuffalo_flect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/flect",
        sum = "h1:3GQ53z7E3o00C/yy7Ko8VXqQXoJGLkrTQCLTF1EjoXU=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_gobuffalo_genny",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/genny",
        sum = "h1:iQ0D6SpNXIxu52WESsD+KoQ7af2e3nCfnSBoSF/hKe0=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_gobuffalo_gitgen",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/gitgen",
        sum = "h1:mSVZ4vj4khv+oThUfS+SQU3UuFIZ5Zo6UNcvK8E8Mz8=",
        version = "v0.0.0-20190315122116-cc086187d211",
    )
    go_repository(
        name = "com_github_gobuffalo_gogen",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/gogen",
        sum = "h1:dLg+zb+uOyd/mKeQUYIbwbNmfRsr9hd/WtYWepmayhI=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_gobuffalo_logger",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/logger",
        sum = "h1:nnZNpxYo0zx+Aj9RfMPBm+x9zAU2OayFh/xrAWi34HU=",
        version = "v1.0.6",
    )
    go_repository(
        name = "com_github_gobuffalo_mapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/mapi",
        sum = "h1:fq9WcL1BYrm36SzK6+aAnZ8hcp+SrmnDyAxhNx8dvJk=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_gobuffalo_packd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/packd",
        sum = "h1:U2wXfRr4E9DH8IdsDLlRFwTZTK7hLfq9qT/QHXGVe/0=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_gobuffalo_packr_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/packr/v2",
        sum = "h1:xE1yzvnO56cUC0sTpKR3DIbxZgB54AftTFMhB2XEWlY=",
        version = "v2.8.3",
    )
    go_repository(
        name = "com_github_gobuffalo_syncx",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobuffalo/syncx",
        sum = "h1:tpom+2CJmpzAWj5/VEHync2rJGi+epHNIeRSWjzGA+4=",
        version = "v0.0.0-20190224160051-33c29581e754",
    )
    go_repository(
        name = "com_github_gobwas_glob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobwas/glob",
        sum = "h1:A4xDbljILXROh+kObIiy5kIaPYD8e96x1tgBhUI5J+Y=",
        version = "v0.2.3",
    )

    go_repository(
        name = "com_github_goccy_go_json",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/goccy/go-json",
        sum = "h1:/pAaQDLHEoCq/5FFmSKBswWmK6H0e8g4159Kc/X/nqk=",
        version = "v0.9.11",
    )

    go_repository(
        name = "com_github_godbus_dbus_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/godbus/dbus/v5",
        sum = "h1:4KLkAxT3aOY8Li4FRJe/KvhoNFFxo0m6fNuFUO8QJUk=",
        version = "v5.1.0",
    )
    go_repository(
        name = "com_github_godror_godror",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/godror/godror",
        sum = "h1:uxGAD7UdnNGjX5gf4NnEIGw0JAPTIFiqAyRBZTPKwXs=",
        version = "v0.24.2",
    )
    go_repository(
        name = "com_github_gofrs_flock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gofrs/flock",
        sum = "h1:+gYjHKf32LDeiEEFhQaotPbLuUXjY5ZqxKgXy7n59aw=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_github_gofrs_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gofrs/uuid",
        sum = "h1:yyYWMnhkhrKwwr8gAOcOCYxOOscHgDS9yZgBrnJfGa0=",
        version = "v4.2.0+incompatible",
    )
    go_repository(
        name = "com_github_gogo_googleapis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gogo/googleapis",
        sum = "h1:kFkMAZBNAn4j7K0GiZr8cRYzejq68VbheufiV3YuyFI=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_gogo_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gogo/protobuf",
        sum = "h1:Ov1cvc58UF3b5XjBnZv7+opcTcQFZebYjWzi34vdm4Q=",
        version = "v1.3.2",
    )

    go_repository(
        name = "com_github_golang_glog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/glog",
        sum = "h1:/d3pCKDPWNnvIWe0vVUpNP32qc8U3PDVxySP/y360qE=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_golang_groupcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/groupcache",
        sum = "h1:oI5xCqsCo564l8iNU+DwB5epxmsaqB+rhGL0m5jtYqE=",
        version = "v0.0.0-20210331224755-41bb18bfe9da",
    )
    go_repository(
        name = "com_github_golang_jwt_jwt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang-jwt/jwt",
        sum = "h1:73Z+4BJcrTC+KczS6WvTPvRGOp1WmfEP4Q1lOd9Z/+c=",
        version = "v3.2.1+incompatible",
    )
    go_repository(
        name = "com_github_golang_jwt_jwt_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang-jwt/jwt/v4",
        sum = "h1:7cYmW1XlMY7h7ii7UhUyChSgS5wUJEnm9uZVTGqOWzg=",
        version = "v4.5.0",
    )
    go_repository(
        name = "com_github_golang_mock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/mock",
        sum = "h1:ErTB+efbowRARo13NNdxyJji2egdxLGQhRaY+DUumQc=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_golang_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/protobuf",
        sum = "h1:KhyjKVUg7Usr/dYsdSqoFveMYd5ko72D+zANwlG1mmg=",
        version = "v1.5.3",
    )
    go_repository(
        name = "com_github_golang_snappy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/snappy",
        sum = "h1:yAGX7huGHXlcLOEtBnF4w7FQwA26wojNCwOYAEhLjQM=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_golang_sql_civil",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang-sql/civil",
        sum = "h1:lXe2qZdvpiX5WZkZR4hgp4KJVfY3nMkvmwbVkpv1rVY=",
        version = "v0.0.0-20190719163853-cb61b32ac6fe",
    )
    go_repository(
        name = "com_github_gomodule_redigo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gomodule/redigo",
        sum = "h1:H5XSIre1MB5NbPYFp+i1NBbb5qN1W8Y8YAQoAYbkm8k=",
        version = "v1.8.2",
    )
    go_repository(
        name = "com_github_google_btree",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/btree",
        sum = "h1:xf4v41cLI2Z6FxbKm+8Bu+m8ifhj15JuZ9sa0jZCMUU=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_google_cadvisor",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/cadvisor",
        sum = "h1:YyKnRy/3myRNGOvF1bNF9FFnpjY7Gky5yKi/ZlN+BSo=",
        version = "v0.47.1",
    )
    go_repository(
        name = "com_github_google_cel_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/cel-go",
        sum = "h1:kjeKudqV0OygrAqA9fX6J55S8gj+Jre2tckIm5RoG4M=",
        version = "v0.12.6",
    )
    go_repository(
        name = "com_github_google_certificate_transparency_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/certificate-transparency-go",
        sum = "h1:hCyXHDbtqlr/lMXU0D4WgbalXL0Zk4dSWWMbPV8VrqY=",
        version = "v1.1.4",
    )
    go_repository(
        name = "com_github_google_flatbuffers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/flatbuffers",
        sum = "h1:ivUb1cGomAB101ZM1T0nOiWz9pSrTMoa9+EiY7igmkM=",
        version = "v2.0.8+incompatible",
    )

    go_repository(
        name = "com_github_google_gnostic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/gnostic",
        sum = "h1:FhTMOKj2VhjpouxvWJAV1TL304uMlb9zcDqkl6cEI54=",
        version = "v0.5.7-v3refs",
    )
    go_repository(
        name = "com_github_google_go_attestation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-attestation",
        sum = "h1:V3P3BqDEVzAAEw0nlx7OhVRynBUUNgL4znTlqg2Udbc=",
        version = "v0.4.4-0.20221011162210-17f9c05652a9",
    )
    go_repository(
        name = "com_github_google_go_cmdtest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-cmdtest",
        sum = "h1:rcv+Ippz6RAtvaGgKxc+8FQIpxHgsF+HBzPyYL2cyVU=",
        version = "v0.4.1-0.20220921163831-55ab3332a786",
    )
    go_repository(
        name = "com_github_google_go_cmp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-cmp",
        sum = "h1:O2Tfq5qg4qc4AmwVlvv0oLiVAGB7enBSJ2x2DqQFi38=",
        version = "v0.5.9",
    )
    go_repository(
        name = "com_github_google_go_containerregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-containerregistry",
        sum = "h1:gMlTWagRJgCJ3EnISyF5+p9phYpFyWEI70Z56T+o2MY=",
        version = "v0.14.1-0.20230409045903-ed5c185df419",
    )
    go_repository(
        name = "com_github_google_go_github_v28",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-github/v28",
        sum = "h1:kORf5ekX5qwXO2mGzXXOjMe/g6ap8ahVe0sBEulhSxo=",
        version = "v28.1.1",
    )

    go_repository(
        name = "com_github_google_go_licenses",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-licenses",
        sum = "h1:MM+VCXf0slYkpWO0mECvdYDVCxZXIQNal5wqUIXEZ/A=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_google_go_querystring",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-querystring",
        sum = "h1:Xkwi/a1rcvNg1PPYe5vI8GbeBY/jrVuDX5ASuANWTrk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_google_go_replayers_grpcreplay",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-replayers/grpcreplay",
        sum = "h1:eNb1y9rZFmY4ax45uEEECSa8fsxGRU+8Bil52ASAwic=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_go_replayers_httpreplay",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-replayers/httpreplay",
        sum = "h1:H91sIMlt1NZzN7R+/ASswyouLJfW0WLW7fhyUFvDEkY=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_google_go_sev_guest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-sev-guest",
        sum = "h1:NajHkAaLqN9/aW7bCFSUplUMtDgk2+HcN7jC2btFtk0=",
        version = "v0.6.1",
    )
    go_repository(
        name = "com_github_google_go_tpm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-tpm",
        replace = "github.com/thomasten/go-tpm",
        sum = "h1:4/WcA7Q05f5to+jVERKcZjU5NJ89DiRHP3UUZHLfAtw=",
        version = "v0.0.0-20230222180349-bb3cc5560299",
    )
    go_repository(
        name = "com_github_google_go_tpm_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-tpm-tools",
        # keep
        patches = [
            "//3rdparty/bazel/com_github_google_go_tpm_tools:com_github_google_go_tpm_tools.patch",
            "//3rdparty/bazel/com_github_google_go_tpm_tools:ms_tpm_20_ref.patch",
            "//3rdparty/bazel/com_github_google_go_tpm_tools:include.patch",
        ],
        replace = "github.com/daniel-weisse/go-tpm-tools",
        sum = "h1:cJp8NluCBP5idNdtfo2nauPeepn94/jIaeSqTTac9zc=",
        version = "v0.0.0-20230105122812-f7474d459dfc",
    )
    go_repository(
        name = "com_github_google_go_tspi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-tspi",
        sum = "h1:ADtq8RKfP+jrTyIWIZDIYcKOMecRqNJFOew2IT0Inus=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_google_gofuzz",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/gofuzz",
        sum = "h1:xRy4A+RhZaiKjJ1bPfwQ8sedCA+YS2YcCHW6ec7JMi0=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_google_licenseclassifier",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/licenseclassifier",
        sum = "h1:TJsAqW6zLRMDTyGmc9TPosfn9OyVlHs8Hrn3pY6ONSY=",
        version = "v0.0.0-20210722185704-3043a050f148",
    )
    go_repository(
        name = "com_github_google_logger",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/logger",
        sum = "h1:+6Z2geNxc9G+4D4oDO9njjjn2d0wN5d7uOo0vOIW1NQ=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_google_martian",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/martian",
        sum = "h1:xmapqc1AyLoB+ddYT6r04bD9lIjlOqGaREovi0SzFaE=",
        version = "v2.1.1-0.20190517191504-25dcb96d9e51+incompatible",
    )
    go_repository(
        name = "com_github_google_martian_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/martian/v3",
        sum = "h1:IqNFLAmvJOgVlpdEBiQbDc2EwKW77amAycfTuWKdfvw=",
        version = "v3.3.2",
    )
    go_repository(
        name = "com_github_google_pprof",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/pprof",
        sum = "h1:lvddKcYTQ545ADhBujtIJmqQrZBDsGo7XIMbAQe/sNY=",
        version = "v0.0.0-20221103000818-d260c55eee4c",
    )
    go_repository(
        name = "com_github_google_renameio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/renameio",
        sum = "h1:GOZbcHa3HfsPKPlmyPyN2KEohoMXOhdMbHrvbpl2QaA=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_renameio_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/renameio/v2",
        sum = "h1:UifI23ZTGY8Tt29JbYFiuyIU3eX+RNFtUwefq9qAhxg=",
        version = "v2.0.0",
    )

    go_repository(
        name = "com_github_google_rpmpack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/rpmpack",
        sum = "h1:Fv9Ni1vIq9+Gv4Sm0Xq+NnPYcnsMbdNhJ4Cu4rkbPBM=",
        version = "v0.0.0-20210518075352-dc539ef4f2ea",
    )
    go_repository(
        name = "com_github_google_s2a_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/s2a-go",
        sum = "h1:FAgZmpLl/SXurPEZyCMPBIiiYeTbqfjlbdnCNTAkbGE=",
        version = "v0.1.3",
    )

    go_repository(
        name = "com_github_google_shlex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/shlex",
        sum = "h1:El6M4kTTCOh6aBiKaUGG7oYTSPP8MxqL4YI3kZKwcP4=",
        version = "v0.0.0-20191202100458-e7afc7fbc510",
    )
    go_repository(
        name = "com_github_google_subcommands",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/subcommands",
        sum = "h1:/eqq+otEXm5vhfBrbREPCSVQbvofip6kIz+mX5TUH7k=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_google_trillian",
        build_file_generation = "on",
        build_file_name = "",  # keep
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/trillian",
        sum = "h1:roGP6G8aaAch7vP08+oitPkvmZzxjTfIkguozqJ04Ok=",
        version = "v1.5.2",
    )
    go_repository(
        name = "com_github_google_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/uuid",
        sum = "h1:t6JiXgmwXMjEs8VusXIJk2BXHsn+wx8BZdTaoZ5fu7I=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_google_wire",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/wire",
        sum = "h1:I7ELFeVBr3yfPIcc8+MWvrjk+3VjbcSzoXm3JVa+jD8=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_googleapis_enterprise_certificate_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/enterprise-certificate-proxy",
        sum = "h1:yk9/cqRKtT9wXZSsRH9aurXEpJX+U6FLtpYTdC3R06k=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_googleapis_gax_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/gax-go",
        sum = "h1:silFMLAnr330+NRuag/VjIGF7TLp/LBrV2CJKFLWEww=",
        version = "v2.0.2+incompatible",
    )
    go_repository(
        name = "com_github_googleapis_gax_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/gax-go/v2",
        sum = "h1:UBtEZqx1bjXtOQ5BVTkuYghXrr3N4V123VKJK67vJZc=",
        version = "v2.8.0",
    )
    go_repository(
        name = "com_github_googleapis_go_type_adapters",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/go-type-adapters",
        sum = "h1:9XdMn+d/G57qq1s8dNc5IesGCXHf6V2HZ2JwRxfA2tA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_googleapis_google_cloud_go_testing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/google-cloud-go-testing",
        sum = "h1:tlyzajkF3030q6M8SvmJSemC9DTHL/xaMa18b65+JM4=",
        version = "v0.0.0-20200911160855-bcd43fbb19e8",
    )
    go_repository(
        name = "com_github_googlecloudplatform_cloudsql_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/GoogleCloudPlatform/cloudsql-proxy",
        sum = "h1:sTOp2Ajiew5XIH92YSdwhYc+bgpUX5j5TKK/Ac8Saw8=",
        version = "v0.0.0-20191009163259-e802c2cb94ae",
    )
    go_repository(
        name = "com_github_googlecloudplatform_k8s_cloud_provider",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/GoogleCloudPlatform/k8s-cloud-provider",
        sum = "h1:Heo1J/ttaQFgGJSVnCZquy3e5eH5j1nqxBuomztB3P0=",
        version = "v1.18.1-0.20220218231025-f11817397a1b",
    )
    go_repository(
        name = "com_github_gophercloud_gophercloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gophercloud/gophercloud",
        sum = "h1:RUKyCMiZoQR3VlVR5E3K7PK1AC3/qppsWYo6dtBiqs8=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_gophercloud_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gophercloud/utils",
        sum = "h1:6Gvua77nKyAiZQpu0N3AsamGu1L6Mlnhp3tAtDcqwvc=",
        version = "v0.0.0-20230523080330-de873b9cf00d",
    )

    go_repository(
        name = "com_github_gopherjs_gopherjs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gopherjs/gopherjs",
        sum = "h1:EGx4pi6eqNxGaHF6qqu48+N2wcFQ5qg5FXgOdqsJ5d8=",
        version = "v0.0.0-20181017120253-0766667cb4d1",
    )
    go_repository(
        name = "com_github_gordonklaus_ineffassign",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gordonklaus/ineffassign",
        sum = "h1:vc7Dmrk4JwS0ZPS6WZvWlwDflgDTA26jItmbSj83nug=",
        version = "v0.0.0-20200309095847-7953dde2c7bf",
    )
    go_repository(
        name = "com_github_goreleaser_goreleaser",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/goreleaser/goreleaser",
        sum = "h1:3ua5fSYvc/doKd5sVTzvXbyqW6H/eyuvy46woLrUFic=",
        version = "v0.134.0",
    )
    go_repository(
        name = "com_github_goreleaser_nfpm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/goreleaser/nfpm",
        sum = "h1:AEnu9XVmupRDTR930Z2rAs31Mj6sLIPxFcR9ESYvgDA=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_gorilla_context",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gorilla/context",
        sum = "h1:AWwleXJkX/nhcU9bZSnZoi3h/qGYqQAGhq6zZe/aQW8=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_gorilla_handlers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gorilla/handlers",
        sum = "h1:9lRY6j8DEeeBT10CvO9hGW0gmky0BprnvDI5vfhUHH4=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_gorilla_mux",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gorilla/mux",
        sum = "h1:i40aqfkR1h2SlN9hojwV5ZA91wcXFOvkdNIeFDP5koI=",
        version = "v1.8.0",
    )

    go_repository(
        name = "com_github_gorilla_websocket",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gorilla/websocket",
        sum = "h1:+/TMaTYc4QFitKJxsQ7Yye35DkWvkdLcvGKqM+x0Ufc=",
        version = "v1.4.2",
    )
    go_repository(
        name = "com_github_gosuri_uitable",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gosuri/uitable",
        sum = "h1:IG2xLKRvErL3uhY6e1BylFzG+aJiwQviDDTfOKeKTpY=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_gregjones_httpcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gregjones/httpcache",
        sum = "h1:+ngKgrYPPJrOjhax5N+uePQ0Fh1Z7PheYoUI/0nzkPA=",
        version = "v0.0.0-20190611155906-901d90724c79",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_middleware",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/go-grpc-middleware",
        sum = "h1:UH//fgunKIs4JdUbpDl1VZCDaL56wXCB/5+wF6uHfaI=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_prometheus",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
        sum = "h1:Ovs26xHkKqVztRpIrF/92BcuyuQ/YW4NSIpoGtfXNho=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/grpc-gateway",
        sum = "h1:gmcG1KaJ57LophUzW0Hy8NmPhnMZb4M0+kPpLofRdBo=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/grpc-gateway/v2",
        sum = "h1:gDLXvp5S9izjldquuoAhDzccbskOL6tDC5jMSyx3zxE=",
        version = "v2.15.2",
    )
    go_repository(
        name = "com_github_hashicorp_consul_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/consul/api",
        sum = "h1:HXNYlRkkM/t+Y/Yhxtwcy02dlYwIaoxzvxPnS+cqy78=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_hashicorp_consul_sdk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/consul/sdk",
        sum = "h1:UOxjlb4xVNF93jak1mzzoBatyFju9nrkxpVwIp/QqxQ=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_hashicorp_errwrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/errwrap",
        sum = "h1:OxrOeh75EUXMY8TBjag2fzXGZ40LB6IKw45YeGUDY2I=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_checkpoint",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-checkpoint",
        sum = "h1:MFYpPZCnQqQTE18jFwSII6eUQrD/oxMFp3mlgcqk5mU=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_cleanhttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-cleanhttp",
        sum = "h1:035FKYIWjmULyFRBKPs8TBQoi0x6d9G4xc9neXJWAZQ=",
        version = "v0.5.2",
    )

    go_repository(
        name = "com_github_hashicorp_go_hclog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-hclog",
        sum = "h1:vDwF1DFNZhntP4DAjuTpOw3uEgMUpXh1pB5fW9DqHpo=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_immutable_radix",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-immutable-radix",
        sum = "h1:AKDB1HM5PWEA7i4nhcpwOrO2byshxBjXVn/J/3+z5/0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/v2",
        sum = "h1:9Q2lu1YbbmiAgvYZ7Pr31RdlVonUpX+mmDL7Z7qTA2U=",
        version = "v2.0.8",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_wrappers_awskms_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/wrappers/awskms/v2",
        sum = "h1:E3eEWpkofgPNrYyYznfS1+drq4/jFcqHQVNcL7WhUCo=",
        version = "v2.0.7",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_wrappers_azurekeyvault_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/wrappers/azurekeyvault/v2",
        sum = "h1:X27JWuPW6Gmi2l7NMm0pvnp7z7hhtns2TeIOQU93mqI=",
        version = "v2.0.7",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_wrappers_gcpckms_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2",
        sum = "h1:16I8OqBEuxZIowwn3jiLvhlx+z+ia4dJc9stvz0yUBU=",
        version = "v2.0.8",
    )

    go_repository(
        name = "com_github_hashicorp_go_msgpack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-msgpack",
        sum = "h1:zKjpN5BK/P5lMYrLmBHdBULWbJ0XpYR+7NGzqkZzoD4=",
        version = "v0.5.3",
    )
    go_repository(
        name = "com_github_hashicorp_go_multierror",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-multierror",
        sum = "h1:H5DkEtf6CXdFp0N0Em5UCwQpXMWke8IA0+lD48awMYo=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_net",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go.net",
        sum = "h1:sNCoNyDEvN1xa+X0baata4RdcpKwcMS6DH+xwfqPgjw=",
        version = "v0.0.1",
    )

    go_repository(
        name = "com_github_hashicorp_go_retryablehttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-retryablehttp",
        sum = "h1:AcYqCvkpalPnPF2pn0KamgwamS42TqUDDYFRKq/RAd0=",
        version = "v0.7.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_rootcerts",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-rootcerts",
        sum = "h1:jzhAVGtqPKbwpyCPELlgNWhE1znq+qwJtW5Oi2viEzc=",
        version = "v1.0.2",
    )

    go_repository(
        name = "com_github_hashicorp_go_secure_stdlib_awsutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-secure-stdlib/awsutil",
        sum = "h1:W9WN8p6moV1fjKLkeqEgkAMu5rauy9QeYDAmIaPuuiA=",
        version = "v0.1.6",
    )

    go_repository(
        name = "com_github_hashicorp_go_secure_stdlib_parseutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-secure-stdlib/parseutil",
        sum = "h1:UpiO20jno/eV1eVZcxqWnUohyKRe1g8FPV/xH1s/2qs=",
        version = "v0.1.7",
    )
    go_repository(
        name = "com_github_hashicorp_go_secure_stdlib_strutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-secure-stdlib/strutil",
        sum = "h1:kes8mmyCpxJsI7FTwtzRqEy9CdjCtrXrXGuOpxEA7Ts=",
        version = "v0.1.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_sockaddr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-sockaddr",
        sum = "h1:ztczhD1jLxIRjVejw8gFomI1BQZOe2WoVOu0SyteCQc=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_syslog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-syslog",
        sum = "h1:KaodqZuhUoZereWVIYmpUgZysurB1kBLX2j0MwMrUAE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-uuid",
        sum = "h1:2gKiV6YVmrJ1i2CKKa9obLvRieoRGviZFL26PcT/Co8=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_hashicorp_go_version",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-version",
        sum = "h1:feTTfFNnjP967rlCxM/I9g701jU+RN74YKx2mOkIeek=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_hashicorp_golang_lru",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/golang-lru",
        sum = "h1:YDjusn29QI/Das2iO9M0BHnIbxPeyuCHsjMW+lJfyTc=",
        version = "v0.5.4",
    )
    go_repository(
        name = "com_github_hashicorp_hc_install",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/hc-install",
        sum = "h1:SfwMFnEXVVirpwkDuSF5kymUOhrUxrTq3udEseZdOD0=",
        version = "v0.5.2",
    )
    go_repository(
        name = "com_github_hashicorp_hcl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/hcl",
        sum = "h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_hashicorp_logutils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/logutils",
        sum = "h1:dLEQVugN8vlakKOUE3ihGLTZJRB4j+M2cdTm/ORI65Y=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_mdns",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/mdns",
        sum = "h1:WhIgCr5a7AaVH6jPUwjtRuuE7/RDufnUvzIr48smyxs=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_memberlist",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/memberlist",
        sum = "h1:EmmoJme1matNzb+hMpDuR/0sbJSUisxyqBGG676r31M=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_hashicorp_serf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/serf",
        sum = "h1:YZ7UKsJv+hKjqGVUUbtE3HNj79Eln2oQ75tniF6iPt0=",
        version = "v0.8.2",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_exec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-exec",
        sum = "h1:LAbfDvNQU1l0NOQlTuudjczVhHj061fNX5H8XZxHlH4=",
        version = "v0.18.1",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_json",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-json",
        sum = "h1:/gIyNtR6SFw6h5yzlbDbACyGvIhKtQi8mTsbkNd79lE=",
        version = "v0.15.0",
    )
    go_repository(
        name = "com_github_hashicorp_vault_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/vault/api",
        sum = "h1:LtY/I16+5jVGU8rufyyAkwopgq/HpUnxFBg+QLOAV38=",
        version = "v1.9.1",
    )
    go_repository(
        name = "com_github_hexops_gotextdiff",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hexops/gotextdiff",
        sum = "h1:gitA9+qJrrTCsiCl7+kh75nPqQt1cx4ZkudSTLoUqJM=",
        version = "v1.0.3",
    )

    go_repository(
        name = "com_github_honeycombio_beeline_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/honeycombio/beeline-go",
        sum = "h1:cUDe555oqvw8oD76BQJ8alk7FP0JZ/M/zXpNvOEDLDc=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_honeycombio_libhoney_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/honeycombio/libhoney-go",
        sum = "h1:kPpqoz6vbOzgp7jC6SR7SkNj7rua7rgxvznI6M3KdHc=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_github_howeyc_gopass",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/howeyc/gopass",
        sum = "h1:aY2hhxLhjEAbfXOx2nRJxCXezC6CO2V/yN+OCr1srtk=",
        version = "v0.0.0-20190910152052-7cb4b85ec19c",
    )
    go_repository(
        name = "com_github_hpcloud_tail",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hpcloud/tail",
        sum = "h1:nfCOvKYfkgYP8hkirhJocXT2+zOD8yUNjXaWfTlyFKI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_huandu_xstrings",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/huandu/xstrings",
        sum = "h1:D17IlohoQq4UcpqD7fDk80P7l+lwAmlFaBHgOipl2FU=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_hudl_fargo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hudl/fargo",
        sum = "h1:0U6+BtN6LhaYuTnIJq4Wyq5cpn6O2kWrxAtcqBmYY6w=",
        version = "v1.3.0",
    )

    go_repository(
        name = "com_github_iancoleman_strcase",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/iancoleman/strcase",
        sum = "h1:05I4QRnGpI0m37iZQRuskXh+w77mr6Z41lwQzuHLwW0=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_ianlancetaylor_demangle",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ianlancetaylor/demangle",
        sum = "h1:rcanfLhLDA8nozr/K289V1zcntHr3V+SHlXwzz1ZI2g=",
        version = "v0.0.0-20220319035150-800ac71e25c2",
    )
    go_repository(
        name = "com_github_imdario_mergo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/imdario/mergo",
        sum = "h1:lFzP57bqS/wsqKssCGmtLAb8A0wKjLGrve2q3PPVcBk=",
        version = "v0.3.13",
    )

    go_repository(
        name = "com_github_in_toto_in_toto_golang",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/in-toto/in-toto-golang",
        sum = "h1:tHny7ac4KgtsfrG6ybU8gVOZux2H8jN05AXJ9EBM1XU=",
        version = "v0.9.0",
    )

    go_repository(
        name = "com_github_inconshreveable_mousetrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/inconshreveable/mousetrap",
        sum = "h1:wN+x4NVGpMsO7ErUn/mUI3vEoE6Jt13X2s0bqwp9tc8=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_influxdata_influxdb1_client",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/influxdata/influxdb1-client",
        sum = "h1:/WZQPMZNsjZ7IlCpsLGdQBINg5bxKQ1K1sh6awxLtkA=",
        version = "v0.0.0-20191209144304-8bf82d3c094d",
    )

    go_repository(
        name = "com_github_intel_goresctrl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/intel/goresctrl",
        sum = "h1:K2D3GOzihV7xSBedGxONSlaw/un1LZgWsc9IfqipN4c=",
        version = "v0.3.0",
    )

    go_repository(
        name = "com_github_ishidawataru_sctp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ishidawataru/sctp",
        sum = "h1:qPmlgoeRS18y2dT+iAH5vEKZgIqgiPi2Y8UCu/b7Aq8=",
        version = "v0.0.0-20190723014705-7c296d48a2b5",
    )

    go_repository(
        name = "com_github_jackc_chunkreader_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/chunkreader/v2",
        sum = "h1:i+RDz65UE+mmpjTfyz0MoVTnzeYxroil2G82ki7MGG8=",
        version = "v2.0.1",
    )
    go_repository(
        name = "com_github_jackc_pgconn",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/pgconn",
        sum = "h1:rsDFzIpRk7xT4B8FufgpCCeyjdNpKyghZeSefViE5W8=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_github_jackc_pgio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/pgio",
        sum = "h1:g12B9UwVnzGhueNavwioyEEpAmqMe1E/BN9ES+8ovkE=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_jackc_pgpassfile",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/pgpassfile",
        sum = "h1:/6Hmqy13Ss2zCq62VdNG8tM1wchn8zjSGOBJ6icpsIM=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_jackc_pgproto3_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/pgproto3/v2",
        sum = "h1:brH0pCGBDkBW07HWlN/oSBXrmo3WB0UvZd1pIuDcL8Y=",
        version = "v2.3.0",
    )
    go_repository(
        name = "com_github_jackc_pgservicefile",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/pgservicefile",
        sum = "h1:C8S2+VttkHFdOOCXJe+YGfa4vHYwlt4Zx+IVXQ97jYg=",
        version = "v0.0.0-20200714003250-2b9c44734f2b",
    )
    go_repository(
        name = "com_github_jackc_pgtype",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/pgtype",
        sum = "h1:u4uiGPz/1hryuXzyaBhSk6dnIyyG2683olG2OV+UUgs=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_github_jackc_pgx_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jackc/pgx/v4",
        sum = "h1:JzTglcal01DrghUqt+PmzWsZx/Yh7SC/CTQmSBMTd0Y=",
        version = "v4.16.1",
    )

    go_repository(
        name = "com_github_jarcoal_httpmock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jarcoal/httpmock",
        sum = "h1:cHtVEcTxRSX4J0je7mWPfc9BpDpqzXSJ5HbymZmyHck=",
        version = "v1.0.5",
    )
    go_repository(
        name = "com_github_jbenet_go_context",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jbenet/go-context",
        sum = "h1:BQSFePA1RWJOlocH6Fxy8MmwDt+yVQYULKfN0RoTN8A=",
        version = "v0.0.0-20150711004518-d14ea06fba99",
    )
    go_repository(
        name = "com_github_jedisct1_go_minisign",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jedisct1/go-minisign",
        sum = "h1:ZGiXF8sz7PDk6RgkP+A/SFfUD0ZR/AgG6SpRNEDKZy8=",
        version = "v0.0.0-20211028175153-1c139d1cc84b",
    )
    go_repository(
        name = "com_github_jeffashton_win_pdh",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/JeffAshton/win_pdh",
        sum = "h1:UKkYhof1njT1/xq4SEg5z+VpTgjmNeHwPGRQl7takDI=",
        version = "v0.0.0-20161109143554-76bb4ee9f0ab",
    )

    go_repository(
        name = "com_github_jellydator_ttlcache_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jellydator/ttlcache/v3",
        sum = "h1:cHgCSMS7TdQcoprXnWUptJZzyFsqs18Lt8VVhRuZYVU=",
        version = "v3.0.1",
    )

    go_repository(
        name = "com_github_jessevdk_go_flags",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jessevdk/go-flags",
        sum = "h1:1jKYvbxEjfUl0fmqTCOfonvskHHXMjBySTLW4y9LFvc=",
        version = "v1.5.0",
    )

    go_repository(
        name = "com_github_jhump_protoreflect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jhump/protoreflect",
        sum = "h1:1NQ4FpWMgn3by/n1X0fbeKEUxP1wBt7+Oitpv01HR10=",
        version = "v1.12.0",
    )

    go_repository(
        name = "com_github_jmespath_go_jmespath",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jmespath/go-jmespath",
        sum = "h1:BEgLn5cpjn8UN1mAw4NjwDrS35OdebyEtFe+9YPoQUg=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_jmespath_go_jmespath_internal_testify",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jmespath/go-jmespath/internal/testify",
        sum = "h1:shLQSRRSCCPj3f2gpwzGwWFoC7ycTf1rcQZHOlsJ6N8=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_jmhodges_clock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jmhodges/clock",
        sum = "h1:dYTbLf4m0a5u0KLmPfB6mgxbcV7588bOCx79hxa5Sr4=",
        version = "v0.0.0-20160418191101-880ee4c33548",
    )
    go_repository(
        name = "com_github_jmoiron_sqlx",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jmoiron/sqlx",
        sum = "h1:vFFPA71p1o5gAeqtEAwLU4dnX2napprKtHr7PYIcN3g=",
        version = "v1.3.5",
    )
    go_repository(
        name = "com_github_joho_godotenv",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/joho/godotenv",
        sum = "h1:Zjp+RcGpHhGlrMbJzXTrZZPrWj+1vfm90La1wgB6Bhc=",
        version = "v1.3.0",
    )

    go_repository(
        name = "com_github_jonboulle_clockwork",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jonboulle/clockwork",
        sum = "h1:9BSCMi8C+0qdApAp4auwX0RkLGUjs956h0EkuQymUhg=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_josharian_intern",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/josharian/intern",
        sum = "h1:vlS4z54oSdjm0bgjRigI+G1HpF+tI+9rE5LLzOg8HmY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_josharian_native",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/josharian/native",
        sum = "h1:Ts/E8zCSEsG17dUqv7joXJFybuMLjQfWE04tsBODTxk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jpillora_backoff",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jpillora/backoff",
        sum = "h1:uvFg412JmmHBHw7iwprIxkPMI+sGQ4kzOWsMeHnm2EA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jsimonetti_rtnetlink",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jsimonetti/rtnetlink",
        sum = "h1:Bl3VxrWwi3eNj2pFuG2x3xcIArSAvHf9paz1OXiDT9A=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_json_iterator_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/json-iterator/go",
        sum = "h1:PV8peI4a0ysnczrg+LtxykD8LfKY9ML6u2jnxaEnrnM=",
        version = "v1.1.12",
    )
    go_repository(
        name = "com_github_jstemmer_go_junit_report",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jstemmer/go-junit-report",
        sum = "h1:6QPYqodiu3GuPL+7mfx+NwDdp2eTkp9IfEUpgAwUN0o=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_jtolds_gls",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jtolds/gls",
        sum = "h1:xdiiI2gbIgH/gLH7ADydsJ1uDOEzR8yvV7C0MuV77Wo=",
        version = "v4.20.0+incompatible",
    )

    go_repository(
        name = "com_github_juju_ratelimit",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/juju/ratelimit",
        sum = "h1:+7AIFJVQ0EQgq/K9+0Krm7m530Du7tIz0METWzN0RgY=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_julienschmidt_httprouter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/julienschmidt/httprouter",
        sum = "h1:U0609e9tgbseu3rBINet9P48AI/D3oJs4dN7jwJOQ1U=",
        version = "v1.3.0",
    )

    go_repository(
        name = "com_github_k0kubun_go_ansi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/k0kubun/go-ansi",
        sum = "h1:qGQQKEcAR99REcMpsXCp3lJ03zYT1PkRd3kQGPn9GVg=",
        version = "v0.0.0-20180517002512-3bf9e2903213",
    )
    go_repository(
        name = "com_github_karrick_godirwalk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/karrick/godirwalk",
        sum = "h1:b4kY7nqDdioR/6qnbHQyDvmA17u5G1cZ6J+CZXwSWoI=",
        version = "v1.17.0",
    )

    go_repository(
        name = "com_github_katexochen_sh_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/katexochen/sh/v3",
        sum = "h1:eIG0cBgpy2u2iQDFgi4fY0Yq38kVBzMSetzuU+ZIGRg=",
        version = "v3.6.0",
    )

    go_repository(
        name = "com_github_kevinburke_ssh_config",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kevinburke/ssh_config",
        sum = "h1:x584FjTGwHzMwvHx18PXxbBVzfnxogHaAReU4gf13a4=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_kisielk_errcheck",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kisielk/errcheck",
        sum = "h1:e8esj/e4R+SAOwFwN+n3zr0nYeCyeweozKfO23MvHzY=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_kisielk_gotool",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kisielk/gotool",
        sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_klauspost_asmfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/klauspost/asmfmt",
        sum = "h1:4Ri7ox3EwapiOjCki+hw14RyKk201CN4rzyCJRFLpK4=",
        version = "v1.3.2",
    )

    go_repository(
        name = "com_github_klauspost_compress",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/klauspost/compress",
        sum = "h1:iULayQNOReoYUe+1qtKOqw9CwJv3aNQu8ivo7lw1HU4=",
        version = "v1.16.0",
    )

    go_repository(
        name = "com_github_klauspost_cpuid_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/klauspost/cpuid/v2",
        sum = "h1:lgaqFMSdTdQYdZ04uHyN2d/eKdOMyi2YLSvlQIBFYa4=",
        version = "v2.0.9",
    )

    go_repository(
        name = "com_github_knetic_govaluate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Knetic/govaluate",
        sum = "h1:1G1pk05UrOh0NlF1oeaaix1x8XzrfjIDK47TY0Zehcw=",
        version = "v3.0.1-0.20171022003610-9aa49832a739+incompatible",
    )

    go_repository(
        name = "com_github_konsorten_go_windows_terminal_sequences",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:CE8S1cTafDpPvMhIxNJKvHsGVBgn1xWYf1NbHQhywc8=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_kortschak_utter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kortschak/utter",
        sum = "h1:AJVccwLrdrikvkH0aI5JKlzZIORLpfMeGBQ5tHfIXis=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_kr_fs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/fs",
        sum = "h1:Jskdu9ieNAYnjxsi0LbQp1ulIKZV1LAFgK1tWhpZgl8=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_kr_logfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/logfmt",
        sum = "h1:T+h1c/A9Gawja4Y9mFVWj2vyii2bbUNDw3kt9VxK2EY=",
        version = "v0.0.0-20140226030751-b84e30acd515",
    )
    go_repository(
        name = "com_github_kr_pretty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/pretty",
        sum = "h1:flRD4NNwYAUpkphVc1HcthR4KEIFJ65n8Mw5qdRn3LE=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_kr_pty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/pty",
        sum = "h1:AkaSdXYQOWeaO3neb8EM634ahkXXe3jYbVh/F9lq+GI=",
        version = "v1.1.8",
    )
    go_repository(
        name = "com_github_kr_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/text",
        sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_kylelemons_godebug",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kylelemons/godebug",
        sum = "h1:RPNrshWIDI6G2gRW9EHilWtl7Z6Sb1BR0xunSBf0SNc=",
        version = "v1.1.0",
    )

    go_repository(
        name = "com_github_lann_builder",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lann/builder",
        sum = "h1:SOEGU9fKiNWd/HOJuq6+3iTQz8KNCLtVX6idSoTLdUw=",
        version = "v0.0.0-20180802200727-47ae307949d0",
    )
    go_repository(
        name = "com_github_lann_ps",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lann/ps",
        sum = "h1:P6pPBnrTSX3DEVR4fDembhRWSsG5rVo6hYhAB/ADZrk=",
        version = "v0.0.0-20150810152359-62de8c46ede0",
    )
    go_repository(
        name = "com_github_leodido_go_urn",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/leodido/go-urn",
        sum = "h1:XlAE/cm/ms7TE/VMVoduSpNBoyc2dOxHs5MZSwAN63Q=",
        version = "v1.2.4",
    )
    go_repository(
        name = "com_github_letsencrypt_boulder",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/letsencrypt/boulder",
        sum = "h1:ndns1qx/5dL43g16EQkPV/i8+b3l5bYQwLeoSBe7tS8=",
        version = "v0.0.0-20221109233200-85aa52084eaf",
    )
    go_repository(
        name = "com_github_letsencrypt_challtestsrv",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/letsencrypt/challtestsrv",
        sum = "h1:Lzv4jM+wSgVMCeO5a/F/IzSanhClstFMnX6SfrAJXjI=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_letsencrypt_pkcs11key_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/letsencrypt/pkcs11key/v4",
        sum = "h1:qLc/OznH7xMr5ARJgkZCCWk+EomQkiNTOoOF5LAgagc=",
        version = "v4.0.0",
    )

    go_repository(
        name = "com_github_lib_pq",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lib/pq",
        sum = "h1:YXG7RB+JIjhP29X+OtkiDnYaXQwpS4JEWq7dtCCRUEw=",
        version = "v1.10.9",
    )
    go_repository(
        name = "com_github_libopenstorage_openstorage",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/libopenstorage/openstorage",
        sum = "h1:GLPam7/0mpdP8ZZtKjbfcXJBTIA/T1O6CBErVEFEyIM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_liggitt_tabwriter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/liggitt/tabwriter",
        sum = "h1:9TO3cAIGXtEhnIaL+V+BEER86oLrvS+kWobKpbJuye0=",
        version = "v0.0.0-20181228230101-89fcab3d43de",
    )
    go_repository(
        name = "com_github_lightstep_lightstep_tracer_common_golang_gogo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lightstep/lightstep-tracer-common/golang/gogo",
        sum = "h1:143Bb8f8DuGWck/xpNUOckBVYfFbBTnLevfRZ1aVVqo=",
        version = "v0.0.0-20190605223551-bc2310a04743",
    )
    go_repository(
        name = "com_github_lightstep_lightstep_tracer_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lightstep/lightstep-tracer-go",
        sum = "h1:vi1F1IQ8N7hNWytK9DpJsUfQhGuNSc19z330K6vl4zk=",
        version = "v0.18.1",
    )

    go_repository(
        name = "com_github_lithammer_dedent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lithammer/dedent",
        sum = "h1:VNzHMVCBNG1j0fh3OrsFRkVUwStdDArbgBWoPAffktY=",
        version = "v1.1.0",
    )

    go_repository(
        name = "com_github_lyft_protoc_gen_star",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lyft/protoc-gen-star",
        sum = "h1:erE0rdztuaDq3bpGifD95wfoPrSZc95nGA6tbiNYh6M=",
        version = "v0.6.1",
    )
    go_repository(
        name = "com_github_lyft_protoc_gen_star_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lyft/protoc-gen-star/v2",
        sum = "h1:keaAo8hRuAT0O3DfJ/wM3rufbAjGeJ1lAtWZHDjKGB0=",
        version = "v2.0.1",
    )

    go_repository(
        name = "com_github_lyft_protoc_gen_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lyft/protoc-gen-validate",
        sum = "h1:KNt/RhmQTOLr7Aj8PsJ7mTronaFyx80mRTT9qF261dA=",
        version = "v0.0.13",
    )

    go_repository(
        name = "com_github_magiconair_properties",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/magiconair/properties",
        sum = "h1:IeQXZAiQcpL9mgcAe1Nu6cX9LLw6ExEHKjN0VQdvPDY=",
        version = "v1.8.7",
    )

    go_repository(
        name = "com_github_mailru_easyjson",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mailru/easyjson",
        sum = "h1:UGYAvKxe3sBsEDzO8ZeWOSlIQfWFlxbzLZe7hwFURr0=",
        version = "v0.7.7",
    )
    go_repository(
        name = "com_github_makenowjust_heredoc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/MakeNowJust/heredoc",
        sum = "h1:cXCdzVdstXyiTqTvfqk9SDHpKNjxuom+DOlyEeQ4pzQ=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_markbates_errx",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/markbates/errx",
        sum = "h1:QDFeR+UP95dO12JgW+tgi2UVfo0V8YBHiUIOaeBPiEI=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_markbates_oncer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/markbates/oncer",
        sum = "h1:E83IaVAHygyndzPimgUYJjbshhDTALZyXxvk9FOlQRY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_markbates_safe",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/markbates/safe",
        sum = "h1:yjZkbvRM6IzKj9tlu/zMJLS0n/V351OZWRnF3QfaUxI=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_martinjungblut_go_cryptsetup",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/martinjungblut/go-cryptsetup",
        patches = [
            "//3rdparty/bazel/com_github_martinjungblut_go_cryptsetup:com_github_martinjungblut_go_cryptsetup.patch",  # keep
        ],
        sum = "h1:YDjLk3wsL5ZLhLC4TIwIvT2NkSCAdAV6pzzZaRfj4jk=",
        version = "v0.0.0-20220520180014-fd0874fd07a6",
    )
    go_repository(
        name = "com_github_masterminds_goutils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/goutils",
        sum = "h1:5nUrii3FMTL5diU80unEVvNevw1nH4+ZV4DSLVJLSYI=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_masterminds_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/semver",
        sum = "h1:H65muMkzWKEuNDnfl9d70GUjFniHKHRbFPGBuZ3QEww=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_masterminds_semver_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/semver/v3",
        sum = "h1:3MEsd0SM6jqZojhjLWWeBY+Kcjy9i6MQAeY7YgDP83g=",
        version = "v3.2.0",
    )
    go_repository(
        name = "com_github_masterminds_sprig",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/sprig",
        sum = "h1:z4yfnGrZ7netVz+0EDJ0Wi+5VZCSYp4Z0m2dk6cEM60=",
        version = "v2.22.0+incompatible",
    )
    go_repository(
        name = "com_github_masterminds_sprig_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/sprig/v3",
        sum = "h1:eL2fZNezLomi0uOLqjQoN6BfsDD+fyLtgbJMAj9n6YA=",
        version = "v3.2.3",
    )
    go_repository(
        name = "com_github_masterminds_squirrel",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/squirrel",
        sum = "h1:YPpoceAcxuzIljlr5iWpNKaql7hLeG1KLSrhvdHpkZc=",
        version = "v1.5.3",
    )
    go_repository(
        name = "com_github_masterminds_vcs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/vcs",
        sum = "h1:IIA2aBdXvfbIM+yl/eTnL4hb1XwdpvuQLglAix1gweE=",
        version = "v1.13.3",
    )
    go_repository(
        name = "com_github_matryer_is",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/matryer/is",
        sum = "h1:92UTHpy8CDwaJ08GqLDzhhuixiBUUD1p3AU6PHddz4A=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_mattn_go_colorable",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-colorable",
        sum = "h1:fFA4WZxdEF4tXPZVKMLwD8oUnCTTo08duU7wxecdEvA=",
        version = "v0.1.13",
    )
    go_repository(
        name = "com_github_mattn_go_ieproxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-ieproxy",
        sum = "h1:HfxbT6/JcvIljmERptWhwa8XzP7H3T+Z2N26gTsaDaA=",
        version = "v0.0.0-20190610004146-91bb50d98149",
    )
    go_repository(
        name = "com_github_mattn_go_isatty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-isatty",
        sum = "h1:JITubQf0MOLdlGRuRq+jtsDlekdYPia9ZFsB8h/APPA=",
        version = "v0.0.19",
    )
    go_repository(
        name = "com_github_mattn_go_oci8",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-oci8",
        sum = "h1:aEUDxNAyDG0tv8CA3TArnDQNyc4EhnWlsfxRgDHABHM=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_mattn_go_runewidth",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-runewidth",
        sum = "h1:+xnbZSEeDbOIg5/mE6JF0w6n9duR1l3/WmbinWVwUuU=",
        version = "v0.0.14",
    )
    go_repository(
        name = "com_github_mattn_go_shellwords",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-shellwords",
        sum = "h1:M2zGm7EW6UQJvDeQxo4T51eKPurbeFbe8WtebGE2xrk=",
        version = "v1.0.12",
    )
    go_repository(
        name = "com_github_mattn_go_sqlite3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-sqlite3",
        sum = "h1:vfoHhTN1af61xCRSWzFIWzx2YskyMTwHLrExkBOjvxI=",
        version = "v1.14.15",
    )
    go_repository(
        name = "com_github_mattn_go_zglob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-zglob",
        sum = "h1:xsEx/XUoVlI6yXjqBK062zYhRTZltCNmYPx6v+8DNaY=",
        version = "v0.0.1",
    )

    go_repository(
        name = "com_github_matttproud_golang_protobuf_extensions",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/matttproud/golang_protobuf_extensions",
        sum = "h1:mmDVorXM7PCGKw94cs5zkfA9PSy5pEvNWRP0ET0TIVo=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_mdlayher_ethtool",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/ethtool",
        sum = "h1:Y7LoKqIgD7vmqJ7+6ZVnADuwUO+m3tGXbf2lK0OvjIw=",
        version = "v0.0.0-20221212131811-ba3b4bc2e02c",
    )
    go_repository(
        name = "com_github_mdlayher_genetlink",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/genetlink",
        sum = "h1:roBiPnual+eqtRkKX2Jb8UQN5ZPWnhDCGj/wR6Jlz2w=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_mdlayher_netlink",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/netlink",
        sum = "h1:FdUaT/e33HjEXagwELR8R3/KL1Fq5x3G5jgHLp/BTmg=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_github_mdlayher_socket",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/socket",
        sum = "h1:280wsy40IC9M9q1uPGcLBwXpcTQDtoGwVt+BNoITxIw=",
        version = "v0.4.0",
    )

    go_repository(
        name = "com_github_mgutz_ansi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mgutz/ansi",
        sum = "h1:j7+1HpAFS1zy5+Q4qx1fWh90gTKwiN4QCGoY9TWyyO4=",
        version = "v0.0.0-20170206155736-9520e82c474b",
    )

    go_repository(
        name = "com_github_microsoft_applicationinsights_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/microsoft/ApplicationInsights-Go",
        sum = "h1:G4+H9WNs6ygSCe6sUyxRc2U81TI5Es90b2t/MwX5KqY=",
        version = "v0.4.4",
    )
    go_repository(
        name = "com_github_microsoft_go_winio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Microsoft/go-winio",
        sum = "h1:slsWYD/zyx7lCXoZVlvQrj0hPTM1HI4+v1sIda2yDvg=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_microsoft_hcsshim",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Microsoft/hcsshim",
        sum = "h1:HBytQPxcv8Oy4244zbQbe6hnOnx544eL5QPUqhJldz8=",
        version = "v0.10.0-rc.7",
    )
    go_repository(
        name = "com_github_miekg_dns",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/miekg/dns",
        sum = "h1:DQUfb9uc6smULcREF09Uc+/Gd46YWqJd5DbpPE9xkcA=",
        version = "v1.1.50",
    )
    go_repository(
        name = "com_github_miekg_pkcs11",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/miekg/pkcs11",
        sum = "h1:Ugu9pdy6vAYku5DEpVWVFPYnzV+bxB+iRdbuFSu7TvU=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_minio_asm2plan9s",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/minio/asm2plan9s",
        sum = "h1:AMFGa4R4MiIpspGNG7Z948v4n35fFGB3RR3G/ry4FWs=",
        version = "v0.0.0-20200509001527-cdd76441f9d8",
    )
    go_repository(
        name = "com_github_minio_c2goasm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/minio/c2goasm",
        sum = "h1:+n/aFZefKZp7spd8DFdX7uMikMLXX4oubIzJF4kv/wI=",
        version = "v0.0.0-20190812172519-36a3d3bbc4f3",
    )

    go_repository(
        name = "com_github_minio_sha256_simd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/minio/sha256-simd",
        sum = "h1:v1ta+49hkWZyvaKwrQB8elexRqm6Y0aMLjCNsrYxo6g=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_mistifyio_go_zfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mistifyio/go-zfs",
        sum = "h1:aKW/4cBs+yK6gpqU3K/oIwk9Q/XICqd3zOX/UFuvqmk=",
        version = "v2.1.2-0.20190413222219-f784269be439+incompatible",
    )
    go_repository(
        name = "com_github_mitchellh_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/cli",
        sum = "h1:OxRIeJXpAMztws/XHlN2vu6imG5Dpq+j61AzAX5fLng=",
        version = "v1.1.5",
    )
    go_repository(
        name = "com_github_mitchellh_colorstring",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/colorstring",
        sum = "h1:62I3jR2EmQ4l5rM/4FEfDWcRD+abF5XlKShorW5LRoQ=",
        version = "v0.0.0-20190213212951-d06e56a500db",
    )
    go_repository(
        name = "com_github_mitchellh_copystructure",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/copystructure",
        sum = "h1:vpKXTN4ewci03Vljg/q9QvCGUDttBOGBIa15WveJJGw=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_homedir",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/go-homedir",
        sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_testing_interface",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/go-testing-interface",
        sum = "h1:fzU/JVNcaqHQEcVFAKeR41fkiLdIPrefOvVG1VZ96U0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_wordwrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/go-wordwrap",
        sum = "h1:TLuKupo69TCn6TQSyGxwI1EblZZEsQ0vMlAFQflz0v0=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_mitchellh_gox",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/gox",
        sum = "h1:lfGJxY7ToLJQjHHwi0EX6uYBdK78egf954SQl13PQJc=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_mitchellh_iochan",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/iochan",
        sum = "h1:C+X3KsSTLFVBr/tK1eYN/vs4rJcvsiLU338UhYPJWeY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_mapstructure",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/mapstructure",
        sum = "h1:jeMsZIYE/09sWLaz43PL7Gy6RuMjD2eJVyuac5Z2hdY=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_mitchellh_reflectwalk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/reflectwalk",
        sum = "h1:G2LzWKi524PWgd3mLHV8Y5k7s6XUvT0Gef6zxSIeXaQ=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_mmcloughlin_avo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mmcloughlin/avo",
        sum = "h1:nAco9/aI9Lg2kiuROBY6BhCI/z0t5jEvJfjWbL8qXLU=",
        version = "v0.5.0",
    )

    go_repository(
        name = "com_github_moby_ipvs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/ipvs",
        sum = "h1:ONN4pGaZQgAx+1Scz5RvWV4Q7Gb+mvfRh3NsPS+1XQQ=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_moby_locker",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/locker",
        sum = "h1:fOXqR41zeveg4fFODix+1Ch4mj/gT0NE1XJbp/epuBg=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_moby_spdystream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/spdystream",
        sum = "h1:cjW1zVyyoiM0T7b6UoySUFqzXMoqRckQtXwGPiBhOM8=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_moby_sys_mountinfo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/sys/mountinfo",
        sum = "h1:BzJjoreD5BMFNmD9Rus6gdd1pLuecOFPt8wC+Vygl78=",
        version = "v0.6.2",
    )
    go_repository(
        name = "com_github_moby_sys_sequential",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/sys/sequential",
        sum = "h1:OPvI35Lzn9K04PBbCLW0g4LcFAJgHsvXsRyewg5lXtc=",
        version = "v0.5.0",
    )

    go_repository(
        name = "com_github_moby_sys_signal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/sys/signal",
        sum = "h1:25RW3d5TnQEoKvRbEKUGay6DCQ46IxAVTT9CUMgmsSI=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_moby_sys_symlink",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/sys/symlink",
        sum = "h1:tk1rOM+Ljp0nFmfOIBtlV3rTDlWOwFRhjEeAhZB0nZc=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_moby_term",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/term",
        sum = "h1:HfkjXDfhgVaN5rmueG8cL8KKeFNecRCXFhaJ2qZ5SKA=",
        version = "v0.0.0-20221205130635-1aeaba878587",
    )
    go_repository(
        name = "com_github_modern_go_concurrent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/modern-go/concurrent",
        sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
        version = "v0.0.0-20180306012644-bacd9c7ef1dd",
    )
    go_repository(
        name = "com_github_modern_go_reflect2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/modern-go/reflect2",
        sum = "h1:xBagoLtFs94CBntxluKeaWgTMpvLxC4ur3nMaC9Gz0M=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_modocache_gover",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/modocache/gover",
        sum = "h1:8Q0qkMVC/MmWkpIdlvZgcv2o2jrlF6zqVOh7W5YHdMA=",
        version = "v0.0.0-20171022184752-b58185e213c5",
    )
    go_repository(
        name = "com_github_mohae_deepcopy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mohae/deepcopy",
        sum = "h1:RWengNIwukTxcDr9M+97sNutRR1RKhG96O6jWumTTnw=",
        version = "v0.0.0-20170929034955-c48cc78d4826",
    )
    go_repository(
        name = "com_github_monochromegane_go_gitignore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/monochromegane/go-gitignore",
        sum = "h1:n6/2gBQ3RWajuToeY6ZtZTIKv2v7ThUy5KKusIT0yc0=",
        version = "v0.0.0-20200626010858-205db1a8cc00",
    )
    go_repository(
        name = "com_github_montanaflynn_stats",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/montanaflynn/stats",
        sum = "h1:r3y12KyNxj/Sb/iOE46ws+3mS1+MZca1wlHQFPsY/JU=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_morikuni_aec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/morikuni/aec",
        sum = "h1:nP9CBfwrvYnBRgY6qfDQkygYDmYwOilePFkwzv4dU8A=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_mr_tron_base58",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mr-tron/base58",
        sum = "h1:T/HDJBh4ZCPbU39/+c3rRvE0uKBQlU27+QI8LJ4t64o=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_mrunalp_fileutils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mrunalp/fileutils",
        sum = "h1:NKzVxiH7eSk+OQ4M+ZYW1K6h27RUV3MI6NUTsHhU6Z4=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_munnerz_goautoneg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/munnerz/goautoneg",
        sum = "h1:C3w9PqII01/Oq1c1nUAm88MOHcQC9l5mIlSMApZMrHA=",
        version = "v0.0.0-20191010083416-a7dc8b61c822",
    )

    go_repository(
        name = "com_github_mwitkow_go_conntrack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mwitkow/go-conntrack",
        sum = "h1:KUppIJq7/+SVif2QVs3tOP0zanoHgBEVAwHxUSIzRqU=",
        version = "v0.0.0-20190716064945-2f068394615f",
    )
    go_repository(
        name = "com_github_mwitkow_go_proto_validators",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mwitkow/go-proto-validators",
        sum = "h1:F6LFfmgVnfULfaRsQWBbe7F7ocuHCr9+7m+GAeDzNbQ=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_mxk_go_flowrate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mxk/go-flowrate",
        sum = "h1:y5//uYreIhSUg3J1GEMiLbxo1LJaP8RfCpH6pymGZus=",
        version = "v0.0.0-20140419014527-cca7078d478f",
    )
    go_repository(
        name = "com_github_nats_io_jwt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nats-io/jwt",
        sum = "h1:+RB5hMpXUUA2dfxuhBTEkMOrYmM+gKIZYS1KjSostMI=",
        version = "v0.3.2",
    )

    go_repository(
        name = "com_github_nats_io_nats_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nats-io/nats.go",
        sum = "h1:ik3HbLhZ0YABLto7iX80pZLPw/6dx3T+++MZJwLnMrQ=",
        version = "v1.9.1",
    )
    go_repository(
        name = "com_github_nats_io_nats_server_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nats-io/nats-server/v2",
        sum = "h1:i2Ly0B+1+rzNZHHWtD4ZwKi+OU5l+uQo1iDHZ2PmiIc=",
        version = "v2.1.2",
    )
    go_repository(
        name = "com_github_nats_io_nkeys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nats-io/nkeys",
        sum = "h1:6JrEfig+HzTH85yxzhSVbjHRJv9cn0p6n3IngIcM5/k=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_nats_io_nuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nats-io/nuid",
        sum = "h1:5iA8DT8V7q8WK2EScv2padNa/rTESc1KdnPw4TC2paw=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_nelsam_hel_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nelsam/hel/v2",
        sum = "h1:Z3TAKd9JS3BoKi6fW+d1bKD2Mf0FzTqDUEAwLWzYPRQ=",
        version = "v2.3.3",
    )

    go_repository(
        name = "com_github_niemeyer_pretty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/niemeyer/pretty",
        sum = "h1:fD57ERR4JtEqsWbfPhv4DMiApHyliiK5xCTNVSPiaAs=",
        version = "v0.0.0-20200227124842-a10e7caefd8e",
    )

    go_repository(
        name = "com_github_nishanths_predeclared",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nishanths/predeclared",
        sum = "h1:3f0nxAmdj/VoCGN/ijdMy7bj6SBagaqYg1B0hu8clMA=",
        version = "v0.0.0-20200524104333-86fad755b4d3",
    )

    go_repository(
        name = "com_github_nytimes_gziphandler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/NYTimes/gziphandler",
        sum = "h1:ZUDjpQae29j0ryrS0u/B8HZfJBtBQHjqw2rQ2cqUQ3I=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_oklog_oklog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/oklog/oklog",
        sum = "h1:wVfs8F+in6nTBMkA7CbRw+zZMIB7nNM825cM1wuzoTk=",
        version = "v0.3.2",
    )

    go_repository(
        name = "com_github_oklog_run",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/oklog/run",
        sum = "h1:Ru7dDtJNOyC66gQ5dQmaCa0qIsAUFY3sFpK1Xk8igrw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_oklog_ulid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/oklog/ulid",
        sum = "h1:EGfNDEx6MqHz8B3uNV6QAib1UR2Lm97sHi3ocA6ESJ4=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_olekukonko_tablewriter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/olekukonko/tablewriter",
        sum = "h1:P2Ga83D34wi1o9J6Wh1mRuqd4mF/x/lgBS7N7AbDhec=",
        version = "v0.0.5",
    )
    go_repository(
        name = "com_github_oneofone_xxhash",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/OneOfOne/xxhash",
        sum = "h1:KMrpdQIwFcEqXDklaen+P1axHaj9BSKzvpUUfnHldSE=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_onsi_ginkgo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/ginkgo",
        sum = "h1:OoxbjfXVZyod1fmWYhI7SEyaD8B00ynP3T+D5GiyHOY=",
        version = "v1.10.3",
    )

    go_repository(
        name = "com_github_onsi_ginkgo_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/ginkgo/v2",
        sum = "h1:06xGQy5www2oN160RtEZoTvnP2sPhEfePYmCDc2szss=",
        version = "v2.9.7",
    )
    go_repository(
        name = "com_github_onsi_gomega",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/gomega",
        sum = "h1:fVih9JD6ogIiHUN6ePK7HJidyEDpWGVB5mzM7cWNXoU=",
        version = "v1.27.7",
    )
    go_repository(
        name = "com_github_op_go_logging",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/op/go-logging",
        sum = "h1:lDH9UUVJtmYCjyT0CI4q8xvlXPxeZ0gYCVvWbmPlp88=",
        version = "v0.0.0-20160315200505-970db520ece7",
    )

    go_repository(
        name = "com_github_opencontainers_go_digest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/go-digest",
        sum = "h1:apOUWs51W5PlhuyGyz9FCeeBIOUDA/6nW8Oi/yOhh5U=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_opencontainers_image_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/image-spec",
        sum = "h1:YWuSjZCQAPM8UUBLkYUk1e+rZcvWHJmFb6i6rM44Xs8=",
        version = "v1.1.0-rc2.0.20221005185240-3a7f492d3f1b",
    )
    go_repository(
        name = "com_github_opencontainers_runc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/runc",
        sum = "h1:XbhB8IfG/EsnhNvZtNdLB0GBw92GYEFvKlhaJk9jUgA=",
        version = "v1.1.6",
    )
    go_repository(
        name = "com_github_opencontainers_runtime_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/runtime-spec",
        sum = "h1:wHa9jroFfKGQqFHj0I1fMRKLl0pfj+ynAqBxo3v6u9w=",
        version = "v1.1.0-rc.1",
    )
    go_repository(
        name = "com_github_opencontainers_runtime_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/runtime-tools",
        sum = "h1:DmNGcqH3WDbV5k8OJ+esPWbqUOX5rMLR2PMvziDMJi0=",
        version = "v0.9.1-0.20221107090550-2e043c6bd626",
    )

    go_repository(
        name = "com_github_opencontainers_selinux",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/selinux",
        sum = "h1:+5Zbo97w3Lbmb3PeqQtpmTkMwsW5nRI3YaLpt7tQ7oU=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_github_opentracing_basictracer_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opentracing/basictracer-go",
        sum = "h1:YyUAhaEfjoWXclZVJ9sGoNct7j4TVk7lZWlQw5UXuoo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_opentracing_contrib_go_observer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opentracing-contrib/go-observer",
        sum = "h1:lM6RxxfUMrYL/f8bWEUqdXrANWtrL7Nndbm9iFN0DlU=",
        version = "v0.0.0-20170622124052-a52f23424492",
    )

    go_repository(
        name = "com_github_opentracing_opentracing_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opentracing/opentracing-go",
        sum = "h1:uEJPy/1a5RIPAJ0Ov+OIO8OxWu77jEv+1B0VhjKrZUs=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_openzipkin_contrib_zipkin_go_opentracing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/openzipkin-contrib/zipkin-go-opentracing",
        sum = "h1:ZCnq+JUrvXcDVhX/xRolRBZifmabN1HcS1wrPSvxhrU=",
        version = "v0.4.5",
    )
    go_repository(
        name = "com_github_openzipkin_zipkin_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/openzipkin/zipkin-go",
        sum = "h1:nY8Hti+WKaP0cRsSeQ026wU03QsM762XBeCXBb9NAWI=",
        version = "v0.2.2",
    )

    go_repository(
        name = "com_github_otiai10_copy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/otiai10/copy",
        sum = "h1:IinKAryFFuPONZ7cm6T6E2QX/vcJwSnlaA5lfoaXIiQ=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_otiai10_curr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/otiai10/curr",
        sum = "h1:TJIWdbX0B+kpNagQrjgq8bCMrbhiuX73M2XwgtDMoOI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_otiai10_mint",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/otiai10/mint",
        sum = "h1:VYWnrP5fXmz1MXvjuUvcBrXSjGE6xjON+axB/UrpO3E=",
        version = "v1.3.2",
    )

    go_repository(
        name = "com_github_pact_foundation_pact_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pact-foundation/pact-go",
        sum = "h1:OYkFijGHoZAYbOIb1LWXrwKQbMMRUv1oQ89blD2Mh2Q=",
        version = "v1.0.4",
    )

    go_repository(
        name = "com_github_pascaldekloe_goe",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pascaldekloe/goe",
        sum = "h1:Lgl0gzECD8GnQ5QCWA8o6BtfL6mDH5rQgM4/fX3avOs=",
        version = "v0.0.0-20180627143212-57f6aae5913c",
    )
    go_repository(
        name = "com_github_pborman_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pborman/uuid",
        sum = "h1:+ZZIw58t/ozdjRaXh/3awHfmWRbzYxJoAdNJxe/3pvw=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_pelletier_go_buffruneio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pelletier/go-buffruneio",
        sum = "h1:U4t4R6YkofJ5xHm3dJzuRpPZ0mr5MMCoAWooScCR7aA=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_pelletier_go_toml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pelletier/go-toml",
        sum = "h1:4yBQzkHv+7BHq2PQUZF3Mx0IYxG7LsP222s7Agd3ve8=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_github_pelletier_go_toml_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pelletier/go-toml/v2",
        sum = "h1:nrzqCb7j9cDFj2coyLNLaZuJTLjWjlaz6nvTvIwycIU=",
        version = "v2.0.6",
    )
    go_repository(
        name = "com_github_performancecopilot_speed",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/performancecopilot/speed",
        sum = "h1:2WnRzIquHa5QxaJKShDkLM+sc0JPuwhXzK8OYOyt3Vg=",
        version = "v3.0.0+incompatible",
    )

    go_repository(
        name = "com_github_peterbourgon_diskv",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/peterbourgon/diskv",
        sum = "h1:UBdAOUP5p4RWqPBg048CAvpKN+vxiaj6gdUUzhl4XmI=",
        version = "v2.0.1+incompatible",
    )

    go_repository(
        name = "com_github_phayes_freeport",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/phayes/freeport",
        sum = "h1:Ii+DKncOVM8Cu1Hc+ETb5K+23HdAMvESYE3ZJ5b5cMI=",
        version = "v0.0.0-20220201140144-74d24b5ae9f5",
    )
    go_repository(
        name = "com_github_pierrec_lz4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pierrec/lz4",
        sum = "h1:2xWsjqPFWcplujydGg4WmhC/6fZqK42wMM8aXeqhl0I=",
        version = "v2.0.5+incompatible",
    )
    go_repository(
        name = "com_github_pierrec_lz4_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pierrec/lz4/v4",
        sum = "h1:MO0/ucJhngq7299dKLwIMtgTfbkoSPF6AoMYDd8Q4q0=",
        version = "v4.1.15",
    )

    go_repository(
        name = "com_github_pjbgf_sha1cd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pjbgf/sha1cd",
        sum = "h1:4D5XXmUUBUl/xQ6IjCkEAbqXskkq/4O7LmGn0AqMDs4=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_pkg_browser",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pkg/browser",
        sum = "h1:KoWmjvw+nsYOo29YJK9vDA65RGE3NrOnUtO7a+RF9HU=",
        version = "v0.0.0-20210911075715-681adbf594b8",
    )
    go_repository(
        name = "com_github_pkg_diff",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pkg/diff",
        sum = "h1:aoZm08cpOy4WuID//EZDgcC4zIxODThtZNPirFr42+A=",
        version = "v0.0.0-20210226163009-20ebb0f2a09e",
    )
    go_repository(
        name = "com_github_pkg_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pkg/errors",
        sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_pkg_profile",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pkg/profile",
        sum = "h1:F++O52m40owAmADcojzM+9gyjmMOY/T4oYJkgFDH8RE=",
        version = "v1.2.1",
    )

    go_repository(
        name = "com_github_pkg_sftp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pkg/sftp",
        sum = "h1:I2qBYMChEhIjOgazfJmV3/mZM256btk6wkCDRmW7JYs=",
        version = "v1.13.1",
    )

    go_repository(
        name = "com_github_pmezard_go_difflib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_posener_complete",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/posener/complete",
        sum = "h1:NP0eAhjcjImqslEwo/1hq7gpajME0fTLTezBKDqfXqo=",
        version = "v1.2.3",
    )

    go_repository(
        name = "com_github_poy_onpar",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/poy/onpar",
        sum = "h1:QaNrNiZx0+Nar5dLgTVp5mXkyoVFIbepjyEoGSnhbAY=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_pquerna_cachecontrol",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pquerna/cachecontrol",
        sum = "h1:yJMy84ti9h/+OEWa752kBTKv4XC30OtVVHYv/8cTqKc=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_prometheus_client_golang",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/client_golang",
        sum = "h1:8tXpTmJbyH5lydzFPoxSIJ0J46jdh3tylbvM1xCv0LI=",
        version = "v1.15.1",
    )
    go_repository(
        name = "com_github_prometheus_client_model",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:5lQXD3cAg1OXBf4Wq03gTrXHeaV0TQvGfUooCfx1yqY=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_prometheus_common",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/common",
        sum = "h1:EKsfXEYo4JpWMHH5cg+KOUWeuJSov1Id8zGR8eeI1YM=",
        version = "v0.42.0",
    )
    go_repository(
        name = "com_github_prometheus_procfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:wzCHvIvM5SxWqYvwgVL7yJY8Lz3PKn49KQtpgMYJfhI=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_prometheus_prometheus",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/prometheus",
        sum = "h1:7QPitgO2kOFG8ecuRn9O/4L9+10He72rVRJvMXrE9Hg=",
        version = "v2.5.0+incompatible",
    )
    go_repository(
        name = "com_github_prometheus_tsdb",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/tsdb",
        sum = "h1:YZcsG11NqnK4czYLrWd9mpEuAJIHVQLwdrleYfszMAA=",
        version = "v0.7.1",
    )

    go_repository(
        name = "com_github_protonmail_go_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ProtonMail/go-crypto",
        sum = "h1:wPbRQzjjwFc0ih8puEVAOFGELsn1zoIIYdxvML7mDxA=",
        version = "v0.0.0-20230217124315-7d5c6f04bbb8",
    )
    go_repository(
        name = "com_github_protonmail_go_mime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ProtonMail/go-mime",
        sum = "h1:dS7r5z4iGS0qCjM7UwWdsEMzQesUQbGcXdSm2/tWboA=",
        version = "v0.0.0-20221031134845-8fd9bc37cf08",
    )
    go_repository(
        name = "com_github_protonmail_gopenpgp_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ProtonMail/gopenpgp/v2",
        sum = "h1:97SjlWNAxXl9P22lgwgrZRshQdiEfAht0g3ZoiA1GCw=",
        version = "v2.5.2",
    )
    go_repository(
        name = "com_github_pseudomuto_protoc_gen_doc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pseudomuto/protoc-gen-doc",
        sum = "h1:Ah259kcrio7Ix1Rhb6u8FCaOkzf9qRBqXnvAufg061w=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_pseudomuto_protokit",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pseudomuto/protokit",
        sum = "h1:hlnBDcy3YEDXH7kc9gV+NLaN0cDzhDvD1s7Y6FZ8RpM=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_puerkitobio_purell",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/PuerkitoBio/purell",
        sum = "h1:WEQqlqaGbrPkxLJWfBwQmfEAE1Z7ONdDLqrN38tNFfI=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_puerkitobio_urlesc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/PuerkitoBio/urlesc",
        sum = "h1:d+Bc7a5rLufV/sSk/8dngufqelfh6jnri85riMAaF/M=",
        version = "v0.0.0-20170810143723-de5bf2ad4578",
    )

    go_repository(
        name = "com_github_rcrowley_go_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rcrowley/go-metrics",
        sum = "h1:9ZKAASQSHhDYGoxY8uLVpewe1GDZ2vu2Tr/vTdVAkFQ=",
        version = "v0.0.0-20181016184325-3113b8401b8a",
    )
    go_repository(
        name = "com_github_redis_go_redis_v9",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/redis/go-redis/v9",
        sum = "h1:FC82T+CHJ/Q/PdyLW++GeCO+Ol59Y4T7R4jbgjvktgc=",
        version = "v9.0.4",
    )

    go_repository(
        name = "com_github_rivo_uniseg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rivo/uniseg",
        sum = "h1:8TfxU8dW6PdqD27gjM8MVNuicgxIjxpm4K7x4jp8sis=",
        version = "v0.4.4",
    )

    go_repository(
        name = "com_github_robfig_cron_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/robfig/cron/v3",
        sum = "h1:WdRxkvbJztn8LMz/QEvLN5sBU+xKpSqwwUO1Pjr4qDs=",
        version = "v3.0.1",
    )
    go_repository(
        name = "com_github_rogpeppe_fastuuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rogpeppe/fastuuid",
        sum = "h1:Ppwyp6VYCF1nvBTXL3trRso7mXMlRrw9ooo375wvi2s=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_rogpeppe_go_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rogpeppe/go-internal",
        sum = "h1:TMyTOH3F/DB16zRVcYyreMH6GnZZrwQVAoYjRBZyWFQ=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_rs_cors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rs/cors",
        sum = "h1:l9HGsTsHJcvW14Nk7J9KFz8bzeAWXn3CG6bgt7LsrAE=",
        version = "v1.9.0",
    )

    go_repository(
        name = "com_github_rubenv_sql_migrate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rubenv/sql-migrate",
        sum = "h1:Vx+n4Du8X8VTYuXbhNxdEUoh6wiJERA0GlWocR5FrbA=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_rubiojr_go_vhd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rubiojr/go-vhd",
        sum = "h1:if3/24+h9Sq6eDx8UUz1SO9cT9tizyIsATfB7b4D3tc=",
        version = "v0.0.0-20200706105327-02e210299021",
    )
    go_repository(
        name = "com_github_russross_blackfriday",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/russross/blackfriday",
        sum = "h1:KqfZb0pUVN2lYqZUYRddxF4OR8ZMURnJIG5Y3VRLtww=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_russross_blackfriday_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/russross/blackfriday/v2",
        sum = "h1:JIOH55/0cWyOuilr9/qlrm0BSXldqnqwMsf35Ld67mk=",
        version = "v2.1.0",
    )

    go_repository(
        name = "com_github_ryanuber_columnize",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ryanuber/columnize",
        sum = "h1:UFr9zpz4xgTnIE5yIMtWAMngCdZ9p/+q6lTbgelo80M=",
        version = "v0.0.0-20160712163229-9b3edd62028f",
    )
    go_repository(
        name = "com_github_ryanuber_go_glob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ryanuber/go-glob",
        sum = "h1:iQh3xXAumdQ+4Ufa5b25cRpC5TYKlno6hsv6Cb3pkBk=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_samuel_go_zookeeper",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/samuel/go-zookeeper",
        sum = "h1:p3Vo3i64TCLY7gIfzeQaUJ+kppEO5WQG3cL8iE8tGHU=",
        version = "v0.0.0-20190923202752-2cc03de413da",
    )
    go_repository(
        name = "com_github_santhosh_tekuri_jsonschema_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/santhosh-tekuri/jsonschema/v5",
        sum = "h1:WCcC4vZDS1tYNxjWlwRJZQy28r8CMoggKnxNzxsVDMQ=",
        version = "v5.2.0",
    )

    go_repository(
        name = "com_github_sassoftware_go_rpmutils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sassoftware/go-rpmutils",
        replace = "github.com/rmohr/go-rpmutils",  # keep
        sum = "h1:IPm+rJWi7P9mkoFQjloDEGThxlM71dzDrC2DZSjgnL8=",  # keep
        version = "v0.1.2-0.20201215123907-5acf7436c00d",  # keep
    )

    go_repository(
        name = "com_github_sassoftware_relic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sassoftware/relic",
        sum = "h1:Pwyh1F3I0r4clFJXkSI8bOyJINGqpgjJU3DYAZeI05A=",
        version = "v7.2.1+incompatible",
    )

    go_repository(
        name = "com_github_schollz_progressbar_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/schollz/progressbar/v3",
        sum = "h1:o8rySDYiQ59Mwzy2FELeHY5ZARXZTVJC7iHD6PEFUiE=",
        version = "v3.13.1",
    )

    go_repository(
        name = "com_github_sean_seed",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sean-/seed",
        sum = "h1:nn5Wsu0esKSJiIVhscUtVbo7ada43DJhG55ua/hjS5I=",
        version = "v0.0.0-20170313163322-e2103e2c3529",
    )
    go_repository(
        name = "com_github_sebdah_goldie",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sebdah/goldie",
        sum = "h1:9GNhIat69MSlz/ndaBg48vl9dF5fI+NBB6kfOxgfkMc=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_seccomp_libseccomp_golang",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/seccomp/libseccomp-golang",
        sum = "h1:RpforrEYXWkmGwJHIGnLZ3tTWStkjVVstwzNGqxX2Ds=",
        version = "v0.9.2-0.20220502022130-f33da4d89646",
    )
    go_repository(
        name = "com_github_secure_systems_lab_go_securesystemslib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/secure-systems-lab/go-securesystemslib",
        sum = "h1:T65atpAVCJQK14UA57LMdZGpHi4QYSH/9FZyNGqMYIA=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_segmentio_ksuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/segmentio/ksuid",
        sum = "h1:sBo2BdShXjmcugAMwjugoGUdUV0pcxY5mW4xKRn3v4c=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_sergi_go_diff",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sergi/go-diff",
        sum = "h1:xkr+Oxo4BOQKmkn/B9eMK0g5Kg/983T9DqqPHwYqD+8=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_shibumi_go_pathspec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/shibumi/go-pathspec",
        sum = "h1:QUyMZhFo0Md5B8zV8x2tesohbb5kfbpTi9rBnKh5dkI=",
        version = "v1.3.0",
    )

    go_repository(
        name = "com_github_shopify_logrus_bugsnag",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Shopify/logrus-bugsnag",
        sum = "h1:UrqY+r/OJnIp5u0s1SbQ8dVfLCZJsnvazdBP5hS4iRs=",
        version = "v0.0.0-20171204204709-577dee27f20d",
    )
    go_repository(
        name = "com_github_shopify_sarama",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Shopify/sarama",
        sum = "h1:9oksLxC6uxVPHPVYUmq6xhr1BOF/hHobWH2UzO67z1s=",
        version = "v1.19.0",
    )
    go_repository(
        name = "com_github_shopify_toxiproxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Shopify/toxiproxy",
        sum = "h1:TKdv8HiTLgE5wdJuEML90aBgNWsokNbMijUGhmcoBJc=",
        version = "v2.1.4+incompatible",
    )

    go_repository(
        name = "com_github_shopspring_decimal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/shopspring/decimal",
        sum = "h1:2Usl1nmF/WZucqkFZhnfFYxxxu8LG21F6nPQBE5gKV8=",
        version = "v1.3.1",
    )

    go_repository(
        name = "com_github_shurcool_sanitized_anchor_name",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/shurcooL/sanitized_anchor_name",
        sum = "h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_siderolabs_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/crypto",
        sum = "h1:o1KIR1KyevUcY9nbJlSyQAj7+p+rveGGF8LjAAFMtjc=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_siderolabs_gen",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/gen",
        sum = "h1:V3UsZ2KrsryaTMZGZUHAr1CFdPc2/R1lM6lA4a4zCDo=",
        version = "v0.4.3",
    )
    go_repository(
        name = "com_github_siderolabs_go_api_signature",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/go-api-signature",
        sum = "h1:C5tUzuFsJYidpYyVfJGYpgQwETglA8B62ET4obkLDGE=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_siderolabs_go_blockdevice",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/go-blockdevice",
        sum = "h1:NgpR9XTl/N7WeL59QHBsseDD0Nb8Y2nel+W3u7xHIvY=",
        version = "v0.4.5",
    )
    go_repository(
        name = "com_github_siderolabs_go_debug",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/go-debug",
        sum = "h1:c8styCvp+MO0oPO8q4N1CKSF3fVuAT0qnuUIeZ/BiW0=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_siderolabs_go_pointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/go-pointer",
        sum = "h1:6TshPKep2doDQJAAtHUuHWXbca8ZfyRySjSBT/4GsMU=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_siderolabs_net",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/net",
        sum = "h1:1bOgVay/ijPkJz4qct98nHsiB/ysLQU0KLoBC4qLm7I=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_siderolabs_protoenc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/protoenc",
        sum = "h1:QFxWIAo//12+/bm27GNYoK/TpQGTYsRrrZCu9jSghvU=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_siderolabs_talos_pkg_machinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/talos/pkg/machinery",
        sum = "h1:OcfrGmzip4kPDF7A4dccI1RFg3MACqgxHYYPTdSzzC8=",
        version = "v1.4.4",
    )
    go_repository(
        name = "com_github_sigstore_protobuf_specs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/protobuf-specs",
        sum = "h1:X0l/E2C2c79t/rI/lmSu8WAoKWsQtMqDzAMiDdEMGr8=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_sigstore_rekor",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/rekor",
        sum = "h1:ahlnoEY3zo8Vc+eZLPobamw6YfBTAbI0lthzUQd6qe4=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_sigstore_sigstore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore",
        sum = "h1:jH4AzR7qlEH/EWzm+opSpxCfuUcjHL+LJPuQE7h40WE=",
        version = "v1.6.4",
    )

    go_repository(
        name = "com_github_sirupsen_logrus",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:trlNQbNUG3OdDrDil03MCb1H2o9nJ1x4/5LYw7byDE0=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_github_skeema_knownhosts",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/skeema/knownhosts",
        sum = "h1:Wvr9V0MxhjRbl3f9nMnKnFfiWTJmtECJ9Njkea3ysW0=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_skratchdot_open_golang",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/skratchdot/open-golang",
        sum = "h1:JIAuq3EEf9cgbU6AtGPK4CTG3Zf6CKMNqf0MHTggAUA=",
        version = "v0.0.0-20200116055534-eef842397966",
    )
    go_repository(
        name = "com_github_smartystreets_assertions",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/smartystreets/assertions",
        sum = "h1:UVQPSSmc3qtTi+zPPkCXvZX9VvW/xT/NsRvKfwY81a8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_smartystreets_go_aws_auth",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/smartystreets/go-aws-auth",
        sum = "h1:hp2CYQUINdZMHdvTdXtPOY2ainKl4IoMcpAXEf2xj3Q=",
        version = "v0.0.0-20180515143844-0c1422d1fdb9",
    )
    go_repository(
        name = "com_github_smartystreets_goconvey",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/smartystreets/goconvey",
        sum = "h1:fv0U8FUIMPNf1L9lnHLvLhgicrIVChEkdzIKYqbNC9s=",
        version = "v1.6.4",
    )
    go_repository(
        name = "com_github_smartystreets_gunit",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/smartystreets/gunit",
        sum = "h1:RyPDUFcJbvtXlhJPk7v+wnxZRY2EUokhEYl2EJOPToI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_soheilhy_cmux",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/soheilhy/cmux",
        sum = "h1:jjzc5WVemNEDTLwv9tlmemhC73tI08BNOIGwBOo10Js=",
        version = "v0.1.5",
    )
    go_repository(
        name = "com_github_sony_gobreaker",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sony/gobreaker",
        sum = "h1:oMnRNZXX5j85zso6xCPRNPtmAycat+WcoKbklScLDgQ=",
        version = "v0.4.1",
    )

    go_repository(
        name = "com_github_spaolacci_murmur3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spaolacci/murmur3",
        sum = "h1:qLC7fQah7D6K1B0ujays3HV9gkFtllcxhzImRR7ArPQ=",
        version = "v0.0.0-20180118202830-f09979ecbc72",
    )
    go_repository(
        name = "com_github_spf13_afero",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/afero",
        sum = "h1:stMpOSZFs//0Lv29HduCmli3GUfpFoF3Y1Q/aXj/wVM=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_github_spf13_cast",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/cast",
        sum = "h1:rj3WzYc11XZaIZMPKmwP96zkFEnnAmV8s6XbB2aY32w=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_spf13_cobra",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/cobra",
        sum = "h1:hyqWnYt1ZQShIddO5kBpj3vu05/++x6tJ6dg8EC572I=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_spf13_jwalterweatherman",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/jwalterweatherman",
        sum = "h1:ue6voC5bR5F8YxI5S67j9i582FU4Qvo2bmqnqMYADFk=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_spf13_pflag",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/pflag",
        sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
        version = "v1.0.5",
    )
    go_repository(
        name = "com_github_spf13_viper",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/viper",
        sum = "h1:js3yy885G8xwJa6iOISGFwd+qlUo5AvyXb7CiihdtiU=",
        version = "v1.15.0",
    )

    go_repository(
        name = "com_github_src_d_gcfg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/src-d/gcfg",
        sum = "h1:xXbNR5AlLSA315x2UO+fTSSAXCDf+Ar38/6oyGbDKQ4=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_stefanberger_go_pkcs11uri",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/stefanberger/go-pkcs11uri",
        sum = "h1:lIOOHPEbXzO3vnmx2gok1Tfs31Q8GQqKLc8vVqyQq/I=",
        version = "v0.0.0-20201008174630-78d3cae3a980",
    )
    go_repository(
        name = "com_github_stoewer_go_strcase",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/stoewer/go-strcase",
        sum = "h1:Z2iHWqGXH00XYgqDmNgQbIBxf3wrNq0F3feEy0ainaU=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_streadway_amqp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/streadway/amqp",
        sum = "h1:WhxRHzgeVGETMlmVfqhRn8RIeeNoPr2Czh33I4Zdccw=",
        version = "v0.0.0-20190827072141-edfb9018d271",
    )
    go_repository(
        name = "com_github_streadway_handy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/streadway/handy",
        sum = "h1:AhmOdSHeswKHBjhsLs/7+1voOxT+LLrSk/Nxvk35fug=",
        version = "v0.0.0-20190108123426-d5acb3125c2a",
    )

    go_repository(
        name = "com_github_stretchr_objx",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/stretchr/objx",
        sum = "h1:1zr/of2m5FGMsad5YfcqgdqdWrIhu+EBEJRhR1U7z/c=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_stretchr_testify",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/stretchr/testify",
        sum = "h1:RP3t2pwF7cMEbC1dqtB6poj3niw/9gnV4Cjg5oW5gtY=",
        version = "v1.8.3",
    )
    go_repository(
        name = "com_github_subosito_gotenv",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/subosito/gotenv",
        sum = "h1:X1TuBLAMDFbaTAChgCBLu3DU3UPyELpnF2jjJ2cz/S8=",
        version = "v1.4.2",
    )
    go_repository(
        name = "com_github_syndtr_gocapability",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/syndtr/gocapability",
        sum = "h1:kdXcSzyDtseVEc4yCz2qF8ZrQvIDBJLl4S1c3GCXmoI=",
        version = "v0.0.0-20200815063812-42c35b437635",
    )
    go_repository(
        name = "com_github_syndtr_goleveldb",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/syndtr/goleveldb",
        sum = "h1:vfofYNRScrDdvS342BElfbETmL1Aiz3i2t0zfRj16Hs=",
        version = "v1.0.1-0.20220721030215-126854af5e6d",
    )
    go_repository(
        name = "com_github_tchap_go_patricia_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tchap/go-patricia/v2",
        sum = "h1:6rQp39lgIYZ+MHmdEq4xzuk1t7OdC35z/xm0BGhTkes=",
        version = "v2.3.1",
    )

    go_repository(
        name = "com_github_tedsuo_ifrit",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tedsuo/ifrit",
        sum = "h1:LUUe4cdABGrIJAhl1P1ZpWY76AwukVszFdwkVFVLwIk=",
        version = "v0.0.0-20180802180643-bea94bb476cc",
    )

    go_repository(
        name = "com_github_theupdateframework_go_tuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/theupdateframework/go-tuf",
        sum = "h1:habfDzTmpbzBLIFGWa2ZpVhYvFBoK0C1onC3a4zuPRA=",
        version = "v0.5.2",
    )
    go_repository(
        name = "com_github_tidwall_pretty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tidwall/pretty",
        sum = "h1:RWIZEg2iJ8/g6fDDYzMpobmaoGh5OLl4AXtGUGPcqCs=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_titanous_rocacheck",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/titanous/rocacheck",
        sum = "h1:e/5i7d4oYZ+C1wj2THlRK+oAhjeS/TRQwMfkIuet3w0=",
        version = "v0.0.0-20171023193734-afe73141d399",
    )
    go_repository(
        name = "com_github_tj_assert",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tj/assert",
        sum = "h1:Rw8kxzWo1mr6FSaYXjQELRe88y2KdfynXdnK72rdjtA=",
        version = "v0.0.0-20171129193455-018094318fb0",
    )
    go_repository(
        name = "com_github_tj_go_elastic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tj/go-elastic",
        sum = "h1:eGaGNxrtoZf/mBURsnNQKDR7u50Klgcf2eFDQEnc8Bc=",
        version = "v0.0.0-20171221160941-36157cbbebc2",
    )
    go_repository(
        name = "com_github_tj_go_kinesis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tj/go-kinesis",
        sum = "h1:m74UWYy+HBs+jMFR9mdZU6shPewugMyH5+GV6LNgW8w=",
        version = "v0.0.0-20171128231115-08b17f58cb1b",
    )
    go_repository(
        name = "com_github_tj_go_spin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tj/go-spin",
        sum = "h1:lhdWZsvImxvZ3q1C5OIB7d72DuOwP4O2NdBg9PyzNds=",
        version = "v1.1.0",
    )

    go_repository(
        name = "com_github_tmc_grpc_websocket_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tmc/grpc-websocket-proxy",
        sum = "h1:6fotK7otjonDflCTK0BCfls4SPy3NcCVb5dqqmbRknE=",
        version = "v0.0.0-20220101234140-673ab2c3ae75",
    )
    go_repository(
        name = "com_github_tomasen_realip",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tomasen/realip",
        sum = "h1:fb190+cK2Xz/dvi9Hv8eCYJYvIGUTN2/KLq1pT6CjEc=",
        version = "v0.0.0-20180522021738-f0c99a92ddce",
    )
    go_repository(
        name = "com_github_transparency_dev_merkle",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/transparency-dev/merkle",
        sum = "h1:Q9nBoQcZcgPamMkGn7ghV8XiTZ/kRxn1yCG81+twTK4=",
        version = "v0.0.2",
    )
    go_repository(
        name = "com_github_ugorji_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ugorji/go",
        sum = "h1:/68gy2h+1mWMrwZFeD1kQialdSzAb432dtpeJ42ovdo=",
        version = "v1.1.7",
    )
    go_repository(
        name = "com_github_ugorji_go_codec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ugorji/go/codec",
        sum = "h1:2SvQaVZ1ouYrrKKwoSk2pzd4A9evlKJb9oTL+OaLUSs=",
        version = "v1.1.7",
    )
    go_repository(
        name = "com_github_ulikunitz_xz",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ulikunitz/xz",
        sum = "h1:t92gobL9l3HE202wg3rlk19F6X+JOxl9BBrCCMYEYd8=",
        version = "v0.5.10",
    )
    go_repository(
        name = "com_github_urfave_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/urfave/cli",
        sum = "h1:igJgVw1JdKH+trcLWLeLwZjU9fEfPesQ+9/e4MQ44S8=",
        version = "v1.22.12",
    )

    go_repository(
        name = "com_github_vbatts_tar_split",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vbatts/tar-split",
        sum = "h1:Via6XqJr0hceW4wff3QRzD5gAk/tatMw/4ZA7cTlIME=",
        version = "v0.11.2",
    )
    go_repository(
        name = "com_github_veraison_go_cose",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/veraison/go-cose",
        sum = "h1:AalPS4VGiKavpAzIlBjrn7bhqXiXi4jbMYY/2+UC+4o=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_vishvananda_netlink",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vishvananda/netlink",
        sum = "h1:Llsql0lnQEbHj0I1OuKyp8otXp0r3q0mPkuhwHfStVs=",
        version = "v1.2.1-beta.2",
    )
    go_repository(
        name = "com_github_vishvananda_netns",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vishvananda/netns",
        sum = "h1:Cn05BRLm+iRP/DZxyVSsfVyrzgjDbwHwkVt38qvXnNI=",
        version = "v0.0.2",
    )
    go_repository(
        name = "com_github_vividcortex_gohistogram",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/VividCortex/gohistogram",
        sum = "h1:6+hBz+qvs0JOrrNhhmR7lFxo5sINxBCGXrdtl/UvroE=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_vmihailenco_msgpack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmihailenco/msgpack",
        sum = "h1:wapg9xDUZDzGCNFlwc5SqI1rvcciqcxEHac4CYj89xI=",
        version = "v3.3.3+incompatible",
    )
    go_repository(
        name = "com_github_vmihailenco_msgpack_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmihailenco/msgpack/v4",
        sum = "h1:07s4sz9IReOgdikxLTKNbBdqDMLsjPKXwvCazn8G65U=",
        version = "v4.3.12",
    )
    go_repository(
        name = "com_github_vmihailenco_msgpack_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmihailenco/msgpack/v5",
        sum = "h1:5gO0H1iULLWGhs2H5tbAHIZTV8/cYafcFOr9znI5mJU=",
        version = "v5.3.5",
    )
    go_repository(
        name = "com_github_vmihailenco_tagparser",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmihailenco/tagparser",
        sum = "h1:quXMXlA39OCbd2wAdTsGDlK9RkOk6Wuw+x37wVyIuWY=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_vmihailenco_tagparser_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmihailenco/tagparser/v2",
        sum = "h1:y09buUbR+b5aycVFQs/g70pqKVZNBmxwAhO7/IwNM9g=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_vmware_govmomi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmware/govmomi",
        sum = "h1:Fm8ugPnnlMSTSceDKY9goGvjmqc6eQLPUSUeNXdpeXA=",
        version = "v0.30.0",
    )
    go_repository(
        name = "com_github_vtolstov_go_ioctl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vtolstov/go-ioctl",
        sum = "h1:X6ps8XHfpQjw8dUStzlMi2ybiKQ2Fmdw7UM+TinwvyM=",
        version = "v0.0.0-20151206205506-6be9cced4810",
    )

    go_repository(
        name = "com_github_weppos_publicsuffix_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/weppos/publicsuffix-go",
        sum = "h1:eR9jm8DVMdrDUuVji4eOxPK4r/dANDlDBdISSUUV96s=",
        version = "v0.20.1-0.20221031080346-e4081aa8a6de",
    )

    go_repository(
        name = "com_github_x448_float16",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/x448/float16",
        sum = "h1:qLwI1I70+NjRFUR3zs1JPUCgaCXSh3SW62uAKT1mSBM=",
        version = "v0.8.4",
    )
    go_repository(
        name = "com_github_xanzy_go_gitlab",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xanzy/go-gitlab",
        sum = "h1:+nHztQuCXGSMluKe5Q9IRaPdz6tO8O0gMkQ0vqGpiBk=",
        version = "v0.31.0",
    )
    go_repository(
        name = "com_github_xanzy_ssh_agent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xanzy/ssh-agent",
        sum = "h1:+/15pJfg/RsTxqYcX6fHqOXZwwMP+2VyYWJeWM2qQFM=",
        version = "v0.3.3",
    )
    go_repository(
        name = "com_github_xdg_go_pbkdf2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xdg-go/pbkdf2",
        sum = "h1:Su7DPu48wXMwC3bs7MCNG+z4FhcyEuz5dlvchbq0B0c=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_xdg_go_scram",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xdg-go/scram",
        sum = "h1:VOMT+81stJgXW3CpHyqHN3AXDYIMsx56mEFrB37Mb/E=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_xdg_go_stringprep",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xdg-go/stringprep",
        sum = "h1:kdwGpVNwPFtjs98xCGkHjQtGKh86rDcRZN17QEMCOIs=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonpointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xeipuuv/gojsonpointer",
        sum = "h1:zGWFAtiMcyryUHoUjUJX0/lt1H2+i2Ka2n+D3DImSNo=",
        version = "v0.0.0-20190905194746-02993c407bfb",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonreference",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xeipuuv/gojsonreference",
        sum = "h1:EzJWgHovont7NscjpAxXsDA8S8BMYve8Y5+7cuRE7R0=",
        version = "v0.0.0-20180127040603-bd5ef7bd5415",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonschema",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xeipuuv/gojsonschema",
        sum = "h1:LhYJRs+L4fBtjZUfuSZIKGeVu0QRy8e5Xi7D17UxZ74=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_xhit_go_str2duration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xhit/go-str2duration",
        sum = "h1:BcV5u025cITWxEQKGWr1URRzrcXtu7uk8+luz3Yuhwc=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_xi2_xz",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xi2/xz",
        sum = "h1:nIPpBwaJSVYIxUFsDv3M8ofmx9yWTog9BfvIu0q41lo=",
        version = "v0.0.0-20171230120015-48954b6210f8",
    )
    go_repository(
        name = "com_github_xiang90_probing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xiang90/probing",
        sum = "h1:eY9dn8+vbi4tKz5Qo6v2eYzo7kUS51QINcR5jNpbZS8=",
        version = "v0.0.0-20190116061207-43a291ad63a2",
    )
    go_repository(
        name = "com_github_xlab_treeprint",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xlab/treeprint",
        sum = "h1:G/1DjNkPpfZCFt9CSh6b5/nY4VimlbHF3Rh4obvtzDk=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_xordataexchange_crypt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xordataexchange/crypt",
        sum = "h1:ESFSdwYZvkeru3RtdrYueztKhOBCSAAzS4Gf+k0tEow=",
        version = "v0.0.3-0.20170626215501-b2862e3d0a77",
    )

    go_repository(
        name = "com_github_youmark_pkcs8",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/youmark/pkcs8",
        sum = "h1:splanxYIlg+5LfHAM6xpdFEAYOk8iySO56hMFq6uLyA=",
        version = "v0.0.0-20181117223130-1be2e3e5546d",
    )
    go_repository(
        name = "com_github_ysmood_fetchup",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ysmood/fetchup",
        sum = "h1:Qn8/q5uDW7szclt4sVXCFJ1TXup3hogz94OaLf6kloo=",
        version = "v0.2.2",
    )

    go_repository(
        name = "com_github_ysmood_goob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ysmood/goob",
        sum = "h1:HsxXhyLBeGzWXnqVKtmT9qM7EuVs/XOgkX7T6r1o1AQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_ysmood_gson",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ysmood/gson",
        sum = "h1:QFkWbTH8MxyUTKPkVWAENJhxqdBa4lYTQWqZCiLG6kE=",
        version = "v0.7.3",
    )
    go_repository(
        name = "com_github_ysmood_leakless",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ysmood/leakless",
        sum = "h1:BzLrVoiwxikpgEQR0Lk8NyBN5Cit2b1z+u0mgL4ZJak=",
        version = "v0.8.0",
    )

    go_repository(
        name = "com_github_yuin_goldmark",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/yuin/goldmark",
        sum = "h1:fVcFKWvrslecOb/tg+Cc05dkeYx540o0FuFt3nUVDoE=",
        version = "v1.4.13",
    )

    go_repository(
        name = "com_github_yvasiyarov_go_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/yvasiyarov/go-metrics",
        sum = "h1:+lm10QQTNSBd8DVTNGHx7o/IKu9HYDvLMffDhbyLccI=",
        version = "v0.0.0-20140926110328-57bccd1ccd43",
    )
    go_repository(
        name = "com_github_yvasiyarov_gorelic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/yvasiyarov/gorelic",
        sum = "h1:hlE8//ciYMztlGpl/VA+Zm1AcTPHYkHJPbHqE6WJUXE=",
        version = "v0.0.0-20141212073537-a9bba5b9ab50",
    )
    go_repository(
        name = "com_github_yvasiyarov_newrelic_platform_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/yvasiyarov/newrelic_platform_go",
        sum = "h1:ERexzlUfuTvpE74urLSbIQW0Z/6hF9t8U4NsJLaioAY=",
        version = "v0.0.0-20140908184405-b21fdbd4370f",
    )
    go_repository(
        name = "com_github_zalando_go_keyring",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zalando/go-keyring",
        sum = "h1:w2V9lcx/Uj4l+dzAf1m9s+DJ1O8ROkEHnynonHjTcYE=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_zclconf_go_cty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zclconf/go-cty",
        sum = "h1:It5dfKTTZHe9aeppbNOda3mN7Ag7sg6QkBNm6TkyFa0=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_github_zclconf_go_cty_debug",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zclconf/go-cty-debug",
        sum = "h1:FosyBZYxY34Wul7O/MSKey3txpPYyCqVO5ZyceuQJEI=",
        version = "v0.0.0-20191215020915-b22d67c1ba0b",
    )

    go_repository(
        name = "com_github_zeebo_xxh3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zeebo/xxh3",
        sum = "h1:xZmwmqxHZA8AI603jOQ0tMqmBr9lPeFwGg6d+xy9DC0=",
        version = "v1.0.2",
    )

    go_repository(
        name = "com_github_zmap_zcrypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zmap/zcrypto",
        sum = "h1:+nr36qrZEH0RIYNjcUEnOrCUdcSG3om2ANaFA6iSVWA=",
        version = "v0.0.0-20220402174210-599ec18ecbac",
    )
    go_repository(
        name = "com_github_zmap_zlint_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zmap/zlint/v3",
        sum = "h1:Xs/lrMJY74MpJx/jSx2oVvZBrqlyUyFaLLBRyf68cqg=",
        version = "v3.4.0",
    )
    go_repository(
        name = "com_google_cloud_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go",
        sum = "h1:Zc8gqp3+a9/Eyph2KDmcGaPtbKRIoqq4YTlL4NMD0Ys=",
        version = "v0.110.0",
    )
    go_repository(
        name = "com_google_cloud_go_accessapproval",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/accessapproval",
        sum = "h1:x0cEHro/JFPd7eS4BlEWNTMecIj2HdXjOVB5BtvwER0=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_accesscontextmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/accesscontextmanager",
        sum = "h1:MG60JgnEoawHJrbWw0jGdv6HLNSf6gQvYRiXpuzqgEA=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_aiplatform",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/aiplatform",
        sum = "h1:zTw+suCVchgZyO+k847wjzdVjWmrAuehxdvcZvJwfGg=",
        version = "v1.37.0",
    )
    go_repository(
        name = "com_google_cloud_go_analytics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/analytics",
        sum = "h1:LqAo3tAh2FU9+w/r7vc3hBjU23Kv7GhO/PDIW7kIYgM=",
        version = "v0.19.0",
    )
    go_repository(
        name = "com_google_cloud_go_apigateway",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigateway",
        sum = "h1:ZI9mVO7x3E9RK/BURm2p1aw9YTBSCQe3klmyP1WxWEg=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeconnect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigeeconnect",
        sum = "h1:sWOmgDyAsi1AZ48XRHcATC0tsi9SkPT7DA/+VCfkaeA=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigeeregistry",
        sum = "h1:E43RdhhCxdlV+I161gUY2rI4eOaMzHTA5kNkvRsFXvc=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_apikeys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apikeys",
        sum = "h1:B9CdHFZTFjVti89tmyXXrO+7vSNo2jvZuHG8zD5trdQ=",
        version = "v0.6.0",
    )

    go_repository(
        name = "com_google_cloud_go_appengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/appengine",
        sum = "h1:aBGDKmRIaRRoWJ2tAoN0oVSHoWLhtO9aj/NvUyP4aYs=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_area120",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/area120",
        sum = "h1:ugckkFh4XkHJMPhTIx0CyvdoBxmOpMe8rNs4Ok8GAag=",
        version = "v0.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_artifactregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/artifactregistry",
        sum = "h1:o1Q80vqEB6Qp8WLEH3b8FBLNUCrGQ4k5RFj0sn/sgO8=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_asset",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/asset",
        sum = "h1:YAsssO08BqZ6mncbb6FPlj9h6ACS7bJQUOlzciSfbNk=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_assuredworkloads",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/assuredworkloads",
        sum = "h1:VLGnVFta+N4WM+ASHbhc14ZOItOabDLH1MSoDv+Xuag=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_automl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/automl",
        sum = "h1:50VugllC+U4IGl3tDNcZaWvApHBTrn/TvyHDJ0wM+Uw=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_google_cloud_go_baremetalsolution",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/baremetalsolution",
        sum = "h1:2AipdYXL0VxMboelTTw8c1UJ7gYu35LZYUbuRv9Q28s=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_batch",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/batch",
        sum = "h1:YbMt0E6BtqeD5FvSv1d56jbVsWEzlGm55lYte+M6Mzs=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_beyondcorp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/beyondcorp",
        sum = "h1:UkY2BTZkEUAVrgqnSdOJ4p3y9ZRBPEe1LkjgC8Bj/Pc=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_bigquery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/bigquery",
        sum = "h1:RscMV6LbnAmhAzD893Lv9nXXy2WCaJmbxYPWDLbGqNQ=",
        version = "v1.50.0",
    )
    go_repository(
        name = "com_google_cloud_go_billing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/billing",
        sum = "h1:JYj28UYF5w6VBAh0gQYlgHJ/OD1oA+JgW29YZQU+UHM=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_binaryauthorization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/binaryauthorization",
        sum = "h1:d3pMDBCCNivxt5a4eaV7FwL7cSH0H7RrEnFrTb1QKWs=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_certificatemanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/certificatemanager",
        sum = "h1:5C5UWeSt8Jkgp7OWn2rCkLmYurar/vIWIoSQ2+LaTOc=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_channel",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/channel",
        sum = "h1:GpcQY5UJKeOekYgsX3QXbzzAc/kRGtBq43fTmyKe6Uw=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_google_cloud_go_cloudbuild",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/cloudbuild",
        sum = "h1:GHQCjV4WlPPVU/j3Rlpc8vNIDwThhd1U9qSY/NPZdko=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_clouddms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/clouddms",
        sum = "h1:E7v4TpDGUyEm1C/4KIrpVSOCTm0P6vWdHT0I4mostRA=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_cloudtasks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/cloudtasks",
        sum = "h1:uK5k6abf4yligFgYFnG0ni8msai/dSv6mDmiBulU0hU=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_compute",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/compute",
        sum = "h1:DcTwsFgGev/wV5+q8o2fzgcHOaac+DKGC91ZlvpsQds=",
        version = "v1.19.3",
    )
    go_repository(
        name = "com_google_cloud_go_compute_metadata",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/compute/metadata",
        sum = "h1:mg4jlk7mCAj6xXp9UJ4fjI9VUI5rubuGBW5aJ7UnBMY=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_google_cloud_go_contactcenterinsights",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/contactcenterinsights",
        sum = "h1:jXIpfcH/VYSE1SYcPzO0n1VVb+sAamiLOgCw45JbOQk=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_container",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/container",
        sum = "h1:NKlY/wCDapfVZlbVVaeuu2UZZED5Dy1z4Zx1KhEzm8c=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_google_cloud_go_containeranalysis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/containeranalysis",
        sum = "h1:EQ4FFxNaEAg8PqQCO7bVQfWz9NVwZCUKaM1b3ycfx3U=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_datacatalog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datacatalog",
        sum = "h1:4H5IJiyUE0X6ShQBqgFFZvGGcrwGVndTwUSLP4c52gw=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataflow",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataflow",
        sum = "h1:eYyD9o/8Nm6EttsKZaEGD84xC17bNgSKCu0ZxwqUbpg=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataform",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataform",
        sum = "h1:Dyk+fufup1FR6cbHjFpMuP4SfPiF3LI3JtoIIALoq48=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_datafusion",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datafusion",
        sum = "h1:sZjRnS3TWkGsu1LjYPFD/fHeMLZNXDK6PDHi2s2s/bk=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_datalabeling",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datalabeling",
        sum = "h1:ch4qA2yvddGRUrlfwrNJCr79qLqhS9QBwofPHfFlDIk=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataplex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataplex",
        sum = "h1:RvoZ5T7gySwm1CHzAw7yY1QwwqaGswunmqEssPxU/AM=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataproc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataproc",
        sum = "h1:W47qHL3W4BPkAIbk4SWmIERwsWBaNnWm0P2sdx3YgGU=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataqna",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataqna",
        sum = "h1:yFzi/YU4YAdjyo7pXkBE2FeHbgz5OQQBVDdbErEHmVQ=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_datastore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datastore",
        sum = "h1:iF6I/HaLs3Ado8uRKMvZRvF/ZLkWaWE9i8AiHzbC774=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_datastream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datastream",
        sum = "h1:BBCBTnWMDwwEzQQmipUXxATa7Cm7CA/gKjKcR2w35T0=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_deploy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/deploy",
        sum = "h1:otshdKEbmsi1ELYeCKNYppwV0UH5xD05drSdBm7ouTk=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_dialogflow",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dialogflow",
        sum = "h1:uVlKKzp6G/VtSW0E7IH1Y5o0H48/UOCmqksG2riYCwQ=",
        version = "v1.32.0",
    )
    go_repository(
        name = "com_google_cloud_go_dlp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dlp",
        sum = "h1:1JoJqezlgu6NWCroBxr4rOZnwNFILXr4cB9dMaSKO4A=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_documentai",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/documentai",
        sum = "h1:KM3Xh0QQyyEdC8Gs2vhZfU+rt6OCPF0dwVwxKgLmWfI=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_google_cloud_go_domains",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/domains",
        sum = "h1:2ti/o9tlWL4N+wIuWUNH+LbfgpwxPr8J1sv9RHA4bYQ=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_edgecontainer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/edgecontainer",
        sum = "h1:O0YVE5v+O0Q/ODXYsQHmHb+sYM8KNjGZw2pjX2Ws41c=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_google_cloud_go_errorreporting",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/errorreporting",
        sum = "h1:kj1XEWMu8P0qlLhm3FwcaFsUvXChV/OraZwA70trRR0=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_essentialcontacts",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/essentialcontacts",
        sum = "h1:gIzEhCoOT7bi+6QZqZIzX1Erj4SswMPIteNvYVlu+pM=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_eventarc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/eventarc",
        sum = "h1:fsJmNeqvqtk74FsaVDU6cH79lyZNCYP8Rrv7EhaB/PU=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_filestore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/filestore",
        sum = "h1:ckTEXN5towyTMu4q0uQ1Mde/JwTHur0gXs8oaIZnKfw=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_firestore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/firestore",
        sum = "h1:IBlRyxgGySXu5VuW0RgGFlTtLukSnNkpDiEOMkQkmpA=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_functions",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/functions",
        sum = "h1:pPDqtsXG2g9HeOQLoquLbmvmb82Y4Ezdo1GXuotFoWg=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_gaming",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gaming",
        sum = "h1:7vEhFnZmd931Mo7sZ6pJy7uQPDxF7m7v8xtBheG08tc=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkebackup",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkebackup",
        sum = "h1:za3QZvw6ujR0uyqkhomKKKNoXDyqYGPJies3voUK8DA=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkeconnect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkeconnect",
        sum = "h1:gXYKciHS/Lgq0GJ5Kc9SzPA35NGc3yqu6SkjonpEr2Q=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkehub",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkehub",
        sum = "h1:TqCSPsEBQ6oZSJgEYZ3XT8x2gUadbvfwI32YB0kuHCs=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkemulticloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkemulticloud",
        sum = "h1:8I84Q4vl02rJRsFiinBxl7WCozfdLlUVBQuSrqr9Wtk=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_grafeas",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/grafeas",
        sum = "h1:CYjC+xzdPvbV65gi6Dr4YowKcmLo045pm18L0DhdELM=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_gsuiteaddons",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gsuiteaddons",
        sum = "h1:1mvhXqJzV0Vg5Fa95QwckljODJJfDFXV4pn+iL50zzA=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_iam",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iam",
        sum = "h1:+CmB+K0J/33d0zSQ9SlFWUeCCEn5XJA0ZMZ3pHE9u8k=",
        version = "v0.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_iap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iap",
        sum = "h1:PxVHFuMxmSZyfntKXHXhd8bo82WJ+LcATenq7HLdVnU=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_ids",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/ids",
        sum = "h1:fodnCDtOXuMmS8LTC2y3h8t24U8F3eKWfhi+3LY6Qf0=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_iot",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iot",
        sum = "h1:39W5BFSarRNZfVG0eXI5LYux+OVQT8GkgpHCnrZL2vM=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/kms",
        sum = "h1:8UePKEypK3SQ6g+4mn/s/VgE5L7XOh+FwGGRUqvY3Hw=",
        version = "v1.10.2",
    )
    go_repository(
        name = "com_google_cloud_go_language",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/language",
        sum = "h1:7Ulo2mDk9huBoBi8zCE3ONOoBrL6UXfAI71CLQ9GEIM=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_lifesciences",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/lifesciences",
        sum = "h1:uWrMjWTsGjLZpCTWEAzYvyXj+7fhiZST45u9AgasasI=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_logging",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/logging",
        sum = "h1:CJYxlNNNNAMkHp9em/YEXcfJg+rPDg7YfwoRpMU+t5I=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_longrunning",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/longrunning",
        sum = "h1:v+yFJOfKC3yZdY6ZUI933pIYdhyhV8S3NpWrXWmg7jM=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_google_cloud_go_managedidentities",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/managedidentities",
        sum = "h1:ZRQ4k21/jAhrHBVKl/AY7SjgzeJwG1iZa+mJ82P+VNg=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_maps",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/maps",
        sum = "h1:mv9YaczD4oZBZkM5XJl6fXQ984IkJNHPwkc8MUsdkBo=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_mediatranslation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/mediatranslation",
        sum = "h1:anPxH+/WWt8Yc3EdoEJhPMBRF7EhIdz426A+tuoA0OU=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_memcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/memcache",
        sum = "h1:8/VEmWCpnETCrBwS3z4MhT+tIdKgR1Z4Tr2tvYH32rg=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_metastore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/metastore",
        sum = "h1:QCFhZVe2289KDBQ7WxaHV2rAmPrmRAdLC6gbjUd3HPo=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_monitoring",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/monitoring",
        sum = "h1:2qsrgXGVoRXpP7otZ14eE1I568zAa92sJSDPyOJvwjM=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkconnectivity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networkconnectivity",
        sum = "h1:ZD6b4Pk1jEtp/cx9nx0ZYcL3BKqDa+KixNDZ6Bjs1B8=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkmanagement",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networkmanagement",
        sum = "h1:8KWEUNGcpSX9WwZXq7FtciuNGPdPdPN/ruDm769yAEM=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_networksecurity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networksecurity",
        sum = "h1:sOc42Ig1K2LiKlzG71GUVloeSJ0J3mffEBYmvu+P0eo=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_notebooks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/notebooks",
        sum = "h1:Kg2K3K7CbSXYJHZ1aGQpf1xi5x2GUvQWf2sFVuiZh8M=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_optimization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/optimization",
        sum = "h1:dj8O4VOJRB4CUwZXdmwNViH1OtI0WtWL867/lnYH248=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_google_cloud_go_orchestration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/orchestration",
        sum = "h1:Vw+CEXo8M/FZ1rb4EjcLv0gJqqw89b7+g+C/EmniTb8=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_orgpolicy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/orgpolicy",
        sum = "h1:XDriMWug7sd0kYT1QKofRpRHzjad0bK8Q8uA9q+XrU4=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_osconfig",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/osconfig",
        sum = "h1:PkSQx4OHit5xz2bNyr11KGcaFccL5oqglFPdTboyqwQ=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_oslogin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/oslogin",
        sum = "h1:whP7vhpmc+ufZa90eVpkfbgzJRK/Xomjz+XCD4aGwWw=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_phishingprotection",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/phishingprotection",
        sum = "h1:l6tDkT7qAEV49MNEJkEJTB6vOO/onbSOcNtAT09HPuA=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_policytroubleshooter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/policytroubleshooter",
        sum = "h1:yKAGC4p9O61ttZUswaq9GAn1SZnEzTd0vUYXD7ZBT7Y=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_privatecatalog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/privatecatalog",
        sum = "h1:EPEJ1DpEGXLDnmc7mnCAqFmkwUJbIsaLAiLHVOkkwtc=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_profiler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/profiler",
        sum = "h1:b5got9Be9Ia0HVvyt7PavWxXEht15B9lWnigdvHtxOc=",
        version = "v0.3.1",
    )

    go_repository(
        name = "com_google_cloud_go_pubsub",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/pubsub",
        sum = "h1:vCge8m7aUKBJYOgrZp7EsNDf6QMd2CAlXZqWTn3yq6s=",
        version = "v1.30.0",
    )
    go_repository(
        name = "com_google_cloud_go_pubsublite",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/pubsublite",
        sum = "h1:cb9fsrtpINtETHiJ3ECeaVzrfIVhcGjhhJEjybHXHao=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_recaptchaenterprise",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recaptchaenterprise",
        sum = "h1:u6EznTGzIdsyOsvm+Xkw0aSuKFXQlyjGE9a4exk6iNQ=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_google_cloud_go_recaptchaenterprise_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recaptchaenterprise/v2",
        sum = "h1:6iOCujSNJ0YS7oNymI64hXsjGq60T4FK1zdLugxbzvU=",
        version = "v2.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_recommendationengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recommendationengine",
        sum = "h1:VibRFCwWXrFebEWKHfZAt2kta6pS7Tlimsnms0fjv7k=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_recommender",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recommender",
        sum = "h1:ZnFRY5R6zOVk2IDS1Jbv5Bw+DExCI5rFumsTnMXiu/A=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_redis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/redis",
        sum = "h1:JoAd3SkeDt3rLFAAxEvw6wV4t+8y4ZzfZcZmddqphQ8=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_resourcemanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/resourcemanager",
        sum = "h1:NRM0p+RJkaQF9Ee9JMnUV9BQ2QBIOq/v8M+Pbv/wmCs=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_resourcesettings",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/resourcesettings",
        sum = "h1:8Dua37kQt27CCWHm4h/Q1XqCF6ByD7Ouu49xg95qJzI=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_retail",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/retail",
        sum = "h1:1Dda2OpFNzIb4qWgFZjYlpP7sxX3aLeypKG6A3H4Yys=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_google_cloud_go_run",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/run",
        sum = "h1:ydJQo+k+MShYnBfhaRHSZYeD/SQKZzZLAROyfpeD9zw=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_scheduler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/scheduler",
        sum = "h1:NpQAHtx3sulByTLe2dMwWmah8PWgeoieFPpJpArwFV0=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_secretmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/secretmanager",
        sum = "h1:9QwQ3oMurvmPEmM80spGe2SFGDa+RRgkLIdTm3gMWO8=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_security",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/security",
        sum = "h1:PYvDxopRQBfYAXKAuDpFCKBvDOWPWzp9k/H5nB3ud3o=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_securitycenter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/securitycenter",
        sum = "h1:AF3c2s3awNTMoBtMX3oCUoOMmGlYxGOeuXSYHNBkf14=",
        version = "v1.19.0",
    )
    go_repository(
        name = "com_google_cloud_go_servicecontrol",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/servicecontrol",
        sum = "h1:d0uV7Qegtfaa7Z2ClDzr9HJmnbJW7jn0WhZ7wOX6hLE=",
        version = "v1.11.1",
    )
    go_repository(
        name = "com_google_cloud_go_servicedirectory",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/servicedirectory",
        sum = "h1:SJwk0XX2e26o25ObYUORXx6torSFiYgsGkWSkZgkoSU=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_servicemanagement",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/servicemanagement",
        sum = "h1:fopAQI/IAzlxnVeiKn/8WiV6zKndjFkvi+gzu+NjywY=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_serviceusage",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/serviceusage",
        sum = "h1:rXyq+0+RSIm3HFypctp7WoXxIA563rn206CfMWdqXX4=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_shell",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/shell",
        sum = "h1:wT0Uw7ib7+AgZST9eCDygwTJn4+bHMDtZo5fh7kGWDU=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_spanner",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/spanner",
        sum = "h1:vHFqBMuPdTCwA8b9+IyQbGppQoqx7xJfcSa81d7gtAk=",
        version = "v1.45.1",
    )
    go_repository(
        name = "com_google_cloud_go_speech",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/speech",
        sum = "h1:JEVoWGNnTF128kNty7T4aG4eqv2z86yiMJPT9Zjp+iw=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_google_cloud_go_storage",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/storage",
        sum = "h1:uOdMxAs8HExqBlnLtnQyP0YkvbiDpdGShGKtx6U/oNM=",
        version = "v1.30.1",
    )
    go_repository(
        name = "com_google_cloud_go_storagetransfer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/storagetransfer",
        sum = "h1:5T+PM+3ECU3EY2y9Brv0Sf3oka8pKmsCfpQ07+91G9o=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_talent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/talent",
        sum = "h1:nI9sVZPjMKiO2q3Uu0KhTDVov3Xrlpt63fghP9XjyEM=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_texttospeech",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/texttospeech",
        sum = "h1:H4g1ULStsbVtalbZGktyzXzw6jP26RjVGYx9RaYjBzc=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_tpu",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/tpu",
        sum = "h1:/34T6CbSi+kTv5E19Q9zbU/ix8IviInZpzwz3rsFE+A=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_trace",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/trace",
        sum = "h1:olxC0QHC59zgJVALtgqfD9tGk0lfeCP5/AGXL3Px/no=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_translate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/translate",
        sum = "h1:GvLP4oQ4uPdChBmBaUSa/SaZxCdyWELtlAaKzpHsXdA=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_video",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/video",
        sum = "h1:upIbnGI0ZgACm58HPjAeBMleW3sl5cT84AbYQ8PWOgM=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_google_cloud_go_videointelligence",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/videointelligence",
        sum = "h1:Uh5BdoET8XXqXX2uXIahGb+wTKbLkGH7s4GXR58RrG8=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_vision",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vision",
        sum = "h1:/CsSTkbmO9HC8iQpxbK8ATms3OQaX3YQUeTMGCxlaK4=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_vision_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vision/v2",
        sum = "h1:8C8RXUJoflCI4yVdqhTy9tRyygSHmp60aP363z23HKg=",
        version = "v2.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_vmmigration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vmmigration",
        sum = "h1:Azs5WKtfOC8pxvkyrDvt7J0/4DYBch0cVbuFfCCFt5k=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_vmwareengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vmwareengine",
        sum = "h1:b0NBu7S294l0gmtrT0nOJneMYgZapr5x9tVWvgDoVEM=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_vpcaccess",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vpcaccess",
        sum = "h1:FOe6CuiQD3BhHJWt7E8QlbBcaIzVRddupwJlp7eqmn4=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_google_cloud_go_webrisk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/webrisk",
        sum = "h1:IY+L2+UwxcVm2zayMAtBhZleecdIFLiC+QJMzgb0kT0=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_websecurityscanner",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/websecurityscanner",
        sum = "h1:AHC1xmaNMOZtNqxI9Rmm87IJEyPaRkOxeI0gpAacXGk=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_workflows",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/workflows",
        sum = "h1:FfGp9w0cYnaKZJhUOMqCOJCYT/WlvYBfTQhFWV3sRKI=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_shuralyov_dmitri_gpu_mtl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "dmitri.shuralyov.com/gpu/mtl",
        sum = "h1:VpgP7xuJadIUuKccphEpTJnWhS2jkQyMt6Y7pJCD7fY=",
        version = "v0.0.0-20190408044501-666a987793e9",
    )
    go_repository(
        name = "com_sourcegraph_sourcegraph_appdash",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sourcegraph.com/sourcegraph/appdash",
        sum = "h1:ucqkfpjg9WzSUubAO62csmucvxl4/JeW3F4I4909XkM=",
        version = "v0.0.0-20190731080439-ebfcffb1b5c0",
    )

    go_repository(
        name = "dev_gocloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gocloud.dev",
        sum = "h1:fBy0jwJSmxs0IjT0fE32MO+Mj+307VZQwyHaTyFZbC4=",
        version = "v0.29.0",
    )
    go_repository(
        name = "in_gopkg_alecthomas_kingpin_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/alecthomas/kingpin.v2",
        sum = "h1:jMFz6MfLP0/4fUyZle81rXUoxOBFi19VUFKVDOQfozc=",
        version = "v2.2.6",
    )
    go_repository(
        name = "in_gopkg_alexcesaro_statsd_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/alexcesaro/statsd.v2",
        sum = "h1:FXkZSCZIH17vLCO5sO2UucTHsH9pc+17F6pl3JVCwMc=",
        version = "v2.0.0",
    )
    go_repository(
        name = "in_gopkg_check_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/check.v1",
        sum = "h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=",
        version = "v1.0.0-20201130134442-10cb98267c6c",
    )
    go_repository(
        name = "in_gopkg_cheggaaa_pb_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/cheggaaa/pb.v1",
        sum = "h1:n1tBJnnK2r7g9OW2btFH91V92STTUevLXYFb8gy9EMk=",
        version = "v1.0.28",
    )
    go_repository(
        name = "in_gopkg_errgo_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/errgo.v2",
        sum = "h1:0vLT13EuvQ0hNvakwLuFZ/jYrLp5F3kcWHXdRggjCE8=",
        version = "v2.1.0",
    )
    go_repository(
        name = "in_gopkg_fsnotify_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/fsnotify.v1",
        sum = "h1:xOHLXZwVvI9hhs+cLKq5+I5onOuwQLhQwiu63xxlHs4=",
        version = "v1.4.7",
    )
    go_repository(
        name = "in_gopkg_gcfg_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/gcfg.v1",
        sum = "h1:m8OOJ4ccYHnx2f4gQwpno8nAX5OGOh7RLaaz0pj3Ogs=",
        version = "v1.2.3",
    )
    go_repository(
        name = "in_gopkg_go_playground_assert_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/go-playground/assert.v1",
        sum = "h1:xoYuJVE7KT85PYWrN730RguIQO0ePzVRfFMXadIrXTM=",
        version = "v1.2.1",
    )

    go_repository(
        name = "in_gopkg_go_playground_validator_v9",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/go-playground/validator.v9",
        sum = "h1:SvGtYmN60a5CVKTOzMSyfzWDeZRxRuGvRQyEAKbw1xc=",
        version = "v9.29.1",
    )

    go_repository(
        name = "in_gopkg_inf_v0",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/inf.v0",
        sum = "h1:73M5CoZyi3ZLMOyDlQh031Cx6N9NDJ2Vvfl76EDAgDc=",
        version = "v0.9.1",
    )
    go_repository(
        name = "in_gopkg_ini_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/ini.v1",
        sum = "h1:Dgnx+6+nfE+IfzjUEISNeydPJh9AXNNsWbGP9KzCsOA=",
        version = "v1.67.0",
    )

    go_repository(
        name = "in_gopkg_natefinch_lumberjack_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/natefinch/lumberjack.v2",
        sum = "h1:1Lc07Kr7qY4U2YPouBjpCLxpiyxIVoxqXgkXLknAOE8=",
        version = "v2.0.0",
    )
    go_repository(
        name = "in_gopkg_resty_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/resty.v1",
        sum = "h1:CuXP0Pjfw9rOuY6EP+UvtNvt5DSqHpIxILZKT/quCZI=",
        version = "v1.12.0",
    )
    go_repository(
        name = "in_gopkg_retry_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/retry.v1",
        sum = "h1:a9CArYczAVv6Qs6VGoLMio99GEs7kY9UzSF9+LD+iGs=",
        version = "v1.0.3",
    )

    go_repository(
        name = "in_gopkg_square_go_jose_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/square/go-jose.v2",
        sum = "h1:NGk74WTnPKBNUhNzQX7PYcTLUjoq7mzKk2OKbvwk2iI=",
        version = "v2.6.0",
    )
    go_repository(
        name = "in_gopkg_src_d_go_billy_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/src-d/go-billy.v4",
        sum = "h1:0SQA1pRztfTFx2miS8sA97XvooFeNOmvUenF4o0EcVg=",
        version = "v4.3.2",
    )
    go_repository(
        name = "in_gopkg_src_d_go_git_fixtures_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/src-d/go-git-fixtures.v3",
        sum = "h1:ivZFOIltbce2Mo8IjzUHAFoq/IylO9WHhNOAJK+LsJg=",
        version = "v3.5.0",
    )
    go_repository(
        name = "in_gopkg_src_d_go_git_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/src-d/go-git.v4",
        sum = "h1:SRtFyV8Kxc0UP7aCHcijOMQGPxHSmMOPrzulQWolkYE=",
        version = "v4.13.1",
    )
    go_repository(
        name = "in_gopkg_tomb_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/tomb.v1",
        sum = "h1:uRGJdciOHaEIrze2W8Q3AKkepLTh2hOroT7a+7czfdQ=",
        version = "v1.0.0-20141024135613-dd632973f1e7",
    )
    go_repository(
        name = "in_gopkg_warnings_v0",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/warnings.v0",
        sum = "h1:wFXVbFY8DY5/xOe1ECiWdKCzZlxgshcYVNkBHstARME=",
        version = "v0.1.2",
    )
    go_repository(
        name = "in_gopkg_yaml_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=",
        version = "v2.4.0",
    )
    go_repository(
        name = "in_gopkg_yaml_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=",
        version = "v3.0.1",
    )
    go_repository(
        name = "io_etcd_go_bbolt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/bbolt",
        sum = "h1:j+zJOnnEjF/kyHlDDgGnVL/AIqIJPq8UoB2GSNfkUfQ=",
        version = "v1.3.7",
    )
    go_repository(
        name = "io_etcd_go_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd",
        sum = "h1:VcrIfasaLFkyjk6KNlXQSzO+B0fZcnECiDrKJsfxka0=",
        version = "v0.0.0-20191023171146-3cf2f69b5738",
    )

    go_repository(
        name = "io_etcd_go_etcd_api_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/api/v3",
        sum = "h1:Zf44zJszoU7zRV0X/nStPenegNXoFDWcB/MwrJbA+L4=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_pkg_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/pkg/v3",
        sum = "h1:tPp9YRn/UBFAHdhOQUII9eUs7aOK35eulpMhX4YBd+M=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/v2",
        sum = "h1:IGp9Ozt8awy3qRTXSIYJd/o/cr4oUyrm9MF1RJ2dr/c=",
        version = "v2.305.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/v3",
        sum = "h1:B6ngTKZSWWowHEoaucOKHQR/AtZKaoHLiUpWxOLG4l4=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdctl_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/etcdctl/v3",
        sum = "h1:Hc+yvk0e1VOKSzNNA0ZyNqzOQf/jrAZarx/IFd6KcPk=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdutl_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/etcdutl/v3",
        sum = "h1:KTVdvP+9unMtYifGPcIQRMOhsbhek78W8DiCoAbYbQE=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_pkg_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/pkg/v3",
        sum = "h1:hz6w5Cb4p7dbt642m8Y35Ts9yWPWUCymc3v4Z/aiGEU=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_raft_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/raft/v3",
        sum = "h1:wM4IAfiY1+vrCAkUicIOzkyjpV9MawnAul2KvxeMgy4=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_server_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/server/v3",
        sum = "h1:eK9fU6Pd6IJD1k0u4zAq1NZsSsEOOimlP3kIkpcQrho=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_tests_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/tests/v3",
        sum = "h1:jrWlrEJfFuvI9nqMGFZNLc+OCdtyBdw8mSfp8ZwLIWs=",
        version = "v3.5.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/v3",
        sum = "h1:AQF+i8f8jLeli9Cmf6SIs3KNryFPnzlIzI7weyK784g=",
        version = "v3.5.8",
    )

    go_repository(
        name = "io_filippo_edwards25519",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "filippo.io/edwards25519",
        sum = "h1:0wAIcmJUqRdI8IJ/3eGi5/HwXZWPujYXXlkrQogz0Ek=",
        version = "v1.0.0",
    )
    go_repository(
        name = "io_gitea_code_sdk_gitea",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "code.gitea.io/sdk/gitea",
        sum = "h1:CdI3J82Mqn0mElyEKa5DUSr3Wi2R+qm/6uVtCkSSqSM=",
        version = "v0.11.3",
    )

    go_repository(
        name = "io_k8s_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/api",
        sum = "h1:+H17AJpUMvl+clT+BPnKf0E3ksMAzoBBg7CntpSuADo=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_apiextensions_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiextensions-apiserver",
        sum = "h1:iwhyoeS4xj9Y7v8YExhUwbVuBhMr3Q4bd/laClBV6Bo=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_apimachinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apimachinery",
        sum = "h1:vBjGaKKieaIreI+oQwELalVG4d8f3YAMNpWLzDXkxeg=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiserver",
        sum = "h1:p+tjwrcQEZDrEorCZV2/qE8osGTINPuS5ZNqWAvKm5E=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_cli_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cli-runtime",
        sum = "h1:9HI8gfReNujKXt16tGOAnb8b4NZ5E+e0mQQHKhFGwYw=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_client_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/client-go",
        sum = "h1:vDLSeuYvCHKeoQRhCXjxXO45nHVv2Ip4Fe0MfioMrhE=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_cloud_provider",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cloud-provider",
        replace = "k8s.io/cloud-provider",
        sum = "h1:IiQWyFtdzcPOqvrBZE9FCt0CDCx3GUcZhKkykEgKlM4=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_cluster_bootstrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cluster-bootstrap",
        sum = "h1:OL3onrOwrUD7NQxBUqQwTl1Uu2GQKCkw9BMHpc4PbiA=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_code_generator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/code-generator",
        sum = "h1:RmK0CnU5qRaK6WRtSyWNODmfTZNoJbrizpVcsgbtrvI=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_component_base",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/component-base",
        sum = "h1:neju+7s/r5O4x4/txeUONNTS9r1HsPbyoPBAtHsDCpo=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_component_helpers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/component-helpers",
        sum = "h1:i9TgWJ6TH8lQ9x4ExHOwhVitrRpBOr7Wn8aZLbBWxkc=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_controller_manager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/controller-manager",
        replace = "k8s.io/controller-manager",
        sum = "h1:S7984FVb5ajp8YqMQGAm8zXEUEl0Omw6FJlOiQU2Ne8=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_cri_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cri-api",
        sum = "h1:Vifw8T4ZFzU5pQ5dj5rdDsPOSzmLAvhVcYEJpbjOYLY=",
        version = "v0.26.2",
    )
    go_repository(
        name = "io_k8s_csi_translation_lib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/csi-translation-lib",
        replace = "k8s.io/csi-translation-lib",
        sum = "h1:HbwiOk+M3jIkTC+e5nxUCwmux68OguKV/g9NaHDQhzs=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_dynamic_resource_allocation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/dynamic-resource-allocation",
        replace = "k8s.io/dynamic-resource-allocation",
        sum = "h1:lNt4YOVoJqi+wcBesTVJ3KAfr3HnvLedO1/ZovE26pk=",
        version = "v0.27.2",
    )

    go_repository(
        name = "io_k8s_gengo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/gengo",
        sum = "h1:U9tB195lKdzwqicbJvyJeOXV7Klv+wNAWENRnXEGi08=",
        version = "v0.0.0-20220902162205-c0856e24416d",
    )

    go_repository(
        name = "io_k8s_klog_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/klog/v2",
        sum = "h1:7WCHKK6K8fNhTqfBhISHQ97KrnJNFZMcQvKp7gP/tmg=",
        version = "v2.100.1",
    )
    go_repository(
        name = "io_k8s_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kms",
        sum = "h1:wCdmPCa3kubcVd3AssOeaVjLQSu45k5g/vruJ3iqwDU=",
        version = "v0.27.2",
    )

    go_repository(
        name = "io_k8s_kube_aggregator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-aggregator",
        replace = "k8s.io/kube-aggregator",
        sum = "h1:jfHoPip+qN/fn3OcrYs8/xMuVYvkJHKo0H0DYciqdns=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_kube_controller_manager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-controller-manager",
        replace = "k8s.io/kube-controller-manager",
        sum = "h1:+sPNPN0Fyhycd8iRwpV+zG3eL/uAlekWihgOAZxGZs0=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_kube_openapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-openapi",
        sum = "h1:2kWPakN3i/k81b0gvD5C5FJ2kxm1WrQFanWchyKuqGg=",
        version = "v0.0.0-20230501164219-8b0f38b5fd1f",
    )
    go_repository(
        name = "io_k8s_kube_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-proxy",
        replace = "k8s.io/kube-proxy",
        sum = "h1:nb/ASUpYoXlueURXnY+O2IZkCZmIYOnDprFEeiwwOCY=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_kube_scheduler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-scheduler",
        replace = "k8s.io/kube-scheduler",
        sum = "h1:ZsN8meIkmJ+wnFrvhi5YzIbueBeBz2xx4I/0cKgpnlg=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_kubectl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubectl",
        sum = "h1:sSBM2j94MHBFRWfHIWtEXWCicViQzZsb177rNsKBhZg=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_kubelet",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubelet",
        sum = "h1:vpJnBkqQjxItEhehKG0toXoZ+G+tf4UXAOqtMJy6qgc=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_kubernetes",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubernetes",
        sum = "h1:g4v9oY6u7vBUDEuq4FvC50Bbw2K7GZuvM00IIESWVf4=",
        version = "v1.27.2",
    )
    go_repository(
        name = "io_k8s_legacy_cloud_providers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/legacy-cloud-providers",
        replace = "k8s.io/legacy-cloud-providers",
        sum = "h1:4D56C4lm+Byu4z34f0sGBkMFlUWpPUqYjaawIrXaGZQ=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/metrics",
        sum = "h1:TD6z3dhhN9bgg5YkbTh72bPiC1BsxipBLPBWyC3VQAU=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_mount_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/mount-utils",
        sum = "h1:fEqtBdAv88xpoPr3nR0MgYs6P+2PjXyUTwd4NmqSBjY=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_pod_security_admission",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/pod-security-admission",
        replace = "k8s.io/pod-security-admission",
        sum = "h1:dSGK0ftJwJNHSp5fMAwVuFIMMY1MlzW4k82mjar6G8I=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_sample_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/sample-apiserver",
        replace = "k8s.io/sample-apiserver",
        sum = "h1:fF3AvtOh/D3HOoIIKC+PIkNHyZJP2uy8Wq/CXiOLXQw=",
        version = "v0.27.2",
    )
    go_repository(
        name = "io_k8s_sigs_apiserver_network_proxy_konnectivity_client",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/apiserver-network-proxy/konnectivity-client",
        sum = "h1:trsWhjU5jZrx6UvFu4WzQDrN7Pga4a7Qg+zcfcj64PA=",
        version = "v0.1.2",
    )
    go_repository(
        name = "io_k8s_sigs_controller_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/controller-runtime",
        sum = "h1:ML+5Adt3qZnMSYxZ7gAverBLNPSMQEibtzAgp0UPojU=",
        version = "v0.15.0",
    )
    go_repository(
        name = "io_k8s_sigs_json",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/json",
        sum = "h1:EDPBXCAspyGV4jQlpZSudPeMmr1bNJefnuqLsRAsHZo=",
        version = "v0.0.0-20221116044647-bc3834ca7abd",
    )
    go_repository(
        name = "io_k8s_sigs_kustomize_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/kustomize/api",
        sum = "h1:kejWfLeJhUsTGioDoFNJET5LQe/ajzXhJGYoU+pJsiA=",
        version = "v0.13.2",
    )

    go_repository(
        name = "io_k8s_sigs_kustomize_kustomize_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/kustomize/kustomize/v5",
        sum = "h1:HWXbyKDNwGqol+s/sMNr/vnfNME/EoMdEraP4ZkUQek=",
        version = "v5.0.1",
    )

    go_repository(
        name = "io_k8s_sigs_kustomize_kyaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/kustomize/kyaml",
        sum = "h1:c8iibius7l24G2wVAGZn/Va2wNys03GXLjYVIcFVxKA=",
        version = "v0.14.1",
    )
    go_repository(
        name = "io_k8s_sigs_release_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/release-utils",
        sum = "h1:17LmJrydpUloTCtaoWj95uKlcrUp4h2A9Sa+ZL+lV9w=",
        version = "v0.7.4",
    )
    go_repository(
        name = "io_k8s_sigs_structured_merge_diff_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/structured-merge-diff/v4",
        sum = "h1:PRbqxJClWWYMNV1dhaG4NsibJbArud9kFxnAMREiWFE=",
        version = "v4.2.3",
    )
    go_repository(
        name = "io_k8s_sigs_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:a2VclLzOGrwOHDiV8EfBGhvjHvP46CtW5j6POvhYGGo=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_k8s_system_validators",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/system-validators",
        sum = "h1:tq05tdO9zdJZnNF3SXrq6LE7Knc/KfJm5wk68467JDg=",
        version = "v1.8.0",
    )
    go_repository(
        name = "io_k8s_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/utils",
        sum = "h1:EObNQ3TW2D+WptiYXlApGNLVy0zm/JIBVY9i+M4wpAU=",
        version = "v0.0.0-20230505201702-9f6742963106",
    )

    go_repository(
        name = "io_opencensus_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opencensus.io",
        sum = "h1:y73uSU6J157QMP2kn2r30vwW1A2W2WFwSCGnAVxeaD0=",
        version = "v0.24.0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_aws",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "contrib.go.opencensus.io/exporter/aws",
        sum = "h1:YsbWYxDZkC7x2OxlsDEYvvEXZ3cBI3qBgUK5BqkZvRw=",
        version = "v0.0.0-20181029163544-2befc13012d0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_ocagent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "contrib.go.opencensus.io/exporter/ocagent",
        sum = "h1:TKXjQSRS0/cCDrP7KvkgU6SmILtF/yV2TOs/02K/WZQ=",
        version = "v0.5.0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_stackdriver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "contrib.go.opencensus.io/exporter/stackdriver",
        sum = "h1:zBakwHardp9Jcb8sQHcHpXy/0+JIb1M8KjigCJzx7+4=",
        version = "v0.13.14",
    )
    go_repository(
        name = "io_opencensus_go_contrib_integrations_ocsql",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "contrib.go.opencensus.io/integrations/ocsql",
        sum = "h1:kfg5Yyy1nYUrqzyfW5XX+dzMASky8IJXhtHe0KTYNS4=",
        version = "v0.1.4",
    )
    go_repository(
        name = "io_opencensus_go_contrib_resource",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "contrib.go.opencensus.io/resource",
        sum = "h1:4r2CANuYhKGmYWP02+5E94rLRcS/YeD+KlxSrOsMxk0=",
        version = "v0.1.1",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/contrib",
        sum = "h1:ubFQUn0VCZ0gPwIoJfBJVpeBlyRMxu8Mm/huKWYd9p0=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_github_com_emicklei_go_restful_otelrestful",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful",
        sum = "h1:KQjX0qQ8H21oBUAvFp4ZLKJMMLIluONvSPDAFIGmX58=",
        version = "v0.35.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_google_golang_org_grpc_otelgrpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
        sum = "h1:5jD3teb4Qh7mx/nfzq4jO2WFFpvXD0vYWFDrdvNWmXk=",
        version = "v0.40.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_net_http_otelhttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
        sum = "h1:sxoY9kG1s1WpSYNyzm24rlwH4lnRYFXUVVBmKMBfRgw=",
        version = "v0.35.1",
    )
    go_repository(
        name = "io_opentelemetry_go_otel",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel",
        sum = "h1:/79Huy8wbf5DnIPhemGB+zEPVwnN6fuQybr/SRXa6hM=",
        version = "v1.14.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp",
        sum = "h1:PTNgq9MRmQqqJY0REVbZFvwkYOA85vbdQU/nVfxDyqg=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_internal_retry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/internal/retry",
        sum = "h1:/fXHZHGvro6MVqV34fJzDhi7sHGpX3Ej/Qjmfn003ho=",
        version = "v1.14.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace",
        sum = "h1:TKf2uAs2ueguzLaxOCBXNpHxfO/aC7PAdDsSH0IbeRQ=",
        version = "v1.14.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace_otlptracegrpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc",
        sum = "h1:ap+y8RXX3Mu9apKVtOkM6WSFESLM8K3wNQyOU8sWHcc=",
        version = "v1.14.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace_otlptracehttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp",
        sum = "h1:3jAYbRHQAqzLjd9I4tzxwJ8Pk/N6AqBcF6m1ZHrxG94=",
        version = "v1.14.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_metric",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/metric",
        sum = "h1:pHDQuLQOZwYD+Km0eb657A25NaRzy0a+eLyKfDXedEs=",
        version = "v0.37.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_oteltest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/oteltest",
        sum = "h1:HiITxCawalo5vQzdHfKeZurV8x7ljcqAgiWzF6Vaeaw=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/sdk",
        sum = "h1:PDCppFRDq8A1jL9v6KMI6dYesaq+DFcDZvjsoGvxGzY=",
        version = "v1.14.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk_export_metric",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/sdk/export/metric",
        sum = "h1:c5VRjxCXdQlx1HjzwGdQHzZaVI82b5EbBgOu2ljD92g=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk_metric",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/sdk/metric",
        sum = "h1:7ao1wpzHRVKf0OQ7GIxiQJA6X7DLX9o14gmVon7mMK8=",
        version = "v0.20.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_trace",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/trace",
        sum = "h1:wp2Mmvj41tDsyAJXiWDWpfNsOiIyd38fy85pyKcFq/M=",
        version = "v1.14.0",
    )
    go_repository(
        name = "io_opentelemetry_go_proto_otlp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/proto/otlp",
        sum = "h1:IVN6GR+mhC4s5yfcTbmzHYODqvWAp3ZedA2SJPI1Nnw=",
        version = "v0.19.0",
    )
    go_repository(
        name = "io_rsc_binaryregexp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "rsc.io/binaryregexp",
        sum = "h1:HfqmD5MEmC0zvwBuF187nq9mdnXjXsSivRiXN7SmRkE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "io_rsc_pdf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "rsc.io/pdf",
        sum = "h1:k1MczvYDUvJBe93bYd7wrZLLUEcLZAuF824/I4e5Xr4=",
        version = "v0.1.1",
    )

    go_repository(
        name = "io_rsc_quote_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "rsc.io/quote/v3",
        sum = "h1:9JKUTTIUgS6kzR9mK1YuGKv6Nl+DijDNIc0ghT58FaY=",
        version = "v3.1.0",
    )
    go_repository(
        name = "io_rsc_sampler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "rsc.io/sampler",
        sum = "h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/QiW4=",
        version = "v1.3.0",
    )
    go_repository(
        name = "land_oras_oras_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "oras.land/oras-go",
        sum = "h1:v8PJl+gEAntI1pJ/LCrDgsuk+1PKVavVEPsYIHFE5uY=",
        version = "v1.2.3",
    )
    go_repository(
        name = "net_starlark_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.starlark.net",
        sum = "h1:ghIB+2LQvihWROIGpcAVPq/ce5O2uMQersgxXiOeTS4=",
        version = "v0.0.0-20220223235035-243c74974e97",
    )
    go_repository(
        name = "org_bazil_fuse",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "bazil.org/fuse",
        sum = "h1:FNCRpXiquG1aoyqcIWVFmpTSKVcx2bQD38uZZeGtdlw=",
        version = "v0.0.0-20180421153158-65cc252bf669",
    )
    go_repository(
        name = "org_bitbucket_bertimus9_systemstat",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "bitbucket.org/bertimus9/systemstat",
        sum = "h1:n0aLnh2Jo4nBUBym9cE5PJDG8GT6g+4VuS2Ya2jYYpA=",
        version = "v0.5.0",
    )
    go_repository(
        name = "org_bitbucket_creachadair_shell",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "bitbucket.org/creachadair/shell",
        sum = "h1:Z96pB6DkSb7F3Y3BBnJeOZH2gazyMTWlvecSD4vDqfk=",
        version = "v0.0.7",
    )
    go_repository(
        name = "org_cloudfoundry_code_clock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "code.cloudfoundry.org/clock",
        sum = "h1:5eeuG0BHx1+DHeT3AP+ISKZ2ht1UjGhm581ljqYpVeQ=",
        version = "v0.0.0-20180518195852-02e53af36e6c",
    )

    go_repository(
        name = "org_golang_google_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/api",
        sum = "h1:zDobeejm3E7pEG1mNHvdxvjs5XJoCMzyNH+CmwL94Es=",
        version = "v0.122.0",
    )
    go_repository(
        name = "org_golang_google_appengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/appengine",
        sum = "h1:FZR1q0exgwxzPzp/aF+VccGrSfxfPpkBqjIIEq3ru6c=",
        version = "v1.6.7",
    )
    go_repository(
        name = "org_golang_google_genproto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto",
        sum = "h1:KpwkzHKEF7B9Zxg18WzOa7djJ+Ha5DzthMyZYQfEn2A=",
        version = "v0.0.0-20230410155749-daa745c078e1",
    )
    go_repository(
        name = "org_golang_google_grpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc",
        sum = "h1:3Oj82/tFSCeUrRTg/5E/7d/W5A1tj6Ky1ABAuZuv5ag=",
        version = "v1.55.0",
    )
    go_repository(
        name = "org_golang_google_grpc_cmd_protoc_gen_go_grpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
        sum = "h1:rNBFJjBCOgVr9pWD7rs/knKL4FRTKgpZmsRfV214zcA=",
        version = "v1.3.0",
    )
    go_repository(
        name = "org_golang_google_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/protobuf",
        sum = "h1:kPPoIgf3TsEvrm0PFe15JQ+570QVxYzEvvHqChK+cng=",
        version = "v1.30.0",
    )
    go_repository(
        name = "org_golang_x_arch",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/arch",
        sum = "h1:oMxhUYsO9VsR1dcoVUjJjIGhx1LXol3989T/yZ59Xsw=",
        version = "v0.1.0",
    )

    go_repository(
        name = "org_golang_x_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/crypto",
        sum = "h1:LF6fAI+IutBocDJ2OT0Q1g8plpYljMZ4+lty+dsqw3g=",
        version = "v0.9.0",
    )
    go_repository(
        name = "org_golang_x_exp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/exp",
        sum = "h1:k/i9J1pBpvlfR+9QsetwPyERsqu1GIbi967PQMq3Ivc=",
        version = "v0.0.0-20230522175609-2e198f4a06a1",
    )
    go_repository(
        name = "org_golang_x_exp_typeparams",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/exp/typeparams",
        sum = "h1:Jw5wfR+h9mnIYH+OtGT2im5wV1YGGDora5vTv/aa5bE=",
        version = "v0.0.0-20221208152030-732eee02a75a",
    )

    go_repository(
        name = "org_golang_x_image",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/image",
        sum = "h1:+qEpEAPhDZ1o0x3tHzZTQDArnOixOzGD9HUJfcg0mb4=",
        version = "v0.0.0-20190802002840-cff245a6509b",
    )
    go_repository(
        name = "org_golang_x_lint",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/lint",
        sum = "h1:VLliZ0d+/avPrXXH+OakdXhpJuEoBZuwh1m2j7U6Iug=",
        version = "v0.0.0-20210508222113-6edffad5e616",
    )
    go_repository(
        name = "org_golang_x_mobile",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/mobile",
        sum = "h1:4+4C/Iv2U4fMZBiMCc98MG1In4gJY5YRhtpDNeDeHWs=",
        version = "v0.0.0-20190719004257-d2bd2a29d028",
    )
    go_repository(
        name = "org_golang_x_mod",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/mod",
        sum = "h1:lFO9qtOdlre5W1jxS3r/4szv2/6iXxScdzjoBMXNhYk=",
        version = "v0.10.0",
    )
    go_repository(
        name = "org_golang_x_net",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/net",
        sum = "h1:X2//UzNDwYmtCLn7To6G58Wr6f5ahEAQgKNzv9Y951M=",
        version = "v0.10.0",
    )
    go_repository(
        name = "org_golang_x_oauth2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/oauth2",
        sum = "h1:6dkIjl3j3LtZ/O3sTgZTMsLKSftL/B8Zgq4huOIIUu8=",
        version = "v0.8.0",
    )
    go_repository(
        name = "org_golang_x_sync",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sync",
        sum = "h1:PUR+T4wwASmuSTYdKjYHI5TD22Wy5ogLU5qZCOLxBrI=",
        version = "v0.2.0",
    )
    go_repository(
        name = "org_golang_x_sys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sys",
        sum = "h1:EBmGv8NaZBZTWvrbjNoL6HVt+IVy3QDQpJs7VRIw3tU=",
        version = "v0.8.0",
    )
    go_repository(
        name = "org_golang_x_term",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/term",
        sum = "h1:n5xxQn2i3PC0yLAbjTpNT85q/Kgzcr2gIoX9OrJUols=",
        version = "v0.8.0",
    )
    go_repository(
        name = "org_golang_x_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/text",
        sum = "h1:2sjJmO8cDvYveuX97RDLsxlyUxLl+GHoLxBiRdHllBE=",
        version = "v0.9.0",
    )
    go_repository(
        name = "org_golang_x_time",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/time",
        sum = "h1:rg5rLMjNzMS1RkNLzCG38eapWhnYLFYXDXj2gOlr8j4=",
        version = "v0.3.0",
    )
    go_repository(
        name = "org_golang_x_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/tools",
        sum = "h1:8WMNJAz3zrtPmnYC7ISf5dEn3MT0gY7jBJfw27yrrLo=",
        version = "v0.9.1",
    )
    go_repository(
        name = "org_golang_x_vuln",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/vuln",
        sum = "h1:9GRdj6wAIkDrsMevuolY+SXERPjQPp2P1ysYA0jpZe0=",
        version = "v0.1.0",
    )

    go_repository(
        name = "org_golang_x_xerrors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/xerrors",
        sum = "h1:H2TDz8ibqkAF6YGhCdN3jS9O0/s90v0rJh3X/OLHEUk=",
        version = "v0.0.0-20220907171357-04be3eba64a2",
    )

    go_repository(
        name = "org_libvirt_go_libvirt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "libvirt.org/go/libvirt",
        # keep
        patches = [
            "//3rdparty/bazel/org_libvirt_go_libvirt:go_libvirt.patch",
        ],
        sum = "h1:q4UeuyzSp7GiYUB5T3ytzeXh41hp9JqNQu150NkO+7A=",
        version = "v1.8010.0",
    )

    go_repository(
        name = "org_mongodb_go_mongo_driver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.mongodb.org/mongo-driver",
        sum = "h1:Ql6K6qYHEzB6xvu4+AU0BoRoqf9vFPcc4o7MUIdPW8Y=",
        version = "v1.11.3",
    )
    go_repository(
        name = "org_mozilla_go_pkcs7",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.mozilla.org/pkcs7",
        sum = "h1:A/5uWzF44DlIgdm/PQFwfMkW0JX+cIcQi/SwLAmZP5M=",
        version = "v0.0.0-20200128120323-432b2356ecb1",
    )
    go_repository(
        name = "org_uber_go_atomic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/atomic",
        sum = "h1:9qC72Qh0+3MqyJbAn8YU5xVq1frD8bn3JtD2oXtafVQ=",
        version = "v1.10.0",
    )

    go_repository(
        name = "org_uber_go_goleak",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/goleak",
        sum = "h1:NBol2c7O1ZokfZ0LEU9K6Whx/KnwvepVetCUhtKja4A=",
        version = "v1.2.1",
    )
    go_repository(
        name = "org_uber_go_multierr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/multierr",
        sum = "h1:7fIwc/ZtS0q++VgcfqFDxSBZVv/Xo49/SYnDFupUwlI=",
        version = "v1.9.0",
    )
    go_repository(
        name = "org_uber_go_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/tools",
        sum = "h1:0mgffUl7nfd+FpvXMVz4IDEaUSmT1ysygQC7qYo7sG4=",
        version = "v0.0.0-20190618225709-2cfd321de3ee",
    )
    go_repository(
        name = "org_uber_go_zap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/zap",
        sum = "h1:FiJd5l1UOLj0wCgbSE0rwwXHzEdAZS6hiiSnxJN/D60=",
        version = "v1.24.0",
    )
    go_repository(
        name = "sh_helm_helm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "helm.sh/helm",
        sum = "h1:cSe3FaQOpRWLDXvTObQNj0P7WI98IG5yloU6tQVls2k=",
        version = "v2.17.0+incompatible",
    )
    go_repository(
        name = "sh_helm_helm_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "helm.sh/helm/v3",
        sum = "h1:rOq2TPVzg5jt4q5ermAZGZFxNW2uQhKjRhBneAutMEM=",
        version = "v3.12.0",
    )
    go_repository(
        name = "sm_step_go_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.step.sm/crypto",
        sum = "h1:EzqPTvW1g6kxEnfIf/exDW+MhHGeEhtoNMhQX7P/UwI=",
        version = "v0.30.0",
    )
    go_repository(
        name = "tools_gotest_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gotest.tools/v3",
        sum = "h1:ZazjZUfuVeZGLAmlKKuyv3IKP5orXcwtOwDQH6YVr6o=",
        version = "v3.4.0",
    )
    go_repository(
        name = "xyz_gomodules_jsonpatch_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gomodules.xyz/jsonpatch/v2",
        sum = "h1:8NFhfS6gzxNqjLIYnZxg319wZ5Qjnx4m/CcX+Klzazc=",
        version = "v2.3.0",
    )
