INSTALL_DIR := install
PYTHON_BIN := python
PRECOMMIT_FILE := .pre-commit-config.yaml
VENV := venv
REQUIREMENTS_FILE := requirements.txt
GOLANGCI_LINT_BIN := golangci-lint
GOLANG_BIN := go
GORELEASER_BIN := goreleaser

install: $(INSTALL_DIR)/.dir.stamp $(INSTALL_DIR)/.asdf.stamp $(INSTALL_DIR)/.venv.stamp $(INSTALL_DIR)/.precommit.stamp

$(INSTALL_DIR)/.dir.stamp:
	mkdir -p $(INSTALL_DIR)
	touch $@

$(INSTALL_DIR)/.asdf.stamp:
	asdf install
	touch $@

$(INSTALL_DIR)/.venv.stamp: $(REQUIREMENTS_FILE) $(INSTALL_DIR)/.asdf.stamp
	test -d "$(VENV)" || $(PYTHON_BIN) -m venv "$(VENV)"
	. "$(VENV)/bin/activate"; \
		pip install --upgrade pip; \
		pip install -r "$(REQUIREMENTS_FILE)"
	touch $@

$(INSTALL_DIR)/.precommit.stamp: $(PRECOMMIT_FILE) $(INSTALL_DIR)/.venv.stamp
	. "$(VENV)/bin/activate"; pre-commit install
	touch $@

test:
	$(GOLANGCI_LINT_BIN) config verify
	$(GORELEASER_BIN) check
	$(GOLANG_BIN) test ./... -v

fmt:
	$(GOLANG_BIN) fmt ./...
	$(GOLANG_BIN) mod tidy
	env GOFUMPT_SPLIT_LONG_LINES=on $(GOLANGCI_LINT_BIN) fmt

lint:
	$(GOLANGCI_LINT_BIN) run

release/local:
	$(GORELEASER_BIN) release --snapshot --clean
