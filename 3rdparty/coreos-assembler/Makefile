CONTAINER_RUNTIME ?= docker
IMAGENAME         ?= localhost/coreos-assembler
IMAGETAG          ?= latest
UPSTREAM          ?= https://github.com/coreos/coreos-assembler
COMMIT            ?= b8f61207c7a4ce345a1f136823dac6c628086273
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
	cd $(CLONEDIR) && patch -p1 < $(BASE_PATH)/verity.patch

containerimage:
	cd $(CLONEDIR) && $(CONTAINER_RUNTIME) build -t $(IMAGENAME):$(IMAGETAG) -f Dockerfile .

clean:
	rm -rf $(BUILDDIR)
