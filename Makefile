PLATFORMS := windows/386/.exe windows/amd64/.exe

DEP := $(shell command -v dep 2> /dev/null)
GO := $(shell command -v go 2> /dev/null)

# magical formula:
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
ext = $(word 3, $(temp))

all: check-env install-dependencies $(PLATFORMS) package

check-env:
	@echo "==> Checking prerequisites"
	@echo -n "Checking for go ... "
ifndef GO
	@echo "Not Found"
	$(error "go is unavailable")
endif
	@echo $(GO)
	@echo -n "Checking for dep ... "
ifndef DEP
	@echo "Not Found"
	$(error "dep is unavailable")
endif
	@echo $(DEP)
	@echo ""

clean:
	@echo "==> Clearing previous build data"
	@rm -rf build/ || true
	@$(GO) clean -cache

install-dependencies:
	@echo "==> Installing dependencies"
	@$(DEP) ensure -v
	@echo ""

$(PLATFORMS):
	@echo "==> Building ${os} (${arch}) executable"
	@export GOOS=$(os); export GOARCH=$(arch); $(GO) build -o build/$(os)-$(arch)/soundboard$(ext)

package:
	@echo "==> Creating distribution packages"
	@for dir in build/*; do if [ -d "$$dir" ]; then tar -czvf "$(basename "$$dir").tar.gz" --xform="s,$$dir/,," "$$dir"; fi; done
	@echo ""

.PHONY: all
