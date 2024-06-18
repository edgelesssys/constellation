# Constellation with GPUs on Azure

This document describes the general setup flow to run Constellation on Azure's GPU CVM preview.

## Prerequisites

- A Constellation CLI built on a branch with GPU support.
- A Constellation OS Image built on a branch with GPU support.
- Access to the Azure H100 preview, and the capability to create the resources for Constellation in the `eastus2` region.

## Step-by-Step Guide

1. Using the CLI, create a configuration file:
    ```shell-session
    constellation config generate azure
    ```
2. Create IAM configuration as per your requirements. The region needs to be set to `eastus2`.
    ```shell-session
    constellation iam create azure --update-config --region eastus2 --resourceGroup msanft-gpu-rg --servicePrincipal msanft-gpu-sp
    ```
3. Then, a few patches to the configuration are necessary.
    1. Replace the `.image` field's contents with the canonical URL for the node image (e.g. `ref/feat-gpu-h100-support/stream/debug/v2.17.0-pre.0.20240405174758-ec33feb41903`)
    2. Replace the `.microserviceVersion` field's contents with the CLI version used (i.e. the output of `constellation version`)
    3. Set the `nodeGroups.*.instanceTypes` fields' contents to `Standard_NCC40ads_H100_v5`, which is the GPU-enabled CVM type.
    4. Set the `attestation.azureSEVSNP.bootloaderVersion` to `7` and the `attestation.azureSEVSNP.microcodeVersion` to 62. This is due to the GPU-enabled CVMs using the "Genoa" CPU generation, which is not yet used in any other VMs on Azure and thus not reflected by our API resolving the `latest` versions.
    5. Fetch the measurements for your image:
        ```shell-session
        constellation config fetch-measurements --insecure
        ```
        The  `--insecure` flag is necessary, as the image (most likely) won't be a stable image, and is thus not signed by us.
4. Then, continue creating a cluster as per usual:
    ```shell-session
    constellation apply
    ```

## Deploying a sample workload

To deploy an examplary AI workload, we can use [this Llama-2-7B deployment](https://github.com/chenhunghan/ialacol):

```shell-session
helm repo add ialacol https://chenhunghan.github.io/ialacol
helm repo update
helm install llama-2-7b-chat ialacol/ialacol

kubectl port-forward svc/llama-2-7b-chat 8000:8000
curl -X POST \
     -H 'Content-Type: application/json' \
     -d '{ "messages": [{"role": "user", "content": "How are you?"}], "model": "llama-2-7b-chat.ggmlv3.q4_0.bin", "stream": false}' \
     http://localhost:8000/v1/chat/completions
```

But of course, also any other AI workload can be deployed.

## Known Limitations

- Control Plane nodes currently also need to run in GPU-enabled CVMs, as `eastus2` doesn't offer any other SEV-SNP VM type, and mixing regions or attestation types is currently not supported in Constellation. If you want to use the GPU power of the control plane nodes, you should enable workloads to schedule on them.
- We currently use the ARK returned by Azure's compatibility layer (THIM), which would not be the case in production. We would normally use the ARK given by AMD and hardcode it in `constellation-conf.yaml`, but due to the Genoa CPUs not being present anywhere else, this has not yet been done.
