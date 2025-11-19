---
sidebar_position: 3
---

# Setup

Choose your preferred setup approach based on whether you want to run the daemon for automatic profile switching.

## Understanding HDM flow

In essence, what `hyprdynamicmonitors` does is:
- Aggregate all system information about battery/lid states (optional)
- Aggregate information from Hyprland about the currently connected outputs (think: `hyprctl monitors all`)
- See what the user defined in `config.toml`
  - Iterate through user-defined profiles and find the best matching one
  - For that profile, see what template or static configuration it requires
  - Render that configuration to the `config.general.destination` location

This is a simplified view, but it gives you a good understanding of the system. In reality, it responds to events, aggregates state on startup, sends notifications, and may execute callbacks. See [Configuration](../configuration/overview.md) for more details.

## Choose Your Approach

:::tip
For detailed information about running with or without power events, see the [power events guide](../guides/power-events.md).
:::

:::warning
When running the daemon with power (default) or lid events enabled, ensure that `UPower` is installed and its service is running.
:::

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
   # Or use systemd
   systemctl --user daemon-reload
   systemctl --user enable --now hyprdynamicmonitors.service
   ```

6. **Set up the prepare command** to prevent startup issues:

   :::tip
   The `prepare` command removes disabled monitor entries from your configuration before Hyprland starts, preventing the "no active displays" issue. See [What is hyprdynamicmonitors prepare?](../advanced/prepare) for details.
   :::

   **If using systemd** (recommended):
   ```bash
   systemctl --user enable hyprdynamicmonitors-prepare.service
   ```

   **If running Hyprland manually** (e.g., from TTY with `Hyprland` command):
   ```bash
   # Add to your shell alias or startup script
   hyprdynamicmonitors prepare && Hyprland
   ```

   See the [prepare documentation](../advanced/prepare) for more setup options.

:::info
  Systemd services are recommended in favor of `exec-once`: see [Running with systemd](../advanced/systemd).
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

5. **Set up the prepare command** to prevent startup issues:

   :::tip
   The `prepare` command removes disabled monitor entries from your configuration before Hyprland starts, preventing the "no active displays" issue. See [What is hyprdynamicmonitors prepare?](../advanced/prepare) for details.
   :::

   **If using systemd** (recommended):
   ```bash
   systemctl --user enable hyprdynamicmonitors-prepare.service
   ```

   **If running Hyprland manually** (e.g., from TTY with `Hyprland` command):
   ```bash
   # Add to your shell alias or startup script
   hyprdynamicmonitors prepare && Hyprland
   ```

   See the [prepare documentation](../advanced/prepare) for more setup options.

:::info
  Or use systemd (recommended for production), instead of `exec-once`: see [Running with systemd](../advanced/systemd).
:::

---

### Option C: TUI Only (No Daemon)

**Best for:** Desktop users with stable monitor setups who prefer manual control.

1. **Launch the TUI**:
   ```bash
   hyprdynamicmonitors tui
   ```

2. **Configure and save profiles** in the TUI:
   - Adjust monitors in the Monitors view
   - Press `A` to test changes ephemerally (optional)
   - Switch to the Profile view with `Tab`
   - Press `n` to create a new profile or `a` to update an existing one
   - Press `R` to manually render the configuration to the destination file

3. **Source the configuration** in your `~/.config/hypr/hyprland.conf`:
   ```conf
   source = ~/.config/hypr/monitors.conf
   ```

4. **Set up the prepare command** to prevent startup issues:

   :::tip
   The `prepare` command removes disabled monitor entries from your configuration before Hyprland starts, preventing the "no active displays" issue. See [What is hyprdynamicmonitors prepare?](../advanced/prepare) for details.
   :::

   **If using systemd**:
   ```bash
   systemctl --user enable hyprdynamicmonitors-prepare.service
   ```

   **If running Hyprland manually** (e.g., from TTY with `Hyprland` command):
   ```bash
   # Add to your shell alias or startup script
   hyprdynamicmonitors prepare && Hyprland
   ```

   See the [prepare documentation](../advanced/prepare) for more setup options.

:::info
Without the daemon, you manage profile rendering manually using the `R` key in the Profile view. See the [Running without a daemon guide](../guides/running-without-daemon.md) for detailed workflows.
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
