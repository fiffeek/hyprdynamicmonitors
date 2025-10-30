---
sidebar_position: 1
---

# Templates

Templates use Go template syntax for dynamic configuration generation based on connected monitors, power state, and lid state.

## Template Variables

The following variables are available in all templates:

### .PowerState

Current power state: `"AC"` or `"BAT"`

```go
{{.PowerState}}  # Returns "AC" or "BAT"
```

### .LidState

Current lid state: `"UNKNOWN"`, `"Closed"`, or `"Opened"`

Requires `--enable-lid-events` flag to be set.

```go
{{.LidState}}  # Returns "UNKNOWN", "Closed", or "Opened"
```

### .Monitors

Array of all connected monitors.

Each monitor object has the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | Monitor connector name (e.g., "eDP-1", "DP-1") |
| `ID` | int | Monitor ID |
| `Description` | string | Monitor model/manufacturer string |
| `Disabled` | bool | Whether the monitor is disabled |
| `AvailableModes` | []string | List of available display modes |
| `Mirror` | string | Name of monitor this one mirrors (empty if not mirroring) |
| `CurrentFormat` | string | Current display mode/format |
| `DpmsStatus` | bool | DPMS (power management) status |
| `ActivelyTearing` | bool | Whether the monitor is actively tearing |
| `DirectScanoutTo` | string | Direct scanout target |
| `Solitary` | string | Solitary mode status |

```go
{{range .Monitors}}
# Monitor: {{.Name}} ({{.Description}})
# Current format: {{.CurrentFormat}}
# Disabled: {{.Disabled}}
{{end}}
```

### .ExtraMonitors

Array of connected monitors **not defined** in the profile's required monitors.

Each monitor object has the same fields as `.Monitors` (see above).

Useful for disabling unexpected monitors:

```go
{{range .ExtraMonitors}}
monitor={{.Name}},disable
{{end}}
```

### .RequiredMonitors

Array of connected monitors that **are defined** in the profile's required monitors.

Each monitor object has the same fields as `.Monitors` (see above).

```go
{{range .RequiredMonitors}}
# Required monitor: {{.Name}} ({{.Description}})
monitor={{.Name}},{{.CurrentFormat}},auto,1
{{end}}
```

### .MonitorsByTag

Map of tagged monitors (monitor_tag -> monitor object).

Each monitor object has the same fields as `.Monitors` (see above).

This is the most convenient way to reference specific monitors in templates:

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
{{- $external := index .MonitorsByTag "external" -}}

monitor={{$laptop.Name}},2880x1920@120,0x0,2.0
monitor={{$external.Name}},preferred,auto,1
# External monitor description: {{$external.Description}}
```

## Template Functions

### isOnBattery

Returns true if on battery power.

```go
{{if isOnBattery}}
# Low power configuration
monitor=eDP-1,1920x1080@60,0x0,2.0
{{end}}
```

### isOnAC

Returns true if on AC power.

```go
{{if isOnAC}}
# High performance configuration
monitor=eDP-1,2880x1920@120,0x0,2.0
{{end}}
```

### powerState

Returns the current power state string (`"AC"` or `"BAT"`).

```go
# Current power state: {{powerState}}
```

### isLidClosed

Returns true if the lid is closed.

Requires `--enable-lid-events` flag.

```go
{{if isLidClosed}}
monitor=eDP-1,disable
{{end}}
```

### isLidOpened

Returns true if the lid is opened.

Requires `--enable-lid-events` flag.

```go
{{if isLidOpened}}
monitor=eDP-1,2880x1920@120,0x0,2.0
{{end}}
```

## Static Template Values

Define custom values that are available in templates:

### Global Static Values

```toml
[static_template_values]
default_vrr = "1"
default_res = "2880x1920"
refresh_rate_high = "120.00000"
refresh_rate_low = "60.00000"
```

### Per-Profile Static Values

```toml
[profiles.laptop_only.static_template_values]
default_vrr = "0"        # Override global value
battery_scaling = "1.5"  # Profile-specific value
```

### Using in Templates

```go
monitor=eDP-1,{{.default_res}}@{{.refresh_rate_high}},0x0,1,vrr,{{.default_vrr}}
```

## Complete Examples

### Power-Aware Dual Monitor

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
{{- $external := index .MonitorsByTag "external" -}}

{{if isOnAC}}
# AC Power: High performance
monitor={{$laptop.Name}},2880x1920@120,0x0,2.0
monitor={{$external.Name}},3840x2160@60,2880x0,1
{{else}}
# Battery: Power saving
monitor={{$laptop.Name}},1920x1080@60,0x0,2.0
monitor={{$external.Name}},2560x1440@60,1920x0,1
{{end}}
```

### Lid-Aware Docked Setup

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
{{- $external := index .MonitorsByTag "external" -}}

{{if isLidClosed}}
# Lid closed: external monitor only
monitor={{$laptop.Name}},disable
monitor={{$external.Name}},preferred,0x0,1
{{else}}
# Lid open: dual monitor
monitor={{$laptop.Name}},2880x1920@120,0x0,2.0
monitor={{$external.Name}},preferred,{{$laptop.Width}}x0,1
{{end}}
```

### Disable Extra Monitors

```go
{{- range .RequiredMonitors}}
monitor={{.Name}},preferred,auto,1
{{- end}}

{{- range .ExtraMonitors}}
# Disable unexpected monitor
monitor={{.Name}},disable
{{- end}}
```

### Combined Power and Lid State

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}

{{if isLidClosed}}
monitor={{$laptop.Name}},disable
{{else if isOnAC}}
monitor={{$laptop.Name}},2880x1920@120,0x0,2.0,vrr,1
{{else}}
monitor={{$laptop.Name}},1920x1080@60,0x0,2.0,vrr,0
{{end}}
```

## Whitespace Control

Go templates support whitespace trimming using `-` in template delimiters:

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
# This removes trailing whitespace/newlines after the action

{{- $external := index .MonitorsByTag "external" }}
# This keeps the newline after the action
```

:::caution Common Issue
Using `-}}` at the end of the last variable definition before your configuration can cause the next line to merge with the comment above it. See [FAQ - Comments in Templates](../faq#why-are-my-comments-becoming-part-of-the-configuration-in-templates) for solutions.
:::

## Testing Templates

After creating or modifying a template, test it by reloading the configuration:

```bash
# Trigger a configuration reload
kill -SIGHUP $(pidof hyprdynamicmonitors)

# Or if using systemd
systemctl --user reload hyprdynamicmonitors
```

Then check:
- **Service logs** to see if the template rendered successfully:
  ```bash
  # If using systemd
  journalctl --user -u hyprdynamicmonitors -f

  # Or with --debug flag for more details
  hyprdynamicmonitors --debug run
  ```

- **Destination file** to verify the rendered output:
  ```bash
  cat ~/.config/hypr/monitors.conf
  ```

### Creating Initial Templates

Use the `freeze` command to capture your current setup as a starting point:

```bash
hyprdynamicmonitors freeze --profile-name my-setup
```

This generates a template from your current monitor configuration that you can then customize.

## See Also

- [Profiles](../configuration/profiles) - Configure profiles with templates
- [Monitor Matching](../configuration/monitor-matching) - Using monitor tags
- [Examples - Template Variables](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/template-variables)
- [Examples - Disable Monitors](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/disable-monitors)
