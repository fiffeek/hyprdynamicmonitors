---
sidebar_position: 4
---

# Running with Lid Events

## What are lid events?

Lid events allow `hyprdynamicmonitors` to automatically switch monitor profiles based on your laptop's lid state (open vs. closed). This is useful for scenarios like:
- Using a docked laptop with the lid closed and external monitors only
- Automatically disabling the laptop's internal display when the lid is closed
- Different monitor configurations for clamshell mode vs. open laptop mode

## Do I need lid events?

**Laptop users with external monitors:** Yes, if you want to:
- Use your laptop in clamshell mode (lid closed, external displays only)
- Automatically disable/enable the laptop screen based on lid state
- Have different monitor layouts when docked vs. mobile

**Laptop users without external monitors:** Usually no, since closing the lid typically suspends the laptop anyway.

**Desktop users:** No, desktops don't have lids.

**Only care about connected displays?** If you only want profiles to switch based on which monitors are physically connected (not lid state), you don't need to enable lid events.

## Prerequisites

To use lid events, you need **UPower** installed and running on your system.

Check if UPower is running:
```bash
systemctl status upower
```

If it's not installed, install it using your package manager:
```bash
# Arch Linux
sudo pacman -S upower

# Debian/Ubuntu
sudo apt install upower

# Fedora
sudo dnf install upower
```

After installation, enable and start the service:
```bash
sudo systemctl enable --now upower
```

## Enabling lid events

Lid events are disabled by default. You need to pass `--enable-lid-events` to `run` or `tui` commands to enable them:

```bash
hyprdynamicmonitors run --enable-lid-events
```

## Automatic configuration

`hyprdynamicmonitors` has sane defaults for lid events. This is what you get out of the box:
```toml
[lid_events]

[lid_events.dbus_query_object]
destination = "org.freedesktop.UPower"
path = "/org/freedesktop/UPower"
method = "org.freedesktop.DBus.Properties.Get"
expected_lid_closing_value = "true"

[[lid_events.dbus_query_object.args]]
arg = "org.freedesktop.UPower"

[[lid_events.dbus_query_object.args]]
arg = "LidIsClosed"

[[lid_events.dbus_signal_match_rules]]
interface = "org.freedesktop.DBus.Properties"
object_path = "/org/freedesktop/UPower"
member = "PropertiesChanged"

[[lid_events.dbus_signal_receive_filters]]
name = "org.freedesktop.DBus.Properties.PropertiesChanged"
body = "LidIsClosed"
```

This works on most systems, and you can validate the configuration before using it:

**1) Validate the query**

Run the following command:
```bash
gdbus call --system --dest org.freedesktop.UPower --object-path /org/freedesktop/UPower \
  --method org.freedesktop.DBus.Properties.Get org.freedesktop.UPower LidIsClosed

# alternatively use dbus-send
dbus-send --system --print-reply \
  --dest=org.freedesktop.UPower /org/freedesktop/UPower \
  org.freedesktop.DBus.Properties.Get \
  string:org.freedesktop.UPower string:LidIsClosed
```

Depending on your current lid state, this should output `(<false>,)` or `(<true>,)`.

**2) Validate the events**

Run the following command in the background:
```bash
gdbus monitor --system --dest org.freedesktop.UPower --object-path /org/freedesktop/UPower
```

Then monitor the output while closing/opening your lid. You should see:
```bash
/org/freedesktop/UPower: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower', {'LidIsClosed': <true>}, @as [])
/org/freedesktop/UPower: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower', {'LidIsClosed': <false>}, @as [])
```

If both of these pass (and they should on most systems), then the default configuration is enough and you can just pass `--enable-lid-events`.

**(optional) Smoke test**

You can run:
```bash
hyprdynamicmonitors run --debug --run-once --dry-run --enable-lid-events
```
This command will fail if the configuration is invalid.


## Advanced configuration

:::tip Important
When using `--enable-lid-events`, ensure both `run` and `tui` commands use the same flag. Since profiles may have lid state conditions (e.g., require `Closed` lid), the TUI needs to know the current lid state to correctly display which profile would be active.

The easiest way is to create an alias in your `~/.bashrc` (or `~/.zshrc`, etc.):
```bash
alias hdm="hyprdynamicmonitors --enable-lid-events"
```
:::

### Custom D-Bus query

You can introspect your `UPower` device properties for lid state:
```bash
dbus-send --system --print-reply --dest=org.freedesktop.UPower /org/freedesktop/UPower \
  org.freedesktop.DBus.Introspectable.Introspect | grep Lid
```

Example output:
```xml
<property type="b" name="LidIsClosed" access="read"/>
<property type="b" name="LidIsPresent" access="read"/>
```

This will tell you if you need to tweak the queries. For example, you might see `LidIsOpen` instead of `LidIsClosed`. Let's assume that scenario for this example.

#### Custom configuration

