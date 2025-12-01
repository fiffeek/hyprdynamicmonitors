---
sidebar_position: 5
---

# Notifications

HyprDynamicMonitors can show desktop notifications when configuration changes occur. Notifications are sent via D-Bus using the standard `org.freedesktop.Notifications` interface.

## Configuration

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[notifications]
disabled = false      # Enable/disable notifications (default: false)
timeout_ms = 10000   # Notification timeout in milliseconds (default: 10000)
```

## Disabling Notifications

To disable notifications completely:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[notifications]
disabled = true
```

## Customizing Timeout

To show brief notifications (3 seconds):

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[notifications]
timeout_ms = 3000
```

For longer notifications (15 seconds):

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[notifications]
timeout_ms = 15000
```

## What Gets Notified

Notifications are shown when:
- A new monitor configuration profile is applied
- The monitor configuration changes
- Profile switching occurs

The notification includes:
- The name of the profile being applied
- Basic information about the configuration change

## Requirements

Notifications require:
- Write access to system D-Bus (should be default on most systems)
- A notification daemon running (e.g., `dunst`, `mako`, or desktop environment's built-in notification system)

## Alternative: Custom Callbacks

If you want more control over notifications, consider using [callbacks](./callbacks) instead:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[general]
post_apply_exec = "notify-send 'HyprDynamicMonitors' 'Profile applied successfully'"
```

This allows you to:
- Customize notification text
- Use different notification tools
- Add additional actions or scripts
