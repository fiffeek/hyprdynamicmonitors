INSTALL_DIR := install
PYTHON_BIN := python
PRECOMMIT_FILE := .pre-commit-config.yaml
VENV := venv
REQUIREMENTS_FILE := requirements.txt

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
