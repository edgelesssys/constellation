{
  mkShell,
  git,
  bazel_7,
  go,
}:
mkShell {
  nativeBuildInputs = [
    bazel_7
    git
    go
  ];
}
