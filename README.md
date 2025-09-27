<img src="https://github.com/user-attachments/assets/0effc242-3d3d-4d39-a183-0a567c4da3a9" width="90" style="margin-right:10px" align=left alt="hyprdynamicmonitors logo">
<H1>HyprDynamicMonitors</H1><br>


An event-driven service that automatically manages Hyprland monitor configurations based on connected displays and power state.

## Documentation

<!--ts-->
* [HyprDynamicMonitors](#hyprdynamicmonitors)
   * [Documentation](#documentation)
   * [Features](#features)
   * [Design Philosophy](#design-philosophy)
      * [Reliability Through Restarts](#reliability-through-restarts)
      * [Hot Reloading With Graceful Restart](#hot-reloading-with-graceful-restart)
      * [Hyprland-Native Integration](#hyprland-native-integration)
   * [Installation](#installation)
      * [Binary Release](#binary-release)
      * [AUR](#aur)
      * [Nix](#nix)
      * [Build from Source](#build-from-source)
   * [Usage](#usage)
      * [Command Line](#command-line)
      * [Run command](#run-command)
      * [Validate command](#validate-command)
      * [Freeze command](#freeze-command)
   * [Minimal Example](#minimal-example)
   * [Examples](#examples)
   * [Runtime requirements](#runtime-requirements)
   * [Configuration](#configuration)
      * [Monitor Matching](#monitor-matching)
      * [Configuration File Types](#configuration-file-types)
      * [Template Variables](#template-variables)
      * [Static Template Values](#static-template-values)
      * [Template Functions](#template-functions)
      * [Fallback Profile](#fallback-profile)
      * [User Callbacks (Exec Commands)](#user-callbacks-exec-commands)
      * [Notifications](#notifications)
      * [Hyprland Integration](#hyprland-integration)
      * [Power Events](#power-events)
         * [Disabling power events](#disabling-power-events)
         * [Default power event configuration](#default-power-event-configuration)
         * [Querying](#querying)
         * [Receive Filters](#receive-filters)
         * [Custom D-Bus Configuration](#custom-d-bus-configuration)
         * [Leave Empty Token](#leave-empty-token)
      * [Signals](#signals)
      * [Hot Reloading](#hot-reloading)
   * [Tests](#tests)
      * [Live Testing](#live-testing)
      * [Integration Testing](#integration-testing)
   * [Running with systemd](#running-with-systemd)
      * [Hyprland under systemd](#hyprland-under-systemd)
      * [Run on boot and let restarts do the job](#run-on-boot-and-let-restarts-do-the-job)
      * [Custom systemd target](#custom-systemd-target)
      * [Alternative: Wrapper script](#alternative-wrapper-script)
   * [Development](#development)
      * [Setup Development Environment](#setup-development-environment)
      * [Development Commands](#development-commands)
      * [Development Workflow](#development-workflow)
      * [Release Candidates](#release-candidates)
   * [Alternative software](#alternative-software)
<!--te-->

## Features

- Event-driven architecture responding to monitor and power state changes in real-time
- Profile-based configuration with different settings for different monitor setups
- Template support for dynamic configuration generation
- Hot reloading: automatically detects and applies configuration changes without restart by watching config files (optional)
- Configurable UPower queries for custom power management systems
- Desktop notifications for configuration changes (optional)

## Design Philosophy

HyprDynamicMonitors follows a **fail-fast architecture** designed for reliability and simplicity.

### Reliability Through Restarts
The service intentionally fails quickly on critical issues rather than attempting complex recovery. This design expects the service to run under systemd or a wrapper script that provides automatic restarts. Since configuration is applied on startup, restarts ensure the service remains operational even after encountering errors.

### Hot Reloading With Graceful Restart
For configuration changes, the service provides automatic hot reloading by watching configuration files. When hot reloading encounters issues, it gracefully falls back to the fail-fast behavior, prioritizing reliability over attempting risky recovery scenarios.

### Hyprland-Native Integration
The service leverages Hyprland's native abstractions rather than working directly with Wayland protocols. It detects the desired configuration based on current monitor state and power supply, then either:
- Generates a templated Hyprland config file at the specified destination
- Or creates a symlink to a user-provided static configuration file

Hyprland automatically detects and applies these configuration changes (granted it's not explicitly turned off, if so you have to use
[the callbacks](https://github.com/fiffeek/hyprdynamicmonitors?tab=readme-ov-file#user-callbacks-exec-commands) to `hyprctl reload`), ensuring seamless integration with the compositor's built-in configuration system.

## Installation

### Binary Release

Download the latest binary from GitHub releases:

```bash
# optionally override the destination directory, defaults to ~/.local/bin/
export DESTDIR="$HOME/.bin"
curl -o- https://raw.githubusercontent.com/fiffeek/hyprdynamicmonitors/refs/heads/main/scripts/install.sh | bash
```

### AUR

For Arch Linux users, install from the AUR:

```bash
# Using your preferred AUR helper (replace 'aurHelper' with your choice)
aurHelper="yay"  # or paru, trizen, etc.
$aurHelper -S hyprdynamicmonitors-bin

# Or using makepkg:
git clone https://aur.archlinux.org/hyprdynamicmonitors-bin.git
cd hyprdynamicmonitors-bin
makepkg -si
```

### Nix

For Nix and NixOS users:

```bash
# Run directly from GitHub
nix run github:fiffeek/hyprdynamicmonitors

# Or from specific tag/version (recommended)
nix run github:fiffeek/hyprdynamicmonitors/v1.0.0

# Install to profile
nix profile install github:fiffeek/hyprdynamicmonitors
```

### Build from Source

Requires [asdf](https://asdf-vm.com/) to manage the Go toolchain:
```bash
# Build the binary (output goes to ./dest/)
make

# Install to custom location
make DESTDIR=$HOME/binaries install

# Uninstall from custom location
make DESTDIR=$HOME/binaries uninstall

# Install system-wide (may require sudo)
sudo make DESTDIR=/usr/bin install
```

## Usage

### Command Line

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

### Run command
<!-- START runhelp -->
```text
Run the HyprDynamicMonitors service to continuously monitor for display changes and automatically apply matching configuration profiles.

Usage:
  hyprdynamicmonitors run [flags]

Flags:
      --connect-to-session-bus    Connect to session bus instead of system bus for power events: https://wiki.archlinux.org/title/D-Bus. You can switch as long as you expose power line events in your user session bus.
      --disable-auto-hot-reload   Disable automatic hot reload (no file watchers)
      --disable-power-events      Disable power events (dbus)
      --dry-run                   Show what would be done without making changes
  -h, --help                      help for run
      --run-once                  Run once and exit immediately

Global Flags:
      --config string             Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
      --debug                     Enable debug logging
      --enable-json-logs-format   Enable structured logging
      --verbose                   Enable verbose logging
```
<!-- END runhelp -->

### Validate command
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

**Validate configuration:**
```bash
# Validate default config file
hyprdynamicmonitors validate

# Validate specific config file
hyprdynamicmonitors --config /path/to/config.toml validate

# Validate with debug output
hyprdynamicmonitors --debug validate
```


### Freeze command
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


## Minimal Example

This example sets up basic laptop-only monitor configuration. First, check your display name with `hyprctl monitors`.

**Configuration file** (`~/.config/hyprdynamicmonitors/config.toml`):
```toml
[general]
destination = "$HOME/.config/hypr/monitors.conf"

[power_events]
[power_events.dbus_query_object]
path = "/org/freedesktop/UPower/devices/line_power_ACAD"

[[power_events.dbus_signal_match_rules]]
object_path = "/org/freedesktop/UPower/devices/line_power_ACAD"

[profiles.laptop_only]
config_file = "hyprconfigs/laptop.conf"
config_file_type = "static"

[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"  # Replace with your display name from hyprctl monitors
```

**Monitor configuration** (`~/.config/hyprdynamicmonitors/hyprconfigs/laptop.conf`):
```hyprconfig
monitor=eDP-1,2880x1920@120.00000,0x0,2.0,vrr,1
```

**Run the service** using systemd (recommended - see [Running with systemd](#running-with-systemd)) or add to Hyprland config:
```conf
exec-once = hyprdynamicmonitors
```

**Ensure you source the linked `destination` config file** (in `~/.config/hypr/hyprland.conf`):
```conf
source = ~/.config/hypr/monitors.conf
```

**How it works**: When only the `eDP-1` monitor is detected, the service symlinks `hyprconfigs/laptop.conf` to `$HOME/.config/hypr/monitors.conf`, and Hyprland automatically applies the new configuration.

## Examples

See [`examples/`](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples) directory for complete configuration
examples including basic setups and comprehensive configurations with all features.
Most notably, [`examples/full/config.toml`](https://github.com/fiffeek/hyprdynamicmonitors/blob/main/examples/full/config.toml)
contains all available configuration options reference.
An example for disabling a monitor depending on the
detected setup is in [`examples/disable-monitors`](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/disable-monitors).

## Runtime requirements

- Hyprland with IPC support
- UPower (optional, for power state monitoring)
- Read-only access to system D-Bus (optional for power state monitoring; should already be your default)
- Write access to system D-Bus for notifications (optional; should already be your default)

## Configuration

### Monitor Matching

Profiles match monitors based on their properties. You can match by:

- **Name**: The monitor's connector name (e.g., `eDP-1`, `DP-1`)
- **Description**: The monitor's model/manufacturer string

You also optionally set (required for templating):
- **Tags**: Custom labels you assign to monitors for easier reference

```toml
[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"  # Match by connector name

[[profiles.external_4k.conditions.required_monitors]]
description = "Dell U2720Q"  # Match by monitor model

[[profiles.dual_setup.conditions.required_monitors]]
name = "eDP-1"
monitor_tag = "laptop"  # Assign a tag for template use
```

Use `hyprctl monitors` to see available monitors and their properties.
To understand scoring and profile matching more see [`examples/scoring`](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/scoring).

**Profile Selection**: When multiple profiles have equal scores, the last profile defined in the TOML configuration file is selected.

### Configuration File Types

- **Static**: Creates symlinks to existing configuration files
- **Template**: Processes Go templates with dynamic monitor and power state data

### Template Variables

```go
.PowerState          // "AC" or "BAT"
.Monitors           // Array of connected monitors
.MonitorsByTag      // Map of tagged monitors (monitor_tag -> monitor)
```

### Static Template Values

You can define custom static values that are available in templates. These can be defined globally or per-profile:

**Global static values** (available in all templates):
```toml
[static_template_values]
default_vrr = "1"
default_res = "2880x1920"
refresh_rate_high = "120.00000"
refresh_rate_low = "60.00000"
```

**Per-profile static values** (override global values):
```toml
[profiles.laptop_only.static_template_values]
default_vrr = "0"        # Override global value
battery_scaling = "1.5"  # Profile-specific value
```

**Template usage:**
```go
# Use static values in templates
monitor=eDP-1,{{.default_res}}@{{if isOnAC}}{{.refresh_rate_high}}{{else}}{{.refresh_rate_low}}{{end}},0x0,1,vrr,{{.default_vrr}}
```

To understand template variables more see [`examples/template-variables`](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/template-variables).

### Template Functions

```go
isOnBattery         // Returns true if on battery power
isOnAC              // Returns true if on AC power
powerState          // Returns current power state string
```

### Fallback Profile

When no regular profile matches the current monitor setup and power state, you can define a fallback profile that will be used as a last resort. This is particularly useful for handling unexpected monitor configurations or providing a safe default configuration.

```toml
# Regular profiles with specific conditions
[profiles.laptop_only]
config_file = "hyprconfigs/laptop.conf"

[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"

[profiles.dual_monitor]
config_file = "hyprconfigs/dual.conf"

[[profiles.dual_monitor.conditions.required_monitors]]
name = "eDP-1"

[[profiles.dual_monitor.conditions.required_monitors]]
description = "External Monitor"

# Fallback profile - used when no other profile matches
[fallback_profile]
config_file = "hyprconfigs/fallback.conf"
config_file_type = "static"
```

The fallback profile in `hyprconfigs/fallback.conf` might contain a safe default configuration:
```hyprconfig
# Generic fallback: configure all connected monitors with preferred settings
monitor=,preferred,auto,1
```

**Key characteristics of fallback profiles:**
- Cannot define conditions (since they're used when no conditions match)
- Supports both static and template configuration types
- Only used when no regular profile matches the current setup
- Regular matching profiles always take precedence over the fallback

To understand fallback profiles more see [`examples/fallback`](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/fallback).

### User Callbacks (Exec Commands)

HyprDynamicMonitors supports custom user commands that are executed before and after profile configuration changes. These commands can be defined globally or per-profile, allowing for custom actions like notifications, script execution, or system adjustments.

```toml
[general]
# Global exec commands - run for all profile changes
pre_apply_exec = "notify-send 'HyprDynamicMonitors' 'Switching monitor profile...'"
post_apply_exec = "notify-send 'HyprDynamicMonitors' 'Profile applied successfully'"

# Profile-specific exec commands override global settings
[profiles.gaming_setup]
config_file = "hyprconfigs/gaming.conf"
pre_apply_exec = "notify-send 'Gaming Mode' 'Activating high-performance profile'"
post_apply_exec = "/usr/local/bin/gaming-mode-on.sh"
```

**Key characteristics:**
- **`pre_apply_exec`**: Executed before the new monitor configuration is applied
- **`post_apply_exec`**: Executed after the new monitor configuration is successfully applied
- **Profile-specific commands override global commands** for that profile
- **Failure handling**: If exec commands fail, the service continues operating normally (no interruption to monitor configuration)
- **Shell execution**: Commands are executed through `bash -c`, supporting shell features like pipes and environment variables

To understand callbacks more see [`examples/callbacks`](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/callbacks).

### Notifications

HyprDynamicMonitors can show desktop notifications when configuration changes occur. Notifications are sent via D-Bus using the standard `org.freedesktop.Notifications` interface.

```toml
[notifications]
disabled = false      # Enable/disable notifications (default: false)
timeout_ms = 10000   # Notification timeout in milliseconds (default: 10000)
```

**To disable notifications completely:**
```toml
[notifications]
disabled = true
```

**To show brief notifications:**
```toml
[notifications]
timeout_ms = 3000    # 3 seconds
```

### Hyprland Integration

Add to your Hyprland config (assuming `~/.config/hypr/monitors.conf` is your destination):

```conf
source = ~/.config/hypr/monitors.conf
```

**Important**: Do not set `disable_autoreload = true` in Hyprland settings, or you'll have to reload Hyprland manually after configuration changes.

### Power Events

Power state monitoring uses D-Bus to listen for UPower events. This feature is optional and can be completely disabled.

#### Disabling power events

To disable power state monitoring entirely, start with the flag:

```bash
hyprdynamicmonitors --disable-power-events
```

When disabled, the system defaults to `AC` power state.
No power events will be delivered, no dbus connection will be made.

#### Default power event configuration

By default, the service listens for D-Bus signals:
- **Signal**: `org.freedesktop.DBus.Properties.PropertiesChanged`
- **Interface**: `org.freedesktop.DBus.Properties`
- **Member**: `PropertiesChanged`
- **Path**: `/org/freedesktop/UPower/devices/line_power_ACAD`

These defaults can be overridden in the configuration (or left empty).

You can monitor these events with:
```bash
gdbus monitor -y -d org.freedesktop.UPower | grep -E "PropertiesChanged|Device(Added|Removed)"
```

Example output:
```
/org/freedesktop/UPower/devices/line_power_ACAD: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower.Device', {'UpdateTime': <uint64 1756242314>, 'Online': <true>}, @as [])
# Format: PATH INTERFACE.MEMBER (INFO)
```

#### Querying

On each event, the current power status is queried. Here's the equivalent command:
```bash
dbus-send --system --print-reply --dest=org.freedesktop.UPower \
  /org/freedesktop/UPower/devices/line_power_ACAD \
  org.freedesktop.DBus.Properties.Get string:org.freedesktop.UPower.Device string:Online
```

#### Receive Filters
You can filter received events by name. By default, only `org.freedesktop.DBus.Properties.PropertiesChanged` is matched. This prevents noisy signals. Additionally, power status changes are only propagated when the state actually changes, and template/link replacement only occurs when file contents differ.

#### Custom D-Bus Configuration

You can customize which D-Bus signals to monitor:

```toml
[power_events]
# Custom D-Bus signal match rules
[[power_events.dbus_signal_match_rules]]
interface = "org.freedesktop.DBus.Properties"
member = "PropertiesChanged"
object_path = "/org/freedesktop/UPower/devices/line_power_ACAD"

# Custom signal filters
[[power_events.dbus_signal_receive_filters]]
name = "org.freedesktop.DBus.Properties.PropertiesChanged"

# Custom UPower query for non-standard power managers
[power_events.dbus_query_object]
destination = "org.freedesktop.UPower"
path = "/org/freedesktop/UPower"
method = "org.freedesktop.DBus.Properties.Get"
expected_discharging_value = "true"

[[power_events.dbus_query_object.args]]
arg = "org.freedesktop.UPower"

[[power_events.dbus_query_object.args]]
arg = "OnBattery"
```

The above query settings are equivalent to:
```bash
dbus-send --system --print-reply --dest=org.freedesktop.UPower \
  /org/freedesktop/UPower org.freedesktop.DBus.Properties.Get string:org.freedesktop.UPower string:OnBattery
```
**Note**: This particular query is not recommended for production use!

#### Leave Empty Token

To explicitly remove default values from D-Bus match rules, use the `leaveEmptyToken`:

```toml
[[power_events.dbus_signal_match_rules]]
interface = "leaveEmptyToken"  # Removes interface match
member = "PropertiesChanged"
object_path = "/custom/path"
```


### Signals

- **SIGHUP**: Instantly reloads configuration and reapplies monitor setup
- **SIGUSR1**: Reapplies the monitor setup without reloading the service configuration
- **SIGTERM/SIGINT**: Graceful shutdown

You can trigger an instant reload with:
```bash
kill -SIGHUP $(pidof hyprdynamicmonitors)
# or just restart if running as systemd service
# since the configuration is applied on the startup
systemctl --user reload hyprdynamicmonitors
```

### Hot Reloading

HyprDynamicMonitors automatically watches for changes to configuration files and applies them without requiring a restart. This includes:

- **Configuration file changes**: Modifications to `config.toml` are detected and applied automatically
- **Profile config changes**: Updates to individual profile configuration files (both static and template files)
- **New profile files**: Adding new configuration files referenced by profiles

The service uses file system watching with debounced updates to avoid excessive reloading during rapid changes (e.g., when editors create temporary files).

**Hot reload behavior:**
- Configuration changes are debounced by default (1000ms delay; configurable)
- Only actual content changes trigger reloads (identical content is ignored)
- Service continues running normally during hot reloads

**Disabling hot reload:**
```bash
hyprdynamicmonitors --disable-auto-hot-reload
```

When disabled, you can still use `SIGHUP` signal for manual reloading.

## Tests

### Live Testing

Live tested on:
- Hyprland v0.50.1
- UPower v1.90.9

You can see my configuration [here](https://github.com/fiffeek/.dotfiles.v2/blob/main/ansible/files/framework/dots/hyprdynamicmonitors/config.toml).

### Integration Testing
All features should be covered by integration tests. Run `make test/integration` locally to execute end-to-end CLI tests that build the binary and verify expected outputs. Test fixtures can be regenerated using `make test/integration/regenerate`.

## Running with systemd

For production use, it's recommended to run HyprDynamicMonitors as a systemd user service. This ensures automatic restart on failures and proper integration with session management.

**Important**: Ensure you're properly [pushing environment variables to systemd](https://wiki.hypr.land/Nix/Hyprland-on-Home-Manager/#programs-dont-work-in-systemd-services-but-do-on-the-terminal).

### Hyprland under systemd
If you run [Hyprland under systemd](https://wiki.hypr.land/Useful-Utilities/Systemd-start/), setup is straightforward.
Create `~/.config/systemd/user/hyprdynamicmonitors.service`:

```ini
[Unit]
Description=HyprDynamicMonitors - Dynamic monitor configuration for Hyprland
After=graphical-session.target
Wants=graphical-session.target
PartOf=hyprland-session.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors
Restart=on-failure
RestartSec=5

[Install]
WantedBy=hyprland-session.target
```


Enable and start the service:
```bash
systemctl --user daemon-reload
systemctl --user enable hyprdynamicmonitors
systemctl --user start hyprdynamicmonitors
```

### Run on boot and let restarts do the job
You can essentially just run it on boot and add restarts, e.g.:
```ini
[Unit]
Description=HyprDynamicMonitors - Dynamic monitor configuration for Hyprland
After=default.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```
It will keep failing until Hyprland is ready/launched and environment variables are propagated.

### Custom systemd target
You can also add [a custom systemd target that would be started by Hyprland](https://github.com/fiffeek/.dotfiles.v2/commit/2a0d400b81031e3786a2779c36f70c9771aee884), e.g.
```
exec-once = systemctl --user start hyprland-custom-session.target
bind = $mainMod, X, exec, systemctl --user stop hyprland-session.target
```

Then:
```bash
❯ cat ~/.config/systemd/user/hyprland-custom-session.target
[Unit]
Description=A target for other services when hyprland becomes ready
After=graphical-session-pre.target
Wants=graphical-session-pre.target
BindsTo=graphical-session.target
```
And:
```bash
❯ cat ~/.config/systemd/user/hyprdynamicmonitors.service
[Unit]
Description=Run hyprdynamicmonitors daemon
After=hyprland-custom-session.target
After=dbus.socket
Requires=dbus.socket
PartOf=hyprland-custom-session.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors
Restart=on-failure
RestartSec=5


[Install]
WantedBy=hyprland-custom-session.target
```

### Alternative: Wrapper script

If you prefer a wrapper script approach, create a simple restart loop:

```bash
#!/bin/bash
while true; do
    /usr/bin/hyprdynamicmonitors
    echo "HyprDynamicMonitors exited with code $?, restarting in 5 seconds..."
    sleep 5
done
```
Then execute it from Hyprland:
```
exec-once = /path/to/the/script.sh
```

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

## Alternative software

Most similar tools are more generic, working with any Wayland compositor. In contrast, `hyprdynamicmonitors` is specifically designed for Hyprland (using its IPC) but provides several advantages:

**Advantages of HyprDynamicMonitors:**
- **Full configuration control**: Instead of introducing another configuration format, you work directly with Hyprland's native config syntax
- **Template system**: Dynamic configuration generation based on connected monitors and power state
- **Power state awareness**: Built-in AC/battery detection for laptop users

**Trade-offs:**
- Hyprland-specific (not generic Wayland)
- Requires systemd or wrapper script for production use (fail-fast design)
- More complex setup compared to simpler tools

**Similar Tools:**
- [kanshi](https://sr.ht/~emersion/kanshi/) - Generic Wayland output management
- [shikane](https://github.com/hw0lff/shikane) - Another Wayland output manager
- [nwg-displays](https://github.com/nwg-piotr/nwg-displays) - GUI-based display configuration tool for Sway/Hyprland
- [hyprmon](https://github.com/erans/hyprmon) - TUI-based display configuration tool for Hyprland

`hyprdynamicmonitors` can be used side-by-side with `nwg-displays` or `hyprmon`:
- Tweak the configuration in either tool
- Let `hyprdynamicmonitors` automatically write/link it to your hyprland's configuration directory
