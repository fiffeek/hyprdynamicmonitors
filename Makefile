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

export DIST_DIR
export PATH := $(DIST_DIR):$(PATH)

.PHONY: install test fmt lint release/local record/previews

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
	$(INSTALL_DIR)/.asdf.stamp \
	$(INSTALL_DIR)/.venv.stamp \
	$(INSTALL_DIR)/.npm.stamp \
	$(INSTALL_DIR)/.precommit.stamp \
	$(INSTALL_DIR)/.toc.stamp

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

$(INSTALL_DIR)/.npm.stamp: $(PACKAGE_LOCK) $(INSTALL_DIR)/.asdf.stamp
	@$(NPM_BIN) install
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
	@$(GOLANG_BIN) test ./internal/tui -v -run $(TUI_FLOWS)

test/unit:
	@$(GOLANGCI_LINT_BIN) config verify
	@$(GORELEASER_BIN) check
	@$(GOLANG_BIN) test ./internal/... -v

test/unit/selected:
	$(GOLANG_BIN) test ./internal/$(PACKAGE_SELECTOR) -v -run $(TEST_SELECTOR)

build/test:
	@mkdir -p ./dist/
	@$(GOLANG_BIN) build -v -o $(TEST_EXECUTABLE_NAME) ./main.go

test/integration: build/test
	@HDM_BINARY_PATH=$(TEST_EXECUTABLE_NAME) $(GOLANG_BIN) test -v ./test/... --debug

test/integration/regenerate: build/test
	@HDM_BINARY_PATH=$(TEST_EXECUTABLE_NAME) $(GOLANG_BIN) test -v ./test/... --regenerate

test/integration/selected: build/test
	@HDM_BINARY_PATH=$(TEST_EXECUTABLE_NAME) $(GOLANG_BIN) test -v -run $(TEST_SELECTOR) ./test/... --debug

test: test/unit test/integration

fmt:
	@$(GOLANG_BIN) mod tidy
	@env GOFUMPT_SPLIT_LONG_LINES=on $(GOLANGCI_LINT_BIN) fmt ./...

lint:
	@$(GOLANGCI_LINT_BIN) run

pre-push: fmt lint test/unit test/integration

help/generate: build/test
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME)
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) run
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) validate
	@scripts/autohelp.sh $(TEST_EXECUTABLE_NAME) freeze

# requires vhs to be installed, for now a manual action
record/previews: build/test
	@git checkout -- ./preview/tapes/configs/
	@git clean -fd ./preview/tapes/configs/
	@$(VHS_BIN) ./preview/tapes/$(RECORD_TARGET).tape
