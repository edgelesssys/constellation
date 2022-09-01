![Constellation](docs/banner.svg)

# Always Encrypted K8s

<p>
    <a href="https://github.com/edgelesssys/constellation/blob/master/LICENSE"><img src="https://img.shields.io/github/license/edgelesssys/constellation" alt="Constellation License"></a>
    <a href="https://github.com/edgelesssys/constellation/actions/workflows/e2e-test-azure.yml/badge.svg?branch=main"><img src="https://github.com/edgelesssys/constellation/actions/workflows/e2e-test-azure.yml/badge.svg?branch=main" alt="E2E Test Azure"></a>
    <a href="https://github.com/edgelesssys/constellation/actions/workflows/e2e-test-gcp.yml/badge.svg?branch=main"><img src="https://github.com/edgelesssys/constellation/actions/workflows/e2e-test-gcp.yml/badge.svg?branch=main" alt="E2E Test GCP"></a>
    <a href="https://goreportcard.com/report/github.com/edgelesssys/constellation"><img src="https://goreportcard.com/badge/github.com/edgelesssys/constellation" alt="Go Report"></a>
    <a href="https://discord.gg/rH8QTH56JN"><img src="https://img.shields.io/discord/823900998606651454?color=7389D8&label=discord&logo=discord&logoColor=ffffff" alt="Discord"></a>
    <a href="https://twitter.com/EdgelessSystems"><img src="https://img.shields.io/twitter/follow/EdgelessSystems?label=Follow" alt="Twitter"></a>
</p>

Constellation is a Kubernetes engine that aims to provide the best possible data security. It wraps your K8s cluster into a single *confidential context* that is shielded from the underlying cloud infrastructure. Everything inside is always encrypted, including at runtime in memory. For this, Constellation leverages confidential computing (see our [whitepaper]) and more specifically Confidential VMs.

<img src="docs/concept.svg" alt="Concept" width="65%"/>

## Goals

From a security perspective, Constellation is designed to keep all data always encrypted and to prevent access from the infrastructure layer (i.e., remove the infrastructure from the TCB). This includes access from datacenter employees, privileged cloud admins, and attackers coming through the infrastructure (e.g., malicious co-tenants escalating their privileges). 

From a DevOps perspective, Constellation is designed to work just like what you would expect from a modern K8s engine. 

## Use cases

Encrypting your K8s is good for:

* Increasing the overall security of your clusters
* Increasing the trustworthiness of your SaaS offerings
* Moving sensitive workloads from on-prem to the cloud
* Meeting regulatory requirements

## Features

### 🔒 Everything always encrypted

* Runtime encryption: All nodes run inside AMD SEV-based Confidential VMs (CVMs). Support for Intel TDX will be added in the future.
* Transparent encryption of network and storage: All pod-to-pod traffic and all writes to persistent storage are automatically encrypted ([more][network-encryption])
* Transparent key management: All cryptographic keys are managed within the confidential context ([more][key-management])

### 🔍 Everything verifiable

* "Whole cluster" attestation based on the remote-attestation feature of CVMs ([more][cluster-attestation])
* Confidential computing-optimized node images based on Fedora CoreOS; fully measured and integrity-protected ([more][images])
* Supply chain protection with [Sigstore](https://www.sigstore.dev/) ([more][supply-chain])

### 🚀 Performance and scale

* High availability with multi-master architecture and stacked etcd topology 
* Dynamic cluster autoscaling with verification and secure bootstrapping of new nodes
* Competitive performance ([see K-Bench comparison with AKS and GKE][performance])

### 🧩 Easy to use and integrate

<a href="https://landscape.cncf.io/?selected=constellation"><img src="https://raw.githubusercontent.com/cncf/artwork/1c1a10d9cc7de24235e07c8831923874331ef233/projects/kubernetes/certified-kubernetes/versionless/color/certified-kubernetes-color.svg" align="right" width="100px"></a>

* Constellation is a [CNCF-certified][certified] Kubernetes. It's aligned to Kubernetes' [version support policy][k8s-version-support] and will likely work with your existing workloads and tools.
* ☁️ Support for Azure and GCP, more to come.

## Getting started

If you're already familiar with Kubernetes, it's easy to get started with Constellation:

1. 📦 [Install the CLI][install]
2. ⌨️ [Create a Constellation cluster][create-cluster]
3. 🏎️ [Run your app][examples]

![Constellation Shell](docs/shell-windowframe.svg)

## Documentation

To learn more, see the official [documentation](https://docs.edgeless.systems/constellation).
You may want to start with one of the following sections.

* [Confidential Kubernetes][confidential-kubernetes] (Constellation vs. AKS/GKE + CVMs)
* [Security benefits][security-benefits]
* [Architecture][architecture]

## Support

* Please ask questions via [Discord][discord] or file an [issue][github-issues].
* If you experience errors, please create a [bug report][github-issues].
* Visit our [blog](https://blog.edgeless.systems/) for technical deep-dives and tutorials and follow us on [Twitter] for news.
* Edgeless Systems also offers [Enterprise Support][enterprise-support].

## Contributing

Refer to [`CONTRIBUTING.md`](CONTRIBUTING.md) on how to contribute. The most important points: 
* Pull requests are welcome! You need to agree to our [Contributor License Agreement][cla-assistant].
* Please follow the [Code of Conduct](/CODE_OF_CONDUCT.md). 
* ⚠️ To report a security issue, please write to security@edgeless.systems.

## License

The Constellation source code is licensed under the [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.en.html). Edgeless Systems provides pre-built and signed binaries and images for Constellation. You may use these free of charge to create and run services for internal consumption. You can find more information in the [license] section of the docs.

<!-- refs -->
[architecture]: https://docs.edgeless.systems/constellation/architecture/overview
[certified]: https://www.cncf.io/certification/software-conformance/
[Cilium]: https://cilium.io/
[cla-assistant]: https://cla-assistant.io/edgelesssys/constellation
[cluster-attestation]: https://docs.edgeless.systems/constellation/architecture/attestation#cluster-attestation
[confidential-kubernetes]: https://docs.edgeless.systems/constellation/overview/confidential-kubernetes
[enterprise-support]: https://www.edgeless.systems/products/constellation/
[create-cluster]: https://docs.edgeless.systems/constellation/workflows/create
[documentation]: https://docs.edgeless.systems/constellation/latest
[examples]: https://docs.edgeless.systems/constellation/getting-started/examples
[github-issues]: https://github.com/edgelesssys/constellation/issues/new/choose
[images]: https://docs.edgeless.systems/constellation/architecture/images#constellation-images
[install]: https://docs.edgeless.systems/constellation/getting-started/install
[k8s-version-support]: https://docs.edgeless.systems/constellation/architecture/versions#kubernetes-support-policy
[key-management]: https://docs.edgeless.systems/constellation/architecture/
[license]: https://docs.edgeless.systems/constellation/next/overview/license
[network-encryption]: https://docs.edgeless.systems/constellation/architecture/keys#network-encryption
[supply-chain]: https://docs.edgeless.systems/constellation/architecture/attestation#chain-of-trust
[security-benefits]: https://docs.edgeless.systems/constellation/next/overview/security-benefits
[twitter]: https://twitter.com/EdgelessSystems
[whitepaper]: https://content.edgeless.systems/hubfs/Confidential%20Computing%20Whitepaper.pdf
[performance]: https://docs.edgeless.systems/constellation/next/overview/performance