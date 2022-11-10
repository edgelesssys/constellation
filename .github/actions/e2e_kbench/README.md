# K-Bench

## Continuous Benchmarking
The K-Bench action runs K-Bench benchmarks on Constellation clusters.
The benchmark suite records storage, network, and Kubernetes API benchmarks.

After testing, the action compares the results of the benchmarks to previous results of Constellation on the same cloud provider. That way, it is possible to evaluate performance progression throughout the development.

The data of previous benchmarks is stored in the private S3 artifact store.

In order to support encrypted storage, the action deploys the [Azure CSI](https://github.com/edgelesssys/constellation-azuredisk-csi-driver) and [GCP CSI](https://github.com/edgelesssys/constellation-gcp-compute-persistent-disk-csi-driver) drivers. It uses a [fork](https://github.com/edgelesssys/k-bench) of VMware's K-Bench. The fork deploys volumes that use the `encrypted-storage` storage class. Also, it has support to authenticate against GCP which is required to update the stored records for GKE.

### Displaying Performance Progression
The action creates a summary of the action and attaches it the workflow execution log.

The table compares the current benchmark results of Constellation on the selected cloud provider to the previous records (of Constellation on the cloud provider).

The hashes of the two commits that are the base for the comparison are prepended to the table.

Example table:

<details>

- Commit of current benchmark: 8eb0a6803bc431bcebc2f6766ab2c6376500e106
- Commit of previous benchmark: 8f733daaf5c5509f024745260220d89ef8e6e440

| Benchmark suite | Current | Previous | Ratio |
|-|-|-|-|
| pod_create (ms) | 135 | 198 | 0.682 ⬇️ |
| pod_list (ms) | 100 | 99 | 1.01 ⬆️ |
| pod_get (ms) | 98 | 98 | 1.0 ⬆️ |
| pod_update (ms) | 187 | 132 | 1.417 ⬆️ |
| pod_delete (ms) | 119 | 108 | 1.102 ⬆️ |
| svc_create (ms) | 156 | 149 | 1.047 ⬆️ |
| svc_list (ms) | 97 | 96 | 1.01 ⬆️ |
| svc_get (ms) | 97 | 96 | 1.01 ⬆️ |
| svc_update (ms) | 100 | 101 | 0.99 ⬇️ |
| svc_delete (ms) | 143 | 139 | 1.029 ⬆️ |
| depl_create (ms) | 201 | 218 | 0.922 ⬇️ |
| depl_list (ms) | 101 | 101 | 1.0 ⬆️ |
| depl_update (ms) | 111 | 110 | 1.009 ⬆️ |
| depl_scale (ms) | 391 | 391 | 1.0 ⬆️ |
| depl_delete (ms) | 401 | 402 | 0.998 ⬇️ |
| net_internode_snd (Mbit/s) | 953.0 | 964.0 | 1.01 ⬆️ |
| net_intranode_snd (Mbit/s) | 18500.0 | 18600.0 | 1.01 ⬆️ |
| fio_root_async_R70W30_R (MiB/s) | 0.45 | 0.45| 1.0 ⬆️ |
| fio_root_async_R70W30_W (MiB/s) | 0.20 | 0.20 | 1.0 ⬆️ |
| fio_root_async_R100W0_R (MiB/s) | 0.59 | 0.59 | 1.0 ⬆️ |
| fio_root_async_R0W100_W (MiB/s) | 1.18 | 1.18 | 1.0 ⬆️ |

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
# Setup
git clone https://github.com/edgelesssys/k-bench.git
cd k-bench && git checkout feat/constellation
./install.sh

# Remove the Constellation encrypted storage class
# Remember to revert this change before running K-Bench on Constellation!
yq 'del(.spec.storageClassName)' config/dp_fio/fio_pvc.yaml
yq 'del(.spec.storageClassName)' config/dp_netperf_internode/netperf_pvc.yml
yq 'del(.spec.storageClassName)' config/dp_network_internode/netperf_pvc.yaml
yq 'del(.spec.storageClassName)' config/dp_network_intranode/netperf_pvc.yml

# Run K-Bench
mkdir -p ./out
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml | kubectl apply -f -
./run.sh -r "kbench-AKS" -t "default" -o "./out/"
kubectl delete namespace kbench-pod-namespace --wait=true || true
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml |  kubectl apply -f -
./run.sh -r "kbench-AKS" -t "dp_fio" -o "./out/"
kubectl delete namespace kbench-pod-namespace --wait=true || true
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml |  kubectl apply -f -
./run.sh -r "kbench-AKS" -t "dp_network_internode" -o "./out/"
kubectl delete namespace kbench-pod-namespace --wait=true || true
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml |  kubectl apply -f -
./run.sh -r "kbench-AKS" -t "dp_network_intranode" -o "./out/"

# Benchmarks done, do processing.
mkdir -p "./out/AKS"
mv ./out/results_kbench-AKS_*m/* "./out/kbench-AKS/"

# Parse
git clone https://github.com/edgelesssys/constellation.git
mkdir -p benchmarks
BDIR=benchmarks
EXT_NAME=AKS
KBENCH_RESULTS=k-bench/out/

python constellation/.github/actions/k-bench/evaluate/parse.py

# Upload result to S3
S3_PATH=s3://edgeless-artifact-store/constellation/benchmarks
aws s3 cp benchmarks/AKS.json ${S3_PATH}/AKS.json
```

### GKE
Create a GKE cluster of desired benchmarking settings (region, instance types). If comparing against Constellation clusters with CVM instances, make sure to select the matching CVM instance type on GCP and enable **confidential** VMs as well.

Once the cluster is ready, set up managing access via `kubectl` and take the benchmark:
```bash
# Setup
git clone https://github.com/edgelesssys/k-bench.git
cd k-bench && git checkout feat/constellation
./install.sh

# Remove the Constellation encrypted storage class
# Remember to revert this change before running K-Bench on Constellation!
yq 'del(.spec.storageClassName)' config/dp_fio/fio_pvc.yaml
yq 'del(.spec.storageClassName)' config/dp_netperf_internode/netperf_pvc.yml
yq 'del(.spec.storageClassName)' config/dp_network_internode/netperf_pvc.yaml
yq 'del(.spec.storageClassName)' config/dp_network_intranode/netperf_pvc.yml

# Run K-Bench
mkdir -p ./out
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml | kubectl apply -f -
./run.sh -r "kbench-GKE" -t "default" -o "./out/"
kubectl delete namespace kbench-pod-namespace --wait=true || true
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml |  kubectl apply -f -
./run.sh -r "kbench-GKE" -t "dp_fio" -o "./out/"
kubectl delete namespace kbench-pod-namespace --wait=true || true
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml |  kubectl apply -f -
./run.sh -r "kbench-GKE" -t "dp_network_internode" -o "./out/"
kubectl delete namespace kbench-pod-namespace --wait=true || true
kubectl create namespace kbench-pod-namespace --dry-run=client -o yaml |  kubectl apply -f -
./run.sh -r "kbench-GKE" -t "dp_network_intranode" -o "./out/"

# Benchmarks done, do processing.
mkdir -p "./out/GKE"
mv ./out/results_kbench-GKE_*m/* "./out/kbench-GKE/"

# Parse
git clone https://github.com/edgelesssys/constellation.git
mkdir -p benchmarks
BDIR=benchmarks
EXT_NAME=GKE
KBENCH_RESULTS=k-bench/out/

python constellation/.github/actions/k-bench/evaluate/parse.py

# Upload result to S3
S3_PATH=s3://edgeless-artifact-store/constellation/benchmarks
aws s3 cp benchmarks/GKE.json ${S3_PATH}/GKE.json
```

### Constellation
The action updates the stored Constellation records for the selected cloud provider when running on the main branch.
