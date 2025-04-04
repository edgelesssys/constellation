{
  lib,
  callPackage,
}:
lib.packagesFromDirectoryRecursive {
  inherit callPackage;
  directory = ./by-name;
}
