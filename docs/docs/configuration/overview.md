---
sidebar_position: 1
---

# Overview

HyprDynamicMonitors automatically creates a default configuration file on first run if none exists at the specified path (defaults to `~/.config/hypr/hyprdynamicmonitors.toml`).

## Default Configuration

The default configuration:
- Automatically detects your system's power line using `upower -e`
- Searches for common power line paths (e.g., `line_power_ACAD`, `line_power_AC`, `line_power_ADP1`)
- Falls back to `/org/freedesktop/UPower/devices/line_power_ACAD` if detection fails
- Creates a minimal config with power event monitoring but **no profiles**
- Allows you to start adding profiles immediately without manually configuring power events

## Configuration Structure

The configuration file is written in TOML and consists of several main sections:

### General Settings

```toml
[general]
destination = "$HOME/.config/hypr/monitors.conf"
```

The `destination` specifies where the monitor configuration file will be created or linked.

### Power Events

```toml
[power_events]
[power_events.dbus_query_object]
path = "/org/freedesktop/UPower/devices/line_power_ACAD"

[[power_events.dbus_signal_match_rules]]
object_path = "/org/freedesktop/UPower/devices/line_power_ACAD"
```

Power events monitor your system's power state (AC/Battery) via D-Bus. See [Power Events](./power-events) for details.

### Lid Events

```toml
[lid_events]
# custom config goes here, the defaults should work in most cases
```

Lid events monitor your system's lid state (Opened/Closed) via D-Bus. See [Lid Events](./lid-events) for details.

### Profiles

```toml
[profiles.laptop_only]
config_file = "hyprconfigs/laptop.conf"
config_file_type = "static"

[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"
```

Profiles define different monitor configurations for different setups. Each profile can have:
- Configuration file (static or template)
- Conditions (required monitors, power state, lid state)
- Callbacks (pre/post apply commands)

See [Profiles](./profiles) for details.

### Notifications

```toml
[notifications]
disabled = false
timeout_ms = 10000
```

Configure desktop notifications for configuration changes. See [Notifications](./notifications).

## Next Steps

- [Monitor Matching](./monitor-matching) - Learn how to match monitors
- [Profiles](./profiles) - Configure different monitor setups
- [Templates](../advanced/templates) - Use dynamic configuration generation
- [Examples](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples) - See complete configuration examples
