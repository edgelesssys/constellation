"""
Rules for generating Go source files from proto files.

based on https://github.com/bazelbuild/rules_go/issues/2111#issuecomment-1355927231
"""

load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")
load("@io_bazel_rules_go//go:def.bzl", "GoLibrary", "go_context")
load("@io_bazel_rules_go//proto:compiler.bzl", "GoProtoCompiler")

def _output_go_library_srcs_impl(ctx):
    go = go_context(ctx)

    srcs_of_library = []
    importpath = ""
    for src in ctx.attr.deps:
        lib = src[GoLibrary]
        go_src = go.library_to_source(go, ctx.attr, lib, False)
        if importpath and lib.importpath != importpath:
            fail(
                "importpath of all deps must match, got {} and {}",
                importpath,
                lib.importpath,
            )
        importpath = lib.importpath
        srcs_of_library.extend(go_src.srcs)

    if len(srcs_of_library) != 1:
        fail("expected exactly one src for library, got {}", len(srcs_of_library))

    if not ctx.attr.out:
        fail("must specify out for now")

    # Run a command to copy the src file to the out location.
    _copy(ctx, srcs_of_library[0], ctx.outputs.out)

def _copy(ctx, in_file, out_file):
    # based on https://github.com/bazelbuild/examples/blob/main/rules/shell_command/rules.bzl
    ctx.actions.run_shell(
        # Input files visible to the action.
        inputs = [in_file],
        # Output files that must be created by the action.
        outputs = [out_file],
        progress_message = "Copying {} to {}".format(in_file.path, out_file.path),
        arguments = [in_file.path, out_file.path],
        command = """cp "$1" "$2" """,
    )

output_go_library_srcs = rule(
    implementation = _output_go_library_srcs_impl,
    attrs = {
        "compiler": attr.label(
            providers = [GoProtoCompiler],
            default = "@io_bazel_rules_go//proto:go_proto",
        ),
        "deps": attr.label_list(
            providers = [GoLibrary],
            aspects = [],
        ),
        "out": attr.output(
            doc = ("Name of output .go file. If not specified, the file name " +
                   "of the generated source file will be used."),
            mandatory = False,
        ),
        "_go_context_data": attr.label(
            default = "@io_bazel_rules_go//:go_context_data",
        ),
    },
    toolchains = ["@io_bazel_rules_go//go:toolchain"],
)

def write_go_proto_srcs(name, go_proto_library, src, visibility = None):
    generated_src = "__generated_" + src
    output_go_library_srcs(
        name = name + "_generated",
        deps = [go_proto_library],
        out = generated_src,
        visibility = ["//visibility:private"],
    )

    write_source_files(
        name = name,
        files = {
            src: generated_src,
        },
        diff_test = False,
        visibility = visibility,
    )
