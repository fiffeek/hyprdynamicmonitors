<img src="https://github.com/user-attachments/assets/0effc242-3d3d-4d39-a183-0a567c4da3a9" width="90" style="margin-right:10px" align=left alt="hyprdynamicmonitors logo">
<H1>HyprDynamicMonitors</H1><br>


An event-driven service that automatically manages Hyprland monitor configurations based on connected displays and power state.

## Features

- Event-driven architecture responding to monitor and power state changes in real-time
- Profile-based configuration with different settings for different monitor setups
- Template support for dynamic configuration generation
- Configurable UPower queries for custom power management systems

## Design Philosophy

The service is designed to fail fast on most issues, which means it should be run under systemd or a wrapper script for automatic restarts. It applies the configuration on startup, so restarts keep it operational even when critical failures occur. The reasoning is that it's easier to resume operations from a fresh start than to ensure no events are missed (the program is idle most of the time anyway).

## Installation

### Binary Release

Download the latest binary from releases and place it in your PATH:

TODO

### AUR

TODO


### Minimal example

In `~/.config/hyprdynamicmonitors/config.toml` (assuming you have `eDP-1` display attached, see `hyprctl monitors` to query them):

```toml
[general]
destination = "$HOME/.config/hypr/monitors.conf"

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

### Template Functions

```go
isOnBattery         // Returns true if on battery power
isOnAC              // Returns true if on AC power
powerState          // Returns current power state string
```

## Usage

### Command Line

```bash
hyprdynamicmonitors [options]

Usage of hyprdynamicmonitors:
  -config string
        Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
  -debug
        Enable debug logging
  -dry-run
        Show what would be done without making changes
  -verbose
        Enable verbose logging
  -version
        Show version information
```

### Hyprland Integration

Add to your Hyprland config (assuming `~/.config/hypr/monitors.conf` is your destination):

```conf
source = ~/.config/hypr/monitors.conf
```

**Important**: Do not set `disable_autoreload = true` in Hyprland settings, or you'll have to reload Hyprland manually after configuration changes.

### Signals

- **SIGUSR1**: Triggers a configuration update, NOT hot-reload
- **SIGTERM/SIGINT**: Graceful shutdown

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

To disable power state monitoring entirely:

```toml
[power_events]
disabled = true
```

When disabled, the system defaults to `AC` power state.

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
- No hot-reloading (relies on restarts for robustness)
- Requires systemd or wrapper script for production use

**Similar Tools:**
- [kanshi](https://sr.ht/~emersion/kanshi/) - Generic Wayland output management
- [shikane](https://github.com/hw0lff/shikane) - Another Wayland output manager
- [nwg-displays](https://github.com/nwg-piotr/nwg-displays) - GUI-based display configuration tool
