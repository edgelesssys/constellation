"""Proto targets"""

def proto_targets():
    return [
        "//joinservice/joinproto:write_generated_protos",
        "//bootstrapper/initproto:write_generated_protos",
        "//debugd/service:write_generated_protos",
        "//disk-mapper/recoverproto:write_generated_protos",
        "//keyservice/keyserviceproto:write_generated_protos",
        "//internal/versions/components:write_generated_protos",
        "//upgrade-agent/upgradeproto:write_generated_protos",
        "//verify/verifyproto:write_generated_protos",
    ]
