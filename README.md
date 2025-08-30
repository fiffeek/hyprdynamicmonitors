<img src="https://github.com/user-attachments/assets/0effc242-3d3d-4d39-a183-0a567c4da3a9" width="90" style="margin-right:10px" align=left alt="hyprdynamicmonitors logo">
<H1>HyprDynamicMonitors</H1><br>


An event-driven service that automatically manages Hyprland monitor configurations based on connected displays and power state.

## Features

- Event-driven architecture responding to monitor and power state changes in real-time
- Profile-based configuration with different settings for different monitor setups
- Template support for dynamic configuration generation
- Hot reloading: automatically detects and applies configuration changes without restart by watching config files (optional)
- Configurable UPower queries for custom power management systems
- Desktop notifications for configuration changes (optional)

## Design Philosophy

The service is designed to fail fast on most issues, which means it should be run under systemd or a wrapper script for automatic restarts. It applies the configuration on startup, so restarts keep it operational even when critical failures occur.

For configuration changes, the service provides hot reloading that watches for file modifications and automatically applies updates without restart. When hot reloading encounters issues, the service gracefully falls back to its fail-fast behavior, ensuring reliability over attempting complex recovery scenarios.

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


### Minimal example

In `~/.config/hyprdynamicmonitors/config.toml` (assuming you have `eDP-1` display attached, see `hyprctl monitors` to query them):

```toml
[general]
destination = "$HOME/.config/hypr/monitors.conf"

[power_events]

[power_events.dbus_query_object]
# path to your line_power upower device
path = "/org/freedesktop/UPower/devices/line_power_ACAD"

[[power_events.dbus_signal_match_rules]]
# path to your line_power upower device
object_path = "/org/freedesktop/UPower/devices/line_power_ACAD"

[profiles.laptop_only]
config_file = "hyprconfigs/laptop.conf"
config_file_type = "static"

[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"
```

In `~/.config/hyprdynamicmonitors/hyprconfigs/laptop.conf`:
```hyprconfig
monitor=eDP-1,2880x1920@120.00000,0x0,2.0,vrr,1
```

Then run the service either in Hyprland (`exec-once`) or ideally
use systemd (see [Running with systemd](#running-with-systemd)) or a wrapper script.

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

## Usage

### Command Line

```text

Usage: hyprdynamicmonitors [options] [command]

Commands:
  run      Run the service (default)
  validate Validate configuration file and exit

Options:
  -config string
        Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
  -debug
        Enable debug logging
  -disable-auto-hot-reload
        Disable automatic hot reload (no file watchers)
  -disable-power-events
        Disable power events (dbus)
  -dry-run
        Show what would be done without making changes
  -verbose
        Enable verbose logging
  -version
        Show version information
```

**Validate configuration:**
```bash
# Validate default config file
hyprdynamicmonitors validate

# Validate specific config file
hyprdynamicmonitors -config /path/to/config.toml validate

# Validate with debug output
hyprdynamicmonitors -debug validate
```

### Hyprland Integration

Add to your Hyprland config (assuming `~/.config/hypr/monitors.conf` is your destination):

```conf
source = ~/.config/hypr/monitors.conf
```

**Important**: Do not set `disable_autoreload = true` in Hyprland settings, or you'll have to reload Hyprland manually after configuration changes.

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

## Examples

See `examples/` directory for complete configuration examples including basic setups and comprehensive configurations with all features.

## Runtime requirements

- Hyprland with IPC support
- UPower (optional, for power state monitoring)

## Tests

### Live Testing

Live tested on:
- Hyprland v0.50.1
- UPower v1.90.9

You can see my configuration [here](https://github.com/fiffeek/.dotfiles.v2/blob/main/ansible/files/framework/dots/hyprdynamicmonitors/config.toml).

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

## Power Events Configuration

Power state monitoring uses D-Bus to listen for UPower events. This feature is optional and can be completely disabled.

### Disabling power events

To disable power state monitoring entirely, start with the flag:

```bash
hyprdynamicmonitors --disable-power-events
```

When disabled, the system defaults to `AC` power state.
No power events will be delivered, no dbus connection will be made.

### Default power event configuration

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

### Querying

On each event, the current power status is queried. Here's the equivalent command:
```bash
dbus-send --system --print-reply --dest=org.freedesktop.UPower \
  /org/freedesktop/UPower/devices/line_power_ACAD \
  org.freedesktop.DBus.Properties.Get string:org.freedesktop.UPower.Device string:Online
```

### Receive Filters
You can filter received events by name. By default, only `org.freedesktop.DBus.Properties.PropertiesChanged` is matched. This prevents noisy signals. Additionally, power status changes are only propagated when the state actually changes, and template/link replacement only occurs when file contents differ.

### Custom D-Bus Configuration

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

### Leave Empty Token

To explicitly remove default values from D-Bus match rules, use the `leaveEmptyToken`:

```toml
[[power_events.dbus_signal_match_rules]]
interface = "leaveEmptyToken"  # Removes interface match
member = "PropertiesChanged"
object_path = "/custom/path"
```

## Alternatives

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
- [nwg-displays](https://github.com/nwg-piotr/nwg-displays) - GUI-based display configuration tool
