---
sidebar_position: 6
---

# Callbacks

HyprDynamicMonitors supports custom user commands that are executed before and after profile configuration changes. These commands can be defined globally or per-profile.

## Global Callbacks

Define commands that run for all profile changes:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[general]
pre_apply_exec = "notify-send 'HyprDynamicMonitors' 'Switching monitor profile...'"
post_apply_exec = "notify-send 'HyprDynamicMonitors' 'Profile applied successfully'"
```

## Profile-Specific Callbacks

Override global callbacks for specific profiles:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[profiles.gaming_setup]
config_file = "hyprconfigs/gaming.conf"
pre_apply_exec = "notify-send 'Gaming Mode' 'Activating high-performance profile'"
post_apply_exec = "/usr/local/bin/gaming-mode-on.sh"
```

Profile-specific commands **override** global commands for that profile.

## Callback Types

### pre_apply_exec

Executed **before** the new monitor configuration is applied.

Use cases:
- Send notifications about upcoming changes
- Prepare the system for configuration change
- Save current state
- Close applications that might interfere

Example:
```toml
pre_apply_exec = "killall -SIGUSR1 waybar"  # Reload waybar before change
```

### post_apply_exec

Executed **after** the new monitor configuration is successfully applied.

Use cases:
- Send success notifications
- Trigger dependent scripts
- Reload related applications
- Adjust system settings based on new configuration

Example:
```toml
post_apply_exec = "hyprctl reload && systemctl --user restart waybar"
```

## Shell Execution

Commands are executed through `bash -c`, supporting:
- Shell features (pipes, redirects, environment variables)
- Multiple commands with `&&` or `;`
- Background processes with `&`

Examples:

```toml
# Multiple commands
post_apply_exec = "hyprctl reload && sleep 1 && notify-send 'Done'"

# Using pipes
post_apply_exec = "hyprctl monitors | grep -q eDP-1 && notify-send 'Laptop screen active'"

# Background process
post_apply_exec = "sleep 2 && restart-dependent-apps.sh &"
```

## Failure Handling

If exec commands fail:
- The error is logged
- The service continues operating normally
- Monitor configuration is **not** rolled back

This ensures that a failing callback doesn't interrupt monitor configuration.

## Manual Hyprland Reload

If you have `disable_autoreload = true` in Hyprland settings, use callbacks to reload manually:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[general]
post_apply_exec = "hyprctl reload"
```

## Use Cases

### Conditional Actions

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[profiles.docked]
config_file = "hyprconfigs/docked.conf"
post_apply_exec = "systemctl --user start docked-setup.service"

[profiles.laptop_only]
config_file = "hyprconfigs/laptop.conf"
post_apply_exec = "systemctl --user stop docked-setup.service"
```

### Workspace Management

```toml
post_apply_exec = "hyprctl dispatch workspace 1"  # Switch to workspace 1 after config change
```

### Application Restart

```toml
post_apply_exec = "killall waybar eww && sleep 1 && waybar & eww daemon &"
```

### Logging

```toml
post_apply_exec = "echo \"$(date): Applied profile\" >> ~/hyprdynamicmonitors.log"
```

## See Also

- [Notifications](./notifications) - Built-in notification system
- [Profiles](./profiles) - Profile configuration
- [Examples - Callbacks](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/callbacks)
