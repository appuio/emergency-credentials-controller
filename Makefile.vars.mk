IMG_TAG ?= latest

CURDIR ?= $(shell pwd)
BIN_FILENAME ?= $(CURDIR)/$(PROJECT_ROOT_DIR)/emergency-credentials-controller

# Image URL to use all building/pushing image targets
GHCR_IMG ?= ghcr.io/appuio/emergency-credentials-controller:$(IMG_TAG)
