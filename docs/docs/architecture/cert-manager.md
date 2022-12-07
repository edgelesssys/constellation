# Cert-manager

Constellation ships with cert-manager preinstalled.
The default installation is part of the `kube-system` namespace, as all other Constellation-managed components.
While you are free to install more instances of `cert-manager` into other namespaces, please be aware that any new installation need to use the same version as the one installed with Constellation.
Or rely on the same CRD versions as Constellation's installation.
Also remeber to set the `installCRDs` value to `false` when installing new `cert-manager` instances.
Because CRDs are cluster-wide resources and `cert-manager` uses CRDs to expose it's functionality it will create problems if you have two installations of cert-manager depending on different versions of those CRDs.
