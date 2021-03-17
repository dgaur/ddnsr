#
# Makefile for the ddnsr build
#

# Assume the 'go' tools are available in the PATH if not explicitly overridden
GO ?= go

# The CLI tool is the only artifact
DDNSR := ddnsr

all: $(DDNSR)

$(DDNSR): ddnsr.go dns.go
	@$(GO) build


.PHONY: clean
clean:
	@rm -f $(DDNSR)
	@$(GO) clean


.PHONY: test
test:
	@$(GO) test -v -cover


.PHONY: vet
vet:
	@$(GO) vet
	
