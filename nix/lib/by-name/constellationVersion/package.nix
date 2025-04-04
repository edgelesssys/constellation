# Returns the current Continuum version, as defined in `version.txt`.

{ lib }: builtins.readFile (lib.path.append lib.constellationRepoRoot "version.txt")
