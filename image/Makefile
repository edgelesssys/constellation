SHELL                             = /bin/bash
SRC_PATH                          = $(CURDIR)
BASE_PATH                        ?= $(SRC_PATH)
BOOTSTRAPPER_BINARY              ?= $(BASE_PATH)/../build/bootstrapper
DISK_MAPPER_BINARY               ?= $(BASE_PATH)/../build/disk-mapper
PKI                              ?= $(BASE_PATH)/pki
MKOSI_EXTRA                      ?= $(BASE_PATH)/mkosi.extra
-include $(CURDIR)/config.mk
csps := qemu gcp azure
certs := $(PKI)/PK.cer $(PKI)/KEK.cer $(PKI)/db.cer

.PHONY: all clean inject-bins $(csps)

all: $(csps)

$(csps): %: mkosi.output.%/fedora~36/image.raw

mkosi.output.%/fedora~36/image.raw: mkosi.files/mkosi.%.conf inject-bins inject-certs
	mkosi --config mkosi.files/mkosi.$*.conf build
	secure-boot/signed-shim.sh $@
	@if [ -n $(SUDO_UID) ] && [ -n $(SUDO_GID) ]; then \
		chown -R $(SUDO_UID):$(SUDO_GID) mkosi.output.$*; \
	fi
	@echo "Image is ready: $@"

inject-bins:
	mkdir -p $(MKOSI_EXTRA)/usr/bin
	mkdir -p $(MKOSI_EXTRA)/usr/sbin
	cp $(BOOTSTRAPPER_BINARY) $(MKOSI_EXTRA)/usr/bin/bootstrapper
	cp $(DISK_MAPPER_BINARY) $(MKOSI_EXTRA)/usr/sbin/disk-mapper

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
	rm -rf $(MKOSI_EXTRA)
	mkdir -p $(MKOSI_EXTRA)
