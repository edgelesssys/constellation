"""A rule to create an RPM mirror from a lockfile."""

def _impl(ctx):
    contents = ctx.read(ctx.path(ctx.attr.lockfile))
    lines = contents.split("\n")
    package_hashes = {}
    for line in lines:
        components = line.split("  ")
        if len(components) != 2:
            continue
        package_hashes[components[1]] = components[0]
    for package in sorted(package_hashes.keys()):
        ctx.download(
            "%s%s" % (ctx.attr._cas_base_url, package_hashes[package]),
            output = package,
            sha256 = package_hashes[package],
        )
    ctx.execute([
        ctx.path(ctx.attr._createrepo_c),
        "--revision",
        "0",
        "--set-timestamp-to-revision",
        ".",
    ])
    ctx.template(
        "BUILD.bazel",
        ctx.attr._build_tpl,
    )

rpm_repository = repository_rule(
    implementation = _impl,
    attrs = {
        "lockfile": attr.label(
            mandatory = True,
            allow_single_file = True,
            doc = "The lockfile (SHA256SUMS) to use.",
        ),
        "_build_tpl": attr.label(
            default = Label(":BUILD.bazel.tpl"),
            allow_single_file = True,
        ),
        "_cas_base_url": attr.string(
            default = "https://cdn.confidential.cloud/constellation/cas/sha256/",
            doc = "The base URL for the CAS.",
        ),
        "_createrepo_c": attr.label(
            default = Label("@createrepo_c//:bin/createrepo_c"),
            executable = True,
            cfg = "exec",
            allow_single_file = True,
            doc = "The createrepo_c tool.",
        ),
    },
)
