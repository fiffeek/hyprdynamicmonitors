---
sidebar_position: 2
---

# Nix

:::tip
You can look up Github repositories/code for ideas how to integrate `hyprdynamicmonitors` into your nix setup [here](https://github.com/search?q=hyprdynamicmonitors+path%3A*.nix&type=code&p=1).
:::

This guide covers how to install and configure HyprDynamicMonitors on NixOS and Home Manager.

## Installation

HyprDynamicMonitors can be installed in several ways on Nix-based systems. Choose the method that best fits your setup.

### Quick Start

For trying out the tool without installing:

```bash
# Run directly from GitHub
nix run github:fiffeek/hyprdynamicmonitors

# Or from specific tag/version (recommended)
nix run github:fiffeek/hyprdynamicmonitors/v1.0.0
```

### Imperative Installation

Install to your user profile:

```bash
nix profile install github:fiffeek/hyprdynamicmonitors
```

### Declarative Installation

For declarative NixOS or Home Manager configurations using flakes, add the package to your system.

#### NixOS Configuration

Add the flake input and install the package:

```nix title="flake.nix"
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    hyprdynamicmonitors.url = "github:fiffeek/hyprdynamicmonitors";
  };

  outputs = { self, nixpkgs, hyprdynamicmonitors, ... }@inputs:
  let
    system = "x86_64-linux";
  in {
    nixosConfigurations.yourhost = nixpkgs.lib.nixosSystem {
      specialArgs = { inherit inputs system; };
      modules = [ ./configuration.nix ];
    };
  };
}
```

Then in your `configuration.nix`:

```nix title="configuration.nix"
{ inputs, system, ... }:
{
  environment.systemPackages = [
    inputs.hyprdynamicmonitors.packages.${system}.default
  ];
}
```

#### Home Manager Configuration

Add the flake input and install the package:

```nix title="flake.nix"
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    hyprdynamicmonitors.url = "github:fiffeek/hyprdynamicmonitors";
  };

  outputs = { self, nixpkgs, home-manager, hyprdynamicmonitors, ... }@inputs:
  let
    system = "x86_64-linux";
  in {
    homeConfigurations.youruser = home-manager.lib.homeManagerConfiguration {
      pkgs = import nixpkgs { inherit system; };
      modules = [ ./home.nix ];
      extraSpecialArgs = { inherit inputs system; };
    };
  };
}
```

Then in your `home.nix`:

```nix title="home.nix"
{ inputs, system, ... }:
{
  home.packages = [
    inputs.hyprdynamicmonitors.packages.${system}.default
  ];
}
```

## Hyprland Integration

:::warning Important: Source the Configuration File

After installing HyprDynamicMonitors and setting up your profiles, you **must** source the generated configuration file in your Hyprland config. This is a common step that's easy to miss!

**Without this step, your monitor configurations won't be applied.**

The destination path [can be changed](../configuration/overview#general-settings) but defaults to `$HOME/.config/hypr/monitors.conf`.

Add the following to your Hyprland configuration:

```conf title="~/.config/hypr/hyprland.conf"
# Source the auto-generated monitors configuration
# Adjust the path to match your destination setting in config.toml
source = ~/.config/hypr/monitors.conf
```

Or if using Home Manager with declarative Hyprland config:

```nix
wayland.windowManager.hyprland = {
  enable = true;
  extraConfig = ''
    # Source the auto-generated monitors configuration
    source = ~/.config/hypr/monitors.conf
  '';
};
```

:::

## Systemd Service Integration

For automatic startup and proper integration with your system, it's recommended to run HyprDynamicMonitors as a systemd service. HyprDynamicMonitors provides both NixOS and Home Manager modules for declarative systemd service configuration.

See the [Systemd Integration Guide - Nix section](../advanced/systemd#nix) for detailed setup instructions, configuration options, and examples.

## Using the TUI with Declarative Configuration

:::info Config Path for TUI

When using declarative Nix configuration, the TUI needs to operate on the same config file that the daemon uses. By default, the TUI looks for config at `~/.config/hyprdynamicmonitors/config.toml`.

If your daemon uses a different config location (e.g., when using the NixOS module), specify the config path explicitly:

```bash
# For NixOS module (default path)
hyprdynamicmonitors tui --config /etc/xdg/hyprdynamicmonitors/config.toml

# For Home Manager module (default path)
hyprdynamicmonitors tui --config ~/.config/hyprdynamicmonitors/config.toml
```

**Working with immutable Nix configurations:**

If you're managing your configuration declaratively and don't want to modify the live config directly, use this workflow:

```bash
# Create a temporary working directory
mkdir -p /tmp/hyprdynamicmonitors

# Copy config and hyprconfigs directory to temporary location
cp /etc/xdg/hyprdynamicmonitors/config.toml /tmp/hyprdynamicmonitors/config.toml
cp -r /etc/xdg/hyprdynamicmonitors/hyprconfigs /tmp/hyprdynamicmonitors/

# Run TUI on the temporary copy
hyprdynamicmonitors tui --config /tmp/hyprdynamicmonitors/config.toml

# Copy changes back to your Nix configuration directory
cp /tmp/hyprdynamicmonitors/config.toml ~/your-nix-config/hyprdynamicmonitors-config.toml
cp -r /tmp/hyprdynamicmonitors/hyprconfigs ~/your-nix-config/

# Rebuild your system/home-manager to apply changes
nixos-rebuild switch  # or home-manager switch
```

This approach keeps your declarative configuration as the source of truth while still allowing you to use the TUI for configuration. The `hyprconfigs` directory contains your monitor profiles and templates, so it needs to be copied alongside the config file.

:::

## Next Steps

After installation:

1. ✅ Verify the package is installed: `hyprdynamicmonitors --version`
2. ✅ **Add the `source` line to your Hyprland config** (see warning above)
3. ✅ Configure your monitor profiles using the [TUI](./tui) or [manual configuration](../configuration/overview)
4. ✅ Start the service or run `hyprdynamicmonitors run`

## Troubleshooting

### Command not found

If `hyprdynamicmonitors` is not found after installation, ensure:
- For NixOS: The package is in `environment.systemPackages`
- For Home Manager: The package is in `home.packages`
- Your shell session is restarted or you've run `source ~/.profile`

### Monitors not switching

Check that:
1. ✅ The systemd service is running: `systemctl --user status hyprdynamicmonitors`
2. ✅ You've **sourced the configuration file** in your Hyprland config (see warning above)
3. ✅ The destination path matches between your config and the `source` line in Hyprland

### Service fails to start

Check logs: `journalctl --user -u hyprdynamicmonitors -n 50`

Common issues:
- Missing environment variables (see [systemd environment tip](../advanced/systemd))
- UPower not enabled (required for power/lid events)
- Invalid configuration file

## See Also

- [Systemd Integration](../advanced/systemd) - Detailed systemd service configuration
- [TUI Guide](./tui) - Interactive configuration interface
- [Configuration Overview](../configuration/overview) - Configuration file reference
