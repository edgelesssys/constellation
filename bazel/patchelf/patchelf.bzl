""" Bazel rule for postprocessing elf files with patchelf """

def _patchelf_impl(ctx):
    output = ctx.outputs.out
    ctx.actions.run_shell(
        inputs = [ctx.file.src, ctx.file.rpath],
        tools = [ctx.executable._patchelf_binary],
        outputs = [output],
        arguments = [
            ctx.executable._patchelf_binary.path,
            ctx.file.rpath.path,
            output.path,
            ctx.file.src.path,
        ],
        command = "\"$1\" --set-rpath \"$(cat \"$2\")\" --output \"$3\" \"$4\"",
        progress_message = "Patching ELF binary " + ctx.file.src.basename,
    )
    return DefaultInfo(
        files = depset([output]),
        executable = output,
    )

patchelf = rule(
    implementation = _patchelf_impl,
    attrs = {
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
    executable = True,
)
