---
sidebar_position: 8
---

# Development

This guide covers setting up a development environment and contributing to HyprDynamicMonitors.

## Setup Development Environment

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

## Development Commands

To see all available commands, run:

```bash
make help
```

### Code Quality and Testing

```bash
make fmt          # Format code and tidy modules
make lint         # Run linting checks
make test         # Run all tests (unit + integration)
make pre-push     # Run complete CI pipeline (fmt + lint + test)
```

### Testing Specific Areas

```bash
make test/unit                    # Run only unit tests
make test/integration             # Run only integration tests
make test/integration/regenerate  # Regenerate test fixtures
make test/tui/flows               # Run TUI flow tests
make test/tui/flows/regenerate    # Regenerate TUI flow test fixtures
make coverage                     # Generate coverage report (coverage.html)
```

### Running Selected Tests

Runs with `-debug` for log output:

```bash
# Run subset of integration tests
make TEST_SELECTOR=Test__Run_Binary/power_events_triggers test/integration/selected

# Run subset of unit tests
make TEST_SELECTOR="TestIPC_Run/happy_path$" PACKAGE_SELECTOR=hypr/... test/unit/selected

# Regenerate specific test fixtures
make TEST_SELECTOR="TestIPC_Run" PACKAGE_SELECTOR=hypr/... test/unit/selected/regenerate
```

### Building

```bash
make release/local       # Build release binaries for all platforms
make release/local/rc    # Build release candidate binaries
make build/test          # Build test binary with coverage for integration tests
make build/docs          # Build binary for documentation generation
```

### Documentation

```bash
make docs             # Start documentation development server
make help/generate    # Generate help documentation from binary
make toc/generate     # Generate table of contents for markdown files
```

### Recording Demos and Previews

```bash
make demo                              # Record demo preview (requires vhs)
make record/preview RECORD_TARGET=demo # Record specific preview tape
make record/previews                   # Record all preview tapes
```

### Nix Support (Experimental)

```bash
make nix/build/module        # Build NixOS module VM
make nix/build/homemanager   # Build home-manager VM
make nix/run/module          # Run NixOS module VM
make nix/run/homemanager     # Run home-manager VM
```

## Development Workflow

1. **Initial setup**: `make dev` (one-time setup)
2. **Development cycle**: Make changes, then run `make pre-push` before committing
3. **Testing**: Use `make test` for full test suite, or specific test targets for focused testing
4. **Pre-commit hooks**: Automatically run on commit (installed by `make dev`)

## Project Structure

```
hyprdynamicmonitors/
├── cmd/                      # Command-line interface
├── internal/                 # Internal packages
│   ├── app/                 # Application core
│   ├── config/              # Configuration handling
│   ├── dial/                # Connection dialing utilities
│   ├── filewatcher/         # File watching for config changes
│   ├── generators/          # Code/config generators
│   ├── hypr/                # Hyprland IPC integration
│   ├── matchers/            # Monitor matching logic
│   ├── notifications/       # Desktop notifications
│   ├── power/               # Power event handling
│   ├── profilemaker/        # Profile creation and management
│   ├── reloader/            # Configuration reloading
│   ├── signal/              # Signal handling
│   ├── testutils/           # Testing utilities
│   ├── tui/                 # Terminal UI (Bubble Tea)
│   ├── userconfigupdater/   # User configuration updates
│   └── utils/               # Shared utilities
├── examples/                 # Example configurations
│   ├── basic/               # Basic configuration examples
│   ├── callbacks/           # Callback examples
│   ├── disable-monitors/    # Monitor disabling examples
│   ├── fallback/            # Fallback configuration
│   ├── full/                # Complete configuration example
│   ├── lid-states/          # Laptop lid state handling
│   ├── minimal/             # Minimal configuration
│   ├── power-states/        # Power state examples
│   ├── scoring/             # Monitor scoring examples
│   └── template-variables/  # Template variable usage
├── test/                     # Integration tests
│   └── testdata/            # Test fixtures and data
├── docs/                     # Docusaurus documentation
│   ├── docs/                # Documentation content
│   ├── src/                 # Documentation site source
│   ├── static/              # Static assets
│   └── versioned_docs/      # Versioned documentation
├── scripts/                  # Build and utility scripts
├── infrastructure/           # Infrastructure configs
│   └── systemd/             # Systemd service files
├── nix/                      # Nix configuration
│   ├── homemanager/         # Home Manager integration
│   └── module/              # NixOS module
├── preview/                  # Preview/demo recordings
│   ├── output/              # Generated preview output
│   └── tapes/               # VHS tape files for recordings
└── bin/                      # Development binaries
```

