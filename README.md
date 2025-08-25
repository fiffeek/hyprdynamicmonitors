# HyprDynamicMonitors

An event-driven service that automatically manages Hyprland monitor configurations based on connected displays and power state. The service monitors for display connection/disconnection events and power state changes (AC/Battery), then dynamically switches between predefined Hyprland configuration profiles.

## Features

- **Event-driven architecture**: Responds to monitor and power state changes in real-time
- **Profile-based configuration**: Define different configurations for different monitor setups
- **Power state awareness**: Different settings for AC power vs battery mode
- **Template support**: Dynamic configuration generation with Go templates
- **Static configuration**: Simple symlink-based configuration switching
- **Flexible matching**: Match monitors by name, description, or both
- **Fail-fast design**: Designed to be managed by systemd or a wrapper script with automatic restart on failure

## Installation

> **TODO**: Installation instructions will be provided once binary releases and AUR packages are available.

## Configuration

The service uses TOML configuration files with three main sections:

### General Section

```toml
[general]
destination = "$HOME/.config/hypr/config.d/dynamic-monitors.conf"
```

- `destination`: Path where the selected configuration will be written/linked

### Power Events Section

```toml
[power_events]
disabled = false  # Set to true to disable power state monitoring

# Optional: Custom D-Bus signal matching (advanced users)
[[power_events.dbus_signal_match_rules]]
interface = "org.freedesktop.DBus.Properties"
object_path = "/org/freedesktop/UPower"
member = "PropertiesChanged"

[[power_events.dbus_signal_receive_filters]]
name = "org.freedesktop.DBus.Properties.PropertiesChanged"
```

**Options:**
- `disabled`: Boolean, disables power state monitoring (defaults to `false`)
- `dbus_signal_match_rules`: Array of D-Bus signal matching rules for UPower events
- `dbus_signal_receive_filters`: Array of signal name filters to process

### Scoring Section

```toml
[scoring]
name_match = 10        # Points for exact monitor name match
description_match = 5  # Points for exact monitor description match
power_state_match = 3  # Points for matching power state requirement
```

The scoring system determines which profile to use when multiple profiles match the current monitor setup. Higher scoring profiles are selected.

### Profiles Section

Profiles define different monitor configurations for different scenarios:

```toml
[profiles.profile_name]
config_file = "path/to/config.conf"
config_file_type = "static"  # or "template"

# Power state requirement (optional)
[profiles.profile_name.conditions]
power_state = "AC"  # "AC", "BAT", or omit for any

# Required monitors
[[profiles.profile_name.conditions.required_monitors]]
name = "eDP-1"  # Exact monitor name match

[[profiles.profile_name.conditions.required_monitors]]
description = "BOE NE135A1M-NY1"  # Exact monitor description match
monitor_tag = "laptop"  # Tag for template usage (optional)
```

**Configuration Types:**

1. **Static** (`config_file_type = "static"`):
   - Creates a symlink to the specified configuration file
   - Use for fixed configurations

2. **Template** (`config_file_type = "template"`):
   - Processes Go templates with monitor and (optionally) power state data
   - Dynamic configuration generation
   - Access to template functions and variables

**Profile Matching:**

- **All required monitors must be connected** for a profile to match (partial matches are discarded)
- **Power state conditions** are optional but provide bonus scoring
- **Scoring determines winner** when multiple profiles match
- **Monitor tags** allow templates to reference specific monitors

## Template System

Templates have access to the following data and functions:

### Template Variables

```go
.PowerState          // Current power state: "AC" or "BAT"
.Monitors           // Array of all connected monitors
.MonitorsByTag      // Map of monitor tags to monitor objects
```

### Monitor Object Fields

```go
.Name               // Monitor name (e.g., "eDP-1", "DP-1")
.ID                 // Monitor ID (e.g., "0", "1")
.Description        // Monitor description (e.g., "BOE NE135A1M-NY1")
```

### Template Functions

```go
isOnBattery         // Returns true if on battery power
isOnAC              // Returns true if on AC power
powerState          // Returns current power state string
```

### Template Example

```go
# Generated for power state: {{.PowerState}}

{{- $laptop := index .MonitorsByTag "laptop" -}}
{{- $external := index .MonitorsByTag "external" -}}

{{- if $laptop }}
{{- if isOnBattery }}
# Battery mode: Lower refresh rate
monitor={{$laptop.Name}},2880x1920@60.00000,0x0,2.0
{{- else }}
# AC mode: Full performance
monitor={{$laptop.Name}},2880x1920@120.00000,0x0,2.0
{{- end }}
{{- end }}

{{- if $external }}
{{- if isOnAC }}
monitor={{$external.Name}},2560x1440@60.00000,auto-left,1.0
{{- else }}
monitor={{$external.Name}},disable
{{- end }}
{{- end }}
```

## Example Configuration

See the complete example in [`examples/basic/config.toml`](examples/basic/config.toml):

- **Multiple profiles** for different monitor combinations
- **Power state conditions** for battery optimization
- **Template usage** for dynamic configuration
- **Monitor matching** by name and description
- **Tagged monitors** for template access

## Usage

### Command Line Options

```bash
hyprdynamicmonitors [options]

Options:
  -config string
        Path to configuration file (default "$HOME/.config/hyprdynamicmonitors/config.toml")
  -dry-run
        Show what would be done without making changes
  -debug
        Enable debug logging
  -verbose
        Enable verbose logging with caller information
  -version
        Show version information
```

### Signals

- **SIGUSR1**: Trigger manual configuration update
- **SIGTERM/SIGINT**: Graceful shutdown

### Running as a Service

The service is designed to fail fast and be managed by systemd. Example service file:

```ini
[Unit]
Description=Hyprland Dynamic Monitor Manager
After=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/bin/hyprdynamicmonitors
Restart=always
RestartSec=5
Environment=XDG_RUNTIME_DIR=%i

[Install]
WantedBy=default.target
```

### Wiring to Hyprland
Simply source from your desired destination (the one used in `General` section in the config), e.g.:
```hyprland.conf
source = ~/.config/hypr/config.d/dynamic-monitors.conf
```

### Development

```bash
# Build
go build -o hyprdynamicmonitors ./cmd

# Run with debug logging
./hyprdynamicmonitors -debug -config examples/basic/config.toml

# Dry run to test configuration
./hyprdynamicmonitors -dry-run -config examples/basic/config.toml
```

## Architecture

### Event-Driven Design

The service consists of several components that communicate via channels:

1. **Monitor Detector**: Listens to Hyprland IPC for display events
2. **Power Detector**: Monitors UPower D-Bus signals for power state changes

### Flow

1. **Initial Setup**: Detect current monitors and power state, apply best matching profile
2. **Event Monitoring**: Listen for monitor connection/disconnection and power state changes
3. **Debounced Updates**: Batch rapid changes (1.5s debounce) to avoid configuration thrashing
4. **Profile Matching**: Score all profiles against current conditions, select highest scoring
5. **Configuration Apply**: Generate and write/link the selected profile configuration

### Error Handling

- **Fail Fast**: The service exits on critical errors (D-Bus unavailable, invalid configuration, invalid events)
- **Comprehensive Logging**: Debug and info level logging for troubleshooting

## Requirements

- **Hyprland**: Window manager with IPC support
- **UPower**: Power state monitoring (optional, can be disabled)
- **D-Bus**: System bus access for power events (optional, see above)

### Development
- **ASDF**: Dowlnoads and sets up the developer environment

## License

[License information to be added]