The custom configuration will be overlaid on the default. In this case, we expect the lid closing value to be `false`, since we're querying `LidIsOpen` instead of `LidIsClosed`.
```toml
[lid_events]

[lid_events.dbus_query_object]
expected_lid_closing_value = "false"

[[lid_events.dbus_query_object.args]]
arg = "org.freedesktop.UPower"

[[lid_events.dbus_query_object.args]]
arg = "LidIsOpen"

[[lid_events.dbus_signal_receive_filters]]
name = "org.freedesktop.DBus.Properties.PropertiesChanged"
body = "LidIsOpen"
```

To configure `dbus_signal_receive_filters`, verify the signal format by running:
```bash
gdbus monitor --system --dest org.freedesktop.UPower --object-path /org/freedesktop/UPower
```

For this example, the above configuration would match:
```bash
/org/freedesktop/UPower: org.freedesktop.DBus.Properties.PropertiesChanged ('org.freedesktop.UPower', {'LidIsOpen': <true>}, @as [])
```

## Using lid events

You can match profiles based on lid state (closed or open) and use template variables in your configuration. See the [lid events configuration section](../configuration/lid-events.md#using-lid-state) for more details.

---

## Troubleshooting

### Error: "LidIsClosed property not found" or lid detection fails

This error means `hyprdynamicmonitors` couldn't find the expected lid property in UPower.

**Quick fix - Check if your laptop has a lid:**
```bash
upower -d | grep -i lid
```

If you see output like:
```
LidIsPresent: yes
LidIsClosed: no
```
Then your system supports lid detection.

**If no lid properties are found:**

Some laptops don't expose lid state through UPower. Check alternative methods:

1. **Check `/proc/acpi/button/lid/`:**
   ```bash
   cat /proc/acpi/button/lid/LID0/state
   ```
   If this shows lid state, your system uses ACPI but UPower might not be configured correctly.

2. **Verify UPower version:**
   ```bash
   upower --version
   ```
   Older versions might not support lid detection. Try updating UPower.

3. **Check system logs when closing the lid:**
   ```bash
   journalctl -f
   ```
   Then close/open your lid and look for relevant events.

**If you see a different property name:**

Follow the [Custom D-Bus query](#custom-d-bus-query) section to configure the correct property name (e.g., `LidIsOpen` instead of `LidIsClosed`).

### Lid events aren't switching profiles

If lid detection is working but profiles aren't switching when you close/open the lid:

1. **Verify your profile has lid state conditions:**
   Check your `config.toml` to ensure profiles specify lid requirements:
   ```toml
   [profiles.docked]
   # ... monitor conditions ...

   [[profiles.docked.conditions.required_lid_states]]
   state = "Closed"  # Only active when lid is closed
   ```

2. **Check current lid state:**
   ```bash
   gdbus call --system --dest org.freedesktop.UPower --object-path /org/freedesktop/UPower \
     --method org.freedesktop.DBus.Properties.Get org.freedesktop.UPower LidIsClosed
   ```
   This should return `(<true>,)` when lid is closed, `(<false>,)` when open.

3. **Test with dry-run:**
   ```bash
   # Close your laptop lid, then run:
   hyprdynamicmonitors run --run-once --dry-run --enable-lid-events

   # Open it, then run:
   hyprdynamicmonitors run --run-once --dry-run --enable-lid-events
   ```
   The output should show different profiles being selected based on lid state.

4. **Verify daemon is using lid events:**
   If you're running the daemon with `exec-once`, make sure you're passing `--enable-lid-events`:
   ```
   exec-once = hyprdynamicmonitors run --enable-lid-events
   ```

5. **Enable debug logging:**
   ```bash
   hyprdynamicmonitors run --debug --enable-lid-events
   ```
   This will show detailed information about lid state changes and profile matching.

### Lid events trigger on boot/resume from suspend

Some systems send lid events during boot or when resuming from suspend, which might cause unwanted profile switches.

**Workaround:**

Add a small delay in your Hyprland config:
```
exec-once = sleep 2 && hyprdynamicmonitors run --enable-lid-events
```

This gives the system time to stabilize before monitoring lid events.

### UPower is not installed

If you get errors about UPower not being available, see the [Prerequisites](#prerequisites) section for installation instructions.

### Laptop suspends before profile can switch

If your laptop suspends immediately when closing the lid (before the profile switches), you may need to adjust your power management settings.

**For systemd-based systems:**

Edit `/etc/systemd/logind.conf`:
```ini
HandleLidSwitch=ignore
HandleLidSwitchExternalPower=ignore
HandleLidSwitchDocked=ignore
```

Then restart the service:
```bash
sudo systemctl restart systemd-logind
```

**Warning:** This disables automatic suspend on lid close. You'll need to manage suspend manually or through other tools.

Alternatively, configure specific behavior for docked vs. undocked in `logind.conf`:
```ini
HandleLidSwitch=suspend
HandleLidSwitchDocked=ignore
```

This suspends only when undocked, allowing profile switching when docked with external monitors.
