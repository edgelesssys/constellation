# Installing cert-manager
:::caution
Please read this section before installing cert-manager.
:::
Constellation ships with cert-manager preinstalled.
The default installation is part of the `kube-system` namespace, as all other Constellation-managed components.
You are free to install more instances of `cert-manager` into other namespaces.
However, please be aware that any new installation need to use the same version as the one installed with Constellation.
Or rely on the same CRD versions as Constellation's installation.
Also remember to set the `installCRDs` value to `false` when installing new `cert-manager` instances.
It will create problems if you have two installations of cert-manager depending on different versions of the installed CRDs.
CRDs are cluster-wide resources and `cert-manager` depends on specific versions of those CRDs for each release.
