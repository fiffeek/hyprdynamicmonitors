---
sidebar_position: 1
---

# CLI Commands

HyprDynamicMonitors provides several commands for managing monitor configurations.

## Global Flags

<!-- START help -->
```text
HyprDynamicMonitors is a service that automatically switches between predefined Hyprland monitor configuration profiles based on connected monitors and power state.

Usage:
  hyprdynamicmonitors [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  freeze      Freeze current monitor configuration as a new profile template
  help        Help about any command
  run         Run the monitor configuration service
  tui         Launch interactive TUI for monitor configuration
  validate    Validate configuration file

Flags:
      --config string             Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
      --debug                     Enable debug logging
      --enable-json-logs-format   Enable structured logging
  -h, --help                      help for hyprdynamicmonitors
      --verbose                   Enable verbose logging
  -v, --version                   version for hyprdynamicmonitors

Use "hyprdynamicmonitors [command] --help" for more information about a command.
```
<!-- END help -->

## run

Run the HyprDynamicMonitors service to continuously monitor for display changes and automatically apply matching configuration profiles.

```bash
hyprdynamicmonitors run [flags]
```

### Flags

<!-- START runhelp -->
```text
Run the HyprDynamicMonitors service to continuously monitor for display changes and automatically apply matching configuration profiles.

Usage:
  hyprdynamicmonitors run [flags]

Flags:
      --connect-to-session-bus    Connect to session bus instead of system bus for power events: https://wiki.archlinux.org/title/D-Bus. You can switch as long as you expose power line events in your user session bus.
      --disable-auto-hot-reload   Disable automatic hot reload (no file watchers)
      --disable-power-events      Disable power events (dbus). Defaults to true if running on desktop, to false otherwise
      --dry-run                   Show what would be done without making changes
      --enable-lid-events         Enable listening to dbus lid events
  -h, --help                      help for run
      --run-once                  Run once and exit immediately

Global Flags:
      --config string             Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
      --debug                     Enable debug logging
      --enable-json-logs-format   Enable structured logging
      --verbose                   Enable verbose logging
```
<!-- END runhelp -->

### Examples

```bash
# Run the service normally
hyprdynamicmonitors run

# Run with dry-run to see what would happen
hyprdynamicmonitors run --dry-run

# Run once and exit (useful for testing)
hyprdynamicmonitors run --run-once

# Disable power events
hyprdynamicmonitors run --disable-power-events

# Enable lid events
hyprdynamicmonitors run --enable-lid-events
```

## validate

Validate the configuration file for syntax errors and logical consistency.

```bash
hyprdynamicmonitors validate [flags]
```

### Flags

<!-- START validatehelp -->
```text
Validate the configuration file for syntax errors and logical consistency.

Usage:
  hyprdynamicmonitors validate [flags]

Flags:
  -h, --help   help for validate

Global Flags:
      --config string             Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
      --debug                     Enable debug logging
      --enable-json-logs-format   Enable structured logging
      --verbose                   Enable verbose logging
```
<!-- END validatehelp -->


### Examples

```bash
# Validate default config file
hyprdynamicmonitors validate

# Validate specific config file
hyprdynamicmonitors --config /path/to/config.toml validate

# Validate with debug output
hyprdynamicmonitors --debug validate
```

## freeze

Freeze the current Hyprland monitor configuration and save it as a new profile template.

This command captures your current monitor setup and creates two artifacts:
1. A Go template file containing the Hyprland configuration
2. A new profile entry in your configuration file that references this template

```bash
hyprdynamicmonitors freeze [flags]
```

### Flags

