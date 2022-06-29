SRC_PATH                          = $(CURDIR)
BASE_PATH                        ?= $(SRC_PATH)
DEPENDENCIES                     ?= $(BASE_PATH)/dependencies
COREOS_BUILD_PATH                ?= $(BASE_PATH)/build
OVERRIDES                        ?= $(COREOS_BUILD_PATH)/overrides
OVERRIDES_RPM                    ?= $(OVERRIDES)/rpm
OVERRIDES_ROOTFS                 ?= $(OVERRIDES)/rootfs
BOOTSTRAPPER_OVERRIDE_PATH        ?= $(OVERRIDES_ROOTFS)/usr/bin/bootstrapper
CILIUM_CLI_OVERRIDE_PATH         ?= $(OVERRIDES_ROOTFS)/usr/bin/cilium
KERNEL_RPM                       ?= $(OVERRIDES_RPM)/kernel.rpm
KERNEL_CORE_RPM                  ?= $(OVERRIDES_RPM)/kernel-core.rpm
KERNEL_MODULES_RPM               ?= $(OVERRIDES_RPM)/kernel-modules.rpm
OVERRIDE_KERNEL                  ?= n
KERNEL_RPM_URL                   ?= "https://kojipkgs.fedoraproject.org/packages/kernel/5.17.0/0.rc3.89.fc36/x86_64/kernel-5.17.0-0.rc3.89.fc36.x86_64.rpm"
KERNEL_CORE_RPM_URL              ?= "https://kojipkgs.fedoraproject.org/packages/kernel/5.17.0/0.rc3.89.fc36/x86_64/kernel-core-5.17.0-0.rc3.89.fc36.x86_64.rpm"
KERNEL_MODULES_RPM_URL           ?= "https://kojipkgs.fedoraproject.org/packages/kernel/5.17.0/0.rc3.89.fc36/x86_64/kernel-modules-5.17.0-0.rc3.89.fc36.x86_64.rpm"
DOWNLOAD_BOOTSTRAPPER             ?= n
BOOTSTRAPPER_BINARY               ?= $(BASE_PATH)/../build/bootstrapper
CILIUM_CLI_BINARY                ?= $(DEPENDENCIES)/cilium
CILIUM_CLI_URL                   ?= "https://github.com/cilium/cilium-cli/releases/download/v0.11.1/cilium-linux-amd64.tar.gz"
DISK_MAPPER_BINARY               ?= $(BASE_PATH)/../build/disk-mapper
DISK_MAPPER_OVERRIDE_PATH        ?= $(OVERRIDES_ROOTFS)/usr/sbin/disk-mapper
IMAGES_PATH                      ?= $(BASE_PATH)/images
CONTAINER_ENGINE                 ?= podman
COSA_ENV                         ?= $(BASE_PATH)/$(CONTAINER_ENGINE).env
COSA_INIT_REPO                   ?= "https://github.com/edgelesssys/constellation-fedora-coreos-config"
COSA_INIT_BRANCH                 ?= "constellation"
COSA_LOCKFILE                    ?= $(BASE_PATH)/cosa.lock
NETRC                            ?= ~/.netrc
GCP_PROJECT                      ?= constellation-images
GCP_BUCKET                       ?= constellation-images
GCP_IMAGE_FAMILY                 ?= constellation
GCP_IMAGE_NAME                   ?= constellation-coreos
GCP_IMAGE_FILENAME               ?= $(GCP_IMAGE_NAME)_gcp.tar.gz
GCP_IMAGE_PATH                   ?= $(IMAGES_PATH)/$(GCP_IMAGE_FILENAME)
GCP_REGION                       ?= europe-west3
AZURE_IMAGE_NAME                 ?= constellation-coreos
AZURE_IMAGE_FILENAME             ?= $(AZURE_IMAGE_NAME)_azure.vhd
AZURE_IMAGE_PATH                 ?= $(IMAGES_PATH)/$(AZURE_IMAGE_FILENAME)
AZURE_RESOURCE_GROUP_NAME        ?= CONSTELLATION-IMAGES
AZURE_REGION                     ?= northeurope
AZURE_GALLERY_NAME               ?= Constellation
AZURE_IMAGE_DEFINITION           ?= $(AZURE_IMAGE_NAME)
AZURE_IMAGE_OFFER                ?= $(AZURE_IMAGE_DEFINITION)
AZURE_IMAGE_VERSION              ?= 0.0.1
AZURE_PUBLISHER                  ?= edgelesssys
AZURE_SKU                        ?= constellation-coreos

