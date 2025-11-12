---
sidebar_position: 1
---

# Overview

HyprDynamicMonitors automatically creates a default configuration file on first run if none exists at the specified path (defaults to `~/.config/hyprdynamicmonitors/config.toml`).

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

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[general]
destination = "$HOME/.config/hypr/monitors.conf"
debounce_time_ms = 1500
pre_apply_exec = "notify-send 'Switching profile...'"
post_apply_exec = "notify-send 'Profile applied'"
```

- `destination` - Where the monitor configuration file will be created or linked
- `debounce_time_ms` - Collect events for this duration before applying changes (prevents configuration thrashing, default: 1500ms)
- `pre_apply_exec` - Command to run before applying configuration (optional)
- `post_apply_exec` - Command to run after applying configuration (optional)

See [Callbacks](./callbacks) for details on exec commands.

### Power Events

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[power_events]
[power_events.dbus_query_object]
path = "/org/freedesktop/UPower/devices/line_power_ACAD"

[[power_events.dbus_signal_match_rules]]
object_path = "/org/freedesktop/UPower/devices/line_power_ACAD"
```

Power events monitor your system's power state (AC/Battery) via D-Bus. See [Power Events](./power-events) for details.

### Lid Events

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[lid_events]
# custom config goes here, the defaults should work in most cases
```

Lid events monitor your system's lid state (Opened/Closed) via D-Bus. See [Lid Events](./lid-events) for details.

### Profiles

```toml title="~/.config/hyprdynamicmonitors/config.toml"
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

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[notifications]
disabled = false
timeout_ms = 10000
```

Configure desktop notifications for configuration changes. See [Notifications](./notifications).

### Hot Reload

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[hot_reload_section]
debounce_time_ms = 1000
```

Hot reload watches configuration files for changes and automatically reloads them. The `debounce_time_ms` setting controls how long to wait before applying changes after detecting a file modification.

### Scoring

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[scoring]
name_match = 10
description_match = 5
power_state_match = 3
lid_state_match = 2
```

Customize the scoring system for profile selection when multiple profiles match. Higher scores win:
- `name_match` - Points for exact monitor name match (e.g., "eDP-1")
- `description_match` - Points for exact monitor description match
- `power_state_match` - Bonus points for matching power state
- `lid_state_match` - Bonus points for matching lid state

See [Monitor Matching](./monitor-matching) for details on how profiles are selected.

### Static Template Values

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[static_template_values]
default_vrr = "1"
default_res = "2880x1920"
```

Define global values that can be used in all templates. These can be overridden per-profile. See [Templates](../advanced/templates) and [Profiles](./profiles) for details.

### Fallback Profile

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[fallback_profile]
config_file = "hyprconfigs/fallback.conf"
config_file_type = "static"
```

Define a fallback profile that will be used when no other profile matches the current state. See [Profiles](./profiles) for details.

## Next Steps

- [Monitor Matching](./monitor-matching) - Learn how to match monitors
- [Profiles](./profiles) - Configure different monitor setups
- [Templates](../advanced/templates) - Use dynamic configuration generation
- [Examples](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples) - See complete configuration examples
