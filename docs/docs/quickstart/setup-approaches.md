---
sidebar_position: 2
---

# Setup

Choose your preferred setup approach based on whether you want to run the daemon for automatic profile switching.

## Choose Your Approach

### Option A: TUI + Daemon (Recommended for Most Users)

**Best for:** Users who want automatic profile switching and an easy visual setup.

The TUI (Terminal User Interface) provides a visual way to configure monitors and automatically creates the daemon configuration for you.

1. **Launch the TUI** (this automatically creates a default configuration):
   ```bash
   hyprdynamicmonitors tui
   ```

2. **Configure your monitors** in the Monitors view:
   - Use arrow keys to select a monitor
   - Press `Enter` to edit it
   - Adjust position, resolution, scale, etc.
   - Press `A` to apply and test the configuration

3. **Save as a profile** by switching to the Profile view:
   - Press `Tab` to switch to HDM Profile view
   - Press `n` to create a new profile
   - Enter a profile name (e.g., "laptop-only")
   - The profile is automatically saved to your configuration

4. **Source the configuration** in your `~/.config/hypr/hyprland.conf`:
   ```conf
   source = ~/.config/hypr/monitors.conf
   ```

5. **Run the daemon** for automatic profile switching:
   ```conf
   # Add to ~/.config/hypr/hyprland.conf
   exec-once = hyprdynamicmonitors run
   ```

:::info
  Or use systemd (recommended for production), instead of `exec-once`: see [Running with systemd](../advanced/systemd).
:::

:::note
**What happens:** The daemon monitors for display and power changes, automatically switching between your saved profiles.
:::

---

### Option B: Manual Configuration + Daemon

**Best for:** Users who prefer manual configuration file editing.

1. **Create the configuration file** at `~/.config/hyprdynamicmonitors/config.toml`:

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

2. **Create the monitor configuration** at `~/.config/hyprdynamicmonitors/hyprconfigs/laptop.conf`:

   ```
   monitor=eDP-1,2880x1920@120.00000,0x0,2.0,vrr,1
   ```

   Replace with your actual monitor settings (check `hyprctl monitors`).

3. **Source the configuration** in your `~/.config/hypr/hyprland.conf`:

   ```conf
   source = ~/.config/hypr/monitors.conf
   ```

4. **Run the daemon**:
   ```conf
   # Add to ~/.config/hypr/hyprland.conf
   exec-once = hyprdynamicmonitors run
   ```

:::info
  Or use systemd (recommended for production), instead of `exec-once`: see [Running with systemd](../advanced/systemd).
:::

---

### Option C: TUI Only (No Daemon)

**Best for:** Users who want manual control without automatic profile switching.

1. **Launch the TUI**:
   ```bash
   hyprdynamicmonitors tui
   ```

2. **Configure monitors visually** and press `A` to apply changes.

3. **Changes are ephemeral** - they apply immediately to Hyprland but won't persist on restart or monitor changes.

:::info
Without the daemon, you won't get automatic profile switching based on connected monitors or power state.
:::

---

## Validation and Testing

Before running the daemon, validate your configuration:

```bash
# Validate configuration file
hyprdynamicmonitors validate

# Test what would happen without making changes
hyprdynamicmonitors run --dry-run

# Run once and exit (for testing)
hyprdynamicmonitors run --run-once
```

## Next Steps

- Learn about [CLI Commands](../usage/commands) for validation and testing
- Explore [Profiles](../configuration/profiles) for power state and multi-monitor setups
- Check out [Templates](../advanced/templates) for dynamic configurations
- See [Examples](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples) for advanced configurations
- Read the [TUI Guide](./tui) for detailed TUI usage
