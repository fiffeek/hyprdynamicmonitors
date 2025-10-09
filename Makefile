INSTALL_DIR := install
PYTHON_BIN := python
PRECOMMIT_FILE := .pre-commit-config.yaml
VENV := venv
REQUIREMENTS_FILE := requirements.txt
PACKAGE_LOCK := package-lock.json
COMMITLINT_FILE := commitlint.config.js
NPM_BIN := npm
GOLANGCI_LINT_BIN := golangci-lint
GOLANG_BIN := go
GORELEASER_BIN := goreleaser
DESTDIR ?= $(HOME)/.local/bin
EXECUTABLE_NAME := hyprdynamicmonitors
TEST_EXECUTABLE_NAME := ./dist/hdmtest
GH_MD_TOC_FILE := "https://raw.githubusercontent.com/ekalinin/github-markdown-toc/master/gh-md-toc"
DEV_BINARIES_DIR := "./bin"
GH_MD_TOC_BIN := "gh-md-toc"
TEST_SELECTOR ?=
PACKAGE_SELECTOR ?= "..."
TUI_FLOWS ?= "TestModel_Update_UserFlows"
VHS_BIN ?= vhs
DIST_DIR := $(abspath dist)
RECORD_TARGET ?= demo
GOTESTSUM := $(GOLANG_BIN) tool gotestsum
GOTESTSUMINTEGRATION  := $(GOTESTSUM)
DOCS_COMMAND_FILE := "docs/docs/usage/commands.md"

ifeq ($(GITHUB_ACTIONS),true)
GOTESTSUM := $(GOTESTSUM) --format=github-actions
GOTESTSUMINTEGRATION := $(GOTESTSUM) --format=github-actions
else
GOTESTSUM := $(GOTESTSUM) --format=pkgname-and-test-fails
GOTESTSUMINTEGRATION := $(GOTESTSUM) --format=testname
endif

export DIST_DIR
export PATH := $(DIST_DIR):$(PATH)

.PHONY: install test fmt lint release/local record/previews docs

release/local: \
	$(INSTALL_DIR)/.dir.stamp \
	$(INSTALL_DIR)/.asdf.stamp
	@$(GORELEASER_BIN) release --snapshot --clean

release/local/rc: \
	$(INSTALL_DIR)/.dir.stamp \
	$(INSTALL_DIR)/.asdf.stamp
	@$(GORELEASER_BIN) release --snapshot --clean --config .goreleaser.rc.yaml

install:
	@mkdir -p "$(DESTDIR)"
	@if [ "$$(uname -m)" = "x86_64" ]; then \
		cp dist/hyprdynamicmonitors_linux_amd64_v1/$(EXECUTABLE_NAME) "$(DESTDIR)/"; \
	elif [ "$$(uname -m)" = "aarch64" ]; then \
		cp dist/hyprdynamicmonitors_linux_arm64_v8.0/$(EXECUTABLE_NAME) "$(DESTDIR)/"; \
	elif [ "$$(uname -m)" = "i686" ] || [ "$$(uname -m)" = "i386" ]; then \
		cp dist/hyprdynamicmonitors_linux_386_sse2/$(EXECUTABLE_NAME) "$(DESTDIR)/"; \
	else \
		echo "Unsupported architecture: $$(uname -m)"; \
		exit 1; \
	fi
	@echo "Installed $(EXECUTABLE_NAME) to $(DESTDIR)"

uninstall:
	@rm "$(DESTDIR)/$(EXECUTABLE_NAME)"
	@echo "Uninstalled $(EXECUTABLE_NAME) from $(DESTDIR)"

dev: \
	$(INSTALL_DIR)/.dir.stamp \
	$(INSTALL_DIR)/.git.stamp \
	$(INSTALL_DIR)/.asdf.stamp \
	$(INSTALL_DIR)/.venv.stamp \
	$(INSTALL_DIR)/.npm.stamp \
	$(INSTALL_DIR)/.precommit.stamp \
	$(INSTALL_DIR)/.toc.stamp

$(INSTALL_DIR)/.git.stamp: $(INSTALL_DIR)/.dir.stamp
	@git lfs install
	@touch $@

$(INSTALL_DIR)/.toc.stamp: $(INSTALL_DIR)/.dir.stamp
	@mkdir -p $(DEV_BINARIES_DIR)
	@wget -q $(GH_MD_TOC_FILE)
	@chmod 755 $(GH_MD_TOC_BIN)
	@mv $(GH_MD_TOC_BIN) $(DEV_BINARIES_DIR)
	@$(DEV_BINARIES_DIR)/$(GH_MD_TOC_BIN) --help >/dev/null
	@touch $@

$(INSTALL_DIR)/.dir.stamp:
	@mkdir -p $(INSTALL_DIR)
	@touch $@

$(INSTALL_DIR)/.asdf.stamp:
	@asdf install
	@touch $@

$(INSTALL_DIR)/.npm.stamp: $(PACKAGE_LOCK) ./docs/$(PACKAGE_LOCK) $(INSTALL_DIR)/.asdf.stamp
	@$(NPM_BIN) install
	@cd ./docs && $(NPM_BIN) install
	@touch $@

$(INSTALL_DIR)/.venv.stamp: $(REQUIREMENTS_FILE) $(INSTALL_DIR)/.asdf.stamp
	@test -d "$(VENV)" || $(PYTHON_BIN) -m venv "$(VENV)"
	. "$(VENV)/bin/activate"; \
		pip install --upgrade pip; \
		pip install -r "$(REQUIREMENTS_FILE)"
	@touch $@

