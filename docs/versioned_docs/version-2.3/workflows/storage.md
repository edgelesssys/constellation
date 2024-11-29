# Use persistent storage

Persistent storage in Kubernetes requires cloud-specific configuration.
For abstraction of container storage, Kubernetes offers [volumes](https://kubernetes.io/docs/concepts/storage/volumes/),
allowing users to mount storage solutions directly into containers.
The [Container Storage Interface (CSI)](https://kubernetes-csi.github.io/docs/) is the standard interface for exposing arbitrary block and file storage systems into containers in Kubernetes.
Cloud service providers (CSPs) offer their own CSI-based solutions for cloud storage.

## Confidential storage

Most cloud storage solutions support encryption, such as [GCE Persistent Disks (PD)](https://cloud.google.com/kubernetes-engine/docs/how-to/using-cmek).
Constellation supports the available CSI-based storage options for Kubernetes engines in Azure and GCP.
However, their encryption takes place in the storage backend and is managed by the CSP.
Thus, using the default CSI drivers for these storage types means trusting the CSP with your persistent data.

To address this, Constellation provides CSI drivers for Azure Disk and GCE PD, offering [encryption on the node level](../architecture/keys.md#storage-encryption). They enable transparent encryption for persistent volumes without needing to trust the cloud backend. Plaintext data never leaves the confidential VM context, offering you confidential storage.

For more details see [encrypted persistent storage](../architecture/encrypted-storage.md).

## CSI drivers

Constellation supports the following drivers, which offer node-level encryption and optional integrity protection.

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

**Constellation CSI driver for Azure Disk**:
Mount Azure [Disk Storage](https://azure.microsoft.com/en-us/services/storage/disks/#overview) into your Constellation cluster. See the instructions on how to [install the Constellation CSI driver](#installation) or check out the [repository](https://github.com/edgelesssys/constellation-azuredisk-csi-driver) for more information. Since Azure Disks are mounted as ReadWriteOnce, they're only available to a single pod.

</TabItem>
<TabItem value="gcp" label="GCP">

**Constellation CSI driver for GCP Persistent Disk**:
Mount [Persistent Disk](https://cloud.google.com/persistent-disk) block storage into your Constellation cluster.
This includes support for [volume snapshots](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/volume-snapshots), which let you create copies of your volume at a specific point in time.
You can use them to bring a volume back to a prior state or provision new volumes.
Follow the instructions on how to [install the Constellation CSI driver](#installation) or check out the [repository](https://github.com/edgelesssys/constellation-gcp-compute-persistent-disk-csi-driver) for information about the configuration.

</TabItem>
<TabItem value="aws" label="AWS">

:::caution

Confidential storage isn't yet implemented for AWS. If you require this feature, [let us know](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&template=feature_request.md)!

You may use other (non-confidential) CSI drivers that are compatible with Kubernetes on AWS.

:::

</TabItem>
</Tabs>

Note that in case the options above aren't a suitable solution for you, Constellation is compatible with all other CSI-based storage options. For example, you can use [Azure Files](https://docs.microsoft.com/en-us/azure/storage/files/storage-files-introduction) or [GCP Filestore](https://cloud.google.com/filestore) with Constellation out of the box. Constellation is just not providing transparent encryption on the node level for these storage types yet.

## Installation

The Constellation CLI automatically installs Constellation's CSI driver for the selected CSP in your cluster.
If you don't need a CSI driver or wish to deploy your own, you can disable the automatic installation by setting `deployCSIDriver` to `false` in your Constellation config file.

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

Azure comes with two storage classes by default.

* `encrypted-rwo`
  * Uses [Standard SSDs](https://learn.microsoft.com/en-us/azure/virtual-machines/disks-types#standard-ssds)
  * ext-4 filesystem
  * Encryption of all data written to disk
* `integrity-encrypted-rwo`
  * Uses [Premium SSDs](https://learn.microsoft.com/en-us/azure/virtual-machines/disks-types#premium-ssds)
  * ext-4 filesystem
  * Encryption of all data written to disk
  * Integrity protection of data written to disk

For more information on encryption algorithms and key sizes, refer to [cryptographic algorithms](../architecture/encrypted-storage.md#cryptographic-algorithms).

:::info

The default storage class is set to `encrypted-rwo` for performance reasons.
If you want integrity-protected storage, set the `storageClassName` parameter of your persistent volume claim to `integrity-encrypted-rwo`.

Alternatively, you can create your own storage class with integrity protection enabled by adding `csi.storage.k8s.io/fstype: ext4-integrity` to the class `parameters`.
Or use another filesystem by specifying another file system type with the suffix `-integrity`, e.g., `csi.storage.k8s.io/fstype: xfs-integrity`.

Note that volume expansion isn't supported for integrity-protected disks.

:::

</TabItem>
<TabItem value="gcp" label="GCP">

GCP comes with two storage classes by default.

* `encrypted-rwo`
  * Uses [standard persistent disks](https://cloud.google.com/compute/docs/disks#pdspecs)
  * ext-4 filesystem
  * Encryption of all data written to disk
* `integrity-encrypted-rwo`
  * Uses [performance (SSD) persistent disks](https://cloud.google.com/compute/docs/disks#pdspecs)
  * ext-4 filesystem
  * Encryption of all data written to disk
  * Integrity protection of data written to disk

For more information on encryption algorithms and key sizes, refer to [cryptographic algorithms](../architecture/encrypted-storage.md#cryptographic-algorithms).

:::info

The default storage class is set to `encrypted-rwo` for performance reasons.
If you want integrity-protected storage, set the `storageClassName` parameter of your persistent volume claim to `integrity-encrypted-rwo`.

Alternatively, you can create your own storage class with integrity protection enabled by adding `csi.storage.k8s.io/fstype: ext4-integrity` to the class `parameters`.
Or use another filesystem by specifying another file system type with the suffix `-integrity`, e.g., `csi.storage.k8s.io/fstype: xfs-integrity`.

Note that volume expansion isn't supported for integrity-protected disks.

:::

</TabItem>
<TabItem value="aws" label="AWS">

:::caution

Confidential storage isn't yet implemented for AWS. If you require this feature, [let us know](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&template=feature_request.md)!

You may use other (non-confidential) CSI drivers that are compatible with Kubernetes on AWS.

:::

</TabItem>
</Tabs>

1. Create a [persistent volume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)

    A [persistent volume claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) is a request for storage with certain properties.
    It can refer to a storage class.
    The following creates a persistent volume claim, requesting 20 GB of storage via the `encrypted-rwo` storage class:

    ```bash
    cat <<EOF | kubectl apply -f -
    kind: PersistentVolumeClaim
    apiVersion: v1
    metadata:
      name: pvc-example
      namespace: default
    spec:
      accessModes:
      - ReadWriteOnce
      storageClassName: encrypted-rwo
      resources:
        requests:
          storage: 20Gi
    EOF
    ```

2. Create a Pod with persistent storage

    You can assign a persistent volume claim to an application in need of persistent storage.
    The mounted volume will persist restarts.
    The following creates a pod that uses the previously created persistent volume claim:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Pod
    metadata:
      name: web-server
      namespace: default
    spec:
      containers:
      - name: web-server
        image: nginx
        volumeMounts:
        - mountPath: /var/lib/www/html
          name: mypvc
      volumes:
      - name: mypvc
        persistentVolumeClaim:
          claimName: pvc-example
          readOnly: false
    EOF
    ```

### Change the default storage class

The default storage class is responsible for all persistent volume claims that don't explicitly request `storageClassName`.
Constellation creates a storage class with encryption enabled and sets this as the default class.
In case you wish to change it, follow the steps below:

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

  1. List the storage classes in your cluster:

      ```bash
      kubectl get storageclass
      ```

      The output is similar to this:

      ```shell-session
      NAME                      PROVISIONER                        RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
      encrypted-rwo (default)   azuredisk.csi.confidential.cloud   Delete          Immediate           true                   1d
      integrity-encrypted-rwo   azuredisk.csi.confidential.cloud   Delete          Immediate           false                  1d
      ```

      The default storage class is marked by `(default)`.

  2. Mark old default storage class as non default

      If you previously used another storage class as the default, you will have to remove that annotation:

      ```bash
      kubectl patch storageclass encrypted-rwo -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
      ```

  3. Mark new class as the default

      ```bash
      kubectl patch storageclass integrity-encrypted-rwo -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
      ```

  4. Verify that your chosen storage class is default:

      ```bash
      kubectl get storageclass
      ```

      The output is similar to this:

      ```shell-session
      NAME                                PROVISIONER                        RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
      encrypted-rwo                       azuredisk.csi.confidential.cloud   Delete          Immediate           true                   1d
      integrity-encrypted-rwo (default)   azuredisk.csi.confidential.cloud   Delete          Immediate           false                  1d
      ```

</TabItem>
<TabItem value="gcp" label="GCP">

  1. List the storage classes in your cluster:

      ```bash
      kubectl get storageclass
      ```

      The output is similar to this:

      ```shell-session
      NAME                      PROVISIONER                  RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
      encrypted-rwo (default)   gcp.csi.confidential.cloud   Delete          Immediate           true                   1d
      integrity-encrypted-rwo   gcp.csi.confidential.cloud   Delete          Immediate           false                  1d
      ```

      The default storage class is marked by `(default)`.

  2. Mark old default storage class as non default

      If you previously used another storage class as the default, you will have to remove that annotation:

      ```bash
      kubectl patch storageclass encrypted-rwo -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
      ```

  3. Mark new class as the default

      ```bash
      kubectl patch storageclass integrity-encrypted-rwo -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
      ```

  4. Verify that your chosen storage class is default:

      ```bash
      kubectl get storageclass
      ```

      The output is similar to this:

      ```shell-session
      NAME                                PROVISIONER                  RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
      encrypted-rwo                       gcp.csi.confidential.cloud   Delete          Immediate           true                   1d
      integrity-encrypted-rwo (default)   gcp.csi.confidential.cloud   Delete          Immediate           false                  1d
      ```

</TabItem>
<TabItem value="aws" label="AWS">

:::caution

Confidential storage isn't yet implemented for AWS. If you require this feature, [let us know](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&template=feature_request.md)!

You may use other (non-confidential) CSI drivers that are compatible with Kubernetes on AWS.

:::

</TabItem>
</Tabs>
