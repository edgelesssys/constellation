SHELL                             = /bin/bash
SRC_PATH                          = $(CURDIR)
BASE_PATH                        ?= $(SRC_PATH)
MKOSI_CACHE                      ?= $(BASE_PATH)/mkosi.cache
REPODIR                          ?= $(BASE_PATH)/repo
ORAS_IMAGE                       ?= ghcr.io/edgelesssys/constellation-rpms
EXTRA_SEARCH_PATHS               ?=
MKOSI                            ?= mkosi

.PHONY: all clean update repo testrepo

all: repo

update: clean-image clean-cache clean-repo
	$(MKOSI) build
	@if [ -n $(SUDO_UID) ] && [ -n $(SUDO_GID) ]; then \
		chown -R $(SUDO_UID):$(SUDO_GID) image.*; \
	fi
	find $(MKOSI_CACHE) -name "*.rpm" -exec cp {} $(REPODIR)/ \;
	cd $(REPODIR) && sha256sum *.rpm > $(REPODIR)/SHA256SUMS

push:
	cd $(REPODIR) && sha256sum -c SHA256SUMS
	sha256sum $(REPODIR)/SHA256SUMS | cut -d' ' -f1 > $(SRC_PATH)/TAG
	cd $(REPODIR) && oras push "$(ORAS_IMAGE):$$(cat $(SRC_PATH)/TAG)" *.rpm

pull: clean-repo
	cd $(REPODIR) && oras pull "$(ORAS_IMAGE):$$(cat $(SRC_PATH)/TAG)"
	cd $(REPODIR) && sha256sum -c SHA256SUMS
	@if [ -n $(SUDO_UID) ] && [ -n $(SUDO_GID) ]; then \
		chown -R $(SUDO_UID):$(SUDO_GID) image.*; \
	fi

repo:
	rm -rf $(REPODIR)/repodata
	cd $(REPODIR) && sha256sum -c SHA256SUMS
	createrepo_c $(REPODIR)
	@if [ -n $(SUDO_UID) ] && [ -n $(SUDO_GID) ]; then \
		chown -R $(SUDO_UID):$(SUDO_GID) $(REPODIR); \
	fi

testrepo: clean-image clean-cache
	cd $(REPODIR) && sha256sum -c SHA256SUMS
	$(MKOSI) build --local-mirror $(REPODIR)
	@if [ -n $(SUDO_UID) ] && [ -n $(SUDO_GID) ]; then \
		chown -R $(SUDO_UID):$(SUDO_GID) image.*; \
	fi

check:
	cd $(REPODIR) && sha256sum -c SHA256SUMS

clean-cache:
	-find $(MKOSI_CACHE) \! -name "mkosi.cache" \! -name ".gitkeep" -delete

clean-repo:
	-find $(REPODIR) \! -name "repo" \! -name ".gitignore" \! -name "SHA256SUMS" -delete

clean-image:
	-$(MKOSI) clean

clean: clean-cache clean-repo clean-image
