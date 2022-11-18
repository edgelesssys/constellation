# PCR-updater

New images result in different PCR values for the image.
This utility program makes it simple to update the expected PCR values of the CLI.

## Usage

To read the PCR state of any running Constellation node, run the following:

```shell
go run main.go -constell-ip <NODE_IP> -constell-port <VERIFY_SERVICE_PORT>
```

The output is similar to the following:

```shell
$ go run main.go -constell-ip 192.0.2.3 -constell-port 30081
connecting to verification service at 192.0.2.3:30081
PCRs:
{
  "0": "DzXCFGCNk8em5ornNZtKi+Wg6Z7qkQfs5CfE3qTkOc8=",
  "1": "XBoRlWuQx6nIDr5vgUL0DlJHy6H6u1dPU3qK2NyToc8=",
  "10": "WLmYFRmDft/ajZJ056CAhpheU6Vbt73aR8eIQpLRGq0=",
  "11": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "12": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "13": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "14": "4tPyJd6A5g09KduV3+nWZQCiEzHAiRT5DulmAqlvpZU=",
  "15": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "16": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "17": "//////////////////////////////////////////8=",
  "18": "//////////////////////////////////////////8=",
  "19": "//////////////////////////////////////////8=",
  "2": "PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=",
  "20": "//////////////////////////////////////////8=",
  "21": "//////////////////////////////////////////8=",
  "22": "//////////////////////////////////////////8=",
  "23": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "3": "PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=",
  "4": "MmkueFj1rP2seH+bjeIRsO4dUnLnMdl7QgtGoAtQH7M=",
  "5": "ExaiapuIfo0KMBo8wj6kPDORLocgnH1C0G/KY8DcV3A=",
  "6": "PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=",
  "7": "UZcW+fhFRMpFkgU+EfKG2s3KdmgEA+TD2quLmthQHbo=",
  "8": "KLSMootYaHBjysWKq9CAYXkXpeYx9PUBimlSEZGJqUM=",
  "9": "gse53SjsqREEdOpImJH4KAb0b8PqIgwI+Ps/XSiFnN4="
}
```

### Extend Config

To set measurement values in Constellation config, use `yaml` format option.
Optionally filter down results measurements per cloud provider:

Azure

```bash
./pcr-reader --constell-ip ${CONSTELLATION_IP} --format yaml | yq e 'del(.[0,6,10,11,12,13,14,15,16,17,18,19,20,21,22,23])' -
```

## Meaning of PCR values

An overview about what data is measured into the different registers can be found [in the TPM spec](https://trustedcomputinggroup.org/wp-content/uploads/TCG_PCClient_PFP_r1p05_v23_pub.pdf#%5B%7B%22num%22%3A157%2C%22gen%22%3A0%7D%2C%7B%22name%22%3A%22XYZ%22%7D%2C33%2C400%2C0%5D).

We use the TPM and its PCRs to verify all nodes of a Constellation run with the same firmware and OS software.

### Azure

PCR[0] measures the firmware volume (FV). Changes to FV also change PCR[0], making it unreliable for attestation.
PCR[6] measures the VM ID. This is unusable for cluster attestation for two reasons:

1. The verification service does not know the VM ID of nodes wanting to join the cluster, so it can not compute the expected PCR[6] for the joining VM
2. A user may attest any node of the cluster without knowing the VM ID

PCR[10] is used by Linux Integrity Measurement Architecture (IMA).
IMA creates runtime measurements based on a measurement policy (which is obsolete for Constellation, since we use dm-verity).
The first entry of the runtime measurements is the `boot_aggregate`. It is a SHA1 hash over PCRs 0 to 7.
As detailed earlier, PCR[6] is different for every VM in Azure, therefore PCR[10] will also be different since it includes PCR[6], meaning we can not use it for attestation.
IMA writing its measurements into PCR[10] can not be disabled without rebuilding the kernel.

### Azure flexible deployment and attestation (FDA)

With FDA CVMs measuring all of the firmware, it should be possible to use all PCRs for attestation since we know, and can choose, what firmware is running.

### GCP confidential VM

GCP uses confidential VMs based on AMD SEV-ES with a vTPM interface.

PCR[0] contains the measurement of a string marking the VM as using ADM SEV-ES.
All firmware measurements seem to be constant.
