""" Bazel rule for uploading a set of OS images to cloud providers. """

def _upload_os_images_impl(ctx):
    executable = ctx.actions.declare_file("upload_os_images_%s.sh" % ctx.label.name)
    files = []
    files.extend(ctx.files.image_dirs)
    files.append(ctx.file._version)
    files.append(ctx.file._upload_cli)
    files.append(ctx.file._measured_boot)
    files.append(ctx.file._uplosi)
    files.append(ctx.file._dissect_toolchain)
    files.append(ctx.file._cosign)
    files.append(ctx.file._rekor_cli)
    files.append(ctx.file._parallel)
    raw_image_paths = []
    for image_dir in ctx.files.image_dirs:
        raw_image_paths.append("%s/constellation.raw" % image_dir.short_path)
    substitutions = {
        "@@COSIGN@@": ctx.executable._cosign.short_path,
        "@@DISSECT_TOOLCHAIN@@": ctx.executable._dissect_toolchain.short_path,
        "@@FILES@@": " ".join(raw_image_paths),
        "@@MEASURED_BOOT@@": ctx.executable._measured_boot.short_path,
        "@@PARALLEL@@": ctx.executable._parallel.short_path,
        "@@REKOR_CLI@@": ctx.executable._rekor_cli.short_path,
        "@@UPLOAD_CLI@@": ctx.executable._upload_cli.short_path,
        "@@UPLOSI@@": ctx.executable._uplosi.short_path,
        "@@VERSION@@": ctx.file._version.short_path,
    }
    ctx.actions.expand_template(
        template = ctx.file._upload_sh_tpl,
        output = executable,
        is_executable = True,
        substitutions = substitutions,
    )
    runfiles = ctx.runfiles(files = files)
    runfiles = runfiles.merge(ctx.attr._uplosi[DefaultInfo].data_runfiles)
    runfiles = runfiles.merge(ctx.attr._dissect_toolchain[DefaultInfo].data_runfiles)
    runfiles = runfiles.merge(ctx.attr._cosign[DefaultInfo].data_runfiles)
    runfiles = runfiles.merge(ctx.attr._rekor_cli[DefaultInfo].data_runfiles)
    runfiles = runfiles.merge(ctx.attr._parallel[DefaultInfo].data_runfiles)
    runfiles = runfiles.merge(ctx.attr._upload_cli[DefaultInfo].data_runfiles)
    runfiles = runfiles.merge(ctx.attr._measured_boot[DefaultInfo].data_runfiles)

    return DefaultInfo(executable = executable, runfiles = runfiles)

upload_os_images = rule(
    implementation = _upload_os_images_impl,
    attrs = {
        "image_dirs": attr.label_list(
            doc = "List of directories containing OS images to upload.",
        ),
        "_cosign": attr.label(
            default = Label("@cosign//:bin/cosign"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
        "_dissect_toolchain": attr.label(
            default = Label("@systemd//:bin/systemd-dissect"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
        "_measured_boot": attr.label(
            default = Label("//image/measured-boot/cmd"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
        "_parallel": attr.label(
            default = Label("@parallel//:bin/parallel"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
        "_rekor_cli": attr.label(
            default = Label("@rekor-cli//:bin/rekor-cli"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
        "_upload_cli": attr.label(
            default = Label("//image/upload"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
        "_upload_sh_tpl": attr.label(
            default = "upload_os_images.sh.in",
            allow_single_file = True,
        ),
        "_uplosi": attr.label(
            default = Label("@uplosi//:bin/uplosi"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
        "_version": attr.label(
            default = Label("//bazel/settings:tag"),
            allow_single_file = True,
        ),
    },
    executable = True,
)
