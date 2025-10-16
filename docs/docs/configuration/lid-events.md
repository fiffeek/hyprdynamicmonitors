---
sidebar_position: 7
---

# Lid Events

Lid state monitoring uses D-Bus to listen for UPower lid events. This feature is optional and needs to be explicitly enabled.

## Enabling Lid Events

To enable lid state monitoring:

```bash
hyprdynamicmonitors run --enable-lid-events
```

When disabled (default):
- The system defaults to `UNKNOWN` lid state
- No lid events will be delivered
- No D-Bus connection for lid events will be made

## Default Configuration

By default, when enabled, the service listens for D-Bus signals:
- **Signal**: `org.freedesktop.DBus.Properties.PropertiesChanged`
- **Interface**: `org.freedesktop.DBus.Properties`
- **Member**: `PropertiesChanged`
- **Path**: `/org/freedesktop/UPower`

### Monitoring Lid Events

You can monitor lid events with:

```bash
gdbus monitor --system --dest org.freedesktop.UPower --object-path /org/freedesktop/UPower
```

Example output:
```
/org/freedesktop/UPower: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower', {'LidIsClosed': <true>}, @as [])
/org/freedesktop/UPower: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower', {'LidIsClosed': <false>}, @as [])
```

## Querying Lid Status

On each event, the current lid status is queried. Equivalent command:

```bash
dbus-send --system --print-reply \
  --dest=org.freedesktop.UPower /org/freedesktop/UPower \
  org.freedesktop.DBus.Properties.Get \
  string:org.freedesktop.UPower string:LidIsClosed
```

## Custom D-Bus Configuration

You can customize which D-Bus signals to monitor for lid events:

```toml
[lid_events]
[[lid_events.dbus_signal_match_rules]]
interface = "org.freedesktop.DBus.Properties"
member = "PropertiesChanged"
object_path = "/org/freedesktop/UPower"

[[lid_events.dbus_signal_receive_filters]]
name = "org.freedesktop.DBus.Properties.PropertiesChanged"
```

See the [lid-states example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/lid-states) for a complete configuration.

## Receive Filters

By default, the service matches:
- `org.freedesktop.DBus.Properties.PropertiesChanged` name
- `LidIsClosed` body

This prevents noisy signals. Lid status changes are only propagated when the state actually changes.

## Using Lid State

Lid state is available in templates via functions:

```go
{{if isLidClosed}}
# Disable laptop screen when lid is closed
monitor=eDP-1,disable
{{else}}
monitor=eDP-1,2880x1920@120,0x0,2.0
{{end}}
```

Available template functions:
- `isLidClosed` - Returns true if the lid is closed
- `isLidOpened` - Returns true if the lid is opened
- `.LidState` - Returns the lid state string (`"Closed"`, `"Opened"`, or `"UNKNOWN"`)

## Lid State Values

The `.LidState` variable can have these values:
- `"UNKNOWN"` - Lid events are disabled or state cannot be determined
- `"Closed"` - Lid is closed
- `"Opened"` - Lid is open

## Common Use Cases

### Disable Laptop Screen When Docked with Lid Closed

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
{{- $external := index .MonitorsByTag "external" -}}

{{if isLidClosed}}
monitor={{$laptop.Name}},disable
{{else}}
monitor={{$laptop.Name}},2880x1920@120,0x0,2.0
{{end}}

monitor={{$external.Name}},preferred,auto,1
```

### Different Layouts Based on Lid State

```go
{{if isLidClosed}}
# Closed lid: external monitor only
monitor=eDP-1,disable
monitor=DP-1,3840x2160@60,0x0,1
{{else}}
# Open lid: dual monitor setup
monitor=eDP-1,2880x1920@120,0x0,2.0
monitor=DP-1,3840x2160@60,2880x0,1
{{end}}
```

## See Also

- [Templates](../advanced/templates) - Template syntax and variables
- [Power Events](./power-events) - Similar D-Bus event system
- [Examples - Lid States](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/lid-states)
