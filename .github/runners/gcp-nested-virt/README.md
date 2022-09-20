This folder contains a template for deploying a builder for CoreOS on GCP.

## Manually start a builder instance
```
gcloud compute instances create coreos-builder --enable-nested-virtualization --zone=us-central1-c --boot-disk-size 64GB --machine-type=n2-highmem-4 --image-project="ubuntu-os-cloud" --image="ubuntu-2110-impish-v20220118" --metadata-from-file=user-data=cloud-init.txt
```
