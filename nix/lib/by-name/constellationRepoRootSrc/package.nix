# Returns a package set originating from the root of the Constellation repository.
# The `files` attribute is a list of paths relative to the root of the repository.

{ lib }:
files:
let
  filteredFiles = lib.map (subpath: lib.path.append lib.constellationRepoRoot subpath) files;
in
lib.fileset.toSource {
  root = lib.constellationRepoRoot;
  fileset = lib.fileset.unions filteredFiles;
}
