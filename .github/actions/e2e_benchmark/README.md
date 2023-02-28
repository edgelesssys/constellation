# Perf-Bench

## Continuous Benchmarking
The Benchmark action runs performance benchmarks on Constellation clusters.
The benchmark suite records storage and network benchmarks.

After testing, the action compares the results of the benchmarks to previous results of Constellation on the same cloud provider. That way, it is possible to evaluate performance progression throughout the development.

The data of previous benchmarks is stored in the private S3 artifact store.

In order to support encrypted storage, the action deploys the [Azure CSI](https://github.com/edgelesssys/constellation-azuredisk-csi-driver) and [GCP CSI](https://github.com/edgelesssys/constellation-gcp-compute-persistent-disk-csi-driver) drivers.

For the network benchmark we utilize the `knb` tool of the [k8s-bench-suite](https://github.com/InfraBuilder/k8s-bench-suite).

For the storage benchmark we utilize [kubestr](https://github.com/kastenhq/kubestr) to run FIO tests.

### Displaying Performance Progression
The action creates a summary of the action and attaches it the workflow execution log.

The table compares the current benchmark results of Constellation on the selected cloud provider to the previous records (of Constellation on the cloud provider).

The hashes of the two commits that are the base for the comparison are prepended to the table.

Example table:

<details>

- Commit of current benchmark: 8eb0a6803bc431bcebc2f6766ab2c6376500e106
- Commit of previous benchmark: 8f733daaf5c5509f024745260220d89ef8e6e440

| Benchmark suite | Metric | Current | Previous | Ratio |
|-|-|-|-|-|
| read_iops | iops (IOPS) | 213.6487 | 216.74684 | 0.986 ⬇️ |
| write_iops | iops (IOPS) | 24.412066 | 18.051243 | 1.352 ⬆️ |
| read_bw | bw_kbytes (KiB/s) | 28302.0 | 28530.0 | 0.992 ⬇️ |
| write_bw | bw_kbytes (KiB/s) | 4159.0 | 2584.0 | 1.61 ⬆️ |
| pod2pod | tcp_bw_mbit (MiB/s) | 20450.0 | 929.0 | 22.013 ⬆️ |
| pod2pod | upd_bw_mbit (MiB/s) | 1138.0 | 750.0 | 1.517 ⬆️ |
| pod2svc | tcp_bw_mbit (MiB/s) | 21188.0 | 905.0 | 23.412 ⬆️ |
| pod2svc | upd_bw_mbit (MiB/s) | 1137.0 | 746.0 | 1.524 ⬆️ |

</details>

### Drawing Performance Charts
The action also draws graphs as used in the [Constellation docs](https://docs.edgeless.systems/constellation/next/overview/performance). The graphs compare the performance of Constellation to the performance of managed Kubernetes clusters.

Graphs are created with every run of the benchmarking action. The action attaches them to the `benchmark` artifact of the workflow run.

## Updating Stored Records

### Managed Kubernetes
One must manually update the stored benchmark records of managed Kubernetes:

### AKS
Follow the [Azure documentation](https://learn.microsoft.com/en-us/azure/aks/learn/quick-kubernetes-deploy-portal?tabs=azure-cli) to create an AKS cluster of desired benchmarking settings (region, instance types). If comparing against Constellation clusters with CVM instances, make sure to select the matching CVM instance type on Azure as well.

Once the cluster is ready, set up managing access via `kubectl` and take the benchmark:
```bash
# Setup knb
git clone https://github.com/InfraBuilder/k8s-bench-suite.git
cd k8s-bench-suite
install knb /usr/local/bin
cd ..

# Setup kubestr
HOSTOS="$(go env GOOS)"
HOSTARCH="$(go env GOARCH)"
curl -fsSLO https://github.com/kastenhq/kubestr/releases/download/v${KUBESTR_VER}/kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
tar -xzf kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
install kubestr /usr/local/bin


# Run kubestr
mkdir -p out
kubestr fio -e "out/fio-constellation-aks.json" -o json -s encrypted-rwo -z 400Gi

# Run knb
workers=$(kubectl get nodes | grep worker)
server=$(echo $workers | head -1 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')
client=$(echo $workers | head -2 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')
knb -f "out/knb-constellation-aks.json" -o json --server-node $server --client-node $client


# Benchmarks done, do processing.

# Parse
git clone https://github.com/edgelesssys/constellation.git
mkdir -p benchmarks
BDIR=benchmarks
EXT_NAME=AKS
KBENCH_RESULTS=out/

python constellation/.github/actions/e2e_benchmark/evaluate/parse.py

# Upload result to S3
S3_PATH=s3://edgeless-artifact-store/constellation/benchmarks
aws s3 cp benchmarks/AKS.json ${S3_PATH}/AKS.json
```

### GKE
Create a GKE cluster of desired benchmarking settings (region, instance types). If comparing against Constellation clusters with CVM instances, make sure to select the matching CVM instance type on GCP and enable **confidential** VMs as well.

Once the cluster is ready, set up managing access via `kubectl` and take the benchmark:
```bash
# Setup knb
git clone https://github.com/InfraBuilder/k8s-bench-suite.git
cd k8s-bench-suite
install knb /usr/local/bin
cd ..

# Setup kubestr
HOSTOS="$(go env GOOS)"
HOSTARCH="$(go env GOARCH)"
curl -fsSLO https://github.com/kastenhq/kubestr/releases/download/v${KUBESTR_VER}/kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
tar -xzf kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
install kubestr /usr/local/bin

# Run kubestr
mkdir -p out
kubestr fio -e "out/fio-constellation-gke.json" -o json -s encrypted-rwo -z 400Gi

# Run knb
workers=$(kubectl get nodes | grep worker)
server=$(echo $workers | head -1 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')
client=$(echo $workers | head -2 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')
knb -f "out/knb-constellation-gke.json" -o json --server-node $server --client-node $client


# Parse
git clone https://github.com/edgelesssys/constellation.git
mkdir -p benchmarks
BDIR=benchmarks
EXT_NAME=GKE
KBENCH_RESULTS=out/

python constellation/.github/actions/e2e_benchmark/evaluate/parse.py

# Upload result to S3
S3_PATH=s3://edgeless-artifact-store/constellation/benchmarks
aws s3 cp benchmarks/GKE.json ${S3_PATH}/GKE.json
```

### Constellation
The action updates the stored Constellation records for the selected cloud provider when running on the main branch.
