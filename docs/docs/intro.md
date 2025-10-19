---
sidebar_position: 1
slug: /
---

# Introduction

HyprDynamicMonitors is an event-driven service that automatically manages Hyprland monitor configurations based on connected displays and power state.
It also provides a standalone TUI that can be used for ad-hoc modifications and profile management.

## Preview

### Full TUI Preview

![Demo](/previews/demo.gif)

### Async Events

![Events](/previews/lid_tui.gif)

## Features

- **Event-driven architecture** responding to monitor, power and lid state changes in real-time
- **Interactive TUI** for visual monitor configuration and profile management
- **Profile-based configuration** with different settings for different monitor setups
- **Template support** for dynamic configuration generation
- **Hot reloading**: automatically detects and applies configuration changes without restart by watching config files (optional)
- **Configurable UPower queries** for custom power management systems
- **Desktop notifications** for configuration changes (optional)

## Design Philosophy

HyprDynamicMonitors follows a **fail-fast architecture** designed for reliability and simplicity.

### Reliability Through Restarts

The service intentionally fails quickly on critical issues rather than attempting complex recovery. This design expects the service to run under systemd or a wrapper script that provides automatic restarts. Since configuration is applied on startup, restarts ensure the service remains operational even after encountering errors.

### Hot Reloading With Graceful Restart

For configuration changes, the service provides automatic hot reloading by watching configuration files. When hot reloading encounters issues, it gracefully falls back to the fail-fast behavior, prioritizing reliability over attempting risky recovery scenarios.

### Hyprland-Native Integration

The service leverages Hyprland's native abstractions rather than working directly with Wayland protocols. It detects the desired configuration based on current monitor state and power supply, then either:
- Generates a templated Hyprland config file at the specified destination
- Or creates a symlink to a user-provided static configuration file

Hyprland automatically detects and applies these configuration changes (granted it's not explicitly turned off, if so you have to use [the callbacks](./configuration/callbacks) to `hyprctl reload`), ensuring seamless integration with the compositor's built-in configuration system.

## Quick Start

Ready to get started? Check out the [Quick Start](./category/quick-start) to set up HyprDynamicMonitors.

## Runtime Requirements

- Hyprland with IPC support
- UPower (optional, for power state monitoring)
- Read-only access to system D-Bus (optional for power state monitoring; should already be your default)
- Write access to system D-Bus for notifications (optional; should already be your default)

## Alternative Software

Most similar tools are more generic, working with any Wayland compositor. In contrast, `hyprdynamicmonitors` is specifically designed for Hyprland (using its IPC) but provides several advantages:

**Advantages of HyprDynamicMonitors:**
- **Interactive TUI**: Built-in terminal interface for visual monitor configuration and profile management
- **Full configuration control**: Instead of introducing another configuration format, you work directly with Hyprland's native config syntax
- **Template system**: Dynamic configuration generation based on connected monitors and power state
- **Power state awareness**: Built-in AC/battery detection for laptop users
- **Event-driven automation**: Automatically responds to monitor connect/disconnect events

**Trade-offs:**
- Hyprland-specific (not generic Wayland)
- Requires systemd or wrapper script for production use (fail-fast design)

**Similar Tools:**
- [kanshi](https://sr.ht/~emersion/kanshi/) - Generic Wayland output management
- [shikane](https://github.com/hw0lff/shikane) - Another Wayland output manager
- [nwg-displays](https://github.com/nwg-piotr/nwg-displays) - GUI-based display configuration tool for Sway/Hyprland
- [hyprmon](https://github.com/erans/hyprmon) - TUI-based display configuration tool for Hyprland
