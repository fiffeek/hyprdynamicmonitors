---
sidebar_position: 7
---

# FAQ

## How do I assign workspaces to specific monitors?

You can include workspace assignments directly in your profile configuration files. HyprDynamicMonitors generates or links Hyprland configuration files, so any valid Hyprland configuration syntax works.

### Example with static configuration

In your profile configuration file (e.g., `hyprconfigs/triple-monitor.conf`):

```
# Monitor configuration
monitor=desc:BOE 0x0C6B,preferred,0x0,1
monitor=desc:LG Electronics LG ULTRAWIDE 408NTVSDT319,preferred,1920x0,1
monitor=desc:Dell Inc. DELL P2222H B2RY1H3,preferred,3840x0,1

# Workspace assignments
# The format is: `monitor:${connectorName}`
# You can use templates so that you do not have to hardcode these too
workspace=1,monitor:eDP-1,default:true
workspace=2,monitor:DP-11,default:true
workspace=3,monitor:DP-12,default:true
```

### Example with templates

In `hyprconfigs/dual-monitor.go.tmpl`:

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
{{- $external := index .MonitorsByTag "external" -}}
monitor={{$laptop.Name}},preferred,0x0,1
monitor={{$external.Name}},preferred,1920x0,1

workspace=1,monitor:{{$laptop.Name}},default:true
workspace=2,monitor:{{$external.Name}},default:true
workspace=3,monitor:{{$external.Name}}
```

You can also use the [TUI](./quickstart/tui) to create and edit profiles interactively.

:::info
Hyprland might still force the workspace to appear on a monitor that technically does not exist anymore (see [this discussion](https://github.com/hyprwm/Hyprland/discussions/11180#discussioncomment-14857978)), you can hack this with moving the workspace on Hyprland reload:
```
exec = hyprctl dispatch moveworkspacetomonitor 2 {{$external.Name}}
```

Alternatively, you could also write a [callback](./configuration/callbacks.md) (this is a simplified version, writing a script will be more robust):
```toml
[profiles.gaming_setup]
config_file = "hyprconfigs/gaming.conf"
post_apply_exec = "hyprctl dispatch moveworkspacetomonitor 2 DP-11"
```
:::

## How do I use the TUI to create or edit profiles?

The TUI provides an interactive way to configure monitors and manage profiles. The workflow involves two main views:

### Step 1: Configure monitors in the Monitors view

- Adjust your monitor settings (resolution, position, scale, etc.)
- Apply the configuration with `a` to test it in Hyprland
- Once satisfied with the layout, proceed to save it as a profile

### Step 2: Switch to the HDM Profile view

- Press `Tab` to switch between Monitors view and HDM Profile view
- Choose to either:
  - **Edit existing profile**: Press `a` to apply the current monitor configuration to the selected profile
  - **Create new profile**: Press `n` to create a new profile with the current monitor configuration

### Understanding the generated configuration

The TUI places generated configuration between comment markers in your profile config files:

```
# <<<<< TUI AUTO START
monitor=eDP-1,2880x1920@120,0x0,1.5
monitor=DP-1,3840x2160@60,2880x0,1
# <<<<< TUI AUTO END
```

You can add additional settings outside these markers, and they will be preserved across TUI updates. This is useful for:
- Overriding specific settings
- Adding workspace assignments
- Including additional Hyprland configuration

### Manual editing

From the HDM Profile view, you can manually edit files using your `$EDITOR`:
- Edit the profile configuration file (the Hyprland config)
- Edit the HDM configuration file (the TOML config)

For more details, see the [TUI documentation](./quickstart/tui).

## How to disable all monitors not defined in the profile required monitors?

Use the `.ExtraMonitors` variable in your template:

```go
{{- range .ExtraMonitors}}
monitor={{.Name}},disable
{{- end }}
```

See the [disable monitors example](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/disable-monitors) for a complete configuration.

## Can I use the TUI without running the hyprdynamicmonitors daemon?

Yes, the TUI can be used standalone without the daemon running. However, functionality is limited:

### What works without the daemon

- Experimenting with monitor configurations in the Monitors view
- Applying configurations manually and ephemerally with `a` (changes are temporary until Hyprland restarts)
- Testing different layouts and settings interactively

### What requires a valid configuration

- Saving profiles to HyprDynamicMonitors configuration (HDM Profile view features)
- Persisting monitor configurations that automatically apply on monitor connect/disconnect
- Power state-based profile switching

### Typical standalone usage

```bash
# Use TUI to experiment with monitor layouts
hyprdynamicmonitors tui

