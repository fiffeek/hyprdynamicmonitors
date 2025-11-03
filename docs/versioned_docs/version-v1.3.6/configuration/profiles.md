---
sidebar_position: 3
---

# Profiles

Profiles define different monitor configurations for different setups. Each profile specifies which monitors it applies to and what configuration to use.

## Profile Structure

```toml
[profiles.PROFILE_NAME]
config_file = "path/to/config/file"
config_file_type = "static"  # or "template"

[profiles.PROFILE_NAME.conditions]
power_state = "AC"      # optional: "AC" or "BAT"
lid_state = "Opened"    # optional: "Opened" or "Closed" (requires --enable-lid-events)

[[profiles.PROFILE_NAME.conditions.required_monitors]]
name = "eDP-1"
monitor_tag = "laptop"  # optional
```

## Configuration File Types

### Static Configuration

Static configurations are plain Hyprland config files. The service creates a symlink to them.

```toml
[profiles.laptop_only]
config_file = "hyprconfigs/laptop.conf"
config_file_type = "static"
```

**hyprconfigs/laptop.conf:**
```
monitor=eDP-1,2880x1920@120,0x0,2.0,vrr,1
```

### Template Configuration

Templates use Go template syntax for dynamic configuration generation.

```toml
[profiles.dual_monitor]
config_file = "hyprconfigs/dual.go.tmpl"
config_file_type = "template"

[[profiles.dual_monitor.conditions.required_monitors]]
name = "eDP-1"
monitor_tag = "laptop"

[[profiles.dual_monitor.conditions.required_monitors]]
description = "LG.*ULTRAWIDE"
monitor_tag = "external"
```

**hyprconfigs/dual.go.tmpl:**
```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
{{- $external := index .MonitorsByTag "external" -}}
monitor={{$laptop.Name}},{{if isOnAC}}2880x1920@120{{else}}1920x1080@60{{end}},0x0,2.0
monitor={{$external.Name}},preferred,2880x0,1
```

See [Templates](../advanced/templates) for details on template syntax and variables.

## Profile Conditions

### Required Monitors

Each profile can require specific monitors to be connected. A profile matches when all its required monitors are present.

```toml
# Single monitor
[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"

# Multiple monitors
[[profiles.triple_4k.conditions.required_monitors]]
name = "eDP-1"

[[profiles.triple_4k.conditions.required_monitors]]
description = "Dell U2720Q"

[[profiles.triple_4k.conditions.required_monitors]]
description = "LG 27UK850"
```

### Power State Conditions

You can restrict a profile to only match on specific power states:

```toml
[profiles.performance_mode.conditions]
power_state = "AC"  # Only match when plugged into AC power

[[profiles.performance_mode.conditions.required_monitors]]
name = "eDP-1"

[profiles.battery_saver.conditions]
power_state = "BAT"  # Only match when on battery

[[profiles.battery_saver.conditions.required_monitors]]
name = "eDP-1"
```

Valid values:
- `"AC"` - Only match when on AC power
- `"BAT"` - Only match when on battery power
- Omit `power_state` condition to match on any power state

See [Power States Example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/power-states) for a complete configuration.

### Lid State Conditions

You can restrict a profile to only match on specific lid states (requires `--enable-lid-events` flag):

```toml
[profiles.lid_closed.conditions]
lid_state = "Closed"  # Only match when lid is closed

[[profiles.lid_closed.conditions.required_monitors]]
name = "eDP-1"

[profiles.lid_opened.conditions]
lid_state = "Opened"  # Only match when lid is open

[[profiles.lid_opened.conditions.required_monitors]]
name = "eDP-1"
```

Valid values:
- `"Closed"` - Only match when lid is closed
- `"Opened"` - Only match when lid is open
- Omit `lid_state` condition to match on any lid state

See [Lid States Example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/lid-states) for a complete configuration.

### Combining Conditions

You can combine monitor, power state, and lid state conditions:

```toml
[profiles.docked_ac.conditions]
power_state = "AC"
lid_state = "Closed"

[[profiles.docked_ac.conditions.required_monitors]]
name = "eDP-1"

[[profiles.docked_ac.conditions.required_monitors]]
description = "External Monitor"
```

This profile only matches when:
- On AC power AND
- Lid is closed AND
- Both eDP-1 and an external monitor are connected

## Static Template Values

You can define custom values that are available in templates. These can be defined globally or per-profile.

### Global Static Values

Available in all templates:

```toml
[static_template_values]
default_vrr = "1"
default_res = "2880x1920"
refresh_rate_high = "120.00000"
refresh_rate_low = "60.00000"
```

### Per-Profile Static Values

Override global values or add profile-specific values:

```toml
[profiles.laptop_only.static_template_values]
default_vrr = "0"        # Override global value
battery_scaling = "1.5"  # Profile-specific value
```

### Using in Templates

```go
monitor=eDP-1,{{.default_res}}@{{if isOnAC}}{{.refresh_rate_high}}{{else}}{{.refresh_rate_low}}{{end}},0x0,1,vrr,{{.default_vrr}}
```

## Fallback Profile

When no regular profile matches, you can define a fallback profile:

```toml
[fallback_profile]
config_file = "hyprconfigs/fallback.conf"
config_file_type = "static"
```

The fallback profile:
- Cannot define conditions (used when no conditions match)
- Supports both static and template configuration types
- Only used when no regular profile matches

**Example fallback configuration:**
```
# Generic fallback: configure all connected monitors with preferred settings
monitor=,preferred,auto,1
```

## Profile Callbacks

You can define commands to execute before and after profile application:

```toml
[general]
# Global exec commands - run for all profile changes
pre_apply_exec = "notify-send 'Switching profile...'"
post_apply_exec = "notify-send 'Profile applied'"

# Profile-specific commands override global settings
[profiles.gaming_setup]
config_file = "hyprconfigs/gaming.conf"
pre_apply_exec = "notify-send 'Gaming Mode' 'Activating...'"
post_apply_exec = "/usr/local/bin/gaming-mode-on.sh"
```

See [Callbacks](./callbacks) for details.

## Examples

For complete configuration examples, see:
- [Basic Example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/basic)
- [Full Configuration](https://github.com/fiffeek/hyprdynamicmonitors/blob/main/examples/full/config.toml)
- [Power States Example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/power-states)
- [Lid States Example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/lid-states)
- [Disable Monitors Example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/disable-monitors)
- [Scoring Example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/scoring)
