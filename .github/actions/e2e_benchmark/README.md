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
| pod2pod | udp_bw_mbit (MiB/s) | 1138.0 | 750.0 | 1.517 ⬆️ |
| pod2svc | tcp_bw_mbit (MiB/s) | 21188.0 | 905.0 | 23.412 ⬆️ |
| pod2svc | udp_bw_mbit (MiB/s) | 1137.0 | 746.0 | 1.524 ⬆️ |

</details>

## Updating Stored Records

### Managed Kubernetes
One must manually update the stored benchmark records of managed Kubernetes:

### AKS
Follow the [Azure documentation](https://learn.microsoft.com/en-us/azure/aks/learn/quick-kubernetes-deploy-portal?tabs=azure-cli) to create an AKS cluster of desired benchmarking settings (region, instance types). If comparing against Constellation clusters with CVM instances, make sure to select the instance type on AKS as well.

For example:
```bash
az aks create -g moritz-constellation -n benchmark --node-count 2
az aks get-credentials -g moritz-constellation -n benchmark
```

Once the cluster is ready, set up managing access via `kubectl` and take the benchmark:
```bash
# Setup knb
git clone https://github.com/InfraBuilder/k8s-bench-suite.git
cd k8s-bench-suite
install knb /usr/local/bin
cd ..

# Setup kubestr
case "$(go env GOOS)" in "darwin") HOSTOS="MacOS";; *) HOSTOS="$(go env GOOS)";; esac
HOSTARCH="$(go env GOARCH)"
KUBESTR_VER=0.4.37
curl -fsSLO https://github.com/kastenhq/kubestr/releases/download/v${KUBESTR_VER}/kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
tar -xzf kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
install kubestr /usr/local/bin


# Run kubestr
mkdir -p out
kubestr fio -e "out/fio-AKS.json" -o json -s default -z 400Gi

# Run knb
workers="$(kubectl get nodes | grep nodepool)"
server="$(echo $workers | head -1 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')"
client="$(echo $workers | head -2 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')"
knb -f "out/knb-AKS.json" -o json --server-node $server --client-node $client


# Benchmarks done, do processing.

# Parse
git clone https://github.com/edgelesssys/constellation.git
mkdir -p benchmarks
export BDIR=benchmarks
export CSP=azure
export EXT_NAME=AKS
export BENCH_RESULTS=out/

python constellation/.github/actions/e2e_benchmark/evaluate/parse.py

# Upload result to S3
S3_PATH=s3://edgeless-artifact-store/constellation/benchmarks
aws s3 cp benchmarks/AKS.json ${S3_PATH}/AKS.json
```

### GKE
Create a GKE cluster of desired benchmarking settings (region, instance types). If comparing against Constellation clusters with CVM instances, make sure to select the matching instance type on GKE.
For example:

```bash
gcloud container clusters create benchmark \
    --zone europe-west3-b \
    --node-locations europe-west3-b \
    --machine-type n2d-standard-4 \
    --num-nodes 2
gcloud container clusters get-credentials benchmark --region europe-west3-b
# create storage class for pd-standard
cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: pd-standard
provisioner: pd.csi.storage.gke.io
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
parameters:
  type: pd-standard
EOF
```

Once the cluster is ready, set up managing access via `kubectl` and take the benchmark:
```bash
# Setup knb
git clone https://github.com/InfraBuilder/k8s-bench-suite.git
cd k8s-bench-suite
install knb /usr/local/bin
cd ..

# Setup kubestr
case "$(go env GOOS)" in "darwin") HOSTOS="MacOS";; *) HOSTOS="$(go env GOOS)";; esac
HOSTARCH="$(go env GOARCH)"
KUBESTR_VER=0.4.37
curl -fsSLO https://github.com/kastenhq/kubestr/releases/download/v${KUBESTR_VER}/kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
tar -xzf kubestr_${KUBESTR_VER}_${HOSTOS}_${HOSTARCH}.tar.gz
install kubestr /usr/local/bin

# Run kubestr
mkdir -p out
kubestr fio -e "out/fio-GKE.json" -o json -s pd-standard -z 400Gi

# Run knb
workers="$(kubectl get nodes | grep default-pool)"
server="$(echo $workers | head -1 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')"
client="$(echo $workers | head -2 | tail -1 |cut -d ' ' -f1|tr '\n' ' ')"
knb -f "out/knb-GKE.json" -o json --server-node "$server" --client-node "$client"


# Parse
git clone https://github.com/edgelesssys/constellation.git
mkdir -p benchmarks
export BDIR=benchmarks
export CSP=gcp
export EXT_NAME=GKE
export BENCH_RESULTS=out/

python constellation/.github/actions/e2e_benchmark/evaluate/parse.py

# Upload result to S3
S3_PATH=s3://edgeless-artifact-store/constellation/benchmarks
aws s3 cp benchmarks/GKE.json ${S3_PATH}/GKE.json
```

### Constellation
The action updates the stored Constellation records for the selected cloud provider when running on the main branch.

## Drawing Performance Charts
The action also contains the code to draw graphs as used in the [Constellation docs](https://docs.edgeless.systems/constellation/next/overview/performance).
The graphs compare the performance of Constellation to the performance of managed Kubernetes clusters.
It expects the results of `[AKS.json, GKE.json, constellation-azure.json, constellation-gcp.json]` to be present in the `BDIR` folder.

Graphs can thne be created from using the `graphs.py` script:

```bash
BDIR=benchmarks
python ./graph.py
```
