{ nixpkgs
, symlinkJoin
, pkgs
, bazel_7
, lib
, substituteAll
, writeShellScriptBin
, bash
, coreutils
, diffutils
, file
, findutils
, gawk
, gnugrep
, gnupatch
, gnused
, gnutar
, gzip
, python3
, unzip
, which
, zip
}:
let
  defaultShellUtils = [
    bash
    coreutils
    diffutils
    file
    findutils
    gawk
    gnugrep
    gnupatch
    gnused
    gnutar
    gzip
    python3
    unzip
    which
    zip
  ];
  bazel_with_action_env = bazel_7.overrideAttrs (oldAttrs: {
    # https://github.com/NixOS/nixpkgs/pull/262152#issuecomment-1879053113
    patches = (oldAttrs.patches or [ ]) ++ [
      (substituteAll {
        src = "${nixpkgs}/pkgs/development/tools/build-managers/bazel/bazel_6/actions_path.patch";
        actionsPathPatch = lib.makeBinPath defaultShellUtils;
      })
    ];
  });
in
symlinkJoin {
  name = "bazel_7";
  paths = [ bazel_with_action_env ];
  buildInputs = [ pkgs.makeWrapper ];
  # This wrapper is required in Nix shells to make them work with
  # bazel's --incompatible_sandbox_hermetic_tmp flag.
  # See https://github.com/bazelbuild/bazel/issues/5900
  postBuild = ''
    wrapProgram $out/bin/bazel \
      --unset TMPDIR \
      --unset TMP
  '';
}
