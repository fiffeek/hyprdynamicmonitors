---
sidebar_position: 6
---

# Running without a daemon

In some situations, particularly with desktop computers where monitor configurations remain stable, running the HyprDynamicMonitors daemon may be unnecessary. If you prefer manual control over when monitor configurations are applied and rendered, you can use the TUI exclusively without the daemon.

This guide shows you how to manage your monitor configurations using only the TUI, giving you complete control over when changes are persisted to disk.

:::note
If you need dynamic configuration that responds to monitor or power state changes, consider keeping the daemon enabled or using [the mode pattern guide](./ad-hoc.md) instead.
:::

## TUI terminology

Before diving into the workflows, it's helpful to understand these key concepts:

- **HDM config** - The HyprDynamicMonitors configuration file (typically `~/.config/hyprdynamicmonitors/config.toml`) that stores profiles and settings.

- **Profile** - A saved monitor layout in HDM that includes monitor configurations (position, scale, rotation, etc.) and the required monitors it matches.

- **Applying monitors to profile** - Saving the current monitor layout settings to an existing or new profile in your HDM config file.

- **Rendering** - Generating the Hyprland monitor configuration from HDM profiles and writing it to the destination file (typically `~/.config/hypr/monitors.conf`). When running without a daemon, you need to manually trigger this with the `R` keybind.

- **Ephemeral apply** - Applying monitor settings directly to Hyprland without saving to a profile or rendering to the destination file. Useful for testing layouts on-the-fly.

## Workflows

### Creating a new profile

![Creating a new profile](/previews/create_profile.gif)

When you set up a new monitor configuration:

1. Launch the TUI: `hyprdynamicmonitors tui`
2. In the **Monitors** view:
   - Select monitors with arrow keys and press `Enter` to edit
   - Adjust position (`h/j/k/l`), scale (`s`), rotation (`r`), resolution (`m`), etc.
   - Press `T` frequently to auto-fit monitors in the preview pane
3. Optionally test with ephemeral apply: `A` then confirm with `Y`
4. Switch to **Profile** view with `Tab`
5. If you see "No Matching Profile", press `n` to create a new profile
6. Type the profile name and press `Enter`
7. The profile is saved to your HDM config and a template file is created under `~/.config/hyprdynamicmonitors/hyprconfigs/${profile_name}.go.tmpl`
8. **Without daemon**: Press `R` to manually render the configuration to your destination file
9. The rendered configuration is written to `config.general.destination` (typically `~/.config/hypr/monitors.conf`)

:::tip
Make sure your Hyprland config sources the destination file: `source = ~/.config/hypr/monitors.conf` in `~/.config/hypr/hyprland.conf`
:::

### Adjusting an existing profile

![Editing an existing profile](/previews/edit_existing.gif)

When modifying an existing monitor setup:

1. Launch the TUI: `hyprdynamicmonitors tui`
2. The **Profile** view will show the matching profile name if one exists
3. Switch to **Monitors** view and make your changes
4. Optionally test with ephemeral apply: `A` then `Y`
5. Switch back to **Profile** view with `Tab`
6. Press `a` to apply changes to the existing profile
7. Confirm with `Y`

The TUI will:
- Replace the content between the `>>>>>` TUI markers if they exist (from previous TUI edits or `hyprdynamicmonitors freeze`)
- Or append a new `>>>>>` markers block at the end of the profile file if markers aren't found

8. **Without daemon**: Press `R` to manually render the updated configuration to the destination file

### Rendering

![Rendering and editing config](/previews/render_edit.gif)

When running without the daemon, rendering must be triggered manually:

- **In the TUI**: Press `R` in the **Profile** view to render configuration to the destination file

The rendering process:
1. Reads your HDM config and profiles
2. Matches current monitors to the appropriate profile
3. Generates Hyprland monitor configuration syntax
4. Writes to `config.general.destination` file

:::info
With the daemon running, rendering happens automatically when monitors or configuration changes. Without it, you control when configuration is written to disk.
:::

### Regular workflow

For day-to-day adjustments when running without a daemon, follow this streamlined workflow:

1. **Launch the TUI**: `hyprdynamicmonitors tui`
2. **Adjust monitors**: Make changes in the **Monitors** view
   - Select and edit monitors as needed
   - Use `T` to auto-fit the preview
   - Test with ephemeral apply (`A` then `Y`) if desired
3. **Save to profile**: Switch to **Profile** view with `Tab`
   - For new configurations: press `n` to create a new profile
   - For existing configurations: press `a` to update the matching profile
4. **Render configuration**: Press `R` to write the rendered config to your destination file
5. **Reload Hyprland**: The changes take effect automatically when Hyprland reloads the sourced config file (typically happens immediately)

:::tip Quick workflow
For minor tweaks to an existing profile: `Monitors view` → adjust → `Tab` → `a` → `Y` → `R`
:::

This approach gives you full control over when configuration changes are persisted to disk, without the automatic behavior of the daemon.
