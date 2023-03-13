SHELL                             = /bin/bash
SRC_PATH                          = $(CURDIR)
BASE_PATH                        ?= $(SRC_PATH)
BOOTSTRAPPER_BINARY              ?= $(BASE_PATH)/../build/bootstrapper
DISK_MAPPER_BINARY               ?= $(BASE_PATH)/../build/disk-mapper
UPGRADE_AGENT_BINARY             ?= $(BASE_PATH)/../build/upgrade-agent
DEBUGD_BINARY                    ?= $(BASE_PATH)/../build/debugd
MEASUREMENT_READER_BINARY        ?= $(BASE_PATH)/../build/measurement-reader
PKI                              ?= $(BASE_PATH)/pki
MKOSI_EXTRA                      ?= $(BASE_PATH)/mkosi.extra
IMAGE_VERSION                    ?= v0.0.0
DEBUG                            ?= false
AUTOLOGIN                        ?= false
AUTOLOGIN_ARGS                   := $(if $(filter true,$(AUTOLOGIN)),--autologin) # set "--autologin" if AUTOLOGIN is true
KERNEL_DEBUG_CMDLNE              := $(if $(filter true,$(DEBUG)),constellation.debug) # set "constellation.debug" if DEBUG is true
export INSTALL_DEBUGD            ?= $(DEBUG)
export CONSOLE_MOTD = $(AUTOLOGIN)
-include $(CURDIR)/config.mk
csps := aws azure gcp openstack qemu
certs := $(PKI)/PK.cer $(PKI)/KEK.cer $(PKI)/db.cer

SYSTEMD_FIXED_RPMS := systemd-251.11-2.fc37.x86_64.rpm systemd-libs-251.11-2.fc37.x86_64.rpm systemd-networkd-251.11-2.fc37.x86_64.rpm systemd-pam-251.11-2.fc37.x86_64.rpm systemd-resolved-251.11-2.fc37.x86_64.rpm systemd-udev-251.11-2.fc37.x86_64.rpm
AZURE_FIXED_KERNEL_RPMS := kernel-6.1.18-200.fc37.x86_64.rpm kernel-core-6.1.18-200.fc37.x86_64.rpm kernel-modules-6.1.18-200.fc37.x86_64.rpm
GCP_FIXED_KERNEL_RPMS := kernel-6.1.18-200.fc37.x86_64.rpm kernel-core-6.1.18-200.fc37.x86_64.rpm kernel-modules-6.1.18-200.fc37.x86_64.rpm
PREBUILD_RPMS_SYSTEMD := $(addprefix prebuilt/rpms/systemd/,$(SYSTEMD_FIXED_RPMS))
PREBUILT_RPMS_AZURE := $(addprefix prebuilt/rpms/azure/,$(AZURE_FIXED_KERNEL_RPMS))
PREBUILT_RPMS_GCP := $(addprefix prebuilt/rpms/gcp/,$(GCP_FIXED_KERNEL_RPMS))

.PHONY: all clean inject-bins $(csps)

all: $(csps)

$(csps): %: mkosi.output.%/fedora~37/image.raw

prebuilt/rpms/systemd/%.rpm:
	@echo "Downloading $*"
	@mkdir -p $(@D)
	@curl -fsSL -o $@ https://kojipkgs.fedoraproject.org/packages/systemd/251.11/2.fc37/x86_64/$*.rpm

# Currently, Azure and GCP use the same fixed kernels.
# They will likely derive soon again, but for now we can just copy the file from Azure to save traffic.
prebuilt/rpms/gcp/%.rpm: prebuilt/rpms/azure/%.rpm
	# @echo "Downloading $*"
	# @mkdir -p $(@D)
	# @curl -fsSL -o $@ https://kojipkgs.fedoraproject.org/packages/kernel/6.1.18/200.fc37/x86_64/$*.rpm
	cp prebuilt/rpms/azure/$*.rpm prebuilt/rpms/gcp/$*.rpm

