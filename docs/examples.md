---
layout: default
title: Examples
---


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