-include $(CURDIR)/config.mk

.PHONY: clean all kernel coreos run shell cosa-init cosa-fetch images image-gcp upload-gcp image-azure upload-azure-non-cvm $(BOOTSTRAPPER_OVERRIDE_PATH) $(DISK_MAPPER_OVERRIDE_PATH) $(CILIUM_CLI_OVERRIDE_PATH)
SHELL := /bin/bash

all: coreos images

ifeq ($(DOWNLOAD_BOOTSTRAPPER), y)
$(BOOTSTRAPPER_BINARY):
	wget -q $(BOOTSTRAPPER_URL) -O $@
endif

$(BOOTSTRAPPER_OVERRIDE_PATH): $(BOOTSTRAPPER_BINARY) | cosa-init
	mkdir -p $(COREOS_BUILD_PATH)/overrides/rootfs/usr/bin
	cp $(BOOTSTRAPPER_BINARY) $@
	chmod +x $@

$(CILIUM_CLI_BINARY):
	wget -q $(CILIUM_CLI_URL) -O cilium.tar.gz
	tar xzfC cilium.tar.gz $(DEPENDENCIES)
	rm cilium.tar.gz

$(CILIUM_CLI_OVERRIDE_PATH): $(CILIUM_CLI_BINARY) | cosa-init
	mkdir -p $(COREOS_BUILD_PATH)/overrides/rootfs/usr/bin
	cp $(CILIUM_CLI_BINARY) $@
	chmod +x $@

$(DISK_MAPPER_OVERRIDE_PATH): $(DISK_MAPPER_BINARY) | cosa-init
	mkdir -p $(COREOS_BUILD_PATH)/overrides/rootfs/usr/sbin
	cp $(DISK_MAPPER_BINARY) $@
	chmod +x $@

cosa-init:
	-flock $(COSA_LOCKFILE) -c '. $(COSA_ENV) && cd $(COREOS_BUILD_PATH) && NETRC=$(NETRC) cosa init --branch $(COSA_INIT_BRANCH) $(COSA_INIT_REPO)'

cosa-fetch: cosa-init kernel $(BOOTSTRAPPER_OVERRIDE_PATH) $(DISK_MAPPER_OVERRIDE_PATH) $(CILIUM_CLI_OVERRIDE_PATH)
	flock $(COSA_LOCKFILE) -c '. $(COSA_ENV) && cd $(COREOS_BUILD_PATH) && NETRC=$(NETRC) cosa fetch'

coreos: cosa-fetch $(BOOTSTRAPPER_OVERRIDE_PATH) $(DISK_MAPPER_OVERRIDE_PATH) $(CILIUM_CLI_OVERRIDE_PATH)
	flock $(COSA_LOCKFILE) -c '. $(COSA_ENV) && cd $(COREOS_BUILD_PATH) && NETRC=$(NETRC) cosa build'

run: $(BOOTSTRAPPER_OVERRIDE_PATH) $(DISK_MAPPER_OVERRIDE_PATH) $(CILIUM_CLI_OVERRIDE_PATH)
	flock $(COSA_LOCKFILE) -c '. $(COSA_ENV) && cd $(COREOS_BUILD_PATH) && NETRC=$(NETRC) cosa run --devshell-console'

