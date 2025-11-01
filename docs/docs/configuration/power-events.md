---
sidebar_position: 4
---

# Power Events

Power state monitoring uses D-Bus to listen for UPower events. This allows profiles and templates to react to AC/battery power changes.

![Power events](/previews/power_events.gif)

## Disabling Power Events

To disable power state monitoring entirely:

```bash
hyprdynamicmonitors run --disable-power-events
```

When disabled:
- The system defaults to `AC` power state
- No power events will be delivered
- No D-Bus connection will be made

## Enabling Power Events

Power events are enabled by default on laptops. They are disabled on desktops (the current chassis type is pulled from `/sys/class/dmi/id/chassis_type`).
If you want to enable them anyway, run, e.g.:

```bash
hyprdynamicmonitors run --disable-power-events=false
```

## Default Configuration

By default, the service listens for D-Bus signals:
- **Signal**: `org.freedesktop.DBus.Properties.PropertiesChanged`
- **Interface**: `org.freedesktop.DBus.Properties`
- **Member**: `PropertiesChanged`
- **Path**: `/org/freedesktop/UPower/devices/line_power_ACAD`

### Monitoring Power Events

You can monitor power events with:

```bash
gdbus monitor -y -d org.freedesktop.UPower | grep -E "PropertiesChanged|Device(Added|Removed)"
```

Example output:
```
/org/freedesktop/UPower/devices/line_power_ACAD: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower.Device', {'UpdateTime': <uint64 1756242314>, 'Online': <true>}, @as [])
```

## Querying Power Status

On each event, the current power status is queried. Equivalent command:

```bash
dbus-send --system --print-reply --dest=org.freedesktop.UPower \
  /org/freedesktop/UPower/devices/line_power_ACAD \
  org.freedesktop.DBus.Properties.Get \
  string:org.freedesktop.UPower.Device string:Online
```

## Custom D-Bus Configuration

You can customize which D-Bus signals to monitor and how to query power status.

### Custom Signal Match Rules

```toml
[power_events]
[[power_events.dbus_signal_match_rules]]
interface = "org.freedesktop.DBus.Properties"
member = "PropertiesChanged"
object_path = "/org/freedesktop/UPower/devices/line_power_ACAD"
# sender = "org.freedesktop.UPower"  # Optional: specific sender
```

You can add multiple match rules to listen for different events like `DeviceAdded` and `DeviceRemoved`.

### Custom Signal Filters

Filter received events by name to avoid noisy signals:

```toml
[[power_events.dbus_signal_receive_filters]]
name = "org.freedesktop.DBus.Properties.PropertiesChanged"
```

### Custom UPower Query

For non-standard power managers:

```toml
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

This is equivalent to:
```bash
dbus-send --system --print-reply --dest=org.freedesktop.UPower \
  /org/freedesktop/UPower \
  org.freedesktop.DBus.Properties.Get \
  string:org.freedesktop.UPower string:OnBattery
```

:::warning
The above query configuration is shown for reference only and is not recommended for production use. Stick with the default configuration unless you have a custom power management setup.
:::

## Leave Empty Token

To explicitly remove default values from D-Bus match rules:

```toml
[[power_events.dbus_signal_match_rules]]
interface = "leaveEmptyToken"  # Removes interface match
member = "PropertiesChanged"
object_path = "/custom/path"
```

## Session Bus vs System Bus

By default, the service connects to the system bus. To use the session bus:

```bash
hyprdynamicmonitors run --connect-to-session-bus
```

This is useful if your power line events are exposed in your user session bus instead of the system bus.

## Using Power State

Power state is available in templates via functions:

```go
{{if isOnAC}}
monitor=eDP-1,2880x1920@120,0x0,2.0
{{else}}
monitor=eDP-1,1920x1080@60,0x0,2.0
{{end}}
```

See [Templates](../advanced/templates) for more details.

## Behavior

- Power status changes are only propagated when the state actually changes
- Template/link replacement only occurs when file contents differ
- This prevents unnecessary Hyprland reloads when nothing has changed
