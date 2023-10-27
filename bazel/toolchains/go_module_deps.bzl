"""Go module dependencies for Bazel.

Contains the equivalent of go.mod and go.sum files for Bazel.
"""

load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    """Declare Go module dependencies for Bazel."""
    go_repository(
        name = "build_buf_gen_go_bufbuild_protovalidate_protocolbuffers_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go",
        sum = "h1:tdpHgTbmbvEIARu+bixzmleMi14+3imnpoFXz+Qzjp4=",
        version = "v1.31.0-20230802163732-1c33ebd9ecfa.1",
    )
    go_repository(
        name = "cat_dario_mergo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "dario.cat/mergo",
        sum = "h1:AGCNq9Evsj31mOgNPcLyXc+4PNABt905YmuqPYYpBWk=",
        version = "v1.0.0",
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
        sum = "h1:UoveltGrhghAA7ePc+e+QYDHXrBps2PqFZiHkGR/xK8=",
        version = "v0.0.1-2020.1.4",
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
        sum = "h1:bvDV9vkmnHYOMsOr4WLk+Vo07yKIzd94sVoIqshQ4bU=",
        version = "v0.0.0-20230811130428-ced1acdcaa24",
    )
    go_repository(
        name = "com_github_adamkorcz_go_118_fuzz_build",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/AdamKorcz/go-118-fuzz-build",
        sum = "h1:59MxjQVfjXsBpLy+dbd2/ELV5ofnUkUZBvWSC85sheA=",
        version = "v0.0.0-20230306123547-8075edf89bb0",
    )
    go_repository(
        name = "com_github_adamkorcz_go_fuzz_headers_1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/AdamKorcz/go-fuzz-headers-1",
        sum = "h1:rd389Q26LMy03gG4anandGFC2LW/xvjga5GezeeaxQk=",
        version = "v0.0.0-20230618160516-e936619f9f18",
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
        name = "com_github_agext_levenshtein",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/agext/levenshtein",
        sum = "h1:QmvMAjj2aEICytGiWzmxoE0x2KZvE0fvmqMOfy2tjT8=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_agnivade_levenshtein",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/agnivade/levenshtein",
        sum = "h1:3oJU7J3FGFmyhn8KHjmVaZCN5hxTr7GxgRue+sxIXdQ=",
        version = "v1.0.1",
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
        name = "com_github_alecthomas_kingpin_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/kingpin/v2",
        sum = "h1:H0aULhgmSzN8xQ3nX1uxtdlTHYoPLu5AhHxWrKI6ocU=",
        version = "v2.3.2",
    )

    go_repository(
        name = "com_github_alecthomas_template",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/template",
        sum = "h1:cAKDfWh5VpdgMhJosfJnn5/FoN2SRZ4p7fJNX58YPaU=",
        version = "v0.0.0-20160405071501-a0175ee3bccc",
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
        name = "com_github_alessio_shellescape",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alessio/shellescape",
        sum = "h1:V7yhSDDn8LP4lc4jS8pFkt0zCnzVJlG5JXy9BVKJUX0=",
        version = "v1.4.1",
    )
    go_repository(
        name = "com_github_anatol_vmtest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/anatol/vmtest",
        sum = "h1:t4JGeY9oaF5LB4Rdx9e2wARRRPAYt8Ow4eCf5SwO3fA=",
        version = "v0.0.0-20220413190228-7a42f1f6d7b8",
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
        name = "com_github_antlr_antlr4_runtime_go_antlr_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/antlr/antlr4/runtime/Go/antlr/v4",
        sum = "h1:goHVqTbFX3AIo0tzGr14pgfAW2ZfPChKO21Z9MGf/gk=",
        version = "v4.0.0-20230512164433-5d1fd1a340c9",
    )
    go_repository(
        name = "com_github_apache_arrow_go_v12",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apache/arrow/go/v12",
        sum = "h1:xtZE63VWl7qLdB0JObIXvvhGjoVNrQ9ciIHG2OK5cmc=",
        version = "v12.0.0",
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
        name = "com_github_apparentlymart_go_dump",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apparentlymart/go-dump",
        sum = "h1:ZSTrOEhiM5J5RFxEaFvMZVEAM1KvT1YzbEOwB2EAGjA=",
        version = "v0.0.0-20180507223929-23540a00eaa3",
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
        name = "com_github_asaskevich_govalidator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/asaskevich/govalidator",
        sum = "h1:DklsrG3dyBCFEj5IhUbnKptjxatkF07cF2ak3yi77so=",
        version = "v0.0.0-20230301143203-a9d515a09cc2",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go",
        sum = "h1:uL4EV0gQxotQVYegIoBqK079328MOJqgG95daFYSkAM=",
        version = "v1.44.297",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2",
        sum = "h1:+tefE750oAb7ZQGzla6bLkOwfcQCEtC5y2RqoqCeqKo=",
        version = "v1.18.1",
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
        sum = "h1:Az9uLwmssTE6OGTpsFqOnaGpLnKDqNYOJzWuC6UAYzA=",
        version = "v1.18.27",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_credentials",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/credentials",
        sum = "h1:qmU+yhKmOCyujmuPY7tf5MxR/RKyZrOPO3V4DobiTUk=",
        version = "v1.13.26",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_ec2_imds",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/ec2/imds",
        sum = "h1:LxK/bitrAr4lnh9LnIS6i7zWbCOdMsfzKFBI6LUCS0I=",
        version = "v1.13.4",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_s3_manager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/s3/manager",
        sum = "h1:SAB1UAVaf6nGCu3zyIrV+VWsendXrms1GqtW4zBotKA=",
        version = "v1.11.71",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_configsources",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/configsources",
        sum = "h1:A5UqQEmPaCFpedKouS4v+dHCTUo2sKqhoKO9U5kxyWo=",
        version = "v1.1.34",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_endpoints_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/endpoints/v2",
        sum = "h1:srIVS45eQuewqz6fKKu6ZGXaq6FuFg5NzgQBAM6g8Y4=",
        version = "v2.4.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_ini",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/ini",
        sum = "h1:LWA+3kDM8ly001vJ1X1waCuLJdtTl48gwkPKWy9sosI=",
        version = "v1.3.35",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_v4a",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/v4a",
        sum = "h1:wscW+pnn3J1OYnanMnza5ZVYXLX4cKk5rAvUAl4Qu+c=",
        version = "v1.0.26",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_autoscaling",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/autoscaling",
        sum = "h1:gnNW8xYVF7pKJrIu6WRF2r9NZylc7jLna2O3oPFIii0=",
        version = "v1.28.9",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_cloudfront",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/cloudfront",
        sum = "h1:loRDtQ0vT0+JCB0hQBCfv95tttEzJ1rqSaTDy5cpy0A=",
        version = "v1.26.8",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_cloudwatchlogs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs",
        sum = "h1:zMnh9plMceN5DVuG55IjzEwAS3kbeG0GTNzmbnqI/C8=",
        version = "v1.21.2",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ec2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ec2",
        sum = "h1:P4dyjm49F2kKws0FpouBC6fjVImACXKt752+CWa01lM=",
        version = "v1.102.0",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_elasticloadbalancingv2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2",
        sum = "h1:g/Kzed9qNdvz5p7Av3ffavD19eN11deWqlHgR2JuXuw=",
        version = "v1.19.13",
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
        sum = "h1:zZSLP3v3riMOP14H7b4XP0uyfREDQOYv2cqIrvTXDNQ=",
        version = "v1.1.29",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_presigned_url",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/presigned-url",
        sum = "h1:bkRyG4a929RCnpVSTvLM2j/T4ls015ZhhYApbmYs15s=",
        version = "v1.9.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_s3shared",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/s3shared",
        sum = "h1:dBL3StFxHtpBzJJ/mNEsjXVgfO+7jR0dAIEwLqMapEA=",
        version = "v1.14.3",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/kms",
        sum = "h1:jwmtdM1/l1DRNy5jQrrYpsQm8zwetkgeqhAqefDr1yI=",
        version = "v1.22.2",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_resourcegroupstaggingapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi",
        sum = "h1:6AuIiaZ+oRhprPZw2/siZQcaZRvmKipjGbmGI0BSGsA=",
        version = "v1.14.14",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_s3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/s3",
        sum = "h1:lEmQ1XSD9qLk+NZXbgvLJI/IiTz7OIR2TYUTFH25EI4=",
        version = "v1.36.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_secretsmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/secretsmanager",
        sum = "h1:eW8zPSh7ZLzb7029xCsIEFbnxLvNHPTt7aWwdKjNJc8=",
        version = "v1.19.10",
    )

    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sso",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sso",
        sum = "h1:nneMBM2p79PGWBQovYO/6Xnc2ryRMw3InnDJq1FHkSY=",
        version = "v1.12.12",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ssooidc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ssooidc",
        sum = "h1:2qTR7IFk7/0IN/adSFhYu9Xthr0zVFTgBrmPldILn80=",
        version = "v1.14.12",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sts",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sts",
        sum = "h1:XFJ2Z6sNUUcAz9poj+245DMkrHE4h2j5I9/xD50RHfE=",
        version = "v1.19.2",
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
        sum = "h1:SEy2xmstIphdPwNBUi7uhvjyjhVKISfwjfOJmuy7kg4=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_azidentity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/azidentity",
        sum = "h1:vcYCAze6p19qBW7MhZybIsqD8sMV8js0NyQM8JDnVtg=",
        version = "v1.3.0",
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
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_compute_armcompute_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5",
        sum = "h1:Sg/D8VuUQ+bw+FOYJF+xRKcwizCOP13HL0Se8pWNBzE=",
        version = "v5.1.0",
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
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_network_armnetwork_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4",
        sum = "h1:pqCyNi/Paz03SbWRmGlb5WBzK14aOXVuSJuOTWzOM5M=",
        version = "v4.0.0",
    )

    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_resources_armresources",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources",
        sum = "h1:7CBQ+Ei8SP2c6ydQTGCCrS35bDxgTMfoP2miAwK++OU=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_security_keyvault_azkeys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys",
        sum = "h1:4Kynh6Hn2ekyIsBgNQJb3dn1+/MyvzfUJebti2emB/A=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_security_keyvault_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal",
        sum = "h1:T028gtTPiYt/RMUfs8nVsAL7FDQrfLlrm/NnRG/zcC4=",
        version = "v0.8.0",
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
        sum = "h1:OBhqkivkhkMqLPymWEppkm7vgPQY2XsHoEkaMQ0AdZY=",
        version = "v1.0.0",
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
        sum = "h1:aY2smc3JWyUKOjGYmOKVLX70fPK9ON0rtwQojuIeUHc=",
        version = "v0.42.0",
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
        sum = "h1:Q92kusRqC1XV2MjkWETPvjJVqKetz1OzxZB7mHJLju8=",
        version = "v1.1.0",
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
        name = "com_github_bufbuild_protovalidate_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bufbuild/protovalidate-go",
        sum = "h1:pJr07sYhliyfj/STAM7hU4J3FKpVeLVKvOBmOTN8j+s=",
        version = "v0.2.1",
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
        sum = "h1:o7IhLm0Msx3BaB+n3Ag7L8EVlByGnpq14C4YWiu/gL8=",
        version = "v1.3.2",
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
        sum = "h1:y4OZtCnogmCPw98Zjyt5a6+QwPLGkiQsYW5oUqylYbM=",
        version = "v4.2.1",
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
        sum = "h1:/inchEIKaYC1Akx+H+gqO04wryn5h75LSazbRlnya1k=",
        version = "v0.0.0-20230607035331-e9ce68804cb4",
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
        sum = "h1:f5WFqIVSgo5IZmtTT3qVBo6TzI1ON6sycSBKkymb9L0=",
        version = "v3.0.2",
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
        sum = "h1:oNAVsnhPoy4BTPQivLgTzI9Oleml9l/+eYIDYXRCYo8=",
        version = "v1.7.6",
    )
    go_repository(
        name = "com_github_containerd_continuity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/continuity",
        sum = "h1:v3y/4Yz5jwnvqPKJJ+7Wf93fyWoCB3F5EclWG023MDM=",
        version = "v0.4.2",
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
        sum = "h1:9vqZr0pxwOF5koz6N0N3kJ0zDHokrcPxIR/ZR2YFtOs=",
        version = "v1.2.2",
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
        sum = "h1:3Q4Pt7i8nYwy2KmQWIw2+1hTvwTE/6w9FqcttATPO/4=",
        version = "v2.1.1",
    )

    go_repository(
        name = "com_github_containerd_zfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/zfs",
        sum = "h1:n7OZ7jZumLIqNJqXrEc/paBM840mORnmGdJDmAmJZHM=",
        version = "v1.1.0",
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
        name = "com_github_coreos_go_oidc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-oidc",
        sum = "h1:mh48q/BqXqgjVHpy2ZY7WnWAbenxRjsz9N1i1YxjHAk=",
        version = "v2.2.1+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_oidc_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-oidc/v3",
        sum = "h1:AKVxfYw1Gmkn/w96z0DbT/B/xFnzTd3MkZvWLjF4n/o=",
        version = "v3.6.0",
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
        name = "com_github_coreos_go_systemd_v22",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-systemd/v22",
        sum = "h1:RrqgGjYQKalulkV8NGVIfkXQf6YYmOyiJKk8iXXhfZs=",
        version = "v22.5.0",
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
        sum = "h1:Ugdm7cg7i6ZK6x3xDF1oEu1nfkyfH53EtKeQYTC3kyg=",
        version = "v0.2.4",
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
        name = "com_github_decred_dcrd_dcrec_secp256k1_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/decred/dcrd/dcrec/secp256k1/v4",
        sum = "h1:1iy2qD6JEhHKKhUOA9IWs7mjco7lnw2qx8FsRI2wirE=",
        version = "v4.0.0-20210816181553-5444fa50b93d",
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
        name = "com_github_dgryski_go_rendezvous",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dgryski/go-rendezvous",
        sum = "h1:lO4WD4F/rVNCu3HqELle0jiPLLBs70cWOduZpkS1E78=",
        version = "v0.0.0-20200823014737-9f7001d12a5f",
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
        sum = "h1:fF+XCQCgJjjQNIMjzaSmiKJSCcfcXb3TWTcc7GAneOY=",
        version = "v24.0.6+incompatible",
    )
    go_repository(
        name = "com_github_docker_distribution",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/distribution",
        sum = "h1:T3de5rq0dB1j30rp0sA2rER+m322EBzniBPB6ZIzuh8=",
        version = "v2.8.2+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/docker",
        sum = "h1:hceabKCtUgDqPu+qm0NgsaXf28Ljf4/pWFL7xjWWDgE=",
        version = "v24.0.6+incompatible",
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
        name = "com_github_dustin_go_humanize",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dustin/go-humanize",
        sum = "h1:GzkhY7T5VNhEkwH0PVJgjz+fX1rhBrR7pRT3mDkpeCY=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_edgelesssys_go_azguestattestation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/edgelesssys/go-azguestattestation",
        sum = "h1:1iKB7b+i7svWC0aKXwggi+kHf0K57g8r9hN4VOpJYYg=",
        version = "v0.0.0-20230707101700-a683be600fcf",
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
        name = "com_github_eggsampler_acme_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/eggsampler/acme/v3",
        sum = "h1:5M7vwYRy65iPpCFHZ01RyWXmYT8e8MlcWn/9BUUB7Ro=",
        version = "v3.3.0",
    )
    go_repository(
        name = "com_github_elazarl_goproxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/elazarl/goproxy",
        sum = "h1:RIB4cRk+lBqKK3Oy0r2gRX4ui7tuhiZq2SuTtTCi0/0=",
        version = "v0.0.0-20221015165544-a0805db90819",
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
        sum = "h1:wSUXTlLfiAQRWs2F+p+EKOY9rUyis1MyGqJ2DIk5HpM=",
        version = "v0.11.1",
    )
    go_repository(
        name = "com_github_envoyproxy_protoc_gen_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/envoyproxy/protoc-gen-validate",
        sum = "h1:QkIBuU5k+x7/QXPvPPnWXWlCdaBFApVqftFV6k087DA=",
        version = "v1.0.2",
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
        sum = "h1:kOqh6YHBtK8aywxGerMG2Eq3H6Qgoqeo13Bk2Mv/nBs=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_github_favadi_protoc_go_inject_tag",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/favadi/protoc-go-inject-tag",
        sum = "h1:K3KXxbgRw5WT4f43LbglARGz/8jVsDOS7uMjG4oNvXY=",
        version = "v1.4.0",
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
        name = "com_github_foxboron_go_uefi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/foxboron/go-uefi",
        sum = "h1:SJMQFT74bCrP+kQ24oWhmuyPFHDTavrd3JMIe//2NhU=",
        version = "v0.0.0-20230808201820-18b9ba9cd4c3",
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
        name = "com_github_frankban_quicktest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/frankban/quicktest",
        sum = "h1:dfYrrRyLtiqT9GyKXgdh+k4inNeTvmGbuSgZ3lx3GhA=",
        version = "v1.14.5",
    )
    go_repository(
        name = "com_github_fsnotify_fsnotify",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:8JEhPFa5W2WU7YfeZzPNqzMP6Lwt7L2715Ggo0nosvA=",
        version = "v1.7.0",
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
        sum = "h1:fUmoe+HLsBTctBDoaBwpQo5N+nrCp8g/BjKb/6ZQmYw=",
        version = "v1.1.0",
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
        name = "com_github_ghodss_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ghodss/yaml",
        sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
        version = "v1.0.0",
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
        sum = "h1:+zs/tPmkDkHx3U66DAb0lQFJrpS6731Oaa12ikc+DiI=",
        version = "v1.5.1-0.20230307220236-3a3c6141e376",
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
        sum = "h1:Pz0DHeFij3XFhoBRGUDPzSJ+w2UcK5/0JvF8DRI58r8=",
        version = "v4.3.2-0.20230305113008-0c11038e723f",
    )
    go_repository(
        name = "com_github_go_git_go_git_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-git/go-git/v5",
        sum = "h1:t9AudWVLmqzlo+4bqdf7GY+46SUuRsx59SboFxkq2aE=",
        version = "v5.7.0",
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
        sum = "h1:ItKF/Vbuj31dmV4jxA1qblpSwkl9g1typ24xoe70IGs=",
        version = "v3.1.0",
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
        sum = "h1:Wz+5lgoB0kkuqLEc6NVmwRknTKP6dTGbSqvhZtBI/j0=",
        version = "v0.8.0",
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
        sum = "h1:unTcVm6PispJsMECE3zWgvG4xTiKda1LIR5rCRWLG6M=",
        version = "v0.20.4",
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
        sum = "h1:3sVjiK66+uXK/6oQ8xgcRKcFgQ5KXa2KvnJRumpMGbE=",
        version = "v0.20.2",
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
        sum = "h1:QLMzNJnMGPRNDCbySlcj1x01tzU8/9LTTL9hZZZogBU=",
        version = "v0.22.4",
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
        sum = "h1:9c50NUPC30zyuKprjL3vNZ0m5oG+jU0zvx4AqHGnv4k=",
        version = "v10.14.1",
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
        sum = "h1:oLiKZW721CCMwA5g7977cWfcAKQ+FuosP47Zf1QiDrA=",
        version = "v0.113.3",
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
        sum = "h1:DVjP2PbBOzHyzA+dn3WhHIq4NdVu3Q+pvivFICf/7fo=",
        version = "v1.1.2",
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
        name = "com_github_golang_jwt_jwt_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang-jwt/jwt/v5",
        sum = "h1:1n1XNM9hk7O9mnQoNBGolZvzebBQ7p93ULHRc28XJUE=",
        version = "v5.0.0",
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
        sum = "h1:s2151PDGy/eqpCI80/8dl4VL3xTkqI/YubXLXCFw0mw=",
        version = "v0.17.1",
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
        name = "com_github_google_gnostic_models",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/gnostic-models",
        sum = "h1:yo/ABAfM5IMRsS1VnXjTBvUb61tFIHozhlYvRgGre9I=",
        version = "v0.6.8",
    )

    go_repository(
        name = "com_github_google_go_attestation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-attestation",
        sum = "h1:jXtAWT2sw2Yu8mYU0BC7FDidR+ngxFPSE+pl6IUu3/0=",
        version = "v0.5.0",
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
        sum = "h1:ofyhxvXcZhMsU5ulbFiLKl/XBFqE1GSq7atu8tAmTRI=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_google_go_containerregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-containerregistry",
        sum = "h1:MMkSh+tjSdnmJZO7ljvEqV1DjfekB6VUEAZgy3a+TQE=",
        version = "v0.15.2",
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
        name = "com_github_google_go_pkcs11",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-pkcs11",
        sum = "h1:OF1IPgv+F4NmqmJ98KTjdN97Vs1JxDPB3vbmYzV2dpk=",
        version = "v0.2.1-0.20230907215043-c6f79328ddf9",
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
        replace = "github.com/google/go-sev-guest",
        sum = "h1:6o4Z/vQqNUH+cEagfx1Ez5ElK70iZulEXZwmLnRo44I=",
        version = "v0.0.0-20230928233922-2dcbba0a4b9d",
    )
    go_repository(
        name = "com_github_google_go_tdx_guest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-tdx-guest",
        sum = "h1:MgHcWLCEHVpqR0LXxtgJq2uG0oNNszV8+5XYtKs7+Yg=",
        version = "v0.2.2",
    )

    go_repository(
        name = "com_github_google_go_tpm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-tpm",
        replace = "github.com/thomasten/go-tpm",
        sum = "h1:840nUyrM9df2aLuzWuIkYx/DrUbX4KQZO6B9LD45aWo=",
        version = "v0.0.0-20230629092004-f43f8e2a59eb",
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
        replace = "github.com/msanft/go-tpm-tools",
        sum = "h1:rxzvNtTPknJTrPhcP4op83F+RwmKWI3TGEiyjrBOSLU=",
        version = "v0.0.0-20231009082622-266280269413",
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
        sum = "h1:/CP5g8u/VJHijgedC/Legn3BAbAaWPgecwXBIDzw5no=",
        version = "v2.1.0+incompatible",
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
        sum = "h1:60BLSyTrOV4/haCDW4zb1guZItoSq8foHCXrAnjBo/o=",
        version = "v0.1.7",
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
        sum = "h1:KjJaJ9iWZ3jOFZIf1Lqf4laDRCasjl0BCmnEGxkdLb4=",
        version = "v1.3.1",
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
        sum = "h1:SBWmZhjUDRorQxrN0nwzf+AHBxnbFjViHQS4P0yVpmQ=",
        version = "v0.3.1",
    )

    go_repository(
        name = "com_github_googleapis_gax_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/gax-go/v2",
        sum = "h1:A+gCJKdRfqXkr+BIRGtZLibNXf0m1f9E4HG56etFpas=",
        version = "v2.12.0",
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
        sum = "h1:cDN6XFCLKiiqvYpjQLq9AiM7RDRbIC9450WpPH+yvXo=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_gophercloud_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gophercloud/utils",
        sum = "h1:sH7xkTfYzxIEgzq1tDHIMKRh1vThOEOGNsettdEeLbE=",
        version = "v0.0.0-20231010081019-80377eca5d56",
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
        sum = "h1:+9834+KizmvFV7pXQGSXQTsaWhq2GjuNUt0aUU0YBYw=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_middleware_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/go-grpc-middleware/v2",
        sum = "h1:2cz5kSrxzMYHiWOBbKj8itQm+nRykkB8aMv4ThcHYHA=",
        version = "v2.0.0",
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
        sum = "h1:bI2ocEMgcVlz55Oj1xZNBsVi900c7II+fWDyV9o+13c=",
        version = "v1.5.0",
    )

    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/v2",
        sum = "h1:A51EguZ576URdtcQ0l8mT/tOD948oAtmP1soqIHIFfI=",
        version = "v2.0.10",
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
        name = "com_github_hashicorp_go_multierror",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-multierror",
        sum = "h1:H5DkEtf6CXdFp0N0Em5UCwQpXMWke8IA0+lD48awMYo=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_hashicorp_go_retryablehttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-retryablehttp",
        sum = "h1:ZQgVdpTdAL7WpMIwLzCfbalOcSUdkDZnpUv3/+BxzFA=",
        version = "v0.7.4",
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
        sum = "h1:kWg2vyKl7BRXrNxYziqDJ55n+vtOQ1QsGORjzoeB+uM=",
        version = "v0.2.2",
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
        name = "com_github_hashicorp_hcl_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/hcl/v2",
        sum = "h1:z1XvSUyXd1HP10U4lrLg5e0JMVz6CPaJvAgxM0KNZVY=",
        version = "v2.17.0",
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
        sum = "h1:YjkZLJ7K3inKgMZ0wzCU9OHqc+UqMQyXsPXnf3Cl2as=",
        version = "v1.9.2",
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
        sum = "h1:A9HsByNhogrvm9cWb28sjiS3i7tcKCkflWFEkHfuAgM=",
        version = "v0.0.0-20210920133722-c8aef6fb66ef",
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
        sum = "h1:M8XP7IuFNsqUx6VPK2P9OSmsYsI/YFaGil0uD21V3dM=",
        version = "v0.3.15",
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
        sum = "h1:jrU9BWBgp9o2NcetUVm3dNpQ2SK1zG6aF6WF0wtPajc=",
        version = "v3.7.0",
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
        sum = "h1:IFV2oUNUzZaz+XyusxpLzpzS8Pt5rh0Z16For/djlyI=",
        version = "v1.16.5",
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
        name = "com_github_konsorten_go_windows_terminal_sequences",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:DB17ag19krx9CFsz4o3enTrPXyIXCl+2iCXH/aMAp9s=",
        version = "v1.0.2",
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
        name = "com_github_lestrrat_go_backoff_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lestrrat-go/backoff/v2",
        sum = "h1:oNb5E5isby2kiro9AgdHLv5N5tint1AnDVVf2E2un5A=",
        version = "v2.0.8",
    )
    go_repository(
        name = "com_github_lestrrat_go_blackmagic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lestrrat-go/blackmagic",
        sum = "h1:XzdxDbuQTz0RZZEmdU7cnQxUtFUzgCSPq8RCz4BxIi4=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_lestrrat_go_httpcc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lestrrat-go/httpcc",
        sum = "h1:ydWCStUeJLkpYyjLDHihupbn2tYmZ7m22BGkcvZZrIE=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_lestrrat_go_iter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lestrrat-go/iter",
        sum = "h1:q8faalr2dY6o8bV45uwrxq12bRa1ezKrB6oM9FUgN4A=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_lestrrat_go_jwx",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lestrrat-go/jwx",
        sum = "h1:tAx93jN2SdPvFn08fHNAhqFJazn5mBBOB8Zli0g0otA=",
        version = "v1.2.25",
    )
    go_repository(
        name = "com_github_lestrrat_go_option",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lestrrat-go/option",
        sum = "h1:WqAWL8kh8VcSoD6xjSH34/1m8yxluXQbDeKNfvFeEO4=",
        version = "v1.0.0",
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
        name = "com_github_linuxkit_virtsock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/linuxkit/virtsock",
        sum = "h1:jUp75lepDg0phMUJBCmvaeFDldD2N3S1lBuPwUTszio=",
        version = "v0.0.0-20201010232012-f8cee7dfc7a3",
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
        name = "com_github_lyft_protoc_gen_star_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/lyft/protoc-gen-star/v2",
        sum = "h1:/3+/2sWyXeMLzKd1bX+ixWKgEMsULrIivpDsuaF441o=",
        version = "v2.0.3",
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
        replace = "github.com/daniel-weisse/go-cryptsetup",
        sum = "h1:ToajP6trZoiqlZ3Z4uoG1P02/wtqSw1AcowOXOYjATk=",
        version = "v0.0.0-20230705150314-d8c07bd1723c",
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
        name = "com_github_masterminds_semver_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Masterminds/semver/v3",
        sum = "h1:RN9w6+7QoMeJVGyfmbcgs28Br8cvmnucEXnY0rYXWg0=",
        version = "v3.2.1",
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
        sum = "h1:uUcX/aBc8O7Fg9kaISIUsHXdKuqehiXAMQTYX8afzqM=",
        version = "v1.5.4",
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
        name = "com_github_mattn_go_colorable",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-colorable",
        sum = "h1:fFA4WZxdEF4tXPZVKMLwD8oUnCTTo08duU7wxecdEvA=",
        version = "v0.1.13",
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
        sum = "h1:9/kr64B9VUZrLm5YYwbGtUJnMgqWVOdUAXu6Migciow=",
        version = "v0.6.1",
    )
    go_repository(
        name = "com_github_microsoft_hcsshim",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Microsoft/hcsshim",
        sum = "h1:7EFNIY4igHEXUdj1zXgAyU3fLc7QfOKHbkldRVTBdiM=",
        version = "v0.11.0",
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
        name = "com_github_mistifyio_go_zfs_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mistifyio/go-zfs/v3",
        sum = "h1:YaoXgBePoMA12+S1u/ddkv+QqxcfiZK4prI6HPnkFiU=",
        version = "v3.0.1",
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
        name = "com_github_mitchellh_go_wordwrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/go-wordwrap",
        sum = "h1:TLuKupo69TCn6TQSyGxwI1EblZZEsQ0vMlAFQflz0v0=",
        version = "v1.0.1",
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
        sum = "h1:xt8Q1nalod/v7BqbG21f8mQPqH+xAaC9C3N3wfWbVP0=",
        version = "v0.5.0",
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
        name = "com_github_mohae_deepcopy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mohae/deepcopy",
        sum = "h1:e+l77LJOEqXTIQihQJVkA6ZxPOUmfPM5e4H7rcpgtSk=",
        version = "v0.0.0-20170603005431-491d3605edfb",
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
        name = "com_github_mxk_go_flowrate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mxk/go-flowrate",
        sum = "h1:y5//uYreIhSUg3J1GEMiLbxo1LJaP8RfCpH6pymGZus=",
        version = "v0.0.0-20140419014527-cca7078d478f",
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
        name = "com_github_nytimes_gziphandler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/NYTimes/gziphandler",
        sum = "h1:ZUDjpQae29j0ryrS0u/B8HZfJBtBQHjqw2rQ2cqUQ3I=",
        version = "v1.1.1",
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
        sum = "h1:31czK/TI9sNkxIKfaUfGlU47BAxQ0ztGgd9vPyqimf8=",
        version = "v1.2.8",
    )
    go_repository(
        name = "com_github_onsi_ginkgo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/ginkgo",
        sum = "h1:VkHVNpR4iVnU8XQR6DBm8BqYjN7CRzw+xKUbVVbbW9w=",
        version = "v1.8.0",
    )

    go_repository(
        name = "com_github_onsi_ginkgo_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/ginkgo/v2",
        sum = "h1:0jY9lJquiL8fcf3M4LAXN5aMlS/b2BV86HFFPCPMgE4=",
        version = "v2.13.0",
    )
    go_repository(
        name = "com_github_onsi_gomega",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/gomega",
        sum = "h1:naR28SdDFlqrG6kScpT8VWpu1xWY5nJRCF3XaYyBjhI=",
        version = "v1.27.10",
    )
    go_repository(
        name = "com_github_open_policy_agent_opa",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/open-policy-agent/opa",
        sum = "h1:qocVAKyjrqMjCqsU02S/gHyLr4AQQ9xMtuV1kKnnyhM=",
        version = "v0.42.2",
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
        sum = "h1:Ygwkfw9bpDvs+c9E34SdgGOj41dX/cbdlwvlWt0pnFI=",
        version = "v1.1.0-rc5",
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
        name = "com_github_opentracing_opentracing_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opentracing/opentracing-go",
        sum = "h1:uEJPy/1a5RIPAJ0Ov+OIO8OxWu77jEv+1B0VhjKrZUs=",
        version = "v1.2.0",
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
        sum = "h1:0ctb6s9mE31h0/lhu+J6OPmVeDxJn+kYnJc2jZR9tGQ=",
        version = "v2.0.8",
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
        sum = "h1:yk/hx9hDbrGHovbci4BY+pRMfSuuat626eFsHb7tmT8=",
        version = "v1.16.0",
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
        sum = "h1:+5BrQJwiBB9xsMygAB3TNvpQKOwlkc25LbISbrdOOfY=",
        version = "v0.44.0",
    )
    go_repository(
        name = "com_github_prometheus_procfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:kYK1Va/YMlutzCGazswoHKo//tZVlFpKYh+PymziUAg=",
        version = "v0.10.1",
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
        name = "com_github_protonmail_go_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ProtonMail/go-crypto",
        sum = "h1:ZK3C5DtzV2nVAQTx5S5jQvMeDqWtD1By5mOoyY/xJek=",
        version = "v0.0.0-20230518184743-7afd39499903",
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
        sum = "h1:MkV+77GLUNo5oJ0jf870itWm3D0Sjh7+Za9gazKc5LQ=",
        version = "v0.0.0-20200313005456-10cdbea86bc0",
    )

    go_repository(
        name = "com_github_redis_go_redis_v9",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/redis/go-redis/v9",
        sum = "h1:CuQcn5HIEeK7BgElubPP8CGtE0KakrnbBSTLjathl5o=",
        version = "v9.0.5",
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
        sum = "h1:cWPaGQEPrBb5/AsnsZesgZZ9yb1OQ+GOISoDNXVBh4M=",
        version = "v1.11.0",
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
        sum = "h1:bMDqOnrJVV/6JQgQ/MxOpU+AdO8uzYYA/TxFUBzFtS0=",
        version = "v1.5.2",
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
        name = "com_github_ryanuber_go_glob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ryanuber/go-glob",
        sum = "h1:iQh3xXAumdQ+4Ufa5b25cRpC5TYKlno6hsv6Cb3pkBk=",
        version = "v1.0.0",
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
        name = "com_github_sassoftware_relic_v7",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sassoftware/relic/v7",
        sum = "h1:2ZUM6ovo3STCAp0hZnO9nQY9lOB8OyfneeYIi4YUxMU=",
        version = "v7.5.5",
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
        name = "com_github_shopspring_decimal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/shopspring/decimal",
        sum = "h1:2Usl1nmF/WZucqkFZhnfFYxxxu8LG21F6nPQBE5gKV8=",
        version = "v1.3.1",
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
        sum = "h1:SX7Q6FxTDyX2hxugMgIqyivXWzemgMhHj3AlDbxjuFw=",
        version = "v1.4.6",
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
        sum = "h1:5JK/zKZvcQpL/jBmHvmFj3YbpDMBQnJQ6ygp8xdF3bY=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_sigstore_sigstore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore",
        sum = "h1:fCATemikcBK0cG4+NcM940MfoIgmioY1vC6E66hXxks=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_aws",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/aws",
        sum = "h1:rDHrG/63b3nBq3G9plg7iYnWN6lBhOfq/XultlCZgII=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_azure",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/azure",
        sum = "h1:X3ezwolP+b1jP3R6XPOWhUU0TZKONiv6EIRuySlZGrY=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_gcp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/gcp",
        sum = "h1:mj1KhdzzP1me994bt1UXhq5KZGSR1SoqxTqcT+hfPMk=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_hashivault",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/hashivault",
        sum = "h1:fhOToGY5fC5TY101an8i/oDYpoLzUJ1nUFwhnHA1+XY=",
        version = "v1.7.1",
    )

    go_repository(
        name = "com_github_sirupsen_logrus",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:dueUQJ1C2q9oE3F7wvmSGAaVtTmUizReu6fjN8uqzbQ=",
        version = "v1.9.3",
    )
    go_repository(
        name = "com_github_skeema_knownhosts",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/skeema/knownhosts",
        sum = "h1:MTk78x9FPgDFVFkDLTrsnnfCJl7g1C/nnKvePgrIngE=",
        version = "v1.1.1",
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
        name = "com_github_soheilhy_cmux",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/soheilhy/cmux",
        sum = "h1:jjzc5WVemNEDTLwv9tlmemhC73tI08BNOIGwBOo10Js=",
        version = "v0.1.5",
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
        sum = "h1:EaGW2JJh15aKOejeuJ+wpFSHnbd7GE6Wvp3TsNhb6LY=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_spf13_cast",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/cast",
        sum = "h1:R+kOtfhWQE6TVQzY+4D7wJLBgkdVasCEFxSUBYBYIlA=",
        version = "v1.5.1",
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
        sum = "h1:rGGH0XDZhdUOryiDWjmIvUSWpbNqisK8Wk0Vyefw8hc=",
        version = "v1.16.0",
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
        sum = "h1:g0eASXYtp+yvN9fK8sH94oCIk0fau9uV1/ZdJ0AVEzs=",
        version = "v1.3.0",
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
        sum = "h1:CcVxjf3Q8PM0mHUKJCdn+eZZtm5yQwehR5yeSVQQcUk=",
        version = "v1.8.4",
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
        name = "com_github_tink_crypto_tink_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tink-crypto/tink-go/v2",
        replace = "github.com/derpsteb/tink-go/v2",
        sum = "h1:FVii9oXvddz9sFir5TRYjQKrzJLbVD/hibT+SnRSDzg=",
        version = "v2.0.0-20231002051717-a808e454eed6",
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
        name = "com_github_ulikunitz_xz",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ulikunitz/xz",
        sum = "h1:kpFauv27b6ynzBNT/Xy+1k+fK4WswhN/6PN5WhFAGw8=",
        version = "v0.5.11",
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
        sum = "h1:hLFqsOLQ1SsppQNTMpkpPXClLDfC2A3Zgy9OUU+RVck=",
        version = "v0.11.3",
    )
    go_repository(
        name = "com_github_vektah_gqlparser_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vektah/gqlparser/v2",
        sum = "h1:C02NsyEsL4TXJB7ndonqTfuQOL4XPIu0aAWugdmTgmc=",
        version = "v2.4.5",
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
        name = "com_github_xhit_go_str2duration_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xhit/go-str2duration/v2",
        sum = "h1:lxklc02Drh6ynqX+DdPyp5pCKLUQpRT8bp8Ydu2Bstc=",
        version = "v2.1.0",
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
        sum = "h1:HzHnuAF1plUN2zGlAFHbSQP2qJ0ZAD3XF5XD7OesXRQ=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_yashtewari_glob_intersection",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/yashtewari/glob-intersection",
        sum = "h1:6gJvMYQlTDOL3dMsPF6J0+26vwX9MB8/1q3uAdhmTrg=",
        version = "v0.1.0",
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
        sum = "h1:ulX+SonA0Vma5zUFXtv52Kzip/xe7aj4vqT5AJwQ+ZQ=",
        version = "v0.2.3",
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
        name = "com_github_ysmood_got",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ysmood/got",
        sum = "h1:IrV2uWLs45VXNvZqhJ6g2nIhY+pgIG1CUoOcqfXFl1s=",
        version = "v0.34.1",
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
        sum = "h1:f0xmpYiSrHtSNAVgwip93Cg8tuF45HJM6rHq/A5RI/4=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_zclconf_go_cty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zclconf/go-cty",
        sum = "h1:4GvrUxe/QUDYuJKAav4EYqdM47/kZa672LwmXFmEKT0=",
        version = "v1.13.2",
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
        sum = "h1:tyNdfIxjzaWctIiLYOTalaLKZ17SI44SKFW26QbOhME=",
        version = "v0.110.8",
    )
    go_repository(
        name = "com_google_cloud_go_accessapproval",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/accessapproval",
        sum = "h1:/5YjNhR6lzCvmJZAnByYkfEgWjfAKwYP6nkuTk6nKFE=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_accesscontextmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/accesscontextmanager",
        sum = "h1:WIAt9lW9AXtqw/bnvrEUaE8VG/7bAAeMzRCBGMkc4+w=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_aiplatform",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/aiplatform",
        sum = "h1:J89aj+lqwtjn0qpQBMVaiOmDxBkKDEKUwl+GL19RRpc=",
        version = "v1.50.0",
    )
    go_repository(
        name = "com_google_cloud_go_analytics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/analytics",
        sum = "h1:TFBC1ZAqX9/jL56GEXdLrVe5vT3I22bDVWyDwZX4IEg=",
        version = "v0.21.3",
    )
    go_repository(
        name = "com_google_cloud_go_apigateway",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigateway",
        sum = "h1:aBSwCQPcp9rZ0zVEUeJbR623palnqtvxJlUyvzsKGQc=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeconnect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigeeconnect",
        sum = "h1:6u/jj0P2c3Mcm+H9qLsXI7gYcTiG9ueyQL3n6vCmFJM=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigeeregistry",
        sum = "h1:hgq0ANLDx7t2FDZDJQrCMtCtddR/pjCqVuvQWGrQbXw=",
        version = "v0.7.1",
    )

    go_repository(
        name = "com_google_cloud_go_appengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/appengine",
        sum = "h1:J+aaUZ6IbTpBegXbmEsh8qZZy864ZVnOoWyfa1XSNbI=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_area120",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/area120",
        sum = "h1:wiOq3KDpdqXmaHzvZwKdpoM+3lDcqsI2Lwhyac7stss=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_artifactregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/artifactregistry",
        sum = "h1:k6hNqab2CubhWlGcSzunJ7kfxC7UzpAfQ1UPb9PDCKI=",
        version = "v1.14.1",
    )
    go_repository(
        name = "com_google_cloud_go_asset",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/asset",
        sum = "h1:vlHdznX70eYW4V1y1PxocvF6tEwxJTTarwIGwOhFF3U=",
        version = "v1.14.1",
    )
    go_repository(
        name = "com_google_cloud_go_assuredworkloads",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/assuredworkloads",
        sum = "h1:yaO0kwS+SnhVSTF7BqTyVGt3DTocI6Jqo+S3hHmCwNk=",
        version = "v1.11.1",
    )
    go_repository(
        name = "com_google_cloud_go_automl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/automl",
        sum = "h1:iP9iQurb0qbz+YOOMfKSEjhONA/WcoOIjt6/m+6pIgo=",
        version = "v1.13.1",
    )
    go_repository(
        name = "com_google_cloud_go_baremetalsolution",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/baremetalsolution",
        sum = "h1:3zztyuQHjfU0C0qEsI9LkC3kf5/TQQ3jUJhbmetUoRA=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_batch",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/batch",
        sum = "h1:/4ADpZKoKH300HN2SB6aI7lXX/0hnnbR74wxjLHkyQo=",
        version = "v1.4.1",
    )
    go_repository(
        name = "com_google_cloud_go_beyondcorp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/beyondcorp",
        sum = "h1:VPg+fZXULQjs8LiMeWdLaB5oe8G9sEoZ0I0j6IMiG1Q=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_google_cloud_go_bigquery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/bigquery",
        sum = "h1:hs44Xxov3XLWQiCx2J8lK5U/ihLqnpm4RVVl5fdtLLI=",
        version = "v1.55.0",
    )
    go_repository(
        name = "com_google_cloud_go_billing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/billing",
        sum = "h1:CpagWXb/+QNye+vouomndbc4Gsr0uo+AGR24V16uk8Q=",
        version = "v1.17.0",
    )
    go_repository(
        name = "com_google_cloud_go_binaryauthorization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/binaryauthorization",
        sum = "h1:7L6uUWo/xNCfdVNnnzh2M4x5YA732YPgqRdCG8aKVAU=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_certificatemanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/certificatemanager",
        sum = "h1:uKsohpE0hiobx1Eak9jNcPCznwfB6gvyQCcS28Ah9E8=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_channel",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/channel",
        sum = "h1:Hy2EaOiOB7BS1IJmg2lLilEo8uMfFWTy7RgjTzbUqjM=",
        version = "v1.17.0",
    )
    go_repository(
        name = "com_google_cloud_go_cloudbuild",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/cloudbuild",
        sum = "h1:YTMxmFra7eIjKFgnyQUxOwWNseNqeO38kGh7thy7v4s=",
        version = "v1.14.0",
    )
    go_repository(
        name = "com_google_cloud_go_clouddms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/clouddms",
        sum = "h1:vTcaFaFZTZZ11gXB6aZHdAx+zn30P8YJw4X/S3NC+VQ=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_google_cloud_go_cloudtasks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/cloudtasks",
        sum = "h1:cMh9Q6dkvh+Ry5LAPbD/U2aw6KAqdiU6FttwhbTo69w=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_google_cloud_go_compute",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/compute",
        sum = "h1:tP41Zoavr8ptEqaW6j+LQOnyBBhO7OkOMAGrgLopTwY=",
        version = "v1.23.0",
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
        sum = "h1:YR2aPedGVQPpFBZXJnPkqRj8M//8veIZZH5ZvICoXnI=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_container",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/container",
        sum = "h1:SszQdI0qlyKsImz8/l26rpTZMyqvaH9yfua7rirDZvY=",
        version = "v1.26.0",
    )
    go_repository(
        name = "com_google_cloud_go_containeranalysis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/containeranalysis",
        sum = "h1:/EsoP+UTIjvl4yqrLA4WgUG83kwQhqZmbXEfqirT2LM=",
        version = "v0.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_datacatalog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datacatalog",
        sum = "h1:qGWrlYvWtK+8jD1jhwq5BsGoSr7S4/LOroV7LwXi00g=",
        version = "v1.17.1",
    )
    go_repository(
        name = "com_google_cloud_go_dataflow",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataflow",
        sum = "h1:VzG2tqsk/HbmOtq/XSfdF4cBvUWRK+S+oL9k4eWkENQ=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_google_cloud_go_dataform",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataform",
        sum = "h1:xcWso0hKOoxeW72AjBSIp/UfkvpqHNzzS0/oygHlcqY=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_datafusion",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datafusion",
        sum = "h1:eX9CZoyhKQW6g1Xj7+RONeDj1mV8KQDKEB9KLELX9/8=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_datalabeling",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datalabeling",
        sum = "h1:zxsCD/BLKXhNuRssen8lVXChUj8VxF3ofN06JfdWOXw=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_dataplex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataplex",
        sum = "h1:wqPAP1vRskOoWwNka1yey2wxxCrxRrcxJf78MyFvrbs=",
        version = "v1.9.1",
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
        name = "com_google_cloud_go_dataproc_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataproc/v2",
        sum = "h1:jKijbdsERm2hy/5dFl/LeQN+7CNssLdGXQYBMvMH/M4=",
        version = "v2.2.0",
    )

    go_repository(
        name = "com_google_cloud_go_dataqna",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataqna",
        sum = "h1:ITpUJep04hC9V7C+gcK390HO++xesQFSUJ7S4nSnF3U=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_datastore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datastore",
        sum = "h1:Mq0ApTRdLW3/dyiw+DkjTk0+iGIUvkbzaC8sfPwWTH4=",
        version = "v1.14.0",
    )
    go_repository(
        name = "com_google_cloud_go_datastream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datastream",
        sum = "h1:ra/+jMv36zTAGPfi8TRne1hXme+UsKtdcK4j6bnqQiw=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_deploy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/deploy",
        sum = "h1:A+w/xpWgz99EYzB6e31gMGAI/P5jTZ2UO7veQK5jQ8o=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_dialogflow",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dialogflow",
        sum = "h1:0hBV5ipVbhYNKCyiBoM47bUt+43Kd8eWXhBr+pwUSTw=",
        version = "v1.43.0",
    )
    go_repository(
        name = "com_google_cloud_go_dlp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dlp",
        sum = "h1:tF3wsJ2QulRhRLWPzWVkeDz3FkOGVoMl6cmDUHtfYxw=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_documentai",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/documentai",
        sum = "h1:cBndyac7kPWwSuhUcgdbnqzszfZ57HBEHfD33DIwsBM=",
        version = "v1.22.1",
    )
    go_repository(
        name = "com_google_cloud_go_domains",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/domains",
        sum = "h1:rqz6KY7mEg7Zs/69U6m6LMbB7PxFDWmT3QWNXIqhHm0=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_google_cloud_go_edgecontainer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/edgecontainer",
        sum = "h1:zhHWnLzg6AqzE+I3gzJqiIwHfjEBhWctNQEzqb+FaRo=",
        version = "v1.1.1",
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
        sum = "h1:OEJ0MLXXCW/tX1fkxzEZOsv/wRfyFsvDVNaHWBAvoV0=",
        version = "v1.6.2",
    )
    go_repository(
        name = "com_google_cloud_go_eventarc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/eventarc",
        sum = "h1:xIP3XZi0Xawx8DEfh++mE2lrIi5kQmCr/KcWhJ1q0J4=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_filestore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/filestore",
        sum = "h1:Eiz8xZzMJc5ppBWkuaod/PUdUZGCFR8ku0uS+Ah2fRw=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_firestore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/firestore",
        sum = "h1:/3S4RssUV4GO/kvgJZB+tayjhOfyAHs+KcpJgRVu/Qk=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_functions",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/functions",
        sum = "h1:LtAyqvO1TFmNLcROzHZhV0agEJfBi+zfMZsF4RT/a7U=",
        version = "v1.15.1",
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
        sum = "h1:Kfha8SOF2tqsu4O4jVle66mk7qNdlJ2KhL3E2YyiNZc=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_google_cloud_go_gkeconnect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkeconnect",
        sum = "h1:a1ckRvVznnuvDWESM2zZDzSVFvggeBaVY5+BVB8tbT0=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_gkehub",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkehub",
        sum = "h1:2BLSb8i+Co1P05IYCKATXy5yaaIw/ZqGvVSBTLdzCQo=",
        version = "v0.14.1",
    )
    go_repository(
        name = "com_google_cloud_go_gkemulticloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkemulticloud",
        sum = "h1:MluqhtPVZReoriP5+adGIw+ij/RIeRik8KApCW2WMTw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_google_cloud_go_grafeas",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/grafeas",
        sum = "h1:oyTL/KjiUeBs9eYLw/40cpSZglUC+0F7X4iu/8t7NWs=",
        version = "v0.3.0",
    )

    go_repository(
        name = "com_google_cloud_go_gsuiteaddons",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gsuiteaddons",
        sum = "h1:mi9jxZpzVjLQibTS/XfPZvl+Jr6D5Bs8pGqUjllRb00=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_iam",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iam",
        sum = "h1:gacbrBdWcoVmGLozRuStX45YKvJtzIjJdAolzUs1sm4=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_google_cloud_go_iap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iap",
        sum = "h1:RNhVq/6OMI99/wjPVhqFxjlBxYOBRdaG6rLpBvyaqYY=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_ids",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/ids",
        sum = "h1:khXYmSoDDhWGEVxHl4c4IgbwSRR+qE/L4hzP3vaU9Hc=",
        version = "v1.4.1",
    )
    go_repository(
        name = "com_google_cloud_go_iot",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iot",
        sum = "h1:yrH0OSmicD5bqGBoMlWG8UltzdLkYzNUwNVUVz7OT54=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/kms",
        sum = "h1:lh6qra6oC4AyWe5fUUUBe/S27k12OHAleOOOw6KakdE=",
        version = "v1.15.2",
    )
    go_repository(
        name = "com_google_cloud_go_language",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/language",
        sum = "h1:KnYolG0T5Oex722ZW/sP5QErhVAVNcqpJ16tVJd9RTw=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_lifesciences",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/lifesciences",
        sum = "h1:axkANGx1wiBXHiPcJZAE+TDjjYoJRIDzbHC/WYllCBU=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_google_cloud_go_logging",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/logging",
        sum = "h1:26skQWPeYhvIasWKm48+Eq7oUqdcdbwsCVwz5Ys0FvU=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_longrunning",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/longrunning",
        sum = "h1:Fr7TXftcqTudoyRJa113hyaqlGdiBQkp0Gq7tErFDWI=",
        version = "v0.5.1",
    )
    go_repository(
        name = "com_google_cloud_go_managedidentities",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/managedidentities",
        sum = "h1:2/qZuOeLgUHorSdxSQGtnOu9xQkBn37+j+oZQv/KHJY=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_maps",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/maps",
        sum = "h1:PdfgpBLhAoSzZrQXP+/zBc78fIPLZSJp5y8+qSMn2UU=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_mediatranslation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/mediatranslation",
        sum = "h1:50cF7c1l3BanfKrpnTCaTvhf+Fo6kdF21DG0byG7gYU=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_memcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/memcache",
        sum = "h1:7lkLsF0QF+Mre0O/NvkD9Q5utUNwtzvIYjrOLOs0HO0=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_metastore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/metastore",
        sum = "h1:+9DsxUOHvsqvC0ylrRc/JwzbXJaaBpfIK3tX0Lx8Tcc=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_google_cloud_go_monitoring",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/monitoring",
        sum = "h1:rlndy4K8yknMY9JuGe2aK4SbCh21FXoCdX7SAGHmRgI=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkconnectivity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networkconnectivity",
        sum = "h1:kG2PX6URJ9Kvotfdm+hH8WIhrRY77sAKytUGOz+MgN0=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkmanagement",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networkmanagement",
        sum = "h1:aA6L8aioyM4S6nlPYzp2SvB88lBcByZmqMJM6ReafzU=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_networksecurity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networksecurity",
        sum = "h1:TBLEkMp3AE+6IV/wbIGRNTxnqLXHCTEQWoxRVC18TzY=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_google_cloud_go_notebooks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/notebooks",
        sum = "h1:6x2K1JAWv6RW2yQO6oa+xtKUGOpGQseCmT94vpOt1vc=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_optimization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/optimization",
        sum = "h1:sGvPVtBJUKNYAwldhJvFmnM+EEdOXjDzjcly3g0n0Xg=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_google_cloud_go_orchestration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/orchestration",
        sum = "h1:KmN18kE/xa1n91cM5jhCh7s1/UfIguSCisw7nTMUzgE=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_orgpolicy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/orgpolicy",
        sum = "h1:I/7dHICQkNwym9erHqmlb50LRU588NPCvkfIY0Bx9jI=",
        version = "v1.11.1",
    )
    go_repository(
        name = "com_google_cloud_go_osconfig",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/osconfig",
        sum = "h1:dgyEHdfqML6cUW6/MkihNdTVc0INQst0qSE8Ou1ub9c=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_google_cloud_go_oslogin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/oslogin",
        sum = "h1:LdSuG3xBYu2Sgr3jTUULL1XCl5QBx6xwzGqzoDUw1j0=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_phishingprotection",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/phishingprotection",
        sum = "h1:aK/lNmSd1vtbft/vLe2g7edXK72sIQbqr2QyrZN/iME=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_policytroubleshooter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/policytroubleshooter",
        sum = "h1:pT4qSiL5o0hBSWHDiOcmes/s301PeLLWEhAr/eMQB/g=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_privatecatalog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/privatecatalog",
        sum = "h1:B/18xGo+E0EMS9LOEQ0zXz7F2asMgmVgTYGSI89MHOA=",
        version = "v0.9.1",
    )

    go_repository(
        name = "com_google_cloud_go_pubsub",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/pubsub",
        sum = "h1:6SPCPvWav64tj0sVX/+npCBKhUi/UjJehy9op/V3p2g=",
        version = "v1.33.0",
    )
    go_repository(
        name = "com_google_cloud_go_pubsublite",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/pubsublite",
        sum = "h1:pX+idpWMIH30/K7c0epN6V703xpIcMXWRjKJsz0tYGY=",
        version = "v1.8.1",
    )

    go_repository(
        name = "com_google_cloud_go_recaptchaenterprise_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recaptchaenterprise/v2",
        sum = "h1:IGkbudobsTXAwmkEYOzPCQPApUCsN4Gbq3ndGVhHQpI=",
        version = "v2.7.2",
    )
    go_repository(
        name = "com_google_cloud_go_recommendationengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recommendationengine",
        sum = "h1:nMr1OEVHuDambRn+/y4RmNAmnR/pXCuHtH0Y4tCgGRQ=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_recommender",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recommender",
        sum = "h1:SuzbMJhDAiPro7tR9QP7EX97+TI31urjsIgNh9XQHl8=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_google_cloud_go_redis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/redis",
        sum = "h1:YrjQnCC7ydk+k30op7DSjSHw1yAYhqYXFcOq1bSXRYA=",
        version = "v1.13.1",
    )
    go_repository(
        name = "com_google_cloud_go_resourcemanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/resourcemanager",
        sum = "h1:QIAMfndPOHR6yTmMUB0ZN+HSeRmPjR/21Smq5/xwghI=",
        version = "v1.9.1",
    )
    go_repository(
        name = "com_google_cloud_go_resourcesettings",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/resourcesettings",
        sum = "h1:Fdyq418U69LhvNPFdlEO29w+DRRjwDA4/pFamm4ksAg=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_retail",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/retail",
        sum = "h1:gYBrb9u/Hc5s5lUTFXX1Vsbc/9BEvgtioY6ZKaK0DK8=",
        version = "v1.14.1",
    )
    go_repository(
        name = "com_google_cloud_go_run",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/run",
        sum = "h1:kHeIG8q+N6Zv0nDkBjSOYfK2eWqa5FnaiDPH/7/HirE=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_google_cloud_go_scheduler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/scheduler",
        sum = "h1:yoZbZR8880KgPGLmACOMCiY2tPk+iX4V/dkxqTirlz8=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_secretmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/secretmanager",
        sum = "h1:cLTCwAjFh9fKvU6F13Y4L9vPcx9yiWPyWXE4+zkuEQs=",
        version = "v1.11.1",
    )
    go_repository(
        name = "com_google_cloud_go_security",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/security",
        sum = "h1:jR3itwycg/TgGA0uIgTItcVhA55hKWiNJxaNNpQJaZE=",
        version = "v1.15.1",
    )
    go_repository(
        name = "com_google_cloud_go_securitycenter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/securitycenter",
        sum = "h1:XOGJ9OpnDtqg8izd7gYk/XUhj8ytjIalyjjsR6oyG0M=",
        version = "v1.23.0",
    )

    go_repository(
        name = "com_google_cloud_go_servicedirectory",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/servicedirectory",
        sum = "h1:pBWpjCFVGWkzVTkqN3TBBIqNSoSHY86/6RL0soSQ4z8=",
        version = "v1.11.0",
    )

    go_repository(
        name = "com_google_cloud_go_shell",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/shell",
        sum = "h1:aHbwH9LSqs4r2rbay9f6fKEls61TAjT63jSyglsw7sI=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_spanner",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/spanner",
        sum = "h1:+HY8C4uztU7XyLz3xMi/LCXdetLEOExhvRFJu2NiVXM=",
        version = "v1.49.0",
    )
    go_repository(
        name = "com_google_cloud_go_speech",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/speech",
        sum = "h1:MCagaq8ObV2tr1kZJcJYgXYbIn8Ai5rp42tyGYw9rls=",
        version = "v1.19.0",
    )
    go_repository(
        name = "com_google_cloud_go_storage",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/storage",
        sum = "h1:+S3LjjEN2zZ+L5hOwj4+1OkGCsLVe0NzpXKQ1pSdTCI=",
        version = "v1.31.0",
    )
    go_repository(
        name = "com_google_cloud_go_storagetransfer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/storagetransfer",
        sum = "h1:+ZLkeXx0K0Pk5XdDmG0MnUVqIR18lllsihU/yq39I8Q=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_google_cloud_go_talent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/talent",
        sum = "h1:j46ZgD6N2YdpFPux9mc7OAf4YK3tiBCsbLKc8rQx+bU=",
        version = "v1.6.2",
    )
    go_repository(
        name = "com_google_cloud_go_texttospeech",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/texttospeech",
        sum = "h1:S/pR/GZT9p15R7Y2dk2OXD/3AufTct/NSxT4a7nxByw=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_tpu",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/tpu",
        sum = "h1:kQf1jgPY04UJBYYjNUO+3GrZtIb57MfGAW2bwgLbR3A=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_trace",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/trace",
        sum = "h1:EwGdOLCNfYOOPtgqo+D2sDLZmRCEO1AagRTJCU6ztdg=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_translate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/translate",
        sum = "h1:0na4gC54Lu05ir00dmUSuMkLAojDe1ALq4hBTUkhwjE=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_video",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/video",
        sum = "h1:AkjXyJfQ7DtPyDOAbTMeiGcuKsO8/iKSb3fAmTUHYSg=",
        version = "v1.20.0",
    )
    go_repository(
        name = "com_google_cloud_go_videointelligence",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/videointelligence",
        sum = "h1:MBMWnkQ78GQnRz5lfdTAbBq/8QMCF3wahgtHh3s/J+k=",
        version = "v1.11.1",
    )

    go_repository(
        name = "com_google_cloud_go_vision_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vision/v2",
        sum = "h1:ccK6/YgPfGHR/CyESz1mvIbsht5Y2xRsWCPqmTNydEw=",
        version = "v2.7.2",
    )
    go_repository(
        name = "com_google_cloud_go_vmmigration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vmmigration",
        sum = "h1:gnjIclgqbEMc+cF5IJuPxp53wjBIlqZ8h9hE8Rkwp7A=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_vmwareengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vmwareengine",
        sum = "h1:qsJ0CPlOQu/3MFBGklu752v3AkD+Pdu091UmXJ+EjTA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_google_cloud_go_vpcaccess",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vpcaccess",
        sum = "h1:ram0GzjNWElmbxXMIzeOZUkQ9J8ZAahD6V8ilPGqX0Y=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_google_cloud_go_webrisk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/webrisk",
        sum = "h1:Ssy3MkOMOnyRV5H2bkMQ13Umv7CwB/kugo3qkAX83Fk=",
        version = "v1.9.1",
    )
    go_repository(
        name = "com_google_cloud_go_websecurityscanner",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/websecurityscanner",
        sum = "h1:CfEF/vZ+xXyAR3zC9iaC/QRdf1MEgS20r5UR17Q4gOg=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_google_cloud_go_workflows",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/workflows",
        sum = "h1:cSUlx4PVV9O0vYCl+pHAUmu0996A7eN602d4wjjVHRs=",
        version = "v1.12.0",
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
        name = "com_sslmate_software_src_go_pkcs12",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "software.sslmate.com/src/go-pkcs12",
        sum = "h1:nlFkj7bTysH6VkC4fGphtjXRbezREPgrHuJG20hBGPE=",
        version = "v0.2.0",
    )

    go_repository(
        name = "dev_gocloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gocloud.dev",
        sum = "h1:PRgA+DXUz8/uuTJDA7wc8o2Hwj9yZ2qAsShZ60esbE8=",
        version = "v0.30.0",
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
        sum = "h1:0HIbH907iBTAntm+88IJV2qmJALDAh8sPekI9Vc1fm0=",
        version = "v1.2.0",
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
        sum = "h1:bBRl1b0OH9s/DuPhuXpNl+VtCaJXFZ5/uEFST95x9zc=",
        version = "v2.2.1",
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
        name = "io_etcd_go_etcd_api_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/api/v3",
        sum = "h1:4wSsluwyTbGGmyjJktOf3wFQoTBIURXHnq9n/G/JQHs=",
        version = "v3.5.9",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_pkg_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/pkg/v3",
        sum = "h1:oidDC4+YEuSIQbsR94rY9gur91UPL6DnxDCIYd2IGsE=",
        version = "v3.5.9",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/v2",
        sum = "h1:YZ2OLi0OvR0H75AcgSUajjd5uqKDKocQUqROTG11jIo=",
        version = "v2.305.9",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/v3",
        sum = "h1:r5xghnU7CwbUxD/fbUtRyJGaYNfDun8sp/gTr1hew6E=",
        version = "v3.5.9",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdctl_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/etcdctl/v3",
        sum = "h1:2A+/xUck9vBtimGaU8SQh62wCuvuIuREHSGBXBEY6QE=",
        version = "v3.5.5",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdutl_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/etcdutl/v3",
        sum = "h1:KpsQnj71ai24ScrGXF0iwdVZmJU61GK1IbH5oDvYy3M=",
        version = "v3.5.5",
    )
    go_repository(
        name = "io_etcd_go_etcd_pkg_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/pkg/v3",
        sum = "h1:6R2jg/aWd/zB9+9JxmijDKStGJAPFsX3e6BeJkMi6eQ=",
        version = "v3.5.9",
    )
    go_repository(
        name = "io_etcd_go_etcd_raft_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/raft/v3",
        sum = "h1:ZZ1GIHoUlHsn0QVqiRysAm3/81Xx7+i2d7nSdWxlOiI=",
        version = "v3.5.9",
    )
    go_repository(
        name = "io_etcd_go_etcd_server_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/server/v3",
        sum = "h1:vomEmmxeztLtS5OEH7d0hBAg4cjVIu9wXuNzUZx2ZA0=",
        version = "v3.5.9",
    )
    go_repository(
        name = "io_etcd_go_etcd_tests_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/tests/v3",
        sum = "h1:QMfo2twT9Erol77/aypdJGN1vtuQ4VNSGnb5cRiIRo8=",
        version = "v3.5.5",
    )
    go_repository(
        name = "io_etcd_go_etcd_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/v3",
        sum = "h1:Dd0pMrzlu2T0FsxDSomE4+8PNxpNJFLKP/cMrZiK/9s=",
        version = "v3.5.5",
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
        name = "io_k8s_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/api",
        sum = "h1:9mpl5mOb6vXZvqbQmankOfPIGiudghwCoLl1EYfUZbw=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_apiextensions_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiextensions-apiserver",
        sum = "h1:J6/QRWIKV2/HwBhHRVITMLYoypCoPY1ftigDM0Kn+QU=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_apimachinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apimachinery",
        sum = "h1:KCOJLrc6gu+wV1BYgwik4AF4vXOlVJPdiqn0yAWWwXQ=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiserver",
        sum = "h1:rBeYkLvF94Nku9XfXyUIirsVzCzJBs6jMn3NWeHieyI=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_cli_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cli-runtime",
        sum = "h1:64meB2fDj10/ThIMEJLO29a1oujSm0GQmKzh1RtA/uk=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_client_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/client-go",
        sum = "h1:DNoYI1vGq0slMBN/SWKMZMw0Rq+0EQW6/AK4v9+3VeY=",
        version = "v0.28.2",
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
        sum = "h1:yk1XIWt/mbMgNHFdxd0HyVPq/rnJK7BS3oXj24gHClU=",
        version = "v0.27.3",
    )
    go_repository(
        name = "io_k8s_code_generator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/code-generator",
        sum = "h1:u47guga1rCWLnEnffF09p+cqj8B20oHOLoQ1lb1HGtQ=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_component_base",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/component-base",
        sum = "h1:Yc1yU+6AQSlpJZyvehm/NkJBII72rzlEsd6MkBQ+G0E=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_component_helpers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/component-helpers",
        sum = "h1:r/XJ265PMirW9EcGXr/F+2yWrLPo2I69KdvcY/h9HAo=",
        version = "v0.28.2",
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
        sum = "h1:KWO+U8MfI9drXB/P4oU9VchaWYOlwDglJZVHWMpTT3Q=",
        version = "v0.27.1",
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
        sum = "h1:KhG63LHopCdzs1oKA1j+NWleuIXudgOyCqJo4yi3GaM=",
        version = "v0.28.2",
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
        sum = "h1:LyMgNKD2P8Wn1iAwQU5OhxCKlKJy0sHc+PcDwFB24dQ=",
        version = "v0.0.0-20230717233707-2695361300d9",
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
        sum = "h1:fOWOtU6S0smdNjG1PB9WFbqEIMlkzU5ahyHkc7ESHgM=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_kubelet",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubelet",
        sum = "h1:5WhTV1iiBu9q/rr+gvy65LQ+K/e7dmgcaYjys5ipLqY=",
        version = "v0.27.3",
    )
    go_repository(
        name = "io_k8s_kubernetes",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubernetes",
        sum = "h1:gwufSj7y6X18Q2Gl8v4Ev+AJHdzWkG7A8VNFffS9vu0=",
        version = "v1.27.3",
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
        sum = "h1:Z/oMk5SmiT/Ji1SaWOPfW2l9W831BLO9/XxDq9iS3ak=",
        version = "v0.28.2",
    )
    go_repository(
        name = "io_k8s_mount_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/mount-utils",
        sum = "h1:oubkDKLTZUneW27wgyOmp8a1AAZj04vGmtq+YW8wdvY=",
        version = "v0.27.3",
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
        sum = "h1:XX3Ajgzov2RKUdc5jW3t5jwY7Bo7dcRm+tFxT+NfgY0=",
        version = "v0.13.5-0.20230601165947-6ce0bf390ce3",
    )

    go_repository(
        name = "io_k8s_sigs_kustomize_kustomize_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/kustomize/kustomize/v5",
        sum = "h1:vq2TtoDcQomhy7OxXLUOzSbHMuMYq0Bjn93cDtJEdKw=",
        version = "v5.0.4-0.20230601165947-6ce0bf390ce3",
    )

    go_repository(
        name = "io_k8s_sigs_kustomize_kyaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/kustomize/kyaml",
        sum = "h1:W6cLQc5pnqM7vh3b7HvGNfXrJ/xL6BDMS0v1V/HHg5U=",
        version = "v0.14.3-0.20230601165947-6ce0bf390ce3",
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
        name = "io_opencensus_go_contrib_exporter_stackdriver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "contrib.go.opencensus.io/exporter/stackdriver",
        sum = "h1:bjBKzIf7/TAkxd7L2utGaLM78bmUWlCval5K9UeElbY=",
        version = "v0.13.12",
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
        name = "io_opentelemetry_go_otel_sdk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/sdk",
        sum = "h1:PDCppFRDq8A1jL9v6KMI6dYesaq+DFcDZvjsoGvxGzY=",
        version = "v1.14.0",
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
        sum = "h1:djpBY2/2Cs1PV87GSJlxv4voajVOMZxqqtq9AB8YNvY=",
        version = "v1.2.4",
    )
    go_repository(
        name = "net_starlark_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.starlark.net",
        sum = "h1:VdD38733bfYv5tUZwEIskMM93VanwNIi5bIKnDrJdEY=",
        version = "v0.0.0-20230525235612-a134d8f9ddca",
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
        sum = "h1:HBq4TZlN4/1pNcu0geJZ/Q50vIwIXT532UIMYoo0vOs=",
        version = "v0.148.0",
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
        sum = "h1:SeZZZx0cP0fqUyA+oRzP9k7cSwJlvDFiROO72uwD6i0=",
        version = "v0.0.0-20231002182017-d307bd883b97",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto/googleapis/api",
        sum = "h1:W18sezcAYs+3tDZX4F80yctqa12jcP1PUS2gQu1zTPU=",
        version = "v0.0.0-20231002182017-d307bd883b97",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_bytestream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto/googleapis/bytestream",
        sum = "h1:ptblGJZY+SBRhJRVneg4DiHYKwjgeHedTzrgk8CH7m4=",
        version = "v0.0.0-20231012201019-e917dd12ba7a",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_rpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto/googleapis/rpc",
        sum = "h1:a2MQQVoTo96JC9PMGtGBymLp7+/RzpFc2yX/9WfFg1c=",
        version = "v0.0.0-20231012201019-e917dd12ba7a",
    )

    go_repository(
        name = "org_golang_google_grpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc",
        sum = "h1:Z5Iec2pjwb+LEOqzpB2MR12/eKFhDPhuqW91O+4bwUk=",
        version = "v1.59.0",
    )
    go_repository(
        name = "org_golang_google_grpc_cmd_protoc_gen_go_grpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
        sum = "h1:M1YKkFIboKNieVO5DLUEVzQfGwJD30Nv2jfUgzb5UcE=",
        version = "v1.1.0",
    )
    go_repository(
        name = "org_golang_google_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/protobuf",
        sum = "h1:g0LDEJHgrBl9N9r17Ru3sqWhkIx2NB67okBHPwC7hs8=",
        version = "v1.31.0",
    )

    go_repository(
        name = "org_golang_x_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/crypto",
        sum = "h1:wBqGXzWJW6m1XrIKlAH0Hs1JJ7+9KBwnIO8v66Q9cHc=",
        version = "v0.14.0",
    )
    go_repository(
        name = "org_golang_x_exp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/exp",
        sum = "h1:UA2aFVmmsIlefxMk29Dp2juaUSth8Pyn3Tq5Y5mJGME=",
        version = "v0.0.0-20230626212559-97b1e661b5df",
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
        sum = "h1:I/DsJXRlw/8l/0c24sM9yb0T4z9liZTduXvdAWYiysY=",
        version = "v0.13.0",
    )
    go_repository(
        name = "org_golang_x_net",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/net",
        sum = "h1:pVaXccu2ozPjCXewfr1S7xza/zcXTity9cCdXQYSjIM=",
        version = "v0.17.0",
    )
    go_repository(
        name = "org_golang_x_oauth2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/oauth2",
        sum = "h1:jDDenyj+WgFtmV3zYVoi8aE2BwtXFLWOA67ZfNWftiY=",
        version = "v0.13.0",
    )
    go_repository(
        name = "org_golang_x_sync",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sync",
        sum = "h1:zxkM55ReGkDlKSM+Fu41A+zmbZuaPVbGMzvvdUPznYQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "org_golang_x_sys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sys",
        sum = "h1:Af8nKPmuFypiUBjVoU9V20FiaFXOcuZI21p0ycVYYGE=",
        version = "v0.13.0",
    )
    go_repository(
        name = "org_golang_x_term",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/term",
        sum = "h1:bb+I9cTfFazGW51MZqBVmZy7+JEJMouUHTUSKVQLBek=",
        version = "v0.13.0",
    )
    go_repository(
        name = "org_golang_x_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/text",
        sum = "h1:ablQoSUd0tRdKxZewP80B+BaqeKJuVhuRxj/dkrun3k=",
        version = "v0.13.0",
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
        sum = "h1:jvNa2pY0M4r62jkRQ6RwEZZyPcymeL9XZMLBbV7U2nc=",
        version = "v0.14.0",
    )
    go_repository(
        name = "org_golang_x_vuln",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/vuln",
        sum = "h1:KUas02EjQK5LTuIx1OylBQdKKZ9jeugs+HiqO5HormU=",
        version = "v1.0.1",
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
        sum = "h1:u+CHhs2OhVmu0MWzBDrlbLzQ5QB3ZfWtfT+lD3EaUIs=",
        version = "v1.9004.0",
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
        sum = "h1:ZvwS0R+56ePWxUNi+Atn9dWONBPp/AUETXlHW0DxSjE=",
        version = "v1.11.0",
    )

    go_repository(
        name = "org_uber_go_goleak",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/goleak",
        sum = "h1:2K3zAYmnTNqV73imy9J1T3WC+gmCePx2hEGkimedGto=",
        version = "v1.3.0",
    )
    go_repository(
        name = "org_uber_go_multierr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/multierr",
        sum = "h1:blXXJkSxSSfBVBlC76pxqeO+LN3aDfLQo+309xJstO0=",
        version = "v1.11.0",
    )

    go_repository(
        name = "org_uber_go_zap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/zap",
        sum = "h1:sI7k6L95XOKS281NhVKOFCUNIvv9e0w4BF8N3u+tCRo=",
        version = "v1.26.0",
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
        sum = "h1:DG+XLGzBJeZvMLlMbm6bPDLV1dGaVW9eZsDoUd1/LM0=",
        version = "v3.13.1",
    )
    go_repository(
        name = "sm_step_go_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.step.sm/crypto",
        sum = "h1:EhJpFRNgU3RaNEO3WZ62Kn2gF9NWNglNG4DvSPeuiTs=",
        version = "v0.32.2",
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
