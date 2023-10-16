filegroup(
    name = "repo",
    srcs = glob(["*.rpm", "repodata/*"]),
    visibility = ["//visibility:public"],
)

exports_files(glob(
    ["*.rpm"],
    ["repodata/*"],
))
