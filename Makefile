GO = go

all:

release:
	@GOOS=linux GOARCH=amd64
	@echo "Building $(GOOS) $(GOARCH) release"
	@$(GO) clean
	@$(GO) build
	@tar czvf rehook-`git describe`.tar.gz rehook views/ public/

.PHONY: release
