---
sidebar_position: 2
---

# Signals

HyprDynamicMonitors responds to Unix signals for runtime control.

## SIGHUP

Instantly reloads configuration and reapplies monitor setup.

```bash
kill -SIGHUP $(pidof hyprdynamicmonitors)

# Or with systemd
systemctl --user reload hyprdynamicmonitors
```

Use this when you've made changes to your configuration and want them applied immediately without waiting for automatic hot reload.

## SIGUSR1

Reapplies the monitor setup without reloading the service configuration.

```bash
kill -SIGUSR1 $(pidof hyprdynamicmonitors)
```

This is useful when you want to force a configuration reapplication without changing any settings.

## SIGTERM / SIGINT

Graceful shutdown of the service.

```bash
# SIGTERM
kill -SIGTERM $(pidof hyprdynamicmonitors)

# SIGINT (Ctrl+C)
kill -SIGINT $(pidof hyprdynamicmonitors)

# Or with systemd
systemctl --user stop hyprdynamicmonitors
```

The service will clean up resources and exit gracefully.

## Hot Reload vs SIGHUP

HyprDynamicMonitors automatically watches for changes to configuration files and applies them without requiring a restart. This includes:

- Configuration file changes (`config.toml`)
- Profile config changes (both static and template files)
- New profile files

The service uses file system watching with debounced updates (default 1000ms delay) to avoid excessive reloading.

**When to use SIGHUP:**
- You want immediate config reload (bypass the debounce delay)
- Hot reload is disabled (`--disable-auto-hot-reload`)
- You're testing configuration changes interactively

**When to rely on hot reload:**
- Normal development and operation
- You don't need immediate updates
- You make multiple config changes in quick succession
