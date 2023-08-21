# Deprecations

When removing a feature/flag/command we strive to mark said feature as deprecated one minor version before the removal happens.
I.e. to remove the CLI command `foo` in version v10.6, all invocations of `foo` in v10.5 should print a deprecation warning.
The deprecation should also be noted in the help text of the command to make it show up in the Constellation docs.

The changelog for v10.5 should contain a deprecation warning for `foo`.
The changelog for v10.6, should sort the removal of `foo` in the category `Breaking Changes`.

It may happen that the effort to implement a feature in this manner is deemed to high.
