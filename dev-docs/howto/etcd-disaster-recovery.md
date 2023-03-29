# etcd disaster recovery
When etcd loses quorum and (N-1)/2 control planes are lost, 
etcd will not be able to perform any transactions anymore and will basically stall and cause timeouts.
This makes the Constellation control planes unusable, resulting in a frozen cluster. The worker nodes will continue to function for a bit,
but given that they cannot communicate to the control plane anymore, they will eventually also cease to function correctly.

If the missing control plane nodes are still existent and their state disk still exists, you likely do not need this guide.
If the missing are irrecoverably lost (e.g. scaled up and down the control plane instance set), follow through this guide to bring the cluster back up.

## General concept
1. Create snapshot of a state disk from a remaining control plane node.
2. Download the snapshot and decrypt it locally
3. Follow the [restoring a cluster](https://etcd.io/docs/v3.5/op-guide/recovery/#restoring-a-cluster) guide from etcd.
4. Save the modified virtual disk and upload it back to the CSP.
5. Modify the scale set (or remaining VM singularly, if you can) to use the new uploaded data disk.
6. Reboot, wait a few minutes.
7. Pray it workded ;)

## How I did it once (Azure)

1. If the VM has never been rebooted once after initialization, reboot it once to sync any LUKS passphrase changes to disk (not 100% sure if necessary to sync the change to the passphrase - would have to double-check that later with an experimental cluster)

2. Create a snapshot from the disk using the CLI:
```bash
az snapshot create --resource-group dogfooding --name dogfooding-3 --source /subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/dogfooding/providers/Microsoft.Compute/disks/constell-f2332c74-coconstell-f2332c74-condisk2_dd460a6ae3124aa3a4c23be0ab39634e --location northeurope
```

3. Look up the snapshot online, export it as VHD
Mount the disk: 
```bash
modprobe nbd && sudo qemu-nbd -c /dev/nbd0 /home/nils/Downloads/constellation-disk.vhd
```

4. Get the UUID of the disk: 
```bash 
sudo cryptsetup luksDump /dev/nbd0
```

5. Regenerate the passphrase to unlock the disk (the [code snippet below](#get-disk-decryption-key) might be useful for this)

6. Decrypt the disk: 
```bash
sudo cryptsetup luksOpen /dev/nbd0 constellation-state --key-file passphrase
```

7. Mount the decrypted disk (I just did this via the Nautilus)

8. Find the db file from etcd in `/var/lib/etcd/member/snap/db`

9. Perform the etcd [Restoring a Cluster](https://etcd.io/docs/v3.5/op-guide/recovery/#restoring-a-cluster) step:

```bash
./etcdutl snapshot restore db --initial-cluster constell-f2332c74-control-plane000001=https://10.9.126.0:2380 --initial-advertise-peer-urls https://10.9.126.0:2380  --data-dir recovery --name constell-f2332c74-control-plane000001 --skip-hash-check=true
```
*(replace name & IP with the name & the private IP of the remaining control plane VM - this can be seen in the Azure portal)*

10. Copy the contents of the newly created recovery directory to the mounted state disk and remove any remaining old files. 
**Make sure the permissions are the same as before!**

11. Unmount the partition:
```bash 
sudo umount /your/mount/path
sudo luksClose constellation-state
sudo qemu-nbd -d /dev/nbd0
```

12. Upload the modified VHD back to Azure (I just used Azure Storage Explorer for this).

13. Patch the whole control-plane VMSS to remove LUN 0 from the VMs: 
```bash
az vmss disk detach --lun 0 --resource-group dogfooding --vmss-name constell-f2332c74-control-plane
```

14. Update the VM: 
```bash
az vmss update-instances -g dogfooding --name constell-f2332c74-control-plane --instance-ids 1
```

15. Attach the uploaded disk as LUN 0 (either via CLI or Azure Portal, I just used the Azure Portal).

16. Start the VM and pray it works ;) It can take a few minutes before etcd becomes fully alive again.

17. Patch the state disk definition back to the VMSS (no idea how, haven't done his yet) so newly created VMs in the VMSS have a clean state disk again.

## Get disk decryption key
```golang
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/hkdf"
)

type MasterSecret struct {
	Key  []byte `json:"key"`
	Salt []byte `json:"salt"`
}

func main() {	
	uuid := "4ae66293-57aa-4821-b99c-ebc598a6e5a8" // replace me

	masterSecretRaw, err := os.ReadFile("constellation-mastersecret.json")
	if err != nil {
		panic(err)
	}

	var masterSecret MasterSecret
	if err := json.Unmarshal(masterSecretRaw, &masterSecret); err != nil {
		panic(err)
	}

	dek, err := DeriveKey(masterSecret.Key, masterSecret.Salt, []byte("key-"+uuid), 32)
	if err != nil {
		panic(err)
	}

	fmt.Println(hex.EncodeToString(dek))

	if err := os.WriteFile("passphrase", dek, 0o644); err != nil {
		panic(err)
	}
}

// DeriveKey derives a key from a secret.
func DeriveKey(secret, salt, info []byte, length uint) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, err
	}
	return key, nil
}
```
