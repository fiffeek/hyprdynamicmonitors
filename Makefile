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

.PHONY: install test fmt lint release/local

release/local: \
	$(INSTALL_DIR)/.dir.stamp \
	$(INSTALL_DIR)/.asdf.stamp
	@$(GORELEASER_BIN) release --snapshot --clean

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
	$(INSTALL_DIR)/.precommit.stamp

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

test:
	@$(GOLANGCI_LINT_BIN) config verify
	@$(GORELEASER_BIN) check
	@$(GOLANG_BIN) test ./... -v

fmt:
	@$(GOLANG_BIN) fmt ./...
	@$(GOLANG_BIN) mod tidy
	@env GOFUMPT_SPLIT_LONG_LINES=on $(GOLANGCI_LINT_BIN) fmt

lint:
	@$(GOLANGCI_LINT_BIN) run
