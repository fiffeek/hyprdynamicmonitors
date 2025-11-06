---
sidebar_position: 1
---

# Installation

Choose your preferred installation method below.

## Binary Release

Download the latest binary from GitHub releases:

```bash
# optionally override the destination directory, defaults to ~/.local/bin/
export DESTDIR="$HOME/.bin"
curl -o- https://raw.githubusercontent.com/fiffeek/hyprdynamicmonitors/refs/heads/main/scripts/install.sh | bash
```

## AUR

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

## Nix

For Nix and NixOS users:

```bash
# Run directly from GitHub
nix run github:fiffeek/hyprdynamicmonitors

# Or from specific tag/version (recommended)
nix run github:fiffeek/hyprdynamicmonitors/v1.0.0

# Install to profile
nix profile install github:fiffeek/hyprdynamicmonitors
```

For declarative NixOS or Home Manager configuration using flakes, including package installation and systemd service setup, see the [Nix guide](./nix).

## Build from Source

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

## Next Steps

After installation, proceed to the [Quick Start Guide](../category/quick-start) to set up your first monitor configuration profile.
