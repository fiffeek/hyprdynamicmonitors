---
layout: default
title: Configuration
---

{% raw %}

## Configuration

HyprDynamicMonitors automatically creates a default configuration file on first run if none exists at the specified path (the destination defaults to `~/.config/hypr/monitors.toml`).

The default configuration:
- Automatically detects your system's power line using `upower -e`
- Searches for common power line paths (e.g., `line_power_ACAD`, `line_power_AC`, `line_power_ADP1`)
- Falls back to `/org/freedesktop/UPower/devices/line_power_ACAD` if detection fails
- Creates a minimal config with power event monitoring but **no profiles**
- Allows you to start adding profiles immediately without manually configuring power events

To customize your setup, add profile sections to the generated config file. See the [Minimal Example](#minimal-example) and [Examples](#examples) sections for guidance. Alternatively, use the `TUI`.

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
.PowerState         // "AC" or "BAT"
.LidState           // "UNKNOWN", "Closed" or "Opened"
.Monitors           // Array of all connected monitors returned by `hyprctl monitors`
.ExtraMonitors      // Array of connected monitors not defined in the profile
.RequiredMonitors   // Array of connected monitors defined in the profile
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
isLidClosed         // Returns true if the lid is closed (if --enabling-lid-events is passed)
isLidOpened         // Returns true if the lid is opened (if --enabling-lid-events is passed)
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

### Lid events

Lid state monitoring uses D-Bus to listen for UPower events. This feature is optional and needs to be explicitly enabled.

#### Enabling lid events

To enable lid state monitoring, start with the flag:

```bash
hyprdynamicmonitors --enable-lid-events
```

When disabled:
- the system defaults to `UNKNOWN` lid state.
- no power events will be delivered, no dbus connection will be made.

#### Default lid event configuration

By default, the service listens for D-Bus signals:
- **Signal**: `org.freedesktop.DBus.Properties.PropertiesChanged`
- **Interface**: `org.freedesktop.DBus.Properties`
- **Member**: `PropertiesChanged`
- **Path**: `/org/freedesktop/UPower`

These defaults can be overridden in the configuration (or left empty).

You can monitor these events with:
```bash
gdbus monitor --system --dest org.freedesktop.UPower --object-path /org/freedesktop/UPower
```

Example output:
```
/org/freedesktop/UPower: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower', {'LidIsClosed': <true>}, @as [])
/org/freedesktop/UPower: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower', {'LidIsClosed': <false>}, @as [])
```

#### Querying lid events

On each event, the current power status is queried. Here's the equivalent command:
```bash
dbus-send --system --print-reply \
  --dest=org.freedesktop.UPower /org/freedesktop/UPower \
  org.freedesktop.DBus.Properties.Get string:org.freedesktop.UPower string:LidIsClosed
```

#### Receive lid Filters
You can filter received events by name and body. By default, the service matches:
- `org.freedesktop.DBus.Properties.PropertiesChanged` name.
- `LidIsClosed` body.

This prevents noisy signals. Additionally, lid status changes are only propagated when the state actually changes.


#### Custom D-Bus Configuration for lid events

You can customize which D-Bus signals to monitor, see [lid-states example](./examples/lid-states/config.toml).

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

### TUI

A tui is available for ad-hoc changes as well as persisting the configuration under `hyprdynamicmonitors`.
Refer to the [tui docs](./docs/tui-help.md) for more details.

{% endraw %}
