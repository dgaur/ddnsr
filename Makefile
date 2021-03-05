
DDNSR := ddnsr

GO ?= go


all: $(DDNSR)

$(DDNSR): ddnsr.go dns.go
	@$(GO) build


.PHONY: clean
clean:
	@rm -f $(DDNSR)
	@$(GO) clean


.PHONY: test
test:
	@$(GO) test -v


.PHONY: vet
vet:
	@$(GO) vet
	
