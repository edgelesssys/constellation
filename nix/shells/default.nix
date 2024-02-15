{ mkShell, git, bazel_7 }:
mkShell {
  nativeBuildInputs = [
    bazel_7
    git
  ];
}
