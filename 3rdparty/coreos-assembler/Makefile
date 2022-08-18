CONTAINER_RUNTIME ?= docker
IMAGENAME         ?= localhost/coreos-assembler
IMAGETAG          ?= latest
UPSTREAM          ?= https://github.com/coreos/coreos-assembler
COMMIT            ?= 2dd33ddc36e6b9e1cc01ee0740f29020d203ceb2
SRC_PATH           = $(CURDIR)
BASE_PATH         ?= $(SRC_PATH)
BUILDDIR          ?= $(BASE_PATH)/build
CLONEDIR          ?= $(BUILDDIR)/coreos-assembler


.PHONY: all clone patch containerimage clean

all: clone patch containerimage

clone:
	@mkdir -p $(BUILDDIR)
	git clone $(UPSTREAM) $(CLONEDIR)
	cd $(CLONEDIR) && git checkout $(COMMIT)

patch:
	cd $(CLONEDIR) && patch --verbose -p1 < $(BASE_PATH)/verity.patch

containerimage:
	cd $(CLONEDIR) && $(CONTAINER_RUNTIME) build -t $(IMAGENAME):$(IMAGETAG) -f Dockerfile .

clean:
	rm -rf $(BUILDDIR)
