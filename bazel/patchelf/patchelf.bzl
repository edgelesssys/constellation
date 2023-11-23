""" Bazel rule for postprocessing elf files with patchelf """

def _patchelf_impl(ctx):
    output = ctx.outputs.out
    ctx.actions.run_shell(
        inputs = [ctx.file.src, ctx.file.rpath, ctx.file.interpreter],
        tools = [ctx.executable._patchelf_binary],
        outputs = [output],
        arguments = [
            ctx.executable._patchelf_binary.path,
            ctx.file.rpath.path,
            ctx.file.interpreter.path,
            output.path,
            ctx.file.src.path,
        ],
        command = "\"$1\" --set-rpath \"$(cat \"$2\")\" --set-interpreter \"$(cat \"$3\")\" --output \"$4\" \"$5\"",
        progress_message = "Patching ELF binary " + ctx.file.src.basename,
    )
    return DefaultInfo(
        files = depset([output]),
        executable = output,
    )

patchelf = rule(
    implementation = _patchelf_impl,
    attrs = {
        "interpreter": attr.label(mandatory = True, allow_single_file = True),
        "out": attr.output(mandatory = True),
        "rpath": attr.label(mandatory = True, allow_single_file = True),
        "src": attr.label(mandatory = True, allow_single_file = True),
        "_patchelf_binary": attr.label(
            default = Label("@patchelf//:bin/patchelf"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
    },
    doc = """Uses patchelf to set the rpath and interpreter of an ELF binary.

This is a post processing step for binaries that are built by Bazel
that link shared libraries from Nix.
Native nix binaries always use absolute paths to a specific linker and
embed the rpath for each linked library as an absolute nix store path.
Bazel embeds rpaths using $ORIGIN and uses paths relative to
the binaries output location. Without patchelf, the binary would only
work if we also ship all runfiles with the binary and preserve the relative
directory structure.
See https://github.com/tweag/rules_nixpkgs/issues/449 for more details.
""",
    executable = True,
)
