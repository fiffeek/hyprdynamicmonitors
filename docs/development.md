---
layout: default
title: Development
---

## Development

### Setup Development Environment

Set up the complete development environment with all dependencies:

```bash
make dev
```

This installs:
- asdf version manager with required tool versions
- Go toolchain and dependencies
- Python virtual environment for pre-commit hooks
- Node.js dependencies for commit linting
- Pre-commit hooks configuration
- Documentation generation tools

### Development Commands

**Code quality and testing:**
```bash
make fmt          # Format code and tidy modules
make lint         # Run linting checks
make test         # Run all tests (unit + integration)
make pre-push     # Run complete CI pipeline (fmt + lint + test)
```

**Testing specific areas:**
```bash
make test/unit                    # Run only unit tests
make test/integration             # Run only integration tests
make test/integration/regenerate  # Regenerate test fixtures
```

**Running selected tests** (runs with `-debug` for log output):
```bash
# Run subset of integration tests
make TEST_SELECTOR=Test__Run_Binary/power_events_triggers test/integration/selected

# Run subset of unit tests
make TEST_SELECTOR="TestIPC_Run/happy_path$" PACKAGE_SELECTOR=hypr/... test/unit/selected
```

**Building:**
```bash
make release/local    # Build release binaries for all platforms
make build/test       # Build test binary for integration tests
```

**Documentation:**
```bash
make help/generate    # Generate help documentation from binary
```

### Development Workflow

1. **Initial setup**: `make dev` (one-time setup)
2. **Development cycle**: Make changes, then run `make pre-push` before committing
3. **Testing**: Use `make test` for full test suite, or specific test targets for focused testing
4. **Pre-commit hooks**: Automatically run on commit (installed by `make dev`)

### Release Candidates

Release candidates are published for testing new features before stable releases:

- **GitHub Releases**: RC versions are marked as pre-releases on GitHub (e.g., `v0.2.0-rc1`)
- **AUR Package**: Available as separate `hyprdynamicmonitors-rc-bin` package alongside the stable `hyprdynamicmonitors-bin`
- **Binary Name**: RC builds use `hyprdynamicmonitors-rc` to avoid conflicts with stable installations
- **Parallel Installation**: Both stable and RC versions can be installed simultaneously for testing

To install the RC version from AUR:
```bash
yay -S hyprdynamicmonitors-rc-bin
```

## Tests

### Live Testing

Live tested on:
- Hyprland v0.50.1
- UPower v1.90.9

You can see my configuration [here](https://github.com/fiffeek/.dotfiles.v2/blob/main/ansible/files/framework/dots/hyprdynamicmonitors/config.toml).

### Integration Testing
All features should be covered by integration tests. Run `make test/integration` locally to execute end-to-end CLI tests that build the binary and verify expected outputs. Test fixtures can be regenerated using `make test/integration/regenerate`.

