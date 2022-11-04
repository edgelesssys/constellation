# K-Bench

## Continous Benchmarking
The K-Bench action runs K-Bench benchmarks on Constellation clusters.
The benchmark suite records storage, network, and Kubernetes API benchmarks.

After testing, the action compares the results of the benchmarks to previous results of Constellation on the same cloud provider. That way, it is possible to evaluate performance progression throughout the development.

The data of previous benchmarks is stored in the public S3 bucket.

In order to support encrypted storage, the action deploys the [GCP CSI](https://github.com/edgelesssys/constellation-gcp-compute-persistent-disk-csi-driver) and [Azure CSI](https://github.com/edgelesssys/constellation-azuredisk-csi-driver) drivers. It uses a [fork](https://github.com/edgelesssys/k-bench) of VMware's K-Bench. The fork deploys volumes that use the `encrypted-storage` storage class. Also, it has support to authenticate against GCP which is required to update the stored records for GKE.

### Displaying Performance Progression
The action creates a summary of the action and attaches it the workflow execution log.

The table compares the current benchmark results of Constellation on the selected cloud provider to the previous records (of Constellation on the cloud provider).

Example table:

<details>

| Benchmark suite | Current | Previous | Ratio |
|-|-|-|-|
| pod_create (ms) | 306.521 | 196.762 | 1.558 ⬆️ |
| pod_list (ms) | 157.272 | 2083.966 | 0.075 ⬇️ |
| pod_get (ms) | 157.42899 | 99.745 | 1.578 ⬆️ |
| pod_update (ms) | 176.695 | 150.139 | 1.177 ⬆️ |
| pod_delete (ms) | 165.996 | 449.71698 | 0.369 ⬇️ |
| svc_create (ms) | 208.479 | 407.957 | 0.511 ⬇️ |
| svc_list (ms) | 154.088 | 1398.8309 | 0.11 ⬇️ |
| svc_get (ms) | 156.53699 | 100.264 | 1.561 ⬆️ |
| svc_update (ms) | 157.96 | 127.22 | 1.242 ⬆️ |
| svc_delete (ms) | 193.49901 | 914.974 | 0.211 ⬇️ |
| depl_create (ms) | 347.759 | 212.902 | 1.633 ⬆️ |
| depl_list (ms) | 158.06 | 2060.005 | 0.077 ⬇️ |
| depl_update (ms) | 175.196 | 124.198 | 1.411 ⬆️ |
| depl_scale (ms) | 161.622 | 215.881 | 0.749 ⬇️ |
| depl_delete (ms) | 157.02899 | 178.04 | 0.882 ⬇️ |
| net_internode_snd (Mbit/s) | 979.0 | 1090.0 | 1.113 ⬆️ |
| net_intranode_snd (Mbit/s) | 20700.0 | 30000.0 | 1.449 ⬆️ |
| fio_root_async_R70W30_R (MiB/s) | 0.4541015625 | 5.1953125 | 11.441 ⬆️ |
| fio_root_async_R70W30_W (MiB/s) | 0.2001953125 | 2.2353515625 | 11.166 ⬆️ |
| fio_root_async_R100W0_R (MiB/s) | 0.58984375 | 39.0 | 66.119 ⬆️ |
| fio_root_async_R0W100_W (MiB/s) | 1.1806640625 | 2.3896484375 | 2.024 ⬆️ |

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
S3_PATH=s3://public-edgeless-constellation/benchmarks
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
S3_PATH=s3://public-edgeless-constellation/benchmarks
aws s3 cp benchmarks/GKE.json ${S3_PATH}/GKE.json
```

### Constellation
The action updates the stored Constellation records for the selected cloud provider when running on the main branch.
