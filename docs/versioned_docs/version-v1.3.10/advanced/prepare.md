---
sidebar_position: 4
---

# What is `hyprdynamicmonitors prepare`?

The `prepare` command cleans up your monitor configuration file before Hyprland starts, preventing the "no active displays" issue that can leave Hyprland in an unusable state.

## The Problem

Hyprland currently has a limitation where it cannot start if all monitors are disabled at launch. You can read more about it [here](https://github.com/hyprwm/Hyprland/discussions/12325#discussioncomment-15004584).

This affects `hyprdynamicmonitors` users who disable monitors in their profiles. For example:

1. You're using external displays with your laptop screen disabled (`monitor=eDP-1,disable`)
2. You power off and unplug all external displays
3. You try to start your laptop the next day
4. **Result**: Hyprland shows a blank screen and won't start because the configuration file still has the laptop screen disabled, even though `hyprdynamicmonitors` would correctly reconfigure the monitors once running

## The Solution

The `prepare` command removes all `monitor=...,disable` lines from your destination file (`config.general.destination`) before Hyprland starts. This ensures:

1. Hyprland starts successfully with a clean monitor configuration
2. Once Hyprland is running, `hyprdynamicmonitors run` (if enabled) automatically applies the correct profile
3. If not using the daemon, you can adjust settings manually in the TUI

:::info
This is only relevant if you disable monitors in your profiles. If you don't use `monitor=...,disable` in any configuration, you don't need the prepare command.
:::

For command-line options and examples, see the [prepare command documentation](../usage/commands#prepare).

## How to run `hyprdynamicmonitors prepare`

### Option 1: Using systemd (Recommended)

If you're using the `hyprdynamicmonitors.service` systemd service, enable the prepare service:

```bash
systemctl --user enable hyprdynamicmonitors-prepare.service
```

The service definition ensures it runs before the graphical session starts. The service has a 3-second timeout, so Hyprland will start even if the prepare command takes longer than expected.

:::tip
See the [systemd documentation](./systemd) for complete setup instructions, including both the daemon and prepare services.
:::

### Option 2: Manual execution (Non-systemd)

If you're running Hyprland manually (e.g., from TTY) or using `exec-once` for `hyprdynamicmonitors`, ensure `prepare` runs before Hyprland starts.

**Using a shell alias:**
```bash
# Add to ~/.bashrc or ~/.zshrc
alias hyprstart="hyprdynamicmonitors prepare && Hyprland"
```

**In an existing startup script:**
```bash
#!/bin/bash
# Your existing setup commands
killall waybar  # example cleanup

# Add prepare before launching Hyprland
hyprdynamicmonitors prepare
Hyprland
```

**Direct command:**
```bash
hyprdynamicmonitors prepare && Hyprland
```

You might still, in most cases, be able to just enable the `systemd` `hyprdynamicmonitors-prepare` service since it runs on boot (pulled by `default.target`):
```bash
systemctl --user enable hyprdynamicmonitors-prepare.service
```
But due to relying on timing (you entering the Hyprland session after this runs) adding it to your startup routine is more robust.

:::info
The prepare command is fast (typically completes in milliseconds) and safe to run multiple times. It only modifies your destination file when it finds `monitor=...,disable` lines to remove.
:::

## See Also

- [Setup Approaches](../quickstart/setup-approaches) - Initial setup guide with prepare integration
- [CLI Commands](../usage/commands#prepare) - Command-line options and examples
- [Running with systemd](./systemd) - Complete systemd service setup