# Configure monitors interactively, apply with 'a' to test
# Changes are applied to Hyprland but not persisted
```

If you want to persist configurations created in the TUI, you need a valid HyprDynamicMonitors config file. You can start with a minimal config and build from there using the TUI's profile creation features.

## Why are my comments becoming part of the configuration in templates?

This is due to Go template whitespace trimming behavior. When you use `-}}`, it removes all whitespace (including newlines) after the template action.

### Problem example

```go
# auto generated by hyprdynamicmonitors

{{- $laptop := index .MonitorsByTag "LaptopMonitor" -}}
{{- $external := index .MonitorsByTag "ExternalMonitor" -}}

monitor={{$laptop.Name}}, disable
# change to your preferred external monitor config
monitor={{$external.Name}},preferred,0x0,1
```

The `-}}` after the variable definitions removes the newline, causing `monitor={{$laptop.Name}}, disable` to be pulled up into the comment above it, creating one long comment line.

### Solution 1: Remove trailing dash from the last template action

```go
# auto generated by hyprdynamicmonitors

{{- $laptop := index .MonitorsByTag "LaptopMonitor" -}}
{{- $external := index .MonitorsByTag "ExternalMonitor" }}

monitor={{$laptop.Name}}, disable
# change to your preferred external monitor config
monitor={{$external.Name}},preferred,0x0,1
```

### Solution 2: Define all variables at the start, then add content

```go
{{- $laptop := index .MonitorsByTag "LaptopMonitor" -}}
{{- $external := index .MonitorsByTag "ExternalMonitor" -}}


# auto generated by hyprdynamicmonitors
# change to your preferred external monitor config
monitor={{$laptop.Name}}, disable
monitor={{$external.Name}},preferred,0x0,1
```

### Key points

- `-}}` trims all whitespace **after** the action, including newlines
- This can cause the next line to merge with whatever comes before the template action
- Use `}}` (without trailing dash) on the last variable definition before your config
- Or group all variable definitions at the top, separated from config with blank lines

## How do I test my configuration before applying it?

Use the `--dry-run` flag combined with `--run-once` for instant output:

```bash
hyprdynamicmonitors run --dry-run --run-once
```

This shows what would be done without making actual changes, then exits immediately. It's useful for:
- Testing profile matching logic
- Verifying which profile would be selected
- Checking template rendering output
- Validating configuration before deployment

You can also use `--dry-run` alone to keep the service running and observe how it would respond to monitor or power state changes.

## How do I validate my configuration file?

Use the `validate` command:

```bash
hyprdynamicmonitors validate
```

This checks for:
- TOML syntax errors
- Logical inconsistencies
- Missing required fields
- Invalid configuration values

## Can I use HyprDynamicMonitors with nwg-displays?

Yes! HyprDynamicMonitors can work alongside `nwg-displays`:

1. Use `nwg-displays` to visually configure your monitors
2. Let `hyprdynamicmonitors` automatically write/link the configuration to your Hyprland directory
3. `hyprdynamicmonitors` handles automatic profile switching based on connected monitors

This combines the visual configuration of `nwg-displays` with the automatic profile management of `hyprdynamicmonitors`.

## Is UPower Running? UPower misconfigured or not running: failed to get property from UPower: Object does not exist at path “/org/freedesktop/UPower/devices/line_power_ACAD”

If you got this error, you either want to:
1. If you do not want to use power events (e.g., on a desktop), pass the `--disable-power-events` CLI argument
2. Tweak your [Power Events configuration](./configuration/power-events.md) to use the proper device path (`upower -e` to find your `line_power` device; additionally, it would be helpful if you added it [here](https://github.com/fiffeek/hyprdynamicmonitors/blob/main/internal/utils/power.go#L14) so that others do not encounter this issue going forward)

If you are running on a desktop with a version of `hyprdynamicmonitors` that includes [this patch](https://github.com/fiffeek/hyprdynamicmonitors/pull/80) (`v1.3.5+`),
this error should not occur unless `--disable-power-events=false` is explicitly passed, since power events are disabled
for desktops by default. Please [open an issue](https://github.com/fiffeek/hyprdynamicmonitors/issues) and include the output of `cat /sys/class/dmi/id/chassis_type`.

## See Also

- [Examples](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples) - Complete configuration examples
- [Templates](./advanced/templates) - Template syntax and variables
- [TUI](./quickstart/tui) - Interactive monitor configuration
