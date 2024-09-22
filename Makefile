BINDIR      := $(CURDIR)/bin
BINNAME     ?= what

# ------------------------------------------------------------------------------
#  build

.PHONY: build
build:
	go build -o '$(BINDIR)'/$(BINNAME) ./cmd/what