prebuilt/rpms/azure/%.rpm:
	@echo "Downloading $*"
	@mkdir -p $(@D)
	@curl -fsSL -o $@ https://kojipkgs.fedoraproject.org/packages/kernel/6.1.18/200.fc37/x86_64/$*.rpm

mkosi.output.%/fedora~37/image.raw: mkosi.files/mkosi.%.conf inject-bins inject-certs
	mkosi --config mkosi.files/mkosi.$*.conf \
		--image-version=$(IMAGE_VERSION) \
		$(AUTOLOGIN_ARGS) \
		--environment=INSTALL_DEBUGD \
		--environment=CONSOLE_MOTD \
		--kernel-command-line="$(KERNEL_DEBUG_CMDLNE)" \
		build
	secure-boot/signed-shim.sh $@
	@if [ -n $(SUDO_UID) ] && [ -n $(SUDO_GID) ]; then \
		chown -R $(SUDO_UID):$(SUDO_GID) mkosi.output.$*; \
	fi
	@echo "Image is ready: $@"

inject-bins: $(PREBUILD_RPMS_SYSTEMD) $(PREBUILT_RPMS_AZURE) $(PREBUILT_RPMS_GCP)
	mkdir -p $(MKOSI_EXTRA)/usr/bin
	mkdir -p $(MKOSI_EXTRA)/usr/sbin
	cp $(UPGRADE_AGENT_BINARY) $(MKOSI_EXTRA)/usr/bin/upgrade-agent
	cp $(DISK_MAPPER_BINARY) $(MKOSI_EXTRA)/usr/sbin/disk-mapper
	cp $(MEASUREMENT_READER_BINARY) $(MKOSI_EXTRA)/usr/sbin/measurement-reader
	if [ "$(DEBUG)" = "true" ]; then \
		cp $(DEBUGD_BINARY) $(MKOSI_EXTRA)/usr/bin/debugd; \
		rm -f $(MKOSI_EXTRA)/usr/bin/bootstrapper; \
		rm -f $(MKOSI_EXTRA)/usr/bin/upgrade-agent; \
	else \
		cp $(BOOTSTRAPPER_BINARY) $(MKOSI_EXTRA)/usr/bin/bootstrapper; \
		rm -f $(MKOSI_EXTRA)/usr/bin/debugd; \
	fi

inject-certs: $(certs)
	# for auto enrollment using systemd-boot (not working yet)
	mkdir -p "$(MKOSI_EXTRA)/boot/loader/keys/auto"
	cp $(PKI)/{PK,KEK,db}.cer "$(MKOSI_EXTRA)/boot/loader/keys/auto"
	cp $(PKI)/{MicWinProPCA2011_2011-10-19,MicCorUEFCA2011_2011-06-27,MicCorKEKCA2011_2011-06-24}.crt "$(MKOSI_EXTRA)/boot/loader/keys/auto"
	cp $(PKI)/{PK,KEK,db}.esl "$(MKOSI_EXTRA)/boot/loader/keys/auto"
	cp $(PKI)/{PK,KEK,db}.auth "$(MKOSI_EXTRA)/boot/loader/keys/auto"
	# for manual enrollment using sbkeysync
	mkdir -p $(MKOSI_EXTRA)/etc/secureboot/keys/{db,dbx,KEK,PK}
	cp $(PKI)/db.auth "$(MKOSI_EXTRA)/etc/secureboot/keys/db/"
	cp $(PKI)/KEK.auth "$(MKOSI_EXTRA)/etc/secureboot/keys/KEK/"
	cp $(PKI)/PK.auth "$(MKOSI_EXTRA)/etc/secureboot/keys/PK/"

clean-cache:
	rm -rf mkosi.cache/*

clean-%:
	mkosi --config mkosi.files/mkosi.$*.conf clean

clean:
	rm -rf mkosi.output.*
	rm -rf prebuilt/rpms
	rm -rf $(MKOSI_EXTRA)
	mkdir -p $(MKOSI_EXTRA)
