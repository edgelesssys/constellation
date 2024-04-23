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
        sum = "h1:zjqpY4C7H15HjRPEenkS4SAn3Jy2eRRjkjZbGR30TOg=",
        version = "v0.0.0-20230919221257-8b5d3ce2d11d",
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
        sum = "h1:YB2fHEn0UJagG8T1rrWknE3ZQzWM06O8AMAatNn7lmo=",
        version = "v1.2.3",
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
        name = "com_github_akavel_rsrc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/akavel/rsrc",
        sum = "h1:Zxm8V5eI1hW4gGaYsJQUhxpjkENuG91ki8B4zCrvEsw=",
        version = "v0.10.2",
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
        sum = "h1:f48lwail6p8zpO1bC4TxtqACaGqHYA22qkHjHpqDjYY=",
        version = "v2.4.0",
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
        name = "com_github_anmitsu_go_shlex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/anmitsu/go-shlex",
        sum = "h1:kFOfPq6dUM1hTo4JG6LR5AXSUEsOjtdm0kw0FtQtMJA=",
        version = "v0.0.0-20161002113705-648efa622239",
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
        name = "com_github_antlr_antlr4_runtime_go_antlr_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/antlr/antlr4/runtime/Go/antlr/v4",
        sum = "h1:goHVqTbFX3AIo0tzGr14pgfAW2ZfPChKO21Z9MGf/gk=",
        version = "v4.0.0-20230512164433-5d1fd1a340c9",
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
        name = "com_github_apparentlymart_go_textseg_v12",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apparentlymart/go-textseg/v12",
        sum = "h1:bNEQyAGak9tojivJNkoqWErVCQbjdL7GzRt3F8NvfJ0=",
        version = "v12.0.0",
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
        name = "com_github_apparentlymart_go_textseg_v15",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/apparentlymart/go-textseg/v15",
        sum = "h1:uYvfpb3DyLSCGWnctWKGj857c6ew1u1fNQOlOtuGxQY=",
        version = "v15.0.0",
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
        sum = "h1:BUhSaO2qLk2jkcyLebcvDmbdOunVe/Wq8RsCyI8szL0=",
        version = "v1.50.22",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2",
        sum = "h1:sv7+1JVJxOu/dD/sz/csHX7jFqmP001TIY7aytBWDSQ=",
        version = "v1.25.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_aws_protocol_eventstream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream",
        sum = "h1:2UO6/nT1lCZq1LqM67Oa4tdgP1CvL1sLSxvuD+VrOeE=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_config",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/config",
        sum = "h1:oxvGd/cielb+oumJkQmXI0i5tQCRqfdCHV58AfE0pGY=",
        version = "v1.27.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_credentials",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/credentials",
        sum = "h1:H4WlK2OnVotRmbVgS8Ww2Z4B3/dDHxDS7cW6EiCECN4=",
        version = "v1.17.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_ec2_imds",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/ec2/imds",
        sum = "h1:xWCwjjvVz2ojYTP4kBKUuUh9ZrXfcAXpflhOUUeXg1k=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_s3_manager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/s3/manager",
        sum = "h1:PYWtYuCP+gYQ576MS4QRn7y1+kp+OZzzG7jlwnFj1wQ=",
        version = "v1.16.3",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_configsources",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/configsources",
        sum = "h1:NPs/EqVO+ajwOoq56EfcGKa3L3ruWuazkIw1BqxwOPw=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_endpoints_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/endpoints/v2",
        sum = "h1:ks7KGMVUMoDzcxNWUlEdI+/lokMFD136EL6DWmUOV80=",
        version = "v2.6.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_ini",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/ini",
        sum = "h1:hT8rVHwugYE2lEfdFE0QWVo81lF7jMrYJVDWI+f+VxU=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_v4a",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/v4a",
        sum = "h1:TkbRExyKSVHELwG9gz2+gql37jjec2R5vus9faTomwE=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_autoscaling",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/autoscaling",
        sum = "h1:gdukBEVzo0O/0UiR0ee5zQokJ7RIP0p1jF00ayKHZ4o=",
        version = "v1.39.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_cloudfront",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/cloudfront",
        sum = "h1:fGpjGBqtfTz2mymcChLB42StEw0vHwsHqDFnctmoOQ8=",
        version = "v1.34.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ec2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ec2",
        sum = "h1:OGzK1PwB0sCE2Zwy6ISs/XSul4lrujQf3doXvmGqCwg=",
        version = "v1.148.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_elasticloadbalancingv2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2",
        sum = "h1:V+AqIZnytQg3jgEBIbvLYzxcMagTvC6kzhex0ZbDcTE=",
        version = "v1.29.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_accept_encoding",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding",
        sum = "h1:a33HuFlO0KsveiP90IUJh8Xr/cx9US2PqkSroaLc+o8=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_checksum",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/checksum",
        sum = "h1:UiSyK6ent6OKpkMJN3+k5HZ4sk4UfchEaaW5wv7SblQ=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_presigned_url",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/presigned-url",
        sum = "h1:SHN/umDLTmFTmYfI+gkanz6da3vK8Kvj/5wkqnTHbuA=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_s3shared",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/s3shared",
        sum = "h1:l5puwOHr7IxECuPMIuZG7UKOzAnF24v6t4l+Z5Moay4=",
        version = "v1.17.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/kms",
        sum = "h1:W9PbZAZAEcelhhjb7KuwUtf+Lbc+i7ByYJRuWLlnxyQ=",
        version = "v1.27.9",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_resourcegroupstaggingapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi",
        sum = "h1:TAKRHjyAtRMUeqsPnjzI4EXz3WtIo3IXRhJiIPa4MFo=",
        version = "v1.20.2",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_s3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/s3",
        sum = "h1:UxJGNZ+/VhocG50aui1p7Ub2NjDzijCpg8Y3NuznijM=",
        version = "v1.50.2",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_secretsmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/secretsmanager",
        sum = "h1:Wq73CAj0ktbUHufBTar4uMVzP7JHraTq6ZMloCAQxRk=",
        version = "v1.27.2",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sso",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sso",
        sum = "h1:GokXLGW3JkH/XzEVp1jDVRxty1eNGB7emkjDG1qxGK8=",
        version = "v1.19.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ssooidc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ssooidc",
        sum = "h1:2oxSGiYNxTHsuRuPD9McWvcvR6s61G3ssZLyQzcxQL0=",
        version = "v1.22.1",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sts",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sts",
        sum = "h1:QFT2KUWaVwwGi5/2sQNBOViFpLSkZmiyiHUxE2k6sOU=",
        version = "v1.27.1",
    )
    go_repository(
        name = "com_github_aws_smithy_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/smithy-go",
        sum = "h1:6+kZsCXZwKxZS9RfISnPc4EXlHoyAkm2hPuM8X2BrrQ=",
        version = "v1.20.0",
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
        sum = "h1:c4k2FIYIh4xtwqrQwV0Ct1v5+ehlNXj5NI/MWVsiTkQ=",
        version = "v1.9.2",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_azidentity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/azidentity",
        sum = "h1:sO0/P7g68FrryJzljemN+6GTssUXdANk6aJ7T1ZxnsQ=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/internal",
        sum = "h1:LqbJ/WzJUwBf8UiaSzgX7aMclParm9/5Vgp+TY51uBQ=",
        version = "v1.5.2",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_keyvault_azkeys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys",
        sum = "h1:m/sWOGCREuSBqg2htVQTBY8nOZpyajYztF0vUvSZTuM=",
        version = "v0.10.0",
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
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_compute_armcompute_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5",
        sum = "h1:MxA59PGoCFb+vCwRQi3PhQEwHj4+r2dhuv9HG+vM7iM=",
        version = "v5.5.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_internal_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/internal/v2",
        sum = "h1:PTFGRSlMKCQelWwxUyYVEUqseBJVemLyqWJjvMyt0do=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_network_armnetwork_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5",
        sum = "h1:9CrwzqQ+e8EqD+A2bh547GjBU4K0o30FhiTB981LFNI=",
        version = "v5.0.0",
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
        name = "com_github_azure_azure_sdk_for_go_sdk_resourcemanager_storage_armstorage",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage",
        sum = "h1:AifHbc4mg0x9zW52WOpKbsHaDKuRhlI7TVl47thgQ70=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_security_keyvault_azkeys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys",
        sum = "h1:MyVTgWR8qd/Jw1Le0NZebGBUCLbtak3bJ3z1OlqZBpw=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_security_keyvault_azsecrets",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets",
        sum = "h1:h4Zxgmi9oyZL2l8jeg1iRTqPloHktywWcu0nlJmo1tA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_security_keyvault_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal",
        sum = "h1:D3occbWoio4EBLkbkevetNMAVX197GkzbUMtqjGWn80=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_storage_azblob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob",
        sum = "h1:IfFdxTUDiV58iZqPKgyWiz4X4fCxZeQ1pTQPImLYXpY=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go_sdk_storage_azblob_testdata_perf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/testdata/perf",
        sum = "h1:45Ajiuhu6AeJTFdwxn2OWXZTQOHdXT1U/aezrVu6HIM=",
        version = "v0.0.0-20240208231215-981108a6de20",
    )
    go_repository(
        name = "com_github_azure_go_ansiterm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-ansiterm",
        sum = "h1:L/gRVlceqvL25UVaW/CKtUDjefjrs0SPonmDGUVOYP0=",
        version = "v0.0.0-20230124172434-306776ec8161",
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
        sum = "h1:Yepx8CvFxwNKpH6ja7RZ+sKX+DWYNldbLiALMC3BTz8=",
        version = "v0.9.23",
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
        sum = "h1:XHOnouVk1mxXfQidrMEnLlPk9UMeRtyBTnEFtxkV0kU=",
        version = "v1.2.2",
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
        name = "com_github_bufbuild_protocompile",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bufbuild/protocompile",
        sum = "h1:Uu7WiSQ6Yj9DbkdnOe7U4mNKp58y9WDMKDn28/ZlunY=",
        version = "v0.6.0",
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
        name = "com_github_cavaliercoder_go_rpm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cavaliercoder/go-rpm",
        sum = "h1:jP7ki8Tzx9ThnFPLDhBYAhEpI2+jOURnHQNURgsMvnY=",
        version = "v0.0.0-20200122174316-8cb9fd9c31a8",
    )
    go_repository(
        name = "com_github_cavaliergopher_cpio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cavaliergopher/cpio",
        sum = "h1:KQFSeKmZhv0cr+kawA3a0xTQCU4QxXF1vhU7P7av2KM=",
        version = "v1.0.1",
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
        name = "com_github_chromedp_cdproto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chromedp/cdproto",
        sum = "h1:aPflPkRFkVwbW6dmcVqfgwp1i+UWGFH6VgR1Jim5Ygc=",
        version = "v0.0.0-20230802225258-3cf4e6d46a89",
    )
    go_repository(
        name = "com_github_chromedp_chromedp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chromedp/chromedp",
        sum = "h1:dKtNz4kApb06KuSXoTQIyUC2TrA0fhGDwNZf3bcgfKw=",
        version = "v0.9.2",
    )
    go_repository(
        name = "com_github_chromedp_sysutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chromedp/sysutil",
        sum = "h1:+ZxhTpfpZlmchB58ih/LBHX52ky7w2VhQVKQMucy3Ic=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_chzyer_logex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chzyer/logex",
        sum = "h1:Swpa1K6QvQznwJRcfTfQJmTE72DqScAa40E+fbHEXEE=",
        version = "v1.1.10",
    )
    go_repository(
        name = "com_github_chzyer_readline",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chzyer/readline",
        sum = "h1:upd/6fQk4src78LMRzh5vItIt361/o4uq553V8B5sGI=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_chzyer_test",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/chzyer/test",
        sum = "h1:q763qf9huN11kDQavWsoZXJNW3xEE4JJyHa5Q25/sd8=",
        version = "v0.0.0-20180213035817-a1ea475d72b1",
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
        sum = "h1:qlCDlTPz2n9fu58M0Nh1J/JzcFpfgkFHHX3O35r5vcU=",
        version = "v1.3.7",
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
        sum = "h1:7To3pQ+pZo0i3dsWEbinPNFs5gPSBOsJtx3wTT94VBY=",
        version = "v0.0.0-20231109132714-523115ebc101",
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
        name = "com_github_container_storage_interface_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/container-storage-interface/spec",
        sum = "h1:D0vhF3PLIZwlwZEf2eNbpujGCNwspwTYf2idJRJx4xI=",
        version = "v1.8.0",
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
        sum = "h1:wPYKIeGMN8vaggSKuV1X0wZulpMz4CrgEsZdaCyB6Is=",
        version = "v1.7.13",
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
        name = "com_github_containerd_log",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/log",
        sum = "h1:TCJt7ioM2cr/tfR8GPbGf9/VRAX8D2B4PjzCpfX540I=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_containerd_nri",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/containerd/nri",
        sum = "h1:PjgIBm0RtUiFyEO6JqPBQZRQicbsIz41Fz/5VSC0zgw=",
        version = "v0.4.0",
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
        sum = "h1:2eYKZT7i6yxIfGP3qLJoJ7HAsDJqYB+X68g4NYjSrE0=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_coredns_corefile_migration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coredns/corefile-migration",
        sum = "h1:W/DCETrHDiFo0Wj03EyMkaQ9fwsmSgqTCQDHpceaSsE=",
        version = "v1.0.21",
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
        sum = "h1:0J/ogVOd4y8P0f0xUh8l9t07xRP/d8tccvjHl2dcsSo=",
        version = "v3.9.0",
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
        sum = "h1:3ripxk5ox93TOmYn0WMbddd5XLerG9URonb5XG4GcFU=",
        version = "v0.3.19",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cpuguy83/go-md2man/v2",
        sum = "h1:qMCsGGgs+MAzDFyp9LpAe1Lqy/fY/qCovCm0qnXZOBM=",
        version = "v2.0.3",
    )
    go_repository(
        name = "com_github_creack_pty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/creack/pty",
        sum = "h1:1/QdRyBaHHJP61QkWMXlOIBfsgdDeeKfK8SYVUWJKf0=",
        version = "v1.1.21",
    )
    go_repository(
        name = "com_github_cyberphone_json_canonicalization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cyberphone/json-canonicalization",
        sum = "h1:eHnXnuK47UlSTOQexbzxAZfekVz6i+LKRdj1CU5DPaM=",
        version = "v0.0.0-20231217050601-ba74d44ecf5f",
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
        sum = "h1:dl9cBrupW8+r5250DYkYxocLeZ1Y4vB1kxgtjxw8GQs=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_danwinship_knftables",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/danwinship/knftables",
        sum = "h1:89Ieiia6MMfXWQF9dyaou1CwBU8h8sHa2Zo3OlY2o04=",
        version = "v0.0.13",
    )
    go_repository(
        name = "com_github_data_dog_go_sqlmock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/DATA-DOG/go-sqlmock",
        sum = "h1:OcvFkGmslmlZibjAjaHm3L//6LiuBgolP7OputlJIzU=",
        version = "v1.5.2",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:U9qPSI2PIWSS1VwoXQT9A3Wy9MM3WgvqSxFWenqJduM=",
        version = "v1.1.2-0.20180830191138-d8f796af33cc",
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
        name = "com_github_distribution_distribution_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/distribution/distribution/v3",
        sum = "h1:aBfCb7iqHmDEIp6fBvC/hQUddQfg+3qdYjwzaiP9Hnc=",
        version = "v3.0.0-20221208165359-362910506bc2",
    )
    go_repository(
        name = "com_github_distribution_reference",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/distribution/reference",
        sum = "h1:/FUIFXtfc/x2gpa5/VGfiGLuOIdYa1t65IKK2OFGvA0=",
        version = "v0.5.0",
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
        sum = "h1:KLeNs7zws74oFuVhgZQ5ONGZiXUUdgsdy6/EsX/6284=",
        version = "v25.0.3+incompatible",
    )
    go_repository(
        name = "com_github_docker_distribution",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/distribution",
        sum = "h1:AtKxIZ36LoNK51+Z6RpzLpddBirtxJnzDrHLEKxTAYk=",
        version = "v2.8.3+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/docker",
        sum = "h1:UmQydMduGkrD5nQde1mecF/YnSbTOaPeFIeP5C4W+DE=",
        version = "v25.0.5+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker_credential_helpers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/docker-credential-helpers",
        sum = "h1:j/eKUktUltBtMzKqmfLB0PAgqYyMHOp5vfsD1807oKo=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_github_docker_go_connections",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/go-connections",
        sum = "h1:USnMq7hx7gwdVZq1L49hLXaFtUdTADjXGp+uj1Br63c=",
        version = "v0.5.0",
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
        sum = "h1:UhxFibDNY/bfvqU5CAUmr9zpesgbU6SWc8/B4mflAE4=",
        version = "v0.0.0-20160708172513-aabc10ec26b7",
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
        sum = "h1:TCGUmmH50cQBGXPJsn32APf93fmWQXcSMi7pMbDPtV0=",
        version = "v0.0.0-20240123150912-dcad3c41ec5f",
    )
    go_repository(
        name = "com_github_eggsampler_acme_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/eggsampler/acme/v3",
        sum = "h1:LHWnB3wShVshK1+umL6ObCjnc0MM+D7TE8JINjk8zGY=",
        version = "v3.4.0",
    )
    go_repository(
        name = "com_github_emicklei_go_restful_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/emicklei/go-restful/v3",
        sum = "h1:1onLa9DcsMYO9P+CXaL0dStDqQ2EHHXLiz+BtnqkLAU=",
        version = "v3.11.2",
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
        sum = "h1:fBXyNpNMuTTDdquAq/uisOr2lShz4oaXpDTX2bLe7ls=",
        version = "v5.9.0+incompatible",
    )
    go_repository(
        name = "com_github_evanphx_json_patch_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/evanphx/json-patch/v5",
        sum = "h1:kcBlZQbplgElYIlo/n1hJbls2z/1awpXxpRi0/FOJfg=",
        version = "v5.9.0",
    )
    go_repository(
        name = "com_github_exponent_io_jsonpath",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/exponent-io/jsonpath",
        sum = "h1:Wl78ApPPB2Wvf/TIe2xdyJxTlb6obmF18d8QdkxNDu4=",
        version = "v0.0.0-20210407135951-1de76d718b3f",
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
        sum = "h1:zmkK9Ngbjj+K0yRhTVONQh1p/HknKYSlNT+vZCzyokM=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_github_felixge_httpsnoop",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/felixge/httpsnoop",
        sum = "h1:NFTV2Zj1bL4mc9sqWACXbQFVBBg2W3GPvqp8/ESS2Wg=",
        version = "v1.0.4",
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
        name = "com_github_foxboron_go_uefi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/foxboron/go-uefi",
        sum = "h1:qGlg/7H49H30Eu7nkCBA7YxNmW30ephqBf7xIxlAGuQ=",
        version = "v0.0.0-20240128152106-48be911532c2",
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
        sum = "h1:7Xjx+VpznH+oBnejlPUj8oUpdxnVs4f8XU8WnHkI4W8=",
        version = "v1.14.6",
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
        sum = "h1:JMvZXK8lHDGyLmTQ0ZdGDnVVGuwjbpaumf8p42z0d+c=",
        version = "v1.8.9",
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
        sum = "h1:in2uUcidCuFcDKtdcBxlR0rJ1+fsokWf+uqxgUFjbI0=",
        version = "v1.4.3",
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
        sum = "h1:6zsha5zo/TWhRhwqCD3+EarCAgZ2yN28ipRnGPnwkI0=",
        version = "v0.2.2",
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
        sum = "h1:ZwEMSLRCapFLflTpT7NKaAc7ukJ8ZPEjzlxt8rPN8bk=",
        version = "v1.5.1",
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
        sum = "h1:yEY4yhzCDuMGSv83oGxiBotRzhwhNr8VZyphhiu+mTU=",
        version = "v5.5.0",
    )
    go_repository(
        name = "com_github_go_git_go_git_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-git/go-git/v5",
        sum = "h1:XIZc1p+8YzypNr34itUfSvYJcv+eYdTnTvOZ2vD3cA4=",
        version = "v5.11.0",
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
        sum = "h1:fFKWeig/irsp7XD2zBxvnmA/XaRWp5V3CBsZXJF7G7k=",
        version = "v3.0.3",
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
        sum = "h1:wGYYu3uicYdqXVgoYbvnkrPVXkuLM1p1ifugDMEdRi4=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_go_logr_logr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logr/logr",
        sum = "h1:pKouT5E8xu9zeFC39JXRDukb6JFQPXM5p5I91188VAQ=",
        version = "v1.4.1",
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
        sum = "h1:XGdV8XW8zdwFiwOA2Dryh1gj2KRQyOOoNmBy4EplIcQ=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_go_openapi_analysis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/analysis",
        sum = "h1:ZBmNoP2h5omLKr/srIC9bfqrUGzT6g6gNv03HE9Vpj0=",
        version = "v0.22.2",
    )
    go_repository(
        name = "com_github_go_openapi_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/errors",
        sum = "h1:FhChC/duCnfoLj1gZ0BgaBmzhJC2SL/sJr8a2vAobSY=",
        version = "v0.21.0",
    )
    go_repository(
        name = "com_github_go_openapi_jsonpointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/jsonpointer",
        sum = "h1:mQc3nmndL8ZBzStEo3JYF8wzmeWffDH4VbXz58sAx6Q=",
        version = "v0.20.2",
    )
    go_repository(
        name = "com_github_go_openapi_jsonreference",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/jsonreference",
        sum = "h1:bKlDxQxQJgwpUSgOENiMPzCTBVuc7vTdXSSgNeAhojU=",
        version = "v0.20.4",
    )
    go_repository(
        name = "com_github_go_openapi_loads",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/loads",
        sum = "h1:jDzF4dSoHw6ZFADCGltDb2lE4F6De7aWSpe+IcsRzT0=",
        version = "v0.21.5",
    )
    go_repository(
        name = "com_github_go_openapi_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/runtime",
        sum = "h1:ae53yaOoh+fx/X5Eaq8cRmavHgDma65XPZuvBqvJYto=",
        version = "v0.27.1",
    )
    go_repository(
        name = "com_github_go_openapi_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/spec",
        sum = "h1:7CBlRnw+mtjFGlPDRZmAMnq35cRzI91xj03HVyUi/Do=",
        version = "v0.20.14",
    )
    go_repository(
        name = "com_github_go_openapi_strfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/strfmt",
        sum = "h1:Ew9PnEYc246TwrEspvBdDHS4BVKXy/AOVsfqGDgAcaI=",
        version = "v0.22.0",
    )
    go_repository(
        name = "com_github_go_openapi_swag",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/swag",
        sum = "h1:XX2DssF+mQKM2DHsbgZK74y/zj4mo9I99+89xUmuZCE=",
        version = "v0.22.9",
    )
    go_repository(
        name = "com_github_go_openapi_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/validate",
        sum = "h1:2l7PJLzCis4YUGEoW6eoQw3WhyM65WSIcjX6SQnlfDw=",
        version = "v0.23.0",
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
        name = "com_github_go_redis_redismock_v9",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-redis/redismock/v9",
        sum = "h1:ZrMYQeKPECZPjOj5u9eyOjg8Nnb0BS9lkVIZ6IpsKLw=",
        version = "v9.2.0",
    )
    go_repository(
        name = "com_github_go_rod_rod",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-rod/rod",
        sum = "h1:1x6oqnslwFVuXJbJifgxspJUd3O4ntaGhRLHt+4Er9c=",
        version = "v0.114.5",
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
        name = "com_github_gobwas_glob",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobwas/glob",
        sum = "h1:A4xDbljILXROh+kObIiy5kIaPYD8e96x1tgBhUI5J+Y=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_gobwas_httphead",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobwas/httphead",
        sum = "h1:exrUm0f4YX0L7EBwZHuCF4GDp8aJfVeBrlLQrs6NqWU=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_gobwas_pool",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobwas/pool",
        sum = "h1:xfeeEhW7pwmX8nuLVlqbzVc7udMDrwetjEv+TZIz1og=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_gobwas_ws",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gobwas/ws",
        sum = "h1:F2aeBZrm2NDsc7vbovKrWSogd4wvfAxg0FQ89/iqOTk=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_goccy_go_json",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/goccy/go-json",
        sum = "h1:IcB+Aqpx/iMHu5Yooh7jEzJk1JZ7Pjtmys2ukPr7EeM=",
        version = "v0.9.7",
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
        sum = "h1:X1e7hUd02GDaLWKZj40Z7L0CP0W9TrGgmPQZw6+anBg=",
        version = "v0.40.4",
    )
    go_repository(
        name = "com_github_godror_knownpb",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/godror/knownpb",
        sum = "h1:A4J7jdx7jWBhJm18NntafzSC//iZDHkDi1+juwQ5pTI=",
        version = "v0.1.1",
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
        sum = "h1:3qXRTX8/NbyulANqlc0lchS1gqAVxRgsuW1YrTJupqA=",
        version = "v4.4.0+incompatible",
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
        sum = "h1:d/ix8ftRUorsN+5eMIlF4T6J8CAt9rch3My2winC1Jw=",
        version = "v5.2.0",
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
        sum = "h1:i7eJL8qZTpSEXOPTxNKhASYpMn+8e5Q6AdndVa1dWek=",
        version = "v1.5.4",
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
        sum = "h1:eyYTxKBd+KxI1kh6rst4JSTLUhfHQM34qGpp+0AMlSg=",
        version = "v0.48.1",
    )
    go_repository(
        name = "com_github_google_cel_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/cel-go",
        sum = "h1:6ebJFzu1xO2n7TLtN+UBqShGBhlD85bhvglh5DpcfqQ=",
        version = "v0.17.7",
    )
    go_repository(
        name = "com_github_google_certificate_transparency_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/certificate-transparency-go",
        sum = "h1:IASD+NtgSTJLPdzkthwvAG1ZVbF2WtFg4IvoA68XGSw=",
        version = "v1.1.7",
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
        sum = "h1:jqtOrLk5MNdliTKjPbIPrAaRKJaKW+0LIU2n/brJYms=",
        version = "v0.5.1",
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
        name = "com_github_google_go_configfs_tsm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-configfs-tsm",
        sum = "h1:YnJ9rXIOj5BYD7/0DNnzs8AOp7UcvjfTvt215EWcs98=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_google_go_containerregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-containerregistry",
        sum = "h1:uIsMRBV7m/HDkDxE/nXMnv1q+lOOSPlQ/ywc5JbB8Ic=",
        version = "v0.19.0",
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
        sum = "h1:gl0KvjdsD4RrJzyLefDOvFOUH3NAJri/3qvaL5m83Iw=",
        version = "v0.3.1",
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
        sum = "h1:EQ1rGgyI8IEBApvDH9HPF7ehUd/6H6SxSNKVDF5z/GU=",
        version = "v0.4.3-0.20240112165732-912a43636883",
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
        name = "com_github_google_goterm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/goterm",
        sum = "h1:CVuJwN34x4xM2aT4sIKhmeib40NeBPhRihNjQmpJsA4=",
        version = "v0.0.0-20200907032337-555d40f16ae2",
    )
    go_repository(
        name = "com_github_google_keep_sorted",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/keep-sorted",
        sum = "h1:nsDd3h16Bf1KFNtfvzGoLaei95AMLswikiw1ICDOKPE=",
        version = "v0.3.0",
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
        sum = "h1:RMpPgZTSApbPf7xaVel+QkoGPRLFLrwFO89uDUHEGf0=",
        version = "v0.0.0-20231023181126-ff6d637d2a7b",
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
        sum = "h1:L16KZ3QvkFGpYhmp23iQip+mx1X39foEsqszjMNBm8A=",
        version = "v0.5.0",
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
        sum = "h1:jMBeDBIkINFvS2n6oV5maDqfRlxREAc6CW9QYWQ0qT4=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_google_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/uuid",
        sum = "h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=",
        version = "v1.6.0",
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
        sum = "h1:Vie5ybvEvT75RniqhfFxPRy3Bf7vr3h0cechB90XaQs=",
        version = "v0.3.2",
    )
    go_repository(
        name = "com_github_googleapis_gax_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/gax-go/v2",
        sum = "h1:9F8GV9r9ztXyAi00gsMQHNoF51xPZm8uj1dpYt2ZETM=",
        version = "v2.12.1",
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
        sum = "h1:zC34cGQu69FG7qzJ3WiKW244WfhDC3xxYMeNOX2gtUQ=",
        version = "v0.0.0-20210719221736-1c9a4c676720",
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
        sum = "h1:zKvmHOmHuaZlnx9d2DJpEgbMxrGt/+CJ/bKOKQh9Xzo=",
        version = "v1.9.0",
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
        sum = "h1:TuBL49tXwgrFYWhqrNgrUNEY92u81SPhu7sTdzQEiWY=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_gorilla_websocket",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gorilla/websocket",
        sum = "h1:gmztn0JnHVt9JZquRuzLw3g4wouNVzKL15iLr/zn/QY=",
        version = "v1.5.1",
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
        name = "com_github_grpc_ecosystem_go_grpc_middleware_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/go-grpc-middleware/v2",
        sum = "h1:HcUWd006luQPljE73d5sk+/VgYPGUReEVz2y1/qylwY=",
        version = "v2.0.1",
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
        sum = "h1:6UKoz5ujsI55KNpsJH3UwCq3T8kKbZwNZBNPuTTje8U=",
        version = "v2.18.1",
    )
    go_repository(
        name = "com_github_hashicorp_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/cli",
        sum = "h1:CMOV+/LJfL1tXCOKrgAX0uRKnzjj/mpmqNXloRSy2K8=",
        version = "v1.1.6",
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
        name = "com_github_hashicorp_go_cty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-cty",
        sum = "h1:1/D3zfFHttUKaCaGKZ/dR2roBXv0vKbSCnssIldfQdI=",
        version = "v1.4.1-0.20200414143053-d3edf31b6320",
    )
    go_repository(
        name = "com_github_hashicorp_go_hclog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-hclog",
        sum = "h1:NOtoftovWkDheyUM/8JW3QMiXyxJK3uHRK7wV04nD2I=",
        version = "v1.6.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/v2",
        sum = "h1:WZeXfD26QMWYC35at25KgE021SF9L3u9UMHK8fJAdV0=",
        version = "v2.0.16",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_wrappers_awskms_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/wrappers/awskms/v2",
        sum = "h1:qdxeZvDMRGZ3YSE4Oz0Pp7WUSUn5S6cWZguEOkEVL50=",
        version = "v2.0.9",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_wrappers_azurekeyvault_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/wrappers/azurekeyvault/v2",
        sum = "h1:/7SKkYIhA8cr3l8m1EKT6Q90bPoSVqqVBuQ6HgoMIkw=",
        version = "v2.0.11",
    )
    go_repository(
        name = "com_github_hashicorp_go_kms_wrapping_wrappers_gcpckms_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2",
        sum = "h1:qXOa2uFzT8eORzgfLZSp1dvig2l/70LJIr6634f5HMM=",
        version = "v2.0.11",
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
        name = "com_github_hashicorp_go_plugin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-plugin",
        sum = "h1:wgd4KxHJTVGGqWBq4QPB1i5BZNEx9BR8+OFmHDmTk8A=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_retryablehttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/go-retryablehttp",
        sum = "h1:bJj+Pj19UZMIweq/iie+1u5YCdGrnxCT9yvm0e+Nd5M=",
        version = "v0.7.5",
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
        sum = "h1:I8bynUKMh9I7JdwtW9voJ0xmHvBpxQtLjrMFDYmhOxY=",
        version = "v0.3.0",
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
        sum = "h1:yE/r1yJvWbtrJ0STwScgEnCanb0U9v7zp0Gbkmcoxqs=",
        version = "v0.6.3",
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
        sum = "h1://i05Jqznmb2EXqa39Nsvyan2o5XyMowW5fnCKW5RPI=",
        version = "v2.19.1",
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
        sum = "h1:DIZnPsqzPGuUnq6cH8jWcPunBfY+C+M8JyYF3vpnuEo=",
        version = "v0.20.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_json",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-json",
        sum = "h1:9NQxbLNqPbEMze+S6+YluEdXgJmhQykRyRNd+zTI05U=",
        version = "v0.21.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_plugin_framework",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-plugin-framework",
        sum = "h1:8kcvqJs/x6QyOFSdeAyEgsenVOUeC/IyKpi2ul4fjTg=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_plugin_framework_validators",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-plugin-framework-validators",
        sum = "h1:HOjBuMbOEzl7snOdOoUfE2Jgeto6JOjLVQ39Ls2nksc=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_plugin_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-plugin-go",
        sum = "h1:VSjdVQYNDKR0l2pi3vsFK1PdMQrw6vGOshJXMNFeVc0=",
        version = "v0.21.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_plugin_log",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-plugin-log",
        sum = "h1:i7hOA+vdAItN1/7UrfBqBwvYPQ9TFvymaRGZED3FCV0=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_plugin_sdk_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-plugin-sdk/v2",
        sum = "h1:X7vB6vn5tON2b49ILa4W7mFAsndeqJ7bZFOGbVO+0Cc=",
        version = "v2.30.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_plugin_testing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-plugin-testing",
        sum = "h1:Wsnfh+7XSVRfwcr2jZYHsnLOnZl7UeaOBvsx6dl/608=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_registry_address",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-registry-address",
        sum = "h1:2TAiKJ1A3MAkZlH1YI/aTVcLZRu7JseiXNRHbOAyoTI=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_hashicorp_terraform_svchost",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/terraform-svchost",
        sum = "h1:EZZimZ1GxdqFRinZ1tpJwVxxt49xc/S52uzrw4x0jKQ=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_hashicorp_vault_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/vault/api",
        sum = "h1:/US7sIjWN6Imp4o/Rj1Ce2Nr5bki/AXi9vAW3p2tOJQ=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_hashicorp_yamux",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/yamux",
        sum = "h1:yrQxtgseBDrq9Y652vSRDvsKCJKOUD+GzTS4Y0Y8pvE=",
        version = "v0.1.1",
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
        name = "com_github_howeyc_gopass",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/howeyc/gopass",
        sum = "h1:A9HsByNhogrvm9cWb28sjiS3i7tcKCkflWFEkHfuAgM=",
        version = "v0.0.0-20210920133722-c8aef6fb66ef",
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
        name = "com_github_hugelgupf_vmtest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hugelgupf/vmtest",
        sum = "h1:aa9+0fjwoGotyC8A3QjdITMAX89g/+qvDAhKPrK1NKE=",
        version = "v0.0.0-20240110072021-f6f07acb7aa1",
    )
    go_repository(
        name = "com_github_ianlancetaylor_demangle",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ianlancetaylor/demangle",
        sum = "h1:BA4a7pe6ZTd9F8kXETBoijjFJ/ntaa//1wiH9BZu4zU=",
        version = "v0.0.0-20230524184225-eabc099b10ab",
    )
    go_repository(
        name = "com_github_imdario_mergo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/imdario/mergo",
        sum = "h1:wwQJbIsHYGMUyLSPrEq1CT16AhnhNJQ51+4fdHUnCl4=",
        version = "v0.3.16",
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
        name = "com_github_insomniacslk_dhcp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/insomniacslk/dhcp",
        sum = "h1:9K06NfxkBh25x56yVhWWlKFE8YpicaSfHwoV8SFbueA=",
        version = "v0.0.0-20231206064809-8c70d406f6d2",
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
        sum = "h1:i2fYnDurfLlJH8AyyMOnkLHnHeP8Ff/DDpuZA/D3bPo=",
        version = "v0.0.0-20230406120618-7ff4192f6ff2",
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
        sum = "h1:TMtDYDHKYY15rFihtRfck/bfFqNfvcabqvXAFQfAUpY=",
        version = "v0.0.0-20230811132847-661be99b8267",
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
        sum = "h1:RCgYJqo3jgvhl+fEWvjNW8thxGWsgxi+TPhRir1Y9y8=",
        version = "v3.1.1",
    )
    go_repository(
        name = "com_github_jessevdk_go_flags",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jessevdk/go-flags",
        sum = "h1:4IU2WS7AumrZ/40jfhf4QVDMsQwqA7VEHozFRrGARJA=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_jhump_protoreflect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jhump/protoreflect",
        sum = "h1:6SFRuqU45u9hIZPJAoZ8c28T3nK64BNdp9w6jFonzls=",
        version = "v1.15.3",
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
        sum = "h1:eq4kys+NI0PLngzaHEe7AmPT90XMGIEySD1JfV1PDIs=",
        version = "v1.2.0",
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
        name = "com_github_jonboulle_clockwork",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jonboulle/clockwork",
        sum = "h1:p4Cf1aMWXnXAUh8lVfewRBx1zaTSYKrKMF2g3ST4RZ4=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_josephspurrier_goversioninfo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/josephspurrier/goversioninfo",
        sum = "h1:Puhl12NSHUSALHSuzYwPYQkqa2E1+7SrtAPJorKK0C8=",
        version = "v1.4.0",
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
        sum = "h1:uuaP0hAbW7Y4l0ZRQ6C9zfb7Mg1mbFKry/xzDAfmtLA=",
        version = "v1.1.0",
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
        sum = "h1:Z1BF0fRgcETPEa0Kt0MRk3yV5+kF1FWTni6KUFKrq2I=",
        version = "v1.4.0",
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
        name = "com_github_klauspost_compress",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/klauspost/compress",
        sum = "h1:60eq2E/jlfwQXtvZEeBUYADs+BwKBWURIY+Gj2eRGjI=",
        version = "v1.17.6",
    )
    go_repository(
        name = "com_github_klauspost_cpuid_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/klauspost/cpuid/v2",
        sum = "h1:g0I61F2K2DjRHz1cnxlkNSBIaePVoJIjjnHui8QHbiw=",
        version = "v2.0.4",
    )
    go_repository(
        name = "com_github_klauspost_pgzip",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/klauspost/pgzip",
        sum = "h1:8RXeL5crjEUFnR2/Sn6GJNWtSQ3Dk8pq4CL3jvdDyjU=",
        version = "v1.2.6",
    )
    go_repository(
        name = "com_github_konsorten_go_windows_terminal_sequences",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:mweAR1A6xJ3oS2pRaGiHgQ4OO8tzTaLawm8vnODuwDk=",
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
        sum = "h1:WT9HwE9SGECu3lg4d/dIA+jxlljEa1/ffXKmRjqdmIQ=",
        version = "v1.4.0",
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
        name = "com_github_letsencrypt_borp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/letsencrypt/borp",
        sum = "h1:xS2U6PQYRURk61YN4Y5xvyLbQVyAP/8fpE6hJZdwEWs=",
        version = "v0.0.0-20230707160741-6cc6ce580243",
    )
    go_repository(
        name = "com_github_letsencrypt_boulder",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/letsencrypt/boulder",
        sum = "h1:Y0fwz/hllcpgv9X24KyS/x8O6MdsOx217vAp1XV4Is0=",
        version = "v0.0.0-20240216200101-4eb5e3caa228",
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
        name = "com_github_letsencrypt_validator_v10",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/letsencrypt/validator/v10",
        sum = "h1:HGFsIltYMUiB5eoFSowFzSoXkocM2k9ctmJ57QMGjys=",
        version = "v10.0.0-20230215210743-a0c7dfc17158",
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
        sum = "h1:xfD0iDuEKnDkl03q4limB+vH+GxLEtL/jb4xVJSWWEY=",
        version = "v0.0.20",
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
        sum = "h1:UNAjwbU9l54TA3KzvqLGxwWjHmMgBUVhBiTjelZgg3U=",
        version = "v0.0.15",
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
        sum = "h1:fhGleo2h1p8tVChob4I9HpmVFIAkKGpiukdrgQbWfGI=",
        version = "v1.14.19",
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
        name = "com_github_matttproud_golang_protobuf_extensions_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/matttproud/golang_protobuf_extensions/v2",
        sum = "h1:jWpvCLoY8Z/e3VKvlsiIGKtc+UG6U5vzxaoagmhXfyg=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_mdlayher_ethtool",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/ethtool",
        sum = "h1:XAWHsmKhyPOo42qq/yTPb0eFBGUKKTR1rE0dVrWVQ0Y=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_mdlayher_genetlink",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/genetlink",
        sum = "h1:KdrNKe+CTu+IbZnm/GVUMXSqBBLqcGpRDa0xkQy56gw=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_mdlayher_netlink",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/netlink",
        sum = "h1:/UtM3ofJap7Vl4QWCPDGXY8d3GIY2UGSDbK+QWmY8/g=",
        version = "v1.7.2",
    )
    go_repository(
        name = "com_github_mdlayher_packet",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/packet",
        sum = "h1:3Up1NG6LZrsgDVn6X4L9Ge/iyRyxFEFD9o6Pr3Q1nQY=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_mdlayher_socket",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mdlayher/socket",
        sum = "h1:ilICZmJcQz70vrWVes1MFera4jGiWNocSkykwwoy3XI=",
        version = "v0.5.0",
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
        sum = "h1:68vKo2VN8DE9AdN4tnkWnmdhqdbpUFM8OF3Airm7fz8=",
        version = "v0.11.4",
    )
    go_repository(
        name = "com_github_miekg_dns",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/miekg/dns",
        sum = "h1:ca2Hdkz+cDg/7eNF6V56jjzuZ4aCAE+DbVkILdQWG/4=",
        version = "v1.1.58",
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
        name = "com_github_mitchellh_go_testing_interface",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/go-testing-interface",
        sum = "h1:jrgshOhYAUVNMAJiKbEu7EqAwgJJ2JqpQmpLJOu07cU=",
        version = "v1.14.1",
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
        sum = "h1:/tTvQaSJRr2FshkhXiIpux6fQ2Zvc4j7tAhMTStAG2g=",
        version = "v0.7.1",
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
        name = "com_github_moby_sys_user",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/moby/sys/user",
        sum = "h1:WmZ93f5Ux6het5iituh9x2zAG7NFY9Aqi49jjE1PaQg=",
        version = "v0.1.0",
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
        sum = "h1:F+S7ZlNKnrwHfSwdlgNSkKo67ReVf8o9fel6C3dkm/Q=",
        version = "v0.5.1",
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
        name = "com_github_netflix_go_expect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Netflix/go-expect",
        sum = "h1:+vx7roKuyA63nhn5WAunQHLTznkw5W8b1Xc0dNjp83s=",
        version = "v0.0.0-20220104043353-73e0943537d2",
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
        name = "com_github_nxadm_tail",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/nxadm/tail",
        sum = "h1:8feyoE3OzPrcshW5/MJ4sGESc5cqmGkGCWlco4l0bqY=",
        version = "v1.4.11",
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
        name = "com_github_oklog_run",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/oklog/run",
        sum = "h1:GEenZ1cK0+q0+wsJew9qUg/DyD8k3JzYsZAi5gYi2mA=",
        version = "v1.1.0",
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
        name = "com_github_olareg_olareg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/olareg/olareg",
        sum = "h1:iN0dytB++Fnk2axqdCAfUyqrpHsOv2FfVH5MD3nXscA=",
        version = "v0.0.0-20240206155231-8ba4b6726143",
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
        name = "com_github_onsi_ginkgo_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/ginkgo/v2",
        sum = "h1:vSmGj2Z5YPb9JwCWT6z6ihcUvDhuXLc3sJiqd3jMKAY=",
        version = "v2.14.0",
    )
    go_repository(
        name = "com_github_onsi_gomega",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/gomega",
        sum = "h1:hvMK7xYz4D3HapigLTeGdId/NcfQx1VHMJc60ew99+8=",
        version = "v1.30.0",
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
        sum = "h1:8SG7/vwALn54lVB/0yZ/MMwhFrPYtpEHQb2IpWsCzug=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_opencontainers_runc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/runc",
        sum = "h1:EaL5WeO9lv9wmS6SASjszOeQdSctvpbu0DdBQBizE40=",
        version = "v1.1.10",
    )
    go_repository(
        name = "com_github_opencontainers_runtime_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/opencontainers/runtime-spec",
        sum = "h1:HHUyrt9mwHUjtasSbXSMvs4cyFxh+Bll4AjJ9odEGpg=",
        version = "v1.1.0",
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
        sum = "h1:FnwAJ4oYMvbT/34k9zzHuZNrhlz48GB3/s6at6/MHO4=",
        version = "v2.1.0",
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
        name = "com_github_philhofer_fwd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/philhofer/fwd",
        sum = "h1:GdGcTjf5RNAxwS4QLsiMzJYj5KEvPJD3Abr261yRQXQ=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_pierrec_lz4_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pierrec/lz4/v4",
        sum = "h1:+fL8AQEZtz/ijeNnpduH0bROTu0O3NZAlPjQxGn8LwE=",
        version = "v4.1.14",
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
        sum = "h1:+mdjkGKdHQG3305AYmdv1U2eRNDiU2ErMBj1gwrq8eQ=",
        version = "v0.0.0-20240102092130-5ac0b6a4141c",
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
        sum = "h1:JFZT4XbOU7l77xGSpOdW+pwIMqP044IyjXX6FGyEKFo=",
        version = "v1.13.6",
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
        sum = "h1:HzFfmkOzH5Q8L8G+kSJKUx5dtG87sewO+FoDDqP5Tbk=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_github_prometheus_client_model",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:k1v3CzpSRUTrKMppY35TLwPvxHqBu0bYgxZzqGIgaos=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_prometheus_common",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/common",
        sum = "h1:p5Cz0FNHo7SnWOmWmoRozVcjEp0bIVU8cV7OShpjL1k=",
        version = "v0.47.0",
    )
    go_repository(
        name = "com_github_prometheus_procfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:jluTpSng7V9hY0O2R9DzzJHYb2xULk9VTR1V1R/k6Bo=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_github_prometheus_prometheus",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/prometheus",
        sum = "h1:jWcnuQHz1o1Wu3MZ6nMJDuTI0kU5yJp9pkxh8XEkNvI=",
        version = "v0.47.2",
    )
    go_repository(
        name = "com_github_protonmail_go_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ProtonMail/go-crypto",
        sum = "h1:P5Wd8eQ6zAzT4HpJI67FDKnTSf3xiJGQFqY1agDJPy4=",
        version = "v1.1.0-alpha.0-proton",
    )
    go_repository(
        name = "com_github_protonmail_go_mime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ProtonMail/go-mime",
        sum = "h1:tCbYj7/299ekTTXpdwKYF8eBlsYsDVoggDAuAjoK66k=",
        version = "v0.0.0-20230322103455-7d82a3887f2f",
    )
    go_repository(
        name = "com_github_protonmail_gopenpgp_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ProtonMail/gopenpgp/v2",
        sum = "h1:Vz/8+HViFFnf2A6XX8JOvZMrA6F5puwNvvF21O1mRlo=",
        version = "v2.7.4",
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
        sum = "h1:Yzoz33UZw9I/mFhx4MNrB6Fk+XHO1VukNcCa1+lwyKk=",
        version = "v9.4.0",
    )
    go_repository(
        name = "com_github_regclient_regclient",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/regclient/regclient",
        sum = "h1:d6bXhvz7UYJM+r20ls60RIVdoYh/rp+PygD/dIsJ9UA=",
        version = "v0.5.7",
    )
    go_repository(
        name = "com_github_rivo_uniseg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rivo/uniseg",
        sum = "h1:WUdvkW8uEhrYfLC4ZzdpI2ztxP1I582+49Oc5Mq64VQ=",
        version = "v0.4.7",
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
        sum = "h1:exVL4IDcn6na9z1rAb56Vxr+CgyK3nn3O+epU5NdKM8=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_github_rs_cors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rs/cors",
        sum = "h1:L0uuZVXIKlI1SShY2nhFfo44TYvDPQ1w4oFkUJNfhyo=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_github_rs_xid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rs/xid",
        sum = "h1:mKX4bl4iPYJtEIxp6CYiUuLQ/8DYMoz0PUdtGgMFRVc=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_rs_zerolog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rs/zerolog",
        sum = "h1:FcTR3NnLWW+NnTwwhFWiJSZr4ECLpqCm6QsEnyvbV4A=",
        version = "v1.31.0",
    )
    go_repository(
        name = "com_github_rubenv_sql_migrate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/rubenv/sql-migrate",
        sum = "h1:bo6/sjsan9HaXAsNxYP/jCEDUGibHp8JmOBw7NTGRos=",
        version = "v1.6.1",
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
        name = "com_github_sagikazarmark_locafero",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sagikazarmark/locafero",
        sum = "h1:HApY1R9zGo4DBgr7dqsTH/JJxLTTsOt7u6keLGt6kNQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_sagikazarmark_slog_shim",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sagikazarmark/slog-shim",
        sum = "h1:diDBnUNK9N/354PgrxMywXnAwEr1QZcOr6gto+ugjYE=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_samber_lo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/samber/lo",
        sum = "h1:j2XEAqXKb09Am4ebOg31SpvzUTTs6EN3VfgeLUhPdXM=",
        version = "v1.38.1",
    )
    go_repository(
        name = "com_github_samber_slog_multi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/samber/slog-multi",
        sum = "h1:6BVH9uHGAsiGkbbtQgAOQJMpKgV8unMrHhhJaw+X1EQ=",
        version = "v1.0.2",
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
        sum = "h1:O5s8ewCgq5QYNpv45dK4u6IpBmDM9RIcsbf/G1uXepQ=",
        version = "v7.6.1",
    )
    go_repository(
        name = "com_github_schollz_progressbar_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/schollz/progressbar/v3",
        sum = "h1:VD+MJPCr4s3wdhTc7OEJ/Z3dAeBzJ7yKH/P4lC5yRTI=",
        version = "v3.14.1",
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
        sum = "h1:aA4bp+/Zzi0BnWZ2F1wgNBs5gTpm+na2rWM6M9YjLpY=",
        version = "v0.10.0",
    )
    go_repository(
        name = "com_github_secure_systems_lab_go_securesystemslib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/secure-systems-lab/go-securesystemslib",
        sum = "h1:mr5An6X45Kb2nddcFlbmfHkLguCE9laoZCUzEEpIZXA=",
        version = "v0.8.0",
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
        sum = "h1:PP84WSDDyCCbjYKePcc0IaMSPXDndz8V3cQ9hMRSvpA=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_siderolabs_gen",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/gen",
        sum = "h1:lM69UYggT7yzpubf7hEFaNujPdY55Y9zvQf/NC18GvA=",
        version = "v0.4.7",
    )
    go_repository(
        name = "com_github_siderolabs_go_api_signature",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/go-api-signature",
        sum = "h1:ePXOxBT2fxRICsDntXed9kivmVK269nZe5UXvOxgtnM=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_siderolabs_go_blockdevice",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/go-blockdevice",
        sum = "h1:2bk4WpEEflGxjrNwp57ye24Pr+cYgAiAeNMWiQOuWbQ=",
        version = "v0.4.7",
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
        sum = "h1:BqxEmeWQeMpNP3R6WrPqDatX8sM/r4t97OP8mFmg6GA=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_siderolabs_talos_pkg_machinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/siderolabs/talos/pkg/machinery",
        sum = "h1:xzkHpHqVnio3IL2z44f/dG3TNVvSafZFvuyqlR6J7nY=",
        version = "v1.6.4",
    )
    go_repository(
        name = "com_github_sigstore_protobuf_specs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/protobuf-specs",
        sum = "h1:KIoM7E3C4uaK092q8YoSj/XSf9720f8dlsbYwwOmgEA=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_sigstore_rekor",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/rekor",
        sum = "h1:QoVXcS7NppKY+rpbEFVHr4evGDZBBSh65X0g8PXoUkQ=",
        version = "v1.3.5",
    )
    go_repository(
        name = "com_github_sigstore_sigstore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore",
        sum = "h1:mAVposMb14oplk2h/bayPmIVdzbq2IhCgy4g6R0ZSjo=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_aws",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/aws",
        sum = "h1:rEDdUefulkIQaMJyzLwtgPDLNXBIltBABiFYfb0YmgQ=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_azure",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/azure",
        sum = "h1:DvRWG99QGWZC5mp42SEde2Xke/Q384Idnj2da7yB+Mk=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_gcp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/gcp",
        sum = "h1:lwdRsJv1UbBemuk7w5YfXAQilQxMoFevrzamdPbG0wY=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore_pkg_signature_kms_hashivault",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sigstore/sigstore/pkg/signature/kms/hashivault",
        sum = "h1:9Ki0qudKpc1FQdef7xHO2bkLyTuw+qNUpWRzjBEmF4c=",
        version = "v1.8.1",
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
        sum = "h1:SHWdIUa82uGZz+F+47k8SY4QhhI291cXCpopT1lK2AQ=",
        version = "v1.2.1",
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
        name = "com_github_sourcegraph_conc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sourcegraph/conc",
        sum = "h1:OQTbbt6P72L20UqAkXXuLOj79LfEanQ+YQFNpLA9ySo=",
        version = "v0.3.0",
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
        sum = "h1:WJQKhtpdm3v2IzqG8VMqrr6Rf3UYpEF239Jy9wNepM8=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_github_spf13_cast",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/cast",
        sum = "h1:GEiTHELF+vaR5dhz3VqZfFSzZjYbgeKDpBxQVS4GYJ0=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_spf13_cobra",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/cobra",
        sum = "h1:7aJaZx1B85qltLMc546zn58BxxfZdR/W22ej9CFoEf0=",
        version = "v1.8.0",
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
        sum = "h1:LUXCnvUvSM6FXAsj6nnfc8Q2tp1dIgUfY9Kc8GsSOiQ=",
        version = "v1.18.2",
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
        sum = "h1:9NlTDc1FTs4qu0DDq7AEtTPNw6SVm7uBMsUCUjABIf8=",
        version = "v1.6.0",
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
        name = "com_github_theupdateframework_go_tuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/theupdateframework/go-tuf",
        sum = "h1:CqbQFrWo1ae3/I0UCblSbczevCCbS31Qvs5LdxRWqRI=",
        version = "v0.7.0",
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
        name = "com_github_tinylib_msgp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tinylib/msgp",
        sum = "h1:2gXmtWueD2HefZHQe1QOy9HVzmFrLOVvsXwXBQ0ayy0=",
        version = "v1.1.5",
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
        name = "com_github_ttacon_chalk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ttacon/chalk",
        sum = "h1:OXcKh35JaYsGMRzpvFkLv/MEyPuL49CThT1pZ8aSml4=",
        version = "v0.0.0-20160626202418-22c06c80ed31",
    )
    go_repository(
        name = "com_github_u_root_gobusybox_src",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/u-root/gobusybox/src",
        sum = "h1:AQX6C886dZqnOrXtbP0U59melqbb1+YnCfRYRfr4M3M=",
        version = "v0.0.0-20231224233253-2944a440b6b6",
    )
    go_repository(
        name = "com_github_u_root_u_root",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/u-root/u-root",
        sum = "h1:1AIJqOtdEufYfGb3eRpdaqWONzBOpAwrg1fehbWg+Mg=",
        version = "v0.11.1-0.20230807200058-f87ad7ccb594",
    )
    go_repository(
        name = "com_github_u_root_uio",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/u-root/uio",
        sum = "h1:YcojQL98T/OO+rybuzn2+5KrD5dBwXIvYBvQ2cD3Avg=",
        version = "v0.0.0-20230305220412-3e8cd9d6bf63",
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
        sum = "h1:ebbhrRiGK2i4naQJr+1Xj92HXZCrK7MsyTS/ob3HnAk=",
        version = "v1.22.14",
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
        sum = "h1:Ok0Hr3GMAf8K/1NB4sV65QGgCiukG1w1QD+H5tmt0Ow=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_vincent_petithory_dataurl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vincent-petithory/dataurl",
        sum = "h1:cXw+kPto8NLuJtlMsI152irrVw9fRDX8AbShPRpg2CI=",
        version = "v1.0.0",
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
        sum = "h1:Oeaw1EM2JMxD51g9uhtC0D7erkIjgmj8+JZc26m1YX8=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_vmihailenco_msgpack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmihailenco/msgpack",
        sum = "h1:dSLoQfGFAo3F6OoNhwUmLwVgaUXK79GlxNBwueZn0xI=",
        version = "v4.0.4+incompatible",
    )
    go_repository(
        name = "com_github_vmihailenco_msgpack_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/vmihailenco/msgpack/v5",
        sum = "h1:cQriyiUvjTwOHg8QZaPihLWeRAAVoCpE00IUPn0Bjt8=",
        version = "v5.4.1",
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
        sum = "h1:O3tjSwQBy0XwI5uK1/yVIfQ1LP9bAECEDUfifnyGs9U=",
        version = "v0.30.6",
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
        sum = "h1:h2JizvZl9aIj6za9S5AyrkU+OzIS4CetQthH/ejO+lg=",
        version = "v0.30.2-0.20230730094716-a20f9abcc222",
    )
    go_repository(
        name = "com_github_workiva_go_datastructures",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Workiva/go-datastructures",
        sum = "h1:J6Y/52yX10Xc5JjXmGtWoSSxs3mZnGSaq37xZZh7Yig=",
        version = "v1.0.53",
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
        sum = "h1:FHX5I5B4i4hKRVRBCFRxq1iQRej7WO3hhBuJf+UUySY=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_xdg_go_stringprep",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xdg-go/stringprep",
        sum = "h1:XLI/Ng3O1Atzq0oBs3TWm+5ZVgkq2aqdlvP9JtoZ6c8=",
        version = "v1.0.4",
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
        sum = "h1:S2dVYn90KE98chqDkyE9Z4N61UnQd+KOfgp5Iu53llk=",
        version = "v0.0.0-20221125231312-a49e3df8f510",
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
        name = "com_github_yuin_gopher_lua",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/yuin/gopher-lua",
        sum = "h1:kYKnWBjvbNP4XLT3+bPEwAXJx262OhaHDWDVOPjL46M=",
        version = "v1.1.1",
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
        sum = "h1:kTG7lqmBou0Zkx35r6HJHUQTvaRPr5bIAf3AoHS0izI=",
        version = "v1.14.2",
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
        name = "com_github_zmap_zcrypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zmap/zcrypto",
        sum = "h1:U1b4THKcgOpJ+kILupuznNwPiURtwVW3e9alJvji9+s=",
        version = "v0.0.0-20231219022726-a1f61fb1661c",
    )
    go_repository(
        name = "com_github_zmap_zlint_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/zmap/zlint/v3",
        sum = "h1:vTEaDRtYN0d/1Ax60T+ypvbLQUHwHxbvYRnUMVr35ug=",
        version = "v3.6.0",
    )
    go_repository(
        name = "com_google_cloud_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go",
        sum = "h1:tpFCD7hpHFlQ8yPwT3x+QeXqc2T6+n6T+hmABHfDUSM=",
        version = "v0.112.0",
    )
    go_repository(
        name = "com_google_cloud_go_accessapproval",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/accessapproval",
        sum = "h1:uzmAMSgYcnlHa9X9YSQZ4Q1wlfl4NNkZyQgho1Z6p04=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_accesscontextmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/accesscontextmanager",
        sum = "h1:2GLNaNu9KRJhJBFTIVRoPwk6xE5mUDgD47abBq4Zp/I=",
        version = "v1.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_aiplatform",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/aiplatform",
        sum = "h1:0cSrii1ZeLr16MbBoocyy5KVnrSdiQ3KN/vtrTe7RqE=",
        version = "v1.60.0",
    )
    go_repository(
        name = "com_google_cloud_go_analytics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/analytics",
        sum = "h1:Q+y94XH84jM8SK8O7qiY/PJRexb6n7dRbQ6PiUa4YGM=",
        version = "v0.23.0",
    )
    go_repository(
        name = "com_google_cloud_go_apigateway",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigateway",
        sum = "h1:sPXnpk+6TneKIrjCjcpX5YGsAKy3PTdpIchoj8/74OE=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeconnect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigeeconnect",
        sum = "h1:CrfIKv9Go3fh/QfQgisU3MeP90Ww7l/sVGmr3TpECo8=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/apigeeregistry",
        sum = "h1:C+QU2K+DzDjk4g074ouwHQGkoff1h5OMQp6sblCVreQ=",
        version = "v0.8.3",
    )
    go_repository(
        name = "com_google_cloud_go_appengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/appengine",
        sum = "h1:l2SviT44zWQiOv8bPoMBzW0vOcMO22iO0s+nVtVhdts=",
        version = "v1.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_area120",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/area120",
        sum = "h1:vTs08KPLN/iMzTbxpu5ciL06KcsrVPMjz4IwcQyZ4uY=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_artifactregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/artifactregistry",
        sum = "h1:W9sVlyb1VRcUf83w7aM3yMsnp4HS4PoyGqYQNG0O5lI=",
        version = "v1.14.7",
    )
    go_repository(
        name = "com_google_cloud_go_asset",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/asset",
        sum = "h1:xgFnBP3luSbUcC9RWJvb3Zkt+y/wW6PKwPHr3ssnIP8=",
        version = "v1.17.2",
    )
    go_repository(
        name = "com_google_cloud_go_assuredworkloads",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/assuredworkloads",
        sum = "h1:gCrN3IyvqY3cP0wh2h43d99CgH3G+WYs9CeuFVKChR8=",
        version = "v1.11.5",
    )
    go_repository(
        name = "com_google_cloud_go_automl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/automl",
        sum = "h1:ijiJy9sYWh75WrqImXsfWc1e3HR3iO+ef9fvW03Ig/4=",
        version = "v1.13.5",
    )
    go_repository(
        name = "com_google_cloud_go_baremetalsolution",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/baremetalsolution",
        sum = "h1:LFydisRmS7hQk9P/YhekwuZGqb45TW4QavcrMToWo5A=",
        version = "v1.2.4",
    )
    go_repository(
        name = "com_google_cloud_go_batch",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/batch",
        sum = "h1:2HK4JerwVaIcCh/lJiHwh6+uswPthiMMWhiSWLELayk=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_beyondcorp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/beyondcorp",
        sum = "h1:qs0J0O9Ol2h1yA0AU+r7l3hOCPzs2MjE1d6d/kaHIKo=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_google_cloud_go_bigquery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/bigquery",
        sum = "h1:CpT+/njKuKT3CEmswm6IbhNu9u35zt5dO4yPDLW+nG4=",
        version = "v1.59.1",
    )
    go_repository(
        name = "com_google_cloud_go_billing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/billing",
        sum = "h1:oWUEQvuC4JvtnqLZ35zgzdbuHt4Itbftvzbe6aEyFdE=",
        version = "v1.18.2",
    )
    go_repository(
        name = "com_google_cloud_go_binaryauthorization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/binaryauthorization",
        sum = "h1:1jcyh2uIUwSZkJ/JmL8kd5SUkL/Krbv8zmYLEbAz6kY=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_certificatemanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/certificatemanager",
        sum = "h1:UMBr/twXvH3jcT5J5/YjRxf2tvwTYIfrpemTebe0txc=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_channel",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/channel",
        sum = "h1:/omiBnyFjm4S1ETHoOmJbL7LH7Ljcei4rYG6Sj3hc80=",
        version = "v1.17.5",
    )
    go_repository(
        name = "com_google_cloud_go_cloudbuild",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/cloudbuild",
        sum = "h1:ZB6oOmJo+MTov9n629fiCrO9YZPOg25FZvQ7gIHu5ng=",
        version = "v1.15.1",
    )
    go_repository(
        name = "com_google_cloud_go_clouddms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/clouddms",
        sum = "h1:Sr0Zo5EAcPQiCBgHWICg3VGkcdS/LLP1d9SR7qQBM/s=",
        version = "v1.7.4",
    )
    go_repository(
        name = "com_google_cloud_go_cloudtasks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/cloudtasks",
        sum = "h1:EUt1hIZ9bLv8Iz9yWaCrqgMnIU+Tdh0yXM1MMVGhjfE=",
        version = "v1.12.6",
    )
    go_repository(
        name = "com_google_cloud_go_compute",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/compute",
        sum = "h1:phWcR2eWzRJaL/kOiJwfFsPs4BaKq1j6vnpZrc1YlVg=",
        version = "v1.24.0",
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
        sum = "h1:6Vs/YnDG5STGjlWMEjN/xtmft7MrOTOnOZYUZtGTx0w=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_container",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/container",
        sum = "h1:MAaNH7VRNPWEhvqOypq2j+7ONJKrKzon4v9nS3nLZe0=",
        version = "v1.31.0",
    )
    go_repository(
        name = "com_google_cloud_go_containeranalysis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/containeranalysis",
        sum = "h1:doJ0M1ljS4hS0D2UbHywlHGwB7sQLNrt9vFk9Zyi7vY=",
        version = "v0.11.4",
    )
    go_repository(
        name = "com_google_cloud_go_datacatalog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datacatalog",
        sum = "h1:A0vKYCQdxQuV4Pi0LL9p39Vwvg4jH5yYveMv50gU5Tw=",
        version = "v1.19.3",
    )
    go_repository(
        name = "com_google_cloud_go_dataflow",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataflow",
        sum = "h1:RYHtcPhmE664+F0Je46p+NvFbG8z//KCXp+uEqB4jZU=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_dataform",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataform",
        sum = "h1:5e4eqGrd0iDTCg4Q+VlAao5j2naKAA7xRurNtwmUknU=",
        version = "v0.9.2",
    )
    go_repository(
        name = "com_google_cloud_go_datafusion",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datafusion",
        sum = "h1:HQ/BUOP8OIGJxuztpYvNvlb+/U+/Bfs9SO8tQbh61fk=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_datalabeling",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datalabeling",
        sum = "h1:GpIFRdm0qIZNsxqURFJwHt0ZBJZ0nF/mUVEigR7PH/8=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_dataplex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataplex",
        sum = "h1:fxIfdU8fxzR3clhOoNI7XFppvAmndxDu1AMH+qX9WKQ=",
        version = "v1.14.2",
    )
    go_repository(
        name = "com_google_cloud_go_dataproc_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataproc/v2",
        sum = "h1:/u81Fd+BvCLp+xjctI1DiWVJn6cn9/s3Akc8xPH02yk=",
        version = "v2.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataqna",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dataqna",
        sum = "h1:9ybXs3nr9BzxSGC04SsvtuXaHY0qmJSLIpIAbZo9GqQ=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_datastore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datastore",
        sum = "h1:0P9WcsQeTWjuD1H14JIY7XQscIPQ4Laje8ti96IC5vg=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_google_cloud_go_datastream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/datastream",
        sum = "h1:o1QDKMo/hk0FN7vhoUQURREuA0rgKmnYapB+1M+7Qz4=",
        version = "v1.10.4",
    )
    go_repository(
        name = "com_google_cloud_go_deploy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/deploy",
        sum = "h1:m27Ojwj03gvpJqCbodLYiVmE9x4/LrHGGMjzc0LBfM4=",
        version = "v1.17.1",
    )
    go_repository(
        name = "com_google_cloud_go_dialogflow",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dialogflow",
        sum = "h1:KqG0oxGE71qo0lRVyAoeBozefCvsMfcDzDjoLYSY0F4=",
        version = "v1.49.0",
    )
    go_repository(
        name = "com_google_cloud_go_dlp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/dlp",
        sum = "h1:lTipOuJaSjlYnnotPMbEhKURLC6GzCMDDzVbJAEbmYM=",
        version = "v1.11.2",
    )
    go_repository(
        name = "com_google_cloud_go_documentai",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/documentai",
        sum = "h1:lI62GMEEPO6vXJI9hj+G9WjOvnR0hEjvjokrnex4cxA=",
        version = "v1.25.0",
    )
    go_repository(
        name = "com_google_cloud_go_domains",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/domains",
        sum = "h1:Mml/R6s3vQQvFPpi/9oX3O5dRirgjyJ8cksK8N19Y7g=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_edgecontainer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/edgecontainer",
        sum = "h1:tBY32km78ScpK2aOP84JoW/+wtpx5WluyPUSEE3270U=",
        version = "v1.1.5",
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
        sum = "h1:13eHn5qBnsawxI7mIrv4jRIEmQ1xg0Ztqw5ZGqtUNfA=",
        version = "v1.6.6",
    )
    go_repository(
        name = "com_google_cloud_go_eventarc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/eventarc",
        sum = "h1:ORkd6/UV5FIdA8KZQDLNZYKS7BBOrj0p01DXPmT4tE4=",
        version = "v1.13.4",
    )
    go_repository(
        name = "com_google_cloud_go_filestore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/filestore",
        sum = "h1:X5G4y/vrUo1B8Nsz93qSWTMAcM8LXbGUldq33OdcdCw=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_firestore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/firestore",
        sum = "h1:8aLcKnMPoldYU3YHgu4t2exrKhLQkqaXAGqT0ljrFVw=",
        version = "v1.14.0",
    )
    go_repository(
        name = "com_google_cloud_go_functions",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/functions",
        sum = "h1:IWVylmK5F6hJ3R5zaRW7jI5PrWhCvtBVU4axQLmXSo4=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkebackup",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkebackup",
        sum = "h1:iuE8KNtTsPOc79qeWoNS8zOWoXPD9SAdOmwgxtlCmh8=",
        version = "v1.3.5",
    )
    go_repository(
        name = "com_google_cloud_go_gkeconnect",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkeconnect",
        sum = "h1:17d+ZSSXKqG/RwZCq3oFMIWLPI8Zw3b8+a9/BEVlwH0=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_gkehub",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkehub",
        sum = "h1:RboLNFzf9wEMSo7DrKVBlf+YhK/A/jrLN454L5Tz99Q=",
        version = "v0.14.5",
    )
    go_repository(
        name = "com_google_cloud_go_gkemulticloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gkemulticloud",
        sum = "h1:rsSZAGLhyjyE/bE2ToT5fqo1qSW7S+Ubsc9jFOcbhSI=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_google_cloud_go_gsuiteaddons",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/gsuiteaddons",
        sum = "h1:CZEbaBwmbYdhFw21Fwbo+C35HMe36fTE0FBSR4KSfWg=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_iam",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iam",
        sum = "h1:bEa06k05IO4f4uJonbB5iAgKTPpABy1ayxaIZV/GHVc=",
        version = "v1.1.6",
    )
    go_repository(
        name = "com_google_cloud_go_iap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iap",
        sum = "h1:94zirc2r4t6KzhAMW0R6Dme005eTP6yf7g6vN4IhRrA=",
        version = "v1.9.4",
    )
    go_repository(
        name = "com_google_cloud_go_ids",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/ids",
        sum = "h1:xd4U7pgl3GHV+MABnv1BF4/Vy/zBF7CYC8XngkOLzag=",
        version = "v1.4.5",
    )
    go_repository(
        name = "com_google_cloud_go_iot",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/iot",
        sum = "h1:munTeBlbqI33iuTYgXy7S8lW2TCgi5l1hA4roSIY+EE=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/kms",
        sum = "h1:7caV9K3yIxvlQPAcaFffhlT7d1qpxjB1wHBtjWa13SM=",
        version = "v1.15.7",
    )
    go_repository(
        name = "com_google_cloud_go_language",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/language",
        sum = "h1:iaJZg6K4j/2PvZZVcjeO/btcWWIllVRBhuTFjGO4LXs=",
        version = "v1.12.3",
    )
    go_repository(
        name = "com_google_cloud_go_lifesciences",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/lifesciences",
        sum = "h1:gXvN70m2p+4zgJFzaz6gMKaxTuF9WJ0USYoMLWAOm8g=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_logging",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/logging",
        sum = "h1:iEIOXFO9EmSiTjDmfpbRjOxECO7R8C7b8IXUGOj7xZw=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_longrunning",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/longrunning",
        sum = "h1:GOE6pZFdSrTb4KAiKnXsJBtlE6mEyaW44oKyMILWnOg=",
        version = "v0.5.5",
    )
    go_repository(
        name = "com_google_cloud_go_managedidentities",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/managedidentities",
        sum = "h1:+bpih1piZVLxla/XBqeSUzJBp8gv9plGHIMAI7DLpDM=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_maps",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/maps",
        sum = "h1:EVCZAiDvog9So46460BGbCasPhi613exoaQbpilMVlk=",
        version = "v1.6.4",
    )
    go_repository(
        name = "com_google_cloud_go_mediatranslation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/mediatranslation",
        sum = "h1:c76KdIXljQHSCb/Cy47S8H4s05A4zbK3pAFGzwcczZo=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_memcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/memcache",
        sum = "h1:yeDv5qxRedFosvpMSEswrqUsJM5OdWvssPHFliNFTc4=",
        version = "v1.10.5",
    )
    go_repository(
        name = "com_google_cloud_go_metastore",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/metastore",
        sum = "h1:dR7vqWXlK6IYR8Wbu9mdFfwlVjodIBhd1JRrpZftTEg=",
        version = "v1.13.4",
    )
    go_repository(
        name = "com_google_cloud_go_monitoring",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/monitoring",
        sum = "h1:NfkDLQDG2UR3WYZVQE8kwSbUIEyIqJUPl+aOQdFH1T4=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkconnectivity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networkconnectivity",
        sum = "h1:GBfXFhLyPspnaBE3nI/BRjdhW8vcbpT9QjE/4kDCDdc=",
        version = "v1.14.4",
    )
    go_repository(
        name = "com_google_cloud_go_networkmanagement",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networkmanagement",
        sum = "h1:aLV5GcosBNmd6M8+a0ekB0XlLRexv4fvnJJrYnqeBcg=",
        version = "v1.9.4",
    )
    go_repository(
        name = "com_google_cloud_go_networksecurity",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/networksecurity",
        sum = "h1:+caSxBTj0E8OYVh/5wElFdjEMO1S/rZtE1152Cepchc=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_notebooks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/notebooks",
        sum = "h1:FH48boYmrWVQ6k0Mx/WrnNafXncT5iSYxA8CNyWTgy0=",
        version = "v1.11.3",
    )
    go_repository(
        name = "com_google_cloud_go_optimization",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/optimization",
        sum = "h1:63NZaWyN+5rZEKHPX4ACpw3BjgyeuY8+rCehiCMaGPY=",
        version = "v1.6.3",
    )
    go_repository(
        name = "com_google_cloud_go_orchestration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/orchestration",
        sum = "h1:YHgWMlrPttIVGItgGfuvO2KM7x+y9ivN/Yk92pMm1a4=",
        version = "v1.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_orgpolicy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/orgpolicy",
        sum = "h1:2JbXigqBJVp8Dx5dONUttFqewu4fP0p3pgOdIZAhpYU=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_google_cloud_go_osconfig",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/osconfig",
        sum = "h1:Mo5jGAxOMKH/PmDY7fgY19yFcVbvwREb5D5zMPQjFfo=",
        version = "v1.12.5",
    )
    go_repository(
        name = "com_google_cloud_go_oslogin",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/oslogin",
        sum = "h1:1K4nOT5VEZNt7XkhaTXupBYos5HjzvJMfhvyD2wWdFs=",
        version = "v1.13.1",
    )
    go_repository(
        name = "com_google_cloud_go_phishingprotection",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/phishingprotection",
        sum = "h1:DH3WFLzEoJdW/6xgsmoDqOwT1xddFi7gKu0QGZQhpGU=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_policytroubleshooter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/policytroubleshooter",
        sum = "h1:c0WOzC6hz964QWNBkyKfna8A2jOIx1zzZa43Gx/P09o=",
        version = "v1.10.3",
    )
    go_repository(
        name = "com_google_cloud_go_privatecatalog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/privatecatalog",
        sum = "h1:UZ0assTnATXSggoxUIh61RjTQ4P9zCMk/kEMbn0nMYA=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_profiler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/profiler",
        sum = "h1:ZeRDZbsOBDyRG0OiK0Op1/XWZ3xeLwJc9zjkzczUxyY=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_pubsub",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/pubsub",
        sum = "h1:dfEPuGCHGbWUhaMCTHUFjfroILEkx55iUmKBZTP5f+Y=",
        version = "v1.36.1",
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
        sum = "h1:U3Wfq12X9cVMuTpsWDSURnXF0Z9hSPTHj+xsnXDRLsw=",
        version = "v2.9.2",
    )
    go_repository(
        name = "com_google_cloud_go_recommendationengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recommendationengine",
        sum = "h1:ineqLswaCSBY0csYv5/wuXJMBlxATK6Xc5jJkpiTEdM=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_recommender",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/recommender",
        sum = "h1:LVLYS3r3u0MSCxQSDUtLSkporEGi9OAE6hGvayrZNPs=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_google_cloud_go_redis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/redis",
        sum = "h1:QF0maEdVv0Fj/2roU8sX3NpiDBzP9ICYTO+5F32gQNo=",
        version = "v1.14.2",
    )
    go_repository(
        name = "com_google_cloud_go_resourcemanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/resourcemanager",
        sum = "h1:AZWr1vWVDKGwfLsVhcN+vcwOz3xqqYxtmMa0aABCMms=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_resourcesettings",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/resourcesettings",
        sum = "h1:BTr5MVykJwClASci/7Og4Qfx70aQ4n3epsNLj94ZYgw=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_retail",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/retail",
        sum = "h1:Fn1GuAua1c6crCGqfJ1qMxG1Xh10Tg/x5EUODEHMqkw=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_google_cloud_go_run",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/run",
        sum = "h1:m9WDA7DzTpczhZggwYlZcBWgCRb+kgSIisWn1sbw2rQ=",
        version = "v1.3.4",
    )
    go_repository(
        name = "com_google_cloud_go_scheduler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/scheduler",
        sum = "h1:5U8iXLoQ03qOB+ZXlAecU7fiE33+u3QiM9nh4cd0eTE=",
        version = "v1.10.6",
    )
    go_repository(
        name = "com_google_cloud_go_secretmanager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/secretmanager",
        sum = "h1:82fpF5vBBvu9XW4qj0FU2C6qVMtj1RM/XHwKXUEAfYY=",
        version = "v1.11.5",
    )
    go_repository(
        name = "com_google_cloud_go_security",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/security",
        sum = "h1:wTKJQ10j8EYgvE8Y+KhovxDRVDk2iv/OsxZ6GrLP3kE=",
        version = "v1.15.5",
    )
    go_repository(
        name = "com_google_cloud_go_securitycenter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/securitycenter",
        sum = "h1:/5jjkZ+uGe8hZ7pvd7pO30VW/a+pT2MrrdgOqjyucKQ=",
        version = "v1.24.4",
    )
    go_repository(
        name = "com_google_cloud_go_servicedirectory",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/servicedirectory",
        sum = "h1:da7HFI1229kyzIyuVEzHXip0cw0d+E0s8mjQby0WN+k=",
        version = "v1.11.4",
    )
    go_repository(
        name = "com_google_cloud_go_shell",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/shell",
        sum = "h1:3Fq2hzO0ZSyaqBboJrFkwwf/qMufDtqwwA6ep8EZxEI=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_spanner",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/spanner",
        sum = "h1:fJq+ZfQUDHE+cy1li0bJA8+sy2oiSGhuGqN5nqVaZdU=",
        version = "v1.57.0",
    )
    go_repository(
        name = "com_google_cloud_go_speech",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/speech",
        sum = "h1:nuFc+Kj5B8de75nN4FdPyUbI2SiBoHZG6BLurXL56Q0=",
        version = "v1.21.1",
    )
    go_repository(
        name = "com_google_cloud_go_storage",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/storage",
        sum = "h1:Az68ZRGlnNTpIBbLjSMIV2BDcwwXYlRlQzis0llkpJg=",
        version = "v1.38.0",
    )
    go_repository(
        name = "com_google_cloud_go_storagetransfer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/storagetransfer",
        sum = "h1:dy4fL3wO0VABvzM05ycMUPFHxTPbJz9Em8ikAJVqSbI=",
        version = "v1.10.4",
    )
    go_repository(
        name = "com_google_cloud_go_talent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/talent",
        sum = "h1:JssV0CE3FNujuSWn7SkosOzg7qrMxVnt6txOfGcMSa4=",
        version = "v1.6.6",
    )
    go_repository(
        name = "com_google_cloud_go_texttospeech",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/texttospeech",
        sum = "h1:dxY2Q5mHCbrGa3oPR2O3PCicdnvKa1JmwGQK36EFLOw=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_tpu",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/tpu",
        sum = "h1:C8YyYda8WtNdBoCgFwwBzZd+S6+EScHOxM/z1h0NNp8=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_trace",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/trace",
        sum = "h1:0pr4lIKJ5XZFYD9GtxXEWr0KkVeigc3wlGpZco0X1oA=",
        version = "v1.10.5",
    )
    go_repository(
        name = "com_google_cloud_go_translate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/translate",
        sum = "h1:upovZ0wRMdzZvXnu+RPam41B0mRJ+coRXFP2cYFJ7ew=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_video",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/video",
        sum = "h1:TXwotxkShP1OqgKsbd+b8N5hrIHavSyLGvYnLGCZ7xc=",
        version = "v1.20.4",
    )
    go_repository(
        name = "com_google_cloud_go_videointelligence",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/videointelligence",
        sum = "h1:mYaWH8uhUCXLJCN3gdXswKzRa2+lK0zN6/KsIubm6pE=",
        version = "v1.11.5",
    )
    go_repository(
        name = "com_google_cloud_go_vision_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vision/v2",
        sum = "h1:W52z1b6LdGI66MVhE70g/NFty9zCYYcjdKuycqmlhtg=",
        version = "v2.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_vmmigration",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vmmigration",
        sum = "h1:5v9RT2vWyuw3pK2ox0HQpkoftO7Q7/8591dTxxQc79g=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_vmwareengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vmwareengine",
        sum = "h1:EGdDi9QbqThfZq3ILcDK5g+m9jTevc34AY5tACx5v7k=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_google_cloud_go_vpcaccess",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/vpcaccess",
        sum = "h1:XyL6hTLtEM/eE4F1GEge8xUN9ZCkiVWn44K/YA7z1rQ=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_webrisk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/webrisk",
        sum = "h1:251MvGuC8wisNN7+jqu9DDDZAi38KiMXxOpA/EWy4dE=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_websecurityscanner",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/websecurityscanner",
        sum = "h1:YqWZrZYabG88TZt7364XWRJGhxmxhony2ZUyZEYMF2k=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_workflows",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go/workflows",
        sum = "h1:uHNmUiatTbPQ4H1pabwfzpfEYD4BBnqDHqMm2IesOh4=",
        version = "v1.12.4",
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
        sum = "h1:q5zoXux4xkOZP473e1EZbG8Gq9f0vlg1VNH5Du/ybus=",
        version = "v0.36.0",
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
        name = "in_gopkg_evanphx_json_patch_v5",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/evanphx/json-patch.v5",
        sum = "h1:hx1VU2SGj4F8r9b8GUwJLdc8DNO8sy79ZGui0G05GLo=",
        version = "v5.9.0",
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
        name = "in_gopkg_go_jose_go_jose_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/go-jose/go-jose.v2",
        sum = "h1:nt80fvSDlhKWQgSWyHyy5CfmlQr+asih51R8PTWNKKs=",
        version = "v2.6.3",
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
        name = "io_cncf_tags_container_device_interface",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "tags.cncf.io/container-device-interface",
        sum = "h1:dThE6dtp/93ZDGhqaED2Pu374SOeUkBfuvkLuiTdwzg=",
        version = "v0.6.2",
    )
    go_repository(
        name = "io_cncf_tags_container_device_interface_specs_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "tags.cncf.io/container-device-interface/specs-go",
        sum = "h1:V+tJJN6dqu8Vym6p+Ru+K5mJ49WL6Aoc5SJFSY0RLsQ=",
        version = "v0.6.0",
    )
    go_repository(
        name = "io_etcd_go_bbolt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/bbolt",
        sum = "h1:xs88BrvEv273UsB79e0hcVrlUWmS0a8upikMFhSyAtA=",
        version = "v1.3.8",
    )
    go_repository(
        name = "io_etcd_go_etcd_api_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/api/v3",
        sum = "h1:W4sw5ZoU2Juc9gBWuLk5U6fHfNVyY1WC5g9uiXZio/c=",
        version = "v3.5.12",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_pkg_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/pkg/v3",
        sum = "h1:EYDL6pWwyOsylrQyLp2w+HkQ46ATiOvoEdMarindU2A=",
        version = "v3.5.12",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/v2",
        sum = "h1:MrmRktzv/XF8CvtQt+P6wLUlURaNpSDJHFZhe//2QE4=",
        version = "v2.305.10",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/v3",
        sum = "h1:v5lCPXn1pf1Uu3M4laUE2hp/geOTc5uPcYYsNe1lDxg=",
        version = "v3.5.12",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdctl_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/etcdctl/v3",
        sum = "h1:Wiv+g9i12mXjNgwz/3S8p01U3IPueGUbTcgBCpQ/Fw4=",
        version = "v3.5.10",
    )
    go_repository(
        name = "io_etcd_go_etcd_etcdutl_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/etcdutl/v3",
        sum = "h1:o57fNgdP9Y99wZzpQ5ky5Jb6323/nisMtCOj1+kQwgc=",
        version = "v3.5.10",
    )
    go_repository(
        name = "io_etcd_go_etcd_pkg_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/pkg/v3",
        sum = "h1:WPR8K0e9kWl1gAhB5A7gEa5ZBTNkT9NdNWrR8Qpo1CM=",
        version = "v3.5.10",
    )
    go_repository(
        name = "io_etcd_go_etcd_raft_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/raft/v3",
        sum = "h1:cgNAYe7xrsrn/5kXMSaH8kM/Ky8mAdMqGOxyYwpP0LA=",
        version = "v3.5.10",
    )
    go_repository(
        name = "io_etcd_go_etcd_server_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/server/v3",
        sum = "h1:4NOGyOwD5sUZ22PiWYKmfxqoeh72z6EhYjNosKGLmZg=",
        version = "v3.5.10",
    )
    go_repository(
        name = "io_etcd_go_etcd_tests_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/tests/v3",
        sum = "h1:F1pbXwKxwZ58aBT2+CSL/r8WUCAVhob0y1y8OVJ204s=",
        version = "v3.5.10",
    )
    go_repository(
        name = "io_etcd_go_etcd_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/v3",
        sum = "h1:M147e9UGqf9kZK19ptQOTt3Nue0inAOoUCbk9S+VMZk=",
        version = "v3.5.10",
    )
    go_repository(
        name = "io_filippo_edwards25519",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "filippo.io/edwards25519",
        sum = "h1:FNf4tywRC1HmFuKW5xopWpigGjJKiJSV0Cqo0cJWDaA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "io_k8s_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/api",
        sum = "h1:NiCdQMY1QOp1H8lfRyeEf8eOwV6+0xA6XEE44ohDX2A=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_apiextensions_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiextensions-apiserver",
        sum = "h1:0VuspFG7Hj+SxyF/Z/2T0uFbI5gb5LRgEyUVE3Q4lV0=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_apimachinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apimachinery",
        sum = "h1:+ACVktwyicPz0oc6MTMLwa2Pw3ouLAfAon1wPLtG48o=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiserver",
        sum = "h1:Y1xEMjJkP+BIi0GSEv1BBrf1jLU9UPfAnnGGbbDdp7o=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_cli_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cli-runtime",
        sum = "h1:q2kC3cex4rOBLfPOnMSzV2BIrrQlx97gxHJs21KxKS4=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_client_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/client-go",
        sum = "h1:KmlDtFcrdUzOYrBhXHgKw5ycWzc3ryPX5mQe0SkG3y8=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_cloud_provider",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cloud-provider",
        replace = "k8s.io/cloud-provider",
        sum = "h1:Qgk/jHsSKGRk/ltTlN6e7eaNuuamLROOzVBd0RPp94M=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_cluster_bootstrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cluster-bootstrap",
        sum = "h1:zCYdZ+LWDj4O86FB5tDKckIEsf2qBHjcp78xtjOzD3A=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_code_generator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/code-generator",
        sum = "h1:2LQfayGDhaIlaamXjIjEQlCMy4JNCH9lrzas4DNW1GQ=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_component_base",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/component-base",
        sum = "h1:T7rjd5wvLnPBV1vC4zWd/iWRbV8Mdxs+nGaoaFzGw3s=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_component_helpers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/component-helpers",
        sum = "h1:Y8W70NGeitKxWwhsPo/vEQbQx5VqJV+3xfLpP3V1VxU=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_controller_manager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/controller-manager",
        replace = "k8s.io/controller-manager",
        sum = "h1:kEv9sKLnjDkoSqeouWp2lZ8P33an5wrDJpOMqoyD7pc=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_cri_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cri-api",
        sum = "h1:atenAqOltRsFqcCQlFFpDnl/R4aGfOELoNLTDJfd7t8=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_csi_translation_lib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/csi-translation-lib",
        replace = "k8s.io/csi-translation-lib",
        sum = "h1:we4X1yUlDikvm5Rv0dwMuPHNw6KwjwsQiAuOPWXha8M=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_dynamic_resource_allocation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/dynamic-resource-allocation",
        replace = "k8s.io/dynamic-resource-allocation",
        sum = "h1:JQW5erdoOsvhst7DxMfEpnXhrfm9SmNTnvyaXdqTLAE=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_endpointslice",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/endpointslice",
        replace = "k8s.io/endpointslice",
        sum = "h1:HM+zsyqSALW7FzOVCWYsF+eFabiTGDrZpLEZZX2065U=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_gengo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/gengo",
        sum = "h1:pWEwq4Asjm4vjW7vcsmijwBhOr1/shsbSYiWXmNGlks=",
        version = "v0.0.0-20230829151522-9cce18d56c01",
    )
    go_repository(
        name = "io_k8s_klog_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/klog/v2",
        sum = "h1:QXU6cPEOIslTGvZaXvFWiP9VKyeet3sawzTOvdXb4Vw=",
        version = "v2.120.1",
    )
    go_repository(
        name = "io_k8s_kms",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kms",
        sum = "h1:KJ1zaZt74CgvgV3NR7tnURJ/mJOKC5X3nwon/WdwgxI=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_kube_aggregator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-aggregator",
        replace = "k8s.io/kube-aggregator",
        sum = "h1:N4fmtePxOZ+bwiK1RhVEztOU+gkoVkvterHgpwAuiTw=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_kube_controller_manager",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-controller-manager",
        replace = "k8s.io/kube-controller-manager",
        sum = "h1:25nmyTOdjOLM1QLe4nbu5jvlLSv1ZIPFDvmUUWvbuSw=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_kube_openapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-openapi",
        sum = "h1:QSpdNrZ9uRlV0VkqLvVO0Rqg8ioKi3oSw7O5P7pJV8M=",
        version = "v0.0.0-20240220201932-37d671a357a5",
    )
    go_repository(
        name = "io_k8s_kube_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-proxy",
        replace = "k8s.io/kube-proxy",
        sum = "h1:nZJdLzHTIJ2okftUMsBvEidtH57GAOMMPFKBcA0V+Bg=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_kube_scheduler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-scheduler",
        replace = "k8s.io/kube-scheduler",
        sum = "h1:n4v68EvxYhy7o5Q/LFPgqBEGi7lKoiAxwQ0gQyMoj9M=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_kubectl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubectl",
        sum = "h1:Oqi48gXjikDhrBF67AYuZRTcJV4lg2l42GmvsP7FmYI=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_kubelet",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubelet",
        sum = "h1:SX5hlznTBcGIrS1scaf8r8p6m3e475KMifwt9i12iOk=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_kubernetes",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubernetes",
        sum = "h1:n4VCbX9cUhxHI+zw+m2iZlzT73/mrEJBHIMeauh9g4U=",
        version = "v1.29.4",
    )
    go_repository(
        name = "io_k8s_legacy_cloud_providers",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/legacy-cloud-providers",
        replace = "k8s.io/legacy-cloud-providers",
        sum = "h1:fjGV9OhqseUTp3R8xOm2TBoAxyuRTOS6B2zFTSJ80RE=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/metrics",
        sum = "h1:a6dWcNM+EEowMzMZ8trka6wZtSRIfEA/9oLjuhBksGc=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_mount_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/mount-utils",
        sum = "h1:KcUE0bFHONQC10V3SuLWQ6+l8nmJggw9lKLpDftIshI=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_pod_security_admission",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/pod-security-admission",
        replace = "k8s.io/pod-security-admission",
        sum = "h1:tY/ldtkbBCulMYVSWg6ZDLlgDYDWy6rLj8e/AgmwSj4=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_sample_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/sample-apiserver",
        replace = "k8s.io/sample-apiserver",
        sum = "h1:bUEz09ehjQE/xpgMVkutbBfZhcLvg1BvCMLvJnbLZbc=",
        version = "v0.29.0",
    )
    go_repository(
        name = "io_k8s_sigs_apiserver_network_proxy_konnectivity_client",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/apiserver-network-proxy/konnectivity-client",
        sum = "h1:TgtAeesdhpm2SGwkQasmbeqDo8th5wOBA5h/AjTKA4I=",
        version = "v0.28.0",
    )
    go_repository(
        name = "io_k8s_sigs_controller_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/controller-runtime",
        sum = "h1:FwHwD1CTUemg0pW2otk7/U5/i5m2ymzvOXdbeGOUvw0=",
        version = "v0.17.2",
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
        sum = "h1:/zAR4FOQDCkgSDmVzV2uiFbuy9bhu3jEzthrHCuvm1g=",
        version = "v0.16.0",
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
        sum = "h1:6J33uKSoATlKZH16unr2XOhDI+otoe2sR3M8PDzW3K0=",
        version = "v0.16.0",
    )
    go_repository(
        name = "io_k8s_sigs_release_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/release-utils",
        sum = "h1:JKDOvhCk6zW8ipEOkpTGDH/mW3TI+XqtPp16aaQ79FU=",
        version = "v0.7.7",
    )
    go_repository(
        name = "io_k8s_sigs_structured_merge_diff_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/structured-merge-diff/v4",
        sum = "h1:150L+0vs/8DA78h1u02ooW1/fFq/Lwr+sGiqlzvrtq4=",
        version = "v4.4.1",
    )
    go_repository(
        name = "io_k8s_sigs_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:Mk1wCc2gy/F0THH0TAp1QYyJNzRm2KCLy3o5ASXVI5E=",
        version = "v1.4.0",
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
        sum = "h1:eQ/4ljkx21sObifjzXwlPKpdGLrCfRziVtos3ofG/sQ=",
        version = "v0.0.0-20240102154912-e7106e64919e",
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
        sum = "h1:zBakwHardp9Jcb8sQHcHpXy/0+JIb1M8KjigCJzx7+4=",
        version = "v0.13.14",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_github_com_emicklei_go_restful_otelrestful",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful",
        sum = "h1:Z6SbqeRZAl2OczfkFOqLx1BeYBDYehNjEnqluD7581Y=",
        version = "v0.42.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_google_golang_org_grpc_otelgrpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
        sum = "h1:P+/g8GpuJGYbOp2tAdKrIPUX9JO02q8Q0YNlHolpibA=",
        version = "v0.48.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_net_http_otelhttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
        sum = "h1:doUP+ExOpH3spVTLS0FcWGLnQrPct/hD/bCPbDRUEAU=",
        version = "v0.48.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel",
        sum = "h1:Za4UzOqJYS+MUczKI320AtqZHZb7EqxO00jAHE0jmQY=",
        version = "v1.23.1",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace",
        sum = "h1:cl5P5/GIfFh4t6xyruOgJP5QiA1pw4fYYdv6nc6CBWw=",
        version = "v1.21.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace_otlptracegrpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc",
        sum = "h1:tIqheXEFWAZ7O8A7m+J0aPTmpJN3YQ7qetUAdkkkKpk=",
        version = "v1.21.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace_otlptracehttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp",
        sum = "h1:IeMeyr1aBvBiPVYihXIaeIZba6b8E1bYp7lbdxK8CQg=",
        version = "v1.19.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_metric",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/metric",
        sum = "h1:PQJmqJ9u2QaJLBOELl1cxIdPcpbwzbkjfEyelTl2rlo=",
        version = "v1.23.1",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/sdk",
        sum = "h1:FTt8qirL1EysG6sTQRZ5TokkU8d0ugCj8htOgThZXQ8=",
        version = "v1.21.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_trace",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/otel/trace",
        sum = "h1:4LrmmEd8AU2rFvU1zegmvqW7+kWarxtNOPyeL6HmYY8=",
        version = "v1.23.1",
    )
    go_repository(
        name = "io_opentelemetry_go_proto_otlp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.opentelemetry.io/proto/otlp",
        sum = "h1:T0TX0tmXU8a3CbNXzEKGeU5mIVOdf0oykP+u2lIVU/I=",
        version = "v1.0.0",
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
        sum = "h1:XpYuAwAb0DfQsunIyMfeET92emK8km3W4yEzZvUbsTo=",
        version = "v1.2.5",
    )
    go_repository(
        name = "net_starlark_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.starlark.net",
        sum = "h1:LmbG8Pq7KDGkglKVn8VpZOZj6vb9b8nKEGcg9l03epM=",
        version = "v0.0.0-20240123142251-f86470692795",
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
        name = "org_golang_google_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/api",
        sum = "h1:zd5d4JIIIaYYsfVy1HzoXYZ9rWCSBxxAglbczzo7Bgc=",
        version = "v0.165.0",
    )
    go_repository(
        name = "org_golang_google_appengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/appengine",
        sum = "h1:IhEN5q69dyKagZPYMSdIjS2HqprW324FRQZJcGqPAsM=",
        version = "v1.6.8",
    )
    go_repository(
        name = "org_golang_google_genproto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto",
        sum = "h1:Zmyn5CV/jxzKnF+3d+xzbomACPwLQqVpLTpyXN5uTaQ=",
        version = "v0.0.0-20240221002015-b0ce06bbee7c",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto/googleapis/api",
        sum = "h1:9g7erC9qu44ks7UK4gDNlnk4kOxZG707xKm4jVniy6o=",
        version = "v0.0.0-20240221002015-b0ce06bbee7c",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_bytestream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto/googleapis/bytestream",
        sum = "h1:d9MrRjDhOUCudL8eHNVgZA1ESnBiE7ZKwlZS9foLoRU=",
        version = "v0.0.0-20240205150955-31a09d347014",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_rpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto/googleapis/rpc",
        sum = "h1:NUsgEN92SQQqzfA+YtqYNqYmB3DMMYLlIwUZAQFVFbo=",
        version = "v0.0.0-20240221002015-b0ce06bbee7c",
    )
    go_repository(
        name = "org_golang_google_grpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc",
        sum = "h1:kLAiWrZs7YeDM6MumDe7m3y4aM6wacLzM1Y/wiLP9XY=",
        version = "v1.61.1",
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
        sum = "h1:uNO2rsAINq/JlFpSdYEKIZ0uKD/R9cpdv0T+yoGwGmI=",
        version = "v1.33.0",
    )
    go_repository(
        name = "org_golang_x_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/crypto",
        sum = "h1:g1v0xeRhjcugydODzvb3mEM9SQ0HGp9s/nh3COQ/C30=",
        version = "v0.22.0",
    )
    go_repository(
        name = "org_golang_x_exp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/exp",
        sum = "h1:HinSgX1tJRX3KsL//Gxynpw5CTOAIPhgL4W8PNiIpVE=",
        version = "v0.0.0-20240213143201-ec583247a57a",
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
        sum = "h1:SernR4v+D55NyBH2QiEQrlBAnj1ECL6AGrA5+dPaMY8=",
        version = "v0.15.0",
    )
    go_repository(
        name = "org_golang_x_net",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/net",
        sum = "h1:7EYJ93RZ9vYSZAIb2x3lnuvqO5zneoD6IvWjuhfxjTs=",
        version = "v0.23.0",
    )
    go_repository(
        name = "org_golang_x_oauth2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/oauth2",
        sum = "h1:6m3ZPmLEFdVxKKWnKq4VqZ60gutO35zm+zrAHVmHyDQ=",
        version = "v0.17.0",
    )
    go_repository(
        name = "org_golang_x_sync",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sync",
        sum = "h1:5BMeUDZ7vkXGfEr1x9B4bRcTH4lpkTkpdh0T/J+qjbQ=",
        version = "v0.6.0",
    )
    go_repository(
        name = "org_golang_x_sys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sys",
        sum = "h1:q5f1RH2jigJ1MoAWp2KTp3gm5zAGFUTarQZ5U386+4o=",
        version = "v0.19.0",
    )
    go_repository(
        name = "org_golang_x_telemetry",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/telemetry",
        sum = "h1:+Kc94D8UVEVxJnLXp/+FMfqQARZtWHfVrcRtcG8aT3g=",
        version = "v0.0.0-20240208230135-b75ee8823808",
    )
    go_repository(
        name = "org_golang_x_term",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/term",
        sum = "h1:+ThwsDv+tYfnJFhF4L8jITxu1tdTWRTZpdsWgEgjL6Q=",
        version = "v0.19.0",
    )
    go_repository(
        name = "org_golang_x_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/text",
        sum = "h1:ScX5w1eTa3QqT8oi6+ziP7dTV1S2+ALU0bI+0zXKWiQ=",
        version = "v0.14.0",
    )
    go_repository(
        name = "org_golang_x_time",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/time",
        sum = "h1:o7cqy6amK/52YcAKIPlM3a+Fpj35zvRj2TP+e1xFSfk=",
        version = "v0.5.0",
    )
    go_repository(
        name = "org_golang_x_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/tools",
        sum = "h1:k8NLag8AGHnn+PHbl7g43CtqZAwG60vZkLqgyZgIHgQ=",
        version = "v0.18.0",
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
        sum = "h1:+cNy6SZtPcJQH3LJVLOSmiC7MMxXNOb3PU/VUEz+EhU=",
        version = "v0.0.0-20231012003039-104605ab7028",
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
        sum = "h1:fPVWdvZz8TSmMrTnsStih9ETsHlrzIgSEEiFzOLbhO8=",
        version = "v1.10000.0",
    )
    go_repository(
        name = "org_mongodb_go_mongo_driver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.mongodb.org/mongo-driver",
        sum = "h1:P98w8egYRjYe3XDjxhYJagTokP/H6HzlsnojRgZRd80=",
        version = "v1.14.0",
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
        sum = "h1:aJMhYGrd5QSmlpLMr2MftRKl7t8J8PTZPA732ud/XR8=",
        version = "v1.27.0",
    )
    go_repository(
        name = "sh_elv_src",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "src.elv.sh",
        sum = "h1:pjVeIo9Ba6K1Wy+rlwX91zT7A+xGEmxiNRBdN04gDTQ=",
        version = "v0.16.0-rc1.0.20220116211855-fda62502ad7f",
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
        sum = "h1:V71fv+NGZv0icBlr+in1MJXuUIHCiPG1hW9gEBISTIA=",
        version = "v3.14.2",
    )
    go_repository(
        name = "sm_step_go_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.step.sm/crypto",
        sum = "h1:OmwHm3GJO8S4VGWL3k4+I+Q4P/F2s+j8msvTyGnh1Vg=",
        version = "v0.42.1",
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
        sum = "h1:Ci3iUJyx9UeRx7CeFN8ARgGbkESwJK+KB9lLcWxY/Zw=",
        version = "v2.4.0",
    )
