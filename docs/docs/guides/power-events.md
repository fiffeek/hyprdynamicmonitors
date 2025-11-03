---
sidebar_position: 2
---

# Running with power events

## What are power events?

Power events allow `hyprdynamicmonitors` to automatically switch monitor profiles based on your device's power state (AC power vs. battery). This is useful for scenarios like:
- Switching to a power-saving profile when on battery
- Enabling high refresh rates only when plugged in
- Adjusting monitor scaling based on power state

## Prerequisites

To use power events, you need **UPower** installed and running on your system.

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

## Do I need power events?

**Laptop users:** Yes, if you want to:
- Save battery by reducing refresh rates or resolution when unplugged
- Use high-performance settings (e.g., 144Hz) only when on AC power
- Automatically switch to a minimal monitor setup on battery

**Desktop users:** Usually no, since desktops are always plugged in. Power events are disabled by default on desktops. However, if you have a UPS (battery backup), you might want power events enabled to switch to a power-efficient configuration during power outages.

**Only care about connected displays?** If you only want profiles to switch based on which monitors are physically connected (not power state), you can [disable power events entirely](#disabling-power-events).

## Automatic detection

`hyprdynamicmonitors` automatically detects your device type and configures power events accordingly:
- **On laptops:** Power events are enabled by default
- **On desktops:** Power events are disabled by default

For laptop users (or desktop users who explicitly enable power events), the tool will automatically detect your UPower device by matching against [a well-known list of devices](https://github.com/fiffeek/hyprdynamicmonitors/blob/main/internal/utils/power.go#L14). This means **most laptop users don't need any configuration**.

## Verifying automatic detection

To check that automatic detection is working correctly, run:

```bash
hyprdynamicmonitors run --run-once --dry-run
```

Look for a line like this in the output:
```
INFO[...] Inferred power line    power_line="/org/freedesktop/UPower/devices/line_power_ACAD"
```

(The exact device name may vary, e.g., `line_power_AC`, `line_power_ADP1`, etc.)

**If you see this message:** Everything is working correctly and you don't need to configure anything.

**If you don't see this message:** Automatic detection failed. This could mean:
- Power events are disabled (check if you're on desktop or using `--disable-power-events`)
- Your power device isn't in the recognized list (see [Custom UPower device](#custom-upower-device))
- UPower isn't running or isn't installed (see [Prerequisites](#prerequisites))

---

## Advanced configuration

:::tip Important
When using `--disable-power-events`, ensure both `run` and `tui` commands use the same flag. Since profiles may have power state conditions (e.g., require AC power), the TUI needs to know the current power state to correctly display which profile would be active.
:::

### Custom UPower device

If automatic detection fails or detects the wrong device, you can manually specify which UPower device to use. First, list your available power devices:

```bash
upower -e
```

Then override the power device in your configuration file at `~/.config/hyprdynamicmonitors/config.toml`:
```toml
[power_events]

[power_events.dbus_query_object]
path = "/org/freedesktop/UPower/devices/line_power_XYZ"

[[power_events.dbus_signal_match_rules]]
object_path = "/org/freedesktop/UPower/devices/line_power_XYZ"
```

Replace `/org/freedesktop/UPower/devices/line_power_XYZ` with your actual power device path from `upower -e`.

For more configuration options, see the [Power Events configuration section](../configuration/power-events.md).

### Enabling power events on desktops

Power events are disabled by default on desktops, but you can explicitly enable them. Common reasons to enable:
- Your desktop has a UPS (battery backup) and you want power-saving profiles during outages
- You're testing power-related profile configurations

```bash
hyprdynamicmonitors run --disable-power-events=false
```

### Disabling power events

If you only want to switch profiles based on connected monitors (not power state), you can disable power events:

```bash
hyprdynamicmonitors run --disable-power-events
```

---

## Troubleshooting

### Error: "Object does not exist at path /org/freedesktop/UPower/devices/line_power_ACAD"

This error means `hyprdynamicmonitors` couldn't find the expected power device.

**Quick fix - Disable power events** (if you don't need them):
```bash
hyprdynamicmonitors run --disable-power-events
```

**Full fix - Configure a custom power device** (if you need power events):

1. Check if UPower is running:
   ```bash
   systemctl status upower
   ```
   If not running, start it: `sudo systemctl start upower`

2. List your available power devices:
   ```bash
   upower -e
   ```
   Look for a line containing `line_power` (ignore battery/display devices). Example: `/org/freedesktop/UPower/devices/line_power_AC`

3. Configure the correct device following the [Custom UPower device section](#custom-upower-device) above.

**Help improve automatic detection:** If you have a power device that should be recognized automatically, please contribute by:
- Submitting a PR to add your device to [the recognized device list](https://github.com/fiffeek/hyprdynamicmonitors/blob/main/internal/utils/power.go#L14)
- Opening a [GitHub issue](https://github.com/fiffeek/hyprdynamicmonitors/issues) with your device information

### Power events aren't switching profiles

If automatic detection succeeded but profiles aren't switching when you plug/unplug:

1. **Verify your profile has power state conditions:**
   Check your `config.toml` to ensure profiles specify power requirements:
   ```toml
   [profiles.high_performance]
   # ... monitor conditions ...

   [[profiles.high_performance.conditions.required_power_states]]
   state = "AC"  # Only active when on AC power
   ```

2. **Check current power state:**
   ```bash
   upower -i /org/freedesktop/UPower/devices/line_power_ACAD
   ```
   (Replace `line_power_ACAD` with your actual device from the "Inferred power line" message)

   Look for the `online` property (yes = AC power, no = battery)

3. **Test with dry-run:**
   ```bash
   # Unplug your laptop, then run:
   hyprdynamicmonitors run --run-once --dry-run

   # Plug it back in, then run:
   hyprdynamicmonitors run --run-once --dry-run
   ```
   The output should show different profiles being selected based on power state.

4. **Verify daemon is using the same configuration:**
   If you're running the daemon with `exec-once`, make sure you're not passing conflicting flags like `--disable-power-events`.

### UPower is not installed

If you get errors about UPower not being available, see the [Prerequisites](#prerequisites) section for installation instructions.