shell: $(BOOTSTRAPPER_OVERRIDE_PATH) $(DISK_MAPPER_OVERRIDE_PATH) $(CILIUM_CLI_OVERRIDE_PATH)
	flock $(COSA_LOCKFILE) -c '. $(COSA_ENV) && cd $(COREOS_BUILD_PATH) && NETRC=$(NETRC) cosa shell'

$(GCP_IMAGE_PATH): coreos
	flock $(COSA_LOCKFILE) -c '. $(COSA_ENV) && cd $(COREOS_BUILD_PATH) && NETRC=$(NETRC) cosa buildextend-gcp'
	cp $(COREOS_BUILD_PATH)/builds/latest/x86_64/fedora-coreos-*.dev.*-gcp.x86_64.tar.gz $@
	$(BASE_PATH)/dm-verity/recalculate-dm-verity.sh gcp $@

$(AZURE_IMAGE_PATH): coreos
	flock $(COSA_LOCKFILE) -c '. $(COSA_ENV) && cd $(COREOS_BUILD_PATH) && NETRC=$(NETRC) cosa buildextend-azure'
	cp $(COREOS_BUILD_PATH)/builds/latest/x86_64/fedora-coreos-*.dev.*-azure.x86_64.vhd $@
	$(BASE_PATH)/dm-verity/recalculate-dm-verity.sh azure $@

images: image-gcp image-azure

image-gcp: $(GCP_IMAGE_PATH)

upload-gcp: $(GCP_IMAGE_PATH)
	-gsutil mb -l $(GCP_REGION) gs://$(GCP_BUCKET)
	gsutil pap set enforced gs://$(GCP_BUCKET)
	gsutil cp $(GCP_IMAGE_PATH) gs://$(GCP_BUCKET)
	gcloud compute images create $(GCP_IMAGE_NAME) \
		--family $(GCP_IMAGE_FAMILY) \
		--source-uri=gs://$(GCP_BUCKET)/$(GCP_IMAGE_FILENAME) \
		--guest-os-features=GVNIC,SEV_CAPABLE,VIRTIO_SCSI_MULTIQUEUE,UEFI_COMPATIBLE \
		--labels=bootstrapper-sha1=$$(shasum $(BOOTSTRAPPER_OVERRIDE_PATH) | cut -d " " -f 1),bootstrapper-sha512=$$(sha512sum $(BOOTSTRAPPER_OVERRIDE_PATH) | cut -d " " -f 1 | cut -c-63) \
		--project $(GCP_PROJECT)
	gsutil rm gs://$(GCP_BUCKET)/$(GCP_IMAGE_FILENAME)

image-azure: $(AZURE_IMAGE_PATH)

