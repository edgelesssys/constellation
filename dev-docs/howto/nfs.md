# Deploying NFS in Constellation using Rook

This document describes how to deploy NFS in Constellation using Rook.

## Create a Cluster

The cluster needs at least 3 worker nodes, default machines are powerful enough.

```bash
constellation create --name nfs -c 1 -w 3
```

## Deploy CSI Driver

> **_NOTE:_**  For additional integrity protection, use our [Constellation CSI drivers](https://docs.edgeless.systems/constellation/workflows/storage) with integrity protection enabled. With this option there is no need to enable encryption on Cephs side in the step [Deploy Rook](#deploy-rook).

We need block storage form somewhere. We will use the official Azure CSI for that. We need to create the azure config secret again with the expected fields. Replace "XXX" with the corresponding value from the secret `azureconfig`.

```bash
kubectl create secret generic -n kube-system --from-literal=cloud-config='{"cloud":"AzurePublicCloud","useInstanceMetadata":true,"vmType":"vmss","tenantId":"XXX","subscriptionId":"XXX","resourceGroup":"XXX","location":"XXX", "aadClientId":"XXX","aadClientSecret":"XXX"}' azure-config

helm repo add azuredisk-csi-driver https://raw.githubusercontent.com/kubernetes-sigs/azuredisk-csi-driver/master/charts
helm repo update azuredisk-csi-driver
helm install azuredisk-csi-driver azuredisk-csi-driver/azuredisk-csi-driver --namespace kube-system --set linux.distro=fedora --set controller.cloudConfigSecretName=azure-config --set node.cloudConfigSecretName=azure-config
```

## Deploy the StorageClass

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: managed-premium
provisioner: disk.csi.azure.com
parameters:
  skuName: Premium_LRS
  cachingmode: ReadOnly
  kind: Managed
volumeBindingMode: WaitForFirstConsumer
```

## Deploy Rook

```bash
git clone https://github.com/rook/rook.git
cd rook/deploy/examples
kubectl apply -f common.yaml -f crds.yaml -f operator.yaml
kubectl rollout status -n rook-ceph deployment/rook-ceph-operator
```

Apply the following changes to `cluster-on-pvc.yaml`:

```diff
euler@work:~/projects/rook/deploy/examples$ git diff cluster-on-pvc.yaml
diff --git a/deploy/examples/cluster-on-pvc.yaml b/deploy/examples/cluster-on-pvc.yaml
index ee4976be2..b5cf294cb 100644
--- a/deploy/examples/cluster-on-pvc.yaml
+++ b/deploy/examples/cluster-on-pvc.yaml
@@ -16,7 +16,7 @@ spec:
   mon:
     # Set the number of mons to be started. Generally recommended to be 3.
     # For highest availability, an odd number of mons should be specified.
-    count: 3
+    count: 1
     # The mons should be on unique nodes. For production, at least 3 nodes are recommended for this reason.
     # Mons should only be allowed on the same node for test environments where data loss is acceptable.
     allowMultiplePerNode: false
@@ -28,7 +28,7 @@ spec:
     # size appropriate for monitor data will be used.
     volumeClaimTemplate:
       spec:
-        storageClassName: gp2
+        storageClassName: managed-premium
         resources:
           requests:
             storage: 10Gi
@@ -59,13 +59,13 @@ spec:
         # Certain storage class in the Cloud are slow
         # Rook can configure the OSD running on PVC to accommodate that by tuning some of the Ceph internal
         # Currently, "gp2" has been identified as such
-        tuneDeviceClass: true
+        tuneDeviceClass: false
         # Certain storage class in the Cloud are fast
         # Rook can configure the OSD running on PVC to accommodate that by tuning some of the Ceph internal
         # Currently, "managed-premium" has been identified as such
-        tuneFastDeviceClass: false
+        tuneFastDeviceClass: true
         # whether to encrypt the deviceSet or not
-        encrypted: false
+        encrypted: true
         # Since the OSDs could end up on any node, an effort needs to be made to spread the OSDs
         # across nodes as much as possible. Unfortunately the pod anti-affinity breaks down
         # as soon as you have more than one OSD per node. The topology spread constraints will
@@ -100,7 +100,7 @@ spec:
           topologySpreadConstraints:
             - maxSkew: 1
               # IMPORTANT: If you don't have zone labels, change this to another key such as kubernetes.io/hostname
-              topologyKey: topology.kubernetes.io/zone
+              topologyKey: kubernetes.io/hostname
               whenUnsatisfiable: DoNotSchedule
               labelSelector:
                 matchExpressions:
@@ -127,7 +127,7 @@ spec:
                 requests:
                   storage: 10Gi
               # IMPORTANT: Change the storage class depending on your environment
-              storageClassName: gp2
+              storageClassName: managed-premium
               volumeMode: Block
               accessModes:
                 - ReadWriteOnce
```

Now apply the yaml:

```bash
kubectl apply -f cluster-on-pvc.yaml
```

Verify the health of the ceph cluster:

```bash
$ kubectl apply -f toolbox.yaml
$ kubectl -n rook-ceph exec -it deploy/rook-ceph-tools -- ceph status
  cluster:
    id:     7c220b31-29f7-4f17-a291-3ef39a9553b3
    health: HEALTH_OK

  services:
    mon: 3 daemons, quorum a,b,c (age 2m)
    mgr: a(active, since 72s)
    osd: 3 osds: 3 up (since 61s), 3 in (since 81s)

  data:
    pools:   1 pools, 1 pgs
    objects: 2 objects, 449 KiB
    usage:   62 MiB used, 30 GiB / 30 GiB avail
    pgs:     1 active+clean
```

Deploy the filesystem:

```bash
$ kubectl apply -f filesystem.yaml
$ kubectl -n rook-ceph exec -it deploy/rook-ceph-tools -- ceph status
  cluster:
    id:     7c220b31-29f7-4f17-a291-3ef39a9553b3
    health: HEALTH_OK

  services:
    mon: 3 daemons, quorum a,b,c (age 3m)
    mgr: a(active, since 2m)
    mds: 1/1 daemons up, 1 hot standby
    osd: 3 osds: 3 up (since 2m), 3 in (since 2m)

  data:
    volumes: 1/1 healthy
    pools:   3 pools, 34 pgs
    objects: 24 objects, 451 KiB
    usage:   63 MiB used, 30 GiB / 30 GiB avail
    pgs:     34 active+clean

  io:
    client:   853 B/s rd, 1 op/s rd, 0 op/s wr

  progress:
```

Deploy the StorageClass:

```bash
kubectl apply -f csi/cephfs/storageclass.yaml
```

Rescale the monitor count to 3:

```bash
kubectl -n rook-ceph patch cephcluster rook-ceph --type merge -p '{"spec":{"mon":{"count":3}}}'
```

## Use the NFS

The following deployment will create a PVC based on NFS and mount it into 3 pods.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 5Gi
  storageClassName: rook-cephfs
---
# from https://github.com/Azure/kubernetes-volume-drivers/tree/master/nfs
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-nfs
  labels:
    app: nginx
spec:
  serviceName: statefulset-nfs
  replicas: 3
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: statefulset-nfs
          image: nginx
          command:
            - "/bin/sh"
            - "-c"
            - "sleep 9999999"
          volumeMounts:
            - name: persistent-storage
              mountPath: /mnt/nfs
      volumes:
      - name: persistent-storage
        persistentVolumeClaim:
          claimName: nfs
          readOnly: false
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      app: nginx
```

## Verify Ceph OSD encryption

To verify that Ceph created an encrypted device, [log into a node](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/#ephemeral-container) via `kubectl debug`.

```bash
$ ls /dev/mapper/
control  root  set1-data-1flnzz-block-dmcrypt  state  state_dif

$ cryptsetup status /dev/mapper/set1-data-1flnzz-block-dmcrypt
/dev/mapper/set1-data-1flnzz-block-dmcrypt is active and is in use.
  type:    LUKS2
  cipher:  aes-xts-plain64
  keysize: 512 bits
  key location: dm-crypt
  device:  /dev/sdc
  sector size:  512
  offset:  32768 sectors
  size:    20938752 sectors
  mode:    read/write
  flags:   discards
```
