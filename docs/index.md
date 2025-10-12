---
layout: default
title: Overview
---

# Overview

## Features

- Event-driven architecture responding to monitor and power state changes in real-time
- Interactive TUI for visual monitor configuration and profile management
- Profile-based configuration with different settings for different monitor setups
- Template support for dynamic configuration generation
- Hot reloading: automatically detects and applies configuration changes without restart by watching config files (optional)
- Configurable UPower queries for custom power management systems
- Desktop notifications for configuration changes (optional)

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

Hyprland automatically detects and applies these configuration changes (granted it's not explicitly turned off, if so you have to use
[the callbacks](https://github.com/fiffeek/hyprdynamicmonitors?tab=readme-ov-file#user-callbacks-exec-commands) to `hyprctl reload`), ensuring seamless integration with the compositor's built-in configuration system.


## Installation

### Binary Release

Download the latest binary from GitHub releases:

```bash
# optionally override the destination directory, defaults to ~/.local/bin/
export DESTDIR="$HOME/.bin"
curl -o- https://raw.githubusercontent.com/fiffeek/hyprdynamicmonitors/refs/heads/main/scripts/install.sh | bash
```

### AUR

For Arch Linux users, install from the AUR:

```bash
# Using your preferred AUR helper (replace 'aurHelper' with your choice)
aurHelper="yay"  # or paru, trizen, etc.
$aurHelper -S hyprdynamicmonitors-bin

# Or using makepkg:
git clone https://aur.archlinux.org/hyprdynamicmonitors-bin.git
cd hyprdynamicmonitors-bin
makepkg -si
```

### Nix

For Nix and NixOS users:

```bash
# Run directly from GitHub
nix run github:fiffeek/hyprdynamicmonitors

# Or from specific tag/version (recommended)
nix run github:fiffeek/hyprdynamicmonitors/v1.0.0

# Install to profile
nix profile install github:fiffeek/hyprdynamicmonitors
```

### Build from Source

Requires [asdf](https://asdf-vm.com/) to manage the Go toolchain:
```bash
# Build the binary (output goes to ./dest/)
make

# Install to custom location
make DESTDIR=$HOME/binaries install

# Uninstall from custom location
make DESTDIR=$HOME/binaries uninstall

# Install system-wide (may require sudo)
sudo make DESTDIR=/usr/bin install
```
