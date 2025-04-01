# Install cert-manager

:::caution
If you want to use cert-manager with Constellation, pay attention to the following to avoid potential pitfalls.
:::

Constellation ships with cert-manager preinstalled.
The default installation is part of the `kube-system` namespace, as all other Constellation-managed microservices.
You are free to install more instances of cert-manager into other namespaces.
However, be aware that any new installation needs to use the same version as the one installed with Constellation or rely on the same CRD versions.
Also remember to set the `installCRDs` value to `false` when installing new cert-manager instances.
It will create problems if you have two installations of cert-manager depending on different versions of the installed CRDs.
CRDs are cluster-wide resources and cert-manager depends on specific versions of those CRDs for each release.
