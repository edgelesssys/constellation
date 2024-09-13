""" Bazel rule for building mkosi images. """

load("@bazel_skylib//lib:paths.bzl", "paths")

def _resource_set(_os, _num_inputs):
    return {"cpu": 4, "memory": 4096}

def _mkosi_image_impl(ctx):
    args = ctx.actions.args()
    inputs = []
    outputs = []
    tools = []
    workdir = ctx.file.mkosi_conf.dirname
    env = {}
    args.add("-C", workdir)
    if ctx.attr.distribution:
        args.add("--distribution", ctx.attr.distribution)
    if ctx.attr.architecture:
        args.add("--architecture", ctx.attr.architecture)
    if ctx.attr.output:
        args.add("--output", ctx.attr.output)
    args.add_all(ctx.attr.packages, before_each = "--package")
    for package_dir in ctx.files.package_directories:
        args.add("--package-directory", package_dir.path)
    if len(ctx.files.local_mirror) > 0:
        env["LOCAL_MIRROR"] = ctx.files.local_mirror[0].dirname
    for tree in ctx.files.base_trees:
        args.add("--base-tree", tree.path)
    for tree in ctx.files.skeleton_trees:
        args.add("--skeleton-tree", tree.path)
    for tree in ctx.files.package_manager_trees:
        args.add("--package-manager-tree", tree.path)
    for tree in ctx.files.extra_trees:
        args.add("--extra-tree", tree.path)
    for initrd in ctx.files.initrds:
        inputs.append(initrd)
        args.add("--initrd", initrd.path)
    inputs.extend(ctx.files.mkosi_conf)
    inputs.extend(ctx.files.srcs[:])
    inputs.extend(ctx.files.package_directories[:])
    inputs.extend(ctx.files.base_trees[:])
    inputs.extend(ctx.files.skeleton_trees[:])
    inputs.extend(ctx.files.package_manager_trees[:])
    inputs.extend(ctx.files.extra_trees[:])
    inputs.extend(ctx.files.initrds[:])
    inputs.extend(ctx.files.local_mirror[:])
    if ctx.attr.source_date_epoch:
        args.add("--source-date-epoch", ctx.attr.source_date_epoch)
    if ctx.attr.seed:
        args.add("--seed", ctx.attr.seed)
    if ctx.file.version_file:
        env["VERSION_FILE"] = ctx.file.version_file.path
        inputs.append(ctx.file.version_file)
    outputs.extend(ctx.outputs.outs)
    if ctx.attr.out_dir:
        out_dir = ctx.actions.declare_directory(ctx.attr.out_dir)
        for output in outputs:
            if output.path.startswith(out_dir.path + "/"):
                fail("output {} is nested within output directory {}; outputs cannot be nested within each other!".format(output.path, out_dir.path))
            if output.is_directory and out_dir.path.startswith(output.path + "/"):
                fail("output directory {} is nested within output directory {}; outputs cannot be nested within each other!".format(out_dir.path, output.path))
        outputs.append(out_dir)
        args.add("--output-dir", out_dir.path)
    else:
        args.add("--output-dir", paths.join(ctx.bin_dir.path, ctx.label.package))
    args.add_all(ctx.attr.extra_args)
    for key, value in ctx.attr.env.items():
        args.add("--environment", "{}={}".format(key, value))
    if ctx.attr.kernel_command_line:
        args.add("--kernel-command-line", ctx.attr.kernel_command_line)
    for key, value in ctx.attr.kernel_command_line_dict.items():
        args.add("--kernel-command-line", "{}={}".format(key, value))

    info = ctx.toolchains["@constellation//bazel/mkosi:toolchain_type"].mkosi
    if not info.valid:
        fail("mkosi toolchain is not properly configured: {}".format(info.name))
    if info.label:
        executable_files = info.label[DefaultInfo].files_to_run
        tools.append(executable_files)
        mkosi_bin = executable_files.executable.path
    else:
        mkosi_bin = info.path

    wrapper = ctx.actions.declare_file("mkosi_wrapper.sh")
    ctx.actions.expand_template(
        template = ctx.file._mkosi_wrapper_template,
        output = wrapper,
        substitutions = {
            "@@MKOSI@@": mkosi_bin,
        },
        is_executable = True,
    )
    inputs.append(wrapper)
    ctx.actions.run(
        outputs = outputs,
        inputs = inputs,
        executable = wrapper,
        arguments = [args],
        tools = tools,
        execution_requirements = {"no-remote": "1", "no-sandbox": "1"},
        progress_message = "Building mkosi image " + ctx.label.name,
        env = env,
        resource_set = _resource_set,
    )
    return DefaultInfo(files = depset(outputs), runfiles = ctx.runfiles(outputs))

mkosi_image = rule(
    implementation = _mkosi_image_impl,
    attrs = {
        "architecture": attr.string(),
        "base_trees": attr.label_list(allow_files = True),
        "distribution": attr.string(),
        "env": attr.string_dict(),
        "extra_args": attr.string_list(),
        "extra_trees": attr.label_list(allow_files = True),
        "initrds": attr.label_list(allow_files = True),
        "kernel_command_line": attr.string(),
        "kernel_command_line_dict": attr.string_dict(),
        "local_mirror": attr.label_list(allow_files = True),
        "mkosi_conf": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "main mkosi.conf file",
        ),
        "out_dir": attr.string(),
        "output": attr.string(),
        "outs": attr.output_list(),
        "package_directories": attr.label_list(allow_files = True),
        "package_manager_trees": attr.label_list(allow_files = True),
        "packages": attr.string_list(),
        "seed": attr.string(),
        "skeleton_trees": attr.label_list(allow_files = True),
        "source_date_epoch": attr.string(),
        "srcs": attr.label_list(allow_files = True),
        "version_file": attr.label(allow_single_file = True),
        "_mkosi_wrapper_template": attr.label(
            default = Label("@constellation//bazel/mkosi:mkosi_wrapper.sh.in"),
            allow_single_file = True,
        ),
    },
    executable = False,
    toolchains = ["@constellation//bazel/mkosi:toolchain_type"],
)
