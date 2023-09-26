""" Constellation OS image configuration / variants """

VARIANTS = [
    {
        "attestation_variant": "aws-sev-snp",
        "csp": "aws",
    },
    {
        "attestation_variant": "aws-nitro-tpm",
        "csp": "aws",
    },
    {
        "attestation_variant": "azure-sev-snp",
        "csp": "azure",
    },
    {
        "attestation_variant": "gcp-sev-es",
        "csp": "gcp",
    },
    {
        "attestation_variant": "gcp-sev-snp",
        "csp": "gcp",
    },
    {
        "attestation_variant": "qemu-vtpm",
        "csp": "openstack",
    },
    {
        "attestation_variant": "qemu-vtpm",
        "csp": "qemu",
    },
]

STREAMS = [
    "stable",
    "nightly",
    "console",
    "debug",
]

CSPS = [
    "aws",
    "azure",
    "gcp",
    "openstack",
    "qemu",
]

base_cmdline = "selinux=1 enforcing=0 audit=0"

csp_settings = {
    "aws": {
        "kernel_command_line_dict": {
            "constel.csp": "aws",
            "idle": "poll",
            "mitigations": "auto",
        },
    },
    "azure": {
        "kernel_command_line_dict": {
            "constel.csp": "azure",
            "mitigations": "auto,nosmt",
        },
    },
    "gcp": {
        "kernel_command_line_dict": {
            "constel.csp": "gcp",
            "mitigations": "auto,nosmt",
        },
    },
    "openstack": {
        "autologin": True,
        "kernel_command_line": "console=tty0 console=ttyS0",
        "kernel_command_line_dict": {
            "console": "tty0",
            "constel.csp": "openstack",
            "kvm_amd.sev": "1",
            "mem_encrypt": "on",
            "mitigations": "auto,nosmt",
            "module_blacklist": "qemu_fw_cfg",
        },
    },
    "qemu": {
        "autologin": True,
        "kernel_command_line_dict": {
            "constel.csp": "qemu",
            "mitigations": "auto,nosmt",
        },
    },
}

attestation_variant_settings = {
    "aws-nitro-tpm": {
        "kernel_command_line_dict": {
            "constel.attestation-variant": "aws-nitro-tpm",
        },
    },
    "aws-sev-snp": {
        "kernel_command_line_dict": {
            "constel.attestation-variant": "aws-sev-snp",
        },
    },
    "azure-sev-snp": {
        "kernel_command_line_dict": {
            "constel.attestation-variant": "azure-sev-snp",
        },
    },
    "gcp-sev-es": {
        "kernel_command_line_dict": {
            "constel.attestation-variant": "gcp-sev-es",
        },
    },
    "gcp-sev-snp": {
        "kernel_command_line_dict": {
            "constel.attestation-variant": "gcp-sev-snp",
        },
    },
    "qemu-vtpm": {
        "kernel_command_line_dict": {
            "constel.attestation-variant": "qemu-vtpm",
        },
    },
}

stream_settings = {
    "console": {
        "autologin": True,
    },
    "debug": {
        "autologin": True,
        "kernel_command_line": "constellation.debug",
    },
    "nightly": {},
    "stable": {},
}

def from_settings(csp, attestation_variant, stream, strict = True, default = None):
    """Generates a list of settings dictionaries for the given csp, attestation_variant and stream.

    Args:
      csp: The cloud service provider to use.
      attestation_variant: The attestation variant to use.
      stream: The stream to use.
      strict: If True, fail if any of the given csp, attestation_variant or stream is unknown.
      default: The default value to use if any of the given csp, attestation_variant or stream is unknown.

    Returns:
        A list of settings dictionaries.
    """
    if strict:
        if not csp in csp_settings:
            fail("Unknown csp: " + csp)
        if not attestation_variant in attestation_variant_settings:
            fail("Unknown attestation_variant: " + attestation_variant)
        if not stream in stream_settings:
            fail("Unknown stream: " + stream)
    return [
        csp_settings.get(csp, default),
        attestation_variant_settings.get(attestation_variant, default),
        stream_settings.get(stream, default),
    ]

def constellation_packages(stream):
    base_packages = ["//measurement-reader/cmd:measurement-reader-package"]
    if stream == "debug":
        return ["//debugd/cmd/debugd:debugd-package"] + base_packages
    return [
        "//upgrade-agent/cmd:upgrade-agent-package",
        "//bootstrapper/cmd/bootstrapper:bootstrapper-package",
    ] + base_packages

def autologin(csp, attestation_variant, stream):
    """Generates a boolean indicating whether autologin should be enabled for the given csp, attestation_variant and stream.

    Args:
      csp: The cloud service provider to use.
      attestation_variant: The attestation variant to use.
      stream: The stream to use.

    Returns:
        A boolean indicating whether autologin should be enabled.
    """
    out = None
    for settings in from_settings(csp, attestation_variant, stream):
        if not "autologin" in settings:
            continue
        if out != None and out != settings["autologin"]:
            fail("Inconsistent autologin settings")
        out = settings["autologin"]
    return out

def kernel_command_line(csp, attestation_variant, stream):
    cmdline = base_cmdline
    for settings in from_settings(csp, attestation_variant, stream, default = {}):
        cmdline = append_cmdline(cmdline, settings.get("kernel_command_line", ""))
    return cmdline

def kernel_command_line_dict(csp, attestation_variant, stream):
    commandline_dict = {}
    for settings in from_settings(csp, attestation_variant, stream, default = {}):
        commandline_dict = commandline_dict | settings.get("kernel_command_line_dict", {})
    return commandline_dict

def append_cmdline(current, append):
    """Append a string to an existing commandline, separating them with a space.

    Args:
      current: The existing commandline. May be empty.
      append: The string to append. May be empty.

    Returns:
        The combined commandline.
    """
    if len(current) == 0:
        return append
    if len(append) == 0:
        return current
    return current + " " + append

def images_for_stream(stream):
    return [
        variant["csp"] + "_" + variant["attestation_variant"] + "_" + stream
        for variant in VARIANTS
    ]

def images_for_csp(csp):
    return [
        csp + "_" + variant["attestation_variant"] + "_" + stream
        for variant in VARIANTS
        if variant["csp"] == csp
        for stream in STREAMS
    ]

def images_for_csp_and_stream(csp, stream):
    return [
        csp + "_" + variant["attestation_variant"] + "_" + stream
        for variant in VARIANTS
        if variant["csp"] == csp
    ]