# reference: https://docs.microsoft.com/en-us/azure/virtual-machines/linux/disks-upload-vhd-to-managed-disk-cli
# reference: https://docs.microsoft.com/en-us/azure/virtual-machines/linux/upload-vhd
# reference: https://docs.microsoft.com/en-us/azure/virtual-machines/image-version
upload-azure: $(AZURE_IMAGE_PATH)
	$(eval SIZE := $(shell cat $(AZURE_IMAGE_PATH) | wc -c))
	@echo "creating disk of size $(SIZE)"
	az disk create -n $(AZURE_IMAGE_NAME) -g $(AZURE_RESOURCE_GROUP_NAME) -l $(AZURE_REGION) --hyper-v-generation V2 --os-type Linux --for-upload --upload-size-bytes $(SIZE) --sku standard_lrs --tags bootstrapper-sha512=$$(sha512sum $(BOOTSTRAPPER_OVERRIDE_PATH) | cut -d " " -f 1)
	@echo "waiting for disk to be created"
	az disk wait --created -n $(AZURE_IMAGE_NAME) -g $(AZURE_RESOURCE_GROUP_NAME)
	@echo "created disk $(AZURE_IMAGE_NAME)"
	az disk list --output table --query "[?name == '$(AZURE_IMAGE_NAME)' && resourceGroup == '$(AZURE_RESOURCE_GROUP_NAME)']"
	@echo "Granting temporary write access on disk $(AZURE_IMAGE_NAME)"
	@echo "Copying $(AZURE_IMAGE_PATH) to disk $(AZURE_IMAGE_NAME)"
	azcopy copy "$(AZURE_IMAGE_PATH)" "$$(az disk grant-access -n $(AZURE_IMAGE_NAME) -g $(AZURE_RESOURCE_GROUP_NAME) --access-level Write --duration-in-seconds 86400 | jq -r .accessSas)" --blob-type PageBlob
	@echo "Revoking write access"
	az disk revoke-access -n $(AZURE_IMAGE_NAME) -g $(AZURE_RESOURCE_GROUP_NAME)
	az image create -g $(AZURE_RESOURCE_GROUP_NAME) -l $(AZURE_REGION) -n $(AZURE_IMAGE_NAME) --hyper-v-generation V2 --os-type Linux --tags bootstrapper-sha512=$$(sha512sum $(BOOTSTRAPPER_OVERRIDE_PATH) | cut -d " " -f 1) --source "$$(az disk list --query "[?name == '$(AZURE_IMAGE_NAME)' && resourceGroup == '$(AZURE_RESOURCE_GROUP_NAME)'] | [0].id" --output json | jq -r)"
	@echo "Create shared image gallery (if it does not exist yet)"
	-az sig create -l $(AZURE_REGION) --gallery-name $(AZURE_GALLERY_NAME) --resource-group $(AZURE_RESOURCE_GROUP_NAME)
	@echo "Create image definition (if it does not exist yet)"
	-az sig image-definition create --resource-group $(AZURE_RESOURCE_GROUP_NAME) -l $(AZURE_REGION) --gallery-name $(AZURE_GALLERY_NAME) --gallery-image-definition $(AZURE_IMAGE_DEFINITION) --publisher $(AZURE_PUBLISHER) --offer $(AZURE_IMAGE_OFFER) --sku $(AZURE_SKU) --os-type Linux --os-state generalized --hyper-v-generation V2 --features SecurityType=TrustedLaunch
	az sig image-version create --resource-group $(AZURE_RESOURCE_GROUP_NAME) -l $(AZURE_REGION) --gallery-name $(AZURE_GALLERY_NAME) --gallery-image-definition $(AZURE_IMAGE_DEFINITION) --gallery-image-version $(AZURE_IMAGE_VERSION) --target-regions $(AZURE_REGION) --replica-count 1 --tags bootstrapper-sha512=$$(sha512sum $(BOOTSTRAPPER_OVERRIDE_PATH) | cut -d " " -f 1) --managed-image "$$(az image list --query "[?name == '$(AZURE_IMAGE_NAME)' && resourceGroup == '$(AZURE_RESOURCE_GROUP_NAME)'] | [0].id" --output json | jq -r)"
	@echo "Cleaning up resources"
	az image delete --resource-group $(AZURE_RESOURCE_GROUP_NAME) -n $(AZURE_IMAGE_NAME)
	az disk delete --resource-group $(AZURE_RESOURCE_GROUP_NAME) -n $(AZURE_IMAGE_NAME) --yes

ifeq ($(OVERRIDE_KERNEL), y)
kernel: $(KERNEL_RPM) $(KERNEL_CORE_RPM) $(KERNEL_MODULES_RPM)
$(KERNEL_RPM): | cosa-init
	-mkdir -p $$(dirname $@)
	wget -q $(KERNEL_RPM_URL) -O $@
$(KERNEL_CORE_RPM): | cosa-init
	-mkdir -p $$(dirname $@)
	wget -q $(KERNEL_CORE_RPM_URL) -O $@
$(KERNEL_MODULES_RPM): | cosa-init
	-mkdir -p $$(dirname $@)
	wget -q $(KERNEL_MODULES_RPM_URL) -O $@
else
kernel:
endif

clean:
	-rm $(COSA_LOCKFILE)
	-rm $(IMAGES_PATH)/*
	-rm -r $(COREOS_BUILD_PATH)
	mkdir -p $(COREOS_BUILD_PATH)