$(INSTALL_DIR)/.precommit.stamp: $(PRECOMMIT_FILE) $(INSTALL_DIR)/.venv.stamp
	@. "$(VENV)/bin/activate"; pre-commit install && \
		pre-commit install --hook-type commit-msg
	@touch $@

test/tui/flows/regenerate:
	@$(GOLANG_BIN) test ./internal/tui -v -run $(TUI_FLOWS) -update

test/tui/flows:
	@$(GOTESTSUM) -- ./internal/tui -v -run $(TUI_FLOWS)

test/unit:
	@$(GOLANGCI_LINT_BIN) config verify
	@$(GORELEASER_BIN) check
	@$(GOTESTSUM) -- ./internal/... -v -coverprofile=unit.txt -covermode count

test/unit/selected:
	@$(GOTESTSUM) -- ./internal/$(PACKAGE_SELECTOR) -v -run $(TEST_SELECTOR)

test/unit/selected/regenerate:
	$(GOLANG_BIN) test ./internal/$(PACKAGE_SELECTOR) -v  --regenerate -run $(TEST_SELECTOR)

build/docs:
	@mkdir -p ./dist/
	@$(GOLANG_BIN) build -v -o $(TEST_EXECUTABLE_NAME) ./main.go

build/test:
	@mkdir -p ./dist/
	@$(GOLANG_BIN) build -v -cover -covermode count -o $(TEST_EXECUTABLE_NAME) ./main.go

test/integration: build/test
	@rm -rf .coverdata
	@mkdir -p .coverdata
	@HDM_BINARY_PATH=$(TEST_EXECUTABLE_NAME) $(GOTESTSUMINTEGRATION) --rerun-fails=2 --packages="./test/..." -- -v ./test/... --debug --count=1
	@$(GOLANG_BIN) tool covdata textfmt -i=.coverdata -o=integration.txt

test/integration/regenerate: build/test
	@HDM_BINARY_PATH=$(TEST_EXECUTABLE_NAME) $(GOLANG_BIN) test -v ./test/... --regenerate

test/integration/selected: build/test
	@HDM_BINARY_PATH=$(TEST_EXECUTABLE_NAME) $(GOLANG_BIN) test -v -run $(TEST_SELECTOR) ./test/... --debug

coverage:
	@$(GOLANG_BIN) tool gocovmerge integration.txt unit.txt > coverage.txt
	@grep -v "/testutils/" coverage.txt > coverage.txt.tmp
	@mv coverage.txt.tmp coverage.txt
	@$(GOLANG_BIN) tool cover -html=coverage.txt -o coverage.html
	@$(GOLANG_BIN) tool gocover-cobertura <coverage.txt >coverage.xml

test: test/unit test/integration

fmt:
	@$(GOLANG_BIN) mod tidy
	@env GOFUMPT_SPLIT_LONG_LINES=on $(GOLANGCI_LINT_BIN) fmt ./...

lint:
	@$(GOLANGCI_LINT_BIN) run

pre-push: fmt lint test/unit test/integration

toc/generate:
	@scripts/autotoc.sh

help/generate: build/docs
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) $(DOCS_COMMAND_FILE)
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) $(DOCS_COMMAND_FILE) run
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) $(DOCS_COMMAND_FILE) validate
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) $(DOCS_COMMAND_FILE) freeze
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) $(DOCS_COMMAND_FILE) tui

# requires vhs to be installed, for now a manual action
record/preview: build/docs
	@git checkout -- ./preview/tapes/configs/
	@git clean -fd ./preview/tapes/configs/
	@$(VHS_BIN) ./preview/tapes/$(RECORD_TARGET).tape
	@git checkout -- ./preview/tapes/configs/
	@git clean -fd ./preview/tapes/configs/

demo: record/preview

record/previews: build/docs
	$(MAKE) record/preview RECORD_TARGET=views
	$(MAKE) record/preview RECORD_TARGET=monitor_view
	$(MAKE) record/preview RECORD_TARGET=panning
	$(MAKE) record/preview RECORD_TARGET=zoom
	$(MAKE) record/preview RECORD_TARGET=display_options
	$(MAKE) record/preview RECORD_TARGET=editing
	$(MAKE) record/preview RECORD_TARGET=position
	$(MAKE) record/preview RECORD_TARGET=rotation
	$(MAKE) record/preview RECORD_TARGET=resolution
	$(MAKE) record/preview RECORD_TARGET=scaling
	$(MAKE) record/preview RECORD_TARGET=mirroring
	$(MAKE) record/preview RECORD_TARGET=disable
	$(MAKE) record/preview RECORD_TARGET=vrr
	$(MAKE) record/preview RECORD_TARGET=apply
	$(MAKE) record/preview RECORD_TARGET=create_profile
	$(MAKE) record/preview RECORD_TARGET=edit_existing
	$(MAKE) record/preview RECORD_TARGET=color

docs:
	@cd ./docs && npm run start

# The following nix target require nix locally, just for testing the flake, not in the CI yet
nix/build:
	@cd nix/$(NIX_DIRECTORY) && nix --extra-experimental-features "nix-command flakes" run nixpkgs#nixos-rebuild -- build-vm --flake .#hypr-vm

nix/build/module:
	@$(MAKE) nix/build NIX_DIRECTORY=module

nix/build/homemanager:
	@$(MAKE) nix/build NIX_DIRECTORY=homemanager

nix/run:
	@echo "login: demo"
	@echo "password: demo"
	@cd ./nix/$(NIX_DIRECTORY) && ./result/bin/run-hypr-vm-vm

nix/run/module:
	@$(MAKE) nix/run NIX_DIRECTORY=module

nix/run/homemanager:
	@$(MAKE) nix/run NIX_DIRECTORY=homemanager
