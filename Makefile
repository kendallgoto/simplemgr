golangci-lint ?= ~/go/bin/golangci-lint

BUILD_DIR ?= bin

project_prefix := github.com/kendallgoto/simplemgr
target_prefix := $(project_prefix)/cmd/
targets := $(sort $(notdir $(wildcard cmd/*)))

VERSION ?= $(shell git describe --always)
go_ldflag := github.com/kendallgoto/simplemgr/pkg/globals.Version=$(VERSION)

go_build_flags := -trimpath
go_build_flags += -ldflags "-X $(go_ldflag)"
go_build_flags += -o "$(BUILD_DIR)/"

.PHONY: build
build: $(targets)

.PHONY: $(targets)
$(targets):
	go build $(go_build_flags) -o ./bin/$@$(if $(GOOS),-$(GOOS),)$(if $(GOARCH),-$(GOARCH),) $(addprefix $(target_prefix),$@)

.PHONY: lint
lint:
	$(golangci-lint) run