## Testing

### Unit Tests

Located within the relevant packages:

```bash
make test/unit
```

### Integration Tests

Located in `test/` directory. These build the binary and verify expected outputs:

```bash
make test/integration
```

### Test Fixtures

Integration tests use fixtures that can be regenerated:

```bash
make test/integration/regenerate
```

## Contributing

### Commit Messages

The project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add support for custom monitor tags
fix: resolve template rendering issue
docs: update TUI documentation
test: add integration test for power events
chore: update dependencies
```

Pre-commit hooks will validate commit message format.

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run `make pre-push` to ensure all checks pass
5. Commit with conventional commit messages
6. Push to your fork: `git push origin feat/my-feature`
7. Create a pull request

### Code Style

- Use `gofmt` for formatting (automatically applied by `make fmt`)
- Follow Go best practices
- Add tests for new functionality
- Update documentation as needed

## Release Process

### Release Candidates

Release candidates are published for testing new features:

- **GitHub Releases**: RC versions are marked as pre-releases (e.g., `v0.2.0-rc1`)
- **AUR Package**: Available as `hyprdynamicmonitors-rc-bin` alongside stable `hyprdynamicmonitors-bin`
- **Binary Name**: RC builds use `hyprdynamicmonitors-rc` to avoid conflicts
- **Parallel Installation**: Both stable and RC versions can be installed simultaneously

To install the RC version from AUR:

```bash
yay -S hyprdynamicmonitors-rc-bin
```

### Versioning

The project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality in a backwards compatible manner
- **PATCH**: Backwards compatible bug fixes

## Build System

The project uses Make for build automation. Run `make help` to see all available targets.

### Key Targets

```bash
make help             # Display all available targets and usage
make dev              # Setup development environment
make install          # Install to DESTDIR (default: ~/.local/bin)
make uninstall        # Uninstall from DESTDIR
make test             # Run all tests
make lint             # Run linters
make fmt              # Format code
make pre-push         # Run complete CI pipeline
```

### Environment Variables

- `DESTDIR`: Installation directory (default: `~/.local/bin`)
- `TEST_SELECTOR`: Pattern to select specific tests
- `PACKAGE_SELECTOR`: Package pattern for unit tests (default: `...`)
- `RECORD_TARGET`: Preview tape to record (default: `demo`)
- `TUI_FLOWS`: TUI flow test to run (default: `TestModel_Update_UserFlows`)
- `VHS_BIN`: Path to vhs binary for recording previews (default: `vhs`)

### Examples

```bash
# Run specific unit tests
make TEST_SELECTOR='TestIPC_Run/happy_path$' PACKAGE_SELECTOR=hypr/... test/unit/selected

# Run specific integration tests
make TEST_SELECTOR=Test__Run_Binary/power_events test/integration/selected

# Record a specific preview
make record/preview RECORD_TARGET=demo
```

## Debugging

### Enable Debug Logging

```bash
hyprdynamicmonitors --debug run
```

### Enable Verbose Logging

```bash
hyprdynamicmonitors --verbose run
```

### Structured Logging

```bash
hyprdynamicmonitors --enable-json-logs-format run
```

### Dry Run Mode

Test configuration without applying changes:

```bash
hyprdynamicmonitors run --dry-run
```

## Resources

- [GitHub Repository](https://github.com/fiffeek/hyprdynamicmonitors)
- [Issue Tracker](https://github.com/fiffeek/hyprdynamicmonitors/issues)
- [Examples](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples)

## See Also

- [Installation](./quickstart/installation) - Build from source
- [Examples](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples) - Example configurations