<!-- START freezehelp -->
```text
Freeze the current Hyprland monitor configuration and save it as a new profile template.

This command captures your current monitor setup and creates two artifacts:
1. A Go template file containing the Hyprland configuration
2. A new profile entry in your configuration file that references this template

TEMPLATE FILE:
The Go template will be saved to hyprconfigs/{profile-name}.go.tmpl by default, or to a
custom location specified with --config-file-location. This template can be edited after
creation to customize the configuration.

PROFILE ENTRY:
A new profile with the specified name will be appended to your configuration file. The
profile will automatically require monitors by description (not name) to ensure better
portability across different systems.

PREREQUISITES:
- The profile name must not already exist in your configuration (it will be checked)
- The template file location must not exist (it will be created)
- Hyprland must be running with a valid monitor configuration

This is useful for quickly creating new profiles based on your current working setup.

Usage:
  hyprdynamicmonitors freeze [flags]

Flags:
      --config-file-location string   Where to put the generated config file template (defaults to hyprconfigs/$PROFILE_NAME.go.tmpl)
  -h, --help                          help for freeze
      --profile-name string           What profile name to set the frozen profile to.

Global Flags:
      --config string             Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
      --debug                     Enable debug logging
      --enable-json-logs-format   Enable structured logging
      --verbose                   Enable verbose logging
```
<!-- END freezehelp -->

### Prerequisites

- The profile name must not already exist in your configuration
- The template file location must not exist (it will be created)
- Hyprland must be running with a valid monitor configuration

### Examples

```bash
# Freeze current setup as a new profile
hyprdynamicmonitors freeze --profile-name dual-monitors

# Freeze with custom template location
hyprdynamicmonitors freeze --profile-name triple-4k \
  --config-file-location ~/my-configs/triple.go.tmpl
```

## tui

Launch an interactive terminal-based TUI for managing monitor configurations.

```bash
hyprdynamicmonitors tui [flags]
```

### Flags

<!-- START tuihelp -->
```text
Launch an interactive terminal-based TUI for managing monitor configurations.

Usage:
  hyprdynamicmonitors tui [flags]

Flags:
      --connect-to-session-bus          Connect to session bus instead of system bus for power events: https://wiki.archlinux.org/title/D-Bus. You can switch as long as you expose power line events in your user session bus.
      --disable-power-events            Disable power events (dbus). Defaults to true if running on desktop, to false otherwise
      --enable-lid-events               Enable listening to dbus lid events
  -h, --help                            help for tui
      --hypr-monitors-override string   When used it fill parse the given file as hyprland monitors spec, used for testing.
      --running-under-test              Use test settings such as no styling etc.

Global Flags:
      --config string             Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
      --debug                     Enable debug logging
      --enable-json-logs-format   Enable structured logging
      --verbose                   Enable verbose logging
```
<!-- END tuihelp -->

### Compatibility

The TUI has the same flags as `run` and can be used without the running daemon for ad-hoc changes. When the config is not passed or invalid, you will be unable to persist the configuration in the `hyprdynamicmonitors` config, but you can still experiment with monitors and apply the Hyprland configuration.

Refer to the [TUI Guide](../quickstart/tui) for more details.

### Examples

```bash
# Launch TUI normally
hyprdynamicmonitors tui

# Launch TUI with lid events enabled
hyprdynamicmonitors tui --enable-lid-events

# Launch TUI for testing with a mock monitors file
hyprdynamicmonitors tui --hypr-monitors-override /path/to/monitors-spec.txt

# Launch TUI for testing with a mock monitors file and custom config
hyprdynamicmonitors tui --hypr-monitors-override /path/to/monitors-spec.txt --config /path/to/config.toml
```

## completion

Generate autocompletion scripts for various shells.

```bash
hyprdynamicmonitors completion [bash|zsh|fish|powershell]
```

### Examples

```bash
# Bash
hyprdynamicmonitors completion bash > /etc/bash_completion.d/hyprdynamicmonitors

# Zsh
hyprdynamicmonitors completion zsh > "${fpath[1]}/_hyprdynamicmonitors"

# Fish
hyprdynamicmonitors completion fish > ~/.config/fish/completions/hyprdynamicmonitors.fish
```
