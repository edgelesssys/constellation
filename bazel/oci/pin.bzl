"""
This module contains rules and macros for pinning oci images.
"""

load("@aspect_bazel_lib//lib:jq.bzl", "jq")
load("@bazel_skylib//lib:types.bzl", "types")

def stamp_tags(name, repotags, **kwargs):
    if not types.is_list(repotags):
        fail("repotags should be a list")
    _maybe_quote = lambda x: x if "\"" in x else "\"{}\"".format(x)
    jq(
        name = name,
        srcs = [],
        out = "_{}.tags.txt".format(name),
        args = ["--raw-output"],
        filter = "|".join([
            "$ARGS.named.STAMP as $stamp",
            ",".join([_maybe_quote(t) for t in repotags]),
        ]),
        **kwargs
    )

def _oci_go_source_impl(ctx):
    oci = ctx.file.oci
    inputs = [oci]
    if ctx.attr.repotag_file:
        inputs.append(ctx.file.repotag_file)
    output = ctx.actions.declare_file(ctx.label.name + ".go")
    args = [
        "codegen",
        "--oci-path",
        oci.path,
        "--package",
        ctx.attr.package,
        "--identifier",
        ctx.attr.identifier,
        "--output",
        output.path,
    ]
    if ctx.attr.tag:
        args.append("--image-tag")
        args.append(ctx.attr.tag)
    if ctx.attr.repotag_file:
        args.append("--repoimage-tag-file")
        args.append(ctx.file.repotag_file.path)

    ctx.actions.run(
        inputs = inputs,
        arguments = args,
        outputs = [output],
        executable = ctx.executable._oci_pin,
        mnemonic = "OCIPin",
        progress_message = "Generating OCI pin Go source %{label}",
    )

    return [DefaultInfo(
        files = depset([output]),
    )]

_go_source_attrs = {
    "identifier": attr.string(
        mandatory = True,
        doc = "Identifier to use for the generated Go source.",
    ),
    "image_name": attr.string(
        mandatory = True,
        doc = "Image name to use for the generated Go source.",
    ),
    "oci": attr.label(
        mandatory = True,
        allow_single_file = True,
        doc = "OCI image to extract the digest from.",
    ),
    "package": attr.string(
        mandatory = True,
        doc = "Package to use for the generated Go source.",
    ),
    "repotag_file": attr.label(
        allow_single_file = True,
        doc = "OCI image tag file to use for the generated Go source.",
    ),
    "tag": attr.string(
        doc = "OCI image tag to use for the generated Go source.",
    ),
    "_oci_pin": attr.label(
        allow_single_file = True,
        executable = True,
        cfg = "exec",
        default = Label("//hack/oci-pin"),
    ),
}

oci_go_source = rule(
    implementation = _oci_go_source_impl,
    attrs = _go_source_attrs,
)

def _oci_sum_impl(ctx):
    oci = ctx.file.oci
    inputs = [oci]
    if ctx.attr.repotag_file:
        inputs.append(ctx.file.repotag_file)
    output = ctx.actions.declare_file(ctx.label.name + ".sha256")
    args = [
        "sum",
        "--oci-path",
        oci.path,
        "--output",
        output.path,
    ]
    if ctx.attr.repotag_file:
        args.append("--repoimage-tag-file")
        args.append(ctx.file.repotag_file.path)

    ctx.actions.run(
        inputs = inputs,
        arguments = args,
        outputs = [output],
        executable = ctx.executable._oci_pin,
        mnemonic = "OCISum",
        progress_message = "Generating OCI sum file %{label}",
    )

    return [DefaultInfo(
        files = depset([output]),
    )]

_sum_attrs = {
    "image_name": attr.string(
        mandatory = True,
        doc = "Image name to use for the sum entry.",
    ),
    "oci": attr.label(
        mandatory = True,
        allow_single_file = True,
        doc = "OCI image to extract the digest from.",
    ),
    "repotag_file": attr.label(
        allow_single_file = True,
        doc = "OCI image tag file to use for the sum entry.",
    ),
    "_oci_pin": attr.label(
        allow_single_file = True,
        executable = True,
        cfg = "exec",
        default = Label("//hack/oci-pin"),
    ),
}

oci_sum = rule(
    implementation = _oci_sum_impl,
    attrs = _sum_attrs,
)

def _oci_sum_merge_impl(ctx):
    # TODO: select list of labels
    inputs = ctx.files.sums
    output = ctx.actions.declare_file(ctx.label.name + ".sha256")
    args = [
        "merge",
        "--output",
        output.path,
    ]
    for sum in ctx.files.sums:
        args.append("--input")
        args.append(sum.path)

    ctx.actions.run(
        inputs = inputs,
        arguments = args,
        outputs = [output],
        executable = ctx.executable._oci_pin,
        mnemonic = "OCISumMerge",
        progress_message = "Merging OCI sum files %{label}",
    )

    return [DefaultInfo(
        files = depset([output]),
    )]

_sum_merge_attrs = {
    "sums": attr.label_list(
        doc = "Sum files to merge for the combined sum entry.",
    ),
    "_oci_pin": attr.label(
        allow_single_file = True,
        executable = True,
        cfg = "exec",
        default = Label("//hack/oci-pin"),
    ),
}

oci_sum_merge = rule(
    implementation = _oci_sum_merge_impl,
    attrs = _sum_merge_attrs,
)
