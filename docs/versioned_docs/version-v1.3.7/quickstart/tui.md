---
sidebar_position: 3
---

# TUI

The HyprDynamicMonitors TUI provides an interactive interface for managing Hyprland monitor configurations.

## Launch the TUI

```bash
hyprdynamicmonitors tui
```

The TUI can be used with or without a valid HyprDynamicMonitors configuration:
- **With config**: Full functionality including profile management
- **Without config**: Monitor manipulation and ephemeral application only

## Views

The TUI has two main views that you can switch between using `Tab`:

1. **Monitors View** - Edit monitor layouts, positions, modes, and settings
2. **Profile View** - Manage HyprDynamicMonitors profiles and configuration

![Switching between views](/previews/views.gif)

:::info
The profile view is available only when `--config` points to a valid `hyprdynamicmonitors` configuration. Without it, you can still experiment with monitors and apply settings, but cannot save them under `hyprdynamicmonitors`.
:::

## Global Keybinds

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit the TUI |
| `Tab` | Switch between Monitors and Profile views |

---

## Monitors View

The Monitors view shows connected monitors on the left and a visual preview on the right.

![Monitors View](/previews/monitor_view.gif)

### Navigation

| Key | Action |
|-----|--------|
| `j` / `down` | Move down in the monitor list |
| `k` / `up` | Move up in the monitor list |
| `Enter` | Select a monitor for editing / Deselect when in editing mode |

### Monitor Preview Controls

#### Panning Mode

| Key | Action |
|-----|--------|
| `p` | Toggle panning mode (move freely around the monitor grid) |
| `h/j/k/l` or arrow keys | Pan the preview in panning mode |
| `c` | Center the view back to origin (0,0) |

![Panning mode](/previews/panning.gif)

#### Zoom

| Key | Action |
|-----|--------|
| `+` | Zoom in on the preview |
| `-` | Zoom out on the preview |
| `T` | Zoom and pan monitors to fit the preview screen |

![Zoom](/previews/zoom.gif)

#### Display Options

| Key | Action |
|-----|--------|
| `F` | Toggle fullscreen preview mode |
| `o` | Toggle follow monitor mode (preview auto-centers on selected monitor) |
| `S` | Toggle snapping (when moving monitors, they snap to edges of other monitors) |

![Display options](/previews/display_options.gif)

### Editing a Monitor

Once you select a monitor with `Enter`, it enters **EDITING** mode. In this mode:

![Editing a monitor](/previews/editing.gif)

#### Position

| Key | Action |
|-----|--------|
| `h/j/k/l` or arrow keys | Move the monitor in 50px steps |

- With snapping enabled (default), monitors will snap to edges of other monitors within 50px
- When snapping occurs, visual grid lines show the snap alignment

![Positioning monitors](/previews/position.gif)

#### Rotation

| Key | Action |
|-----|--------|
| `r` | Rotate the monitor by 90 degrees (cycles through 0 → 90 → 180 → 270 → 0) |
| `L` | Flip (mirror) monitor horizontally |

:::note
Cannot rotate disabled monitors
:::

![Rotating monitors](/previews/rotation.gif)

#### Resolution and Refresh Rate

| Key | Action |
|-----|--------|
| `m` | Open mode selection menu |
| `j/k` | Preview different modes (updates preview in real-time) |
| `Enter` | Apply the selected mode |
| `Esc` | Close (applying the last selection) |

Shows all available modes for the selected monitor.

![Changing resolution and refresh rate](/previews/resolution.gif)

#### Scaling

| Key | Action |
|-----|--------|
| `s` | Open scale selector |
| `k` | Increase scale by 0.001 or to the nearest greater valid value |
| `j` | Decrease scale by 0.001 or to the nearest lower valid value |
| `1` | Set scale to 1.0x |
| `2` | Set scale to 2.0x |
| `C` | Enter a custom scale value |
| `e` | Enable/Disable snapping |
| `Enter` | Confirm scale change |
| `Esc` | Close (applying the last selection) |

![Adjusting monitor scaling](/previews/scaling.gif)

#### Mirroring

| Key | Action |
|-----|--------|
| `i` | Open mirror selection menu |
| `j/k` | Navigate through available monitors to mirror |
| `Enter` | Apply mirror setting |
| `Esc` | Close (applying the last selection) |

Select which monitor this monitor should mirror. Mirror loops are prevented automatically.

![Monitor mirroring](/previews/mirroring.gif)

#### Enable/Disable

| Key | Action |
|-----|--------|
| `e` | Toggle monitor (e)nabled/disabled |

- Disabled monitors show as `disable` in the config preview
- Cannot disable the last remaining monitor
- Cannot edit settings of disabled monitors

![Enabling/disabling monitors](/previews/disable.gif)

#### Variable Refresh Rate (VRR)

| Key | Action |
|-----|--------|
| `v` | Toggle VRR on/off for the monitor |

![Toggling VRR](/previews/vrr.gif)

#### Color Profiles Management

| Key | Action |
|-----|--------|
| `C` | Enter color profiles management menu |
| `Esc/Enter` | Back / Select |
| `up/down` | Change the current color preset |
| `b` | Toggle bitdepth (at the moment only `default` and `10` is supported) |
| `r/R` | Adjust SDR brightness (when `hdr` color profile is selected) |
| `t/T` | Adjust SDR saturation (when `hdr` color profile is selected) |

![Color profiles management](/previews/color.gif)

:::caution Known Limitation
Due to [`hyprctl` not yet supporting the output of these](https://github.com/fiffeek/hyprdynamicmonitors/issues/34), when you change the color preset and apply it, then reload the TUI, the color preset will mismatch and show `auto/default/srgb`. This will change as soon as the underlying issue in Hyprland is resolved.
:::

### Applying Changes

| Key | Action |
|-----|--------|
| `A` | Apply current monitor configuration to Hyprland (ephemeral, not persisted on disk) |
| `Y` | Confirm and apply |
| `N` or `Esc` | Cancel |

Shows confirmation prompt before applying.

![Applying changes](/previews/apply.gif)

---

## Profile View

The Profile view shows your HyprDynamicMonitors configuration and lets you save monitor layouts as profiles.

### Profile Management

#### Creating a New Profile

| Key | Action |
|-----|--------|
| `n` | Create new profile from current monitor layout |
| `Enter` | Save the profile |
| `Esc` | Cancel |

Opens profile name input where you can type the profile name.

The profile will include:
- All connected monitors as required monitors (matched by description)
- Current monitor layout, modes, scales, positions, and settings

![Creating a new profile](/previews/create_profile.gif)

#### Editing an Existing Profile

When the current monitors match an existing profile:

| Key | Action |
|-----|--------|
| `a` | Apply edited settings to the matching profile |
| `Y` | Confirm and update the profile |
| `N` or `Esc` | Cancel |
| `e` | Edit the config file in your `$EDITOR` |

Shows confirmation prompt with profile name. After confirmation, the config is reloaded automatically.

![Editing an existing profile](/previews/edit_existing.gif)

### Profile Status Indicators

The TUI shows different status messages:

- **Profile: [name]** - Current monitors match this profile exactly
- **Monitor Count Mismatch** - Monitors match partially but count differs
- **No Matching Profile** - No profile matches current monitor setup

### Configuration

| Key | Action |
|-----|--------|
| `C` | Open the HyprDynamicMonitors config file in your default editor (defined by `$EDITOR`) |

![Opening config file](/previews/hdm_c.gif)

### Rendering Hyprland Configuration

The TUI can generate and edit the Hyprland monitor configuration on-demand:

| Key | Action |
|-----|--------|
| `R` | Render configuration from template/static file to the destination |
| `E` | Edit the rendered configuration file in your `$EDITOR` |

The `R` command writes the rendered output to your configured `config.general.destination`. You can then manually edit this file with `E`.

:::caution Ephemeral Changes
If the `hyprdynamicmonitors` daemon is running, any manual edits to the rendered configuration will be overwritten when the daemon responds to monitor or power state events.
:::

![Rendering and editing config](/previews/render_edit.gif)

---

## Tips

1. **Snapping**: Keep snapping enabled (default) for easier monitor alignment. When snapping is active, monitors automatically align to edges of other monitors within 50px.

2. **Follow Mode**: Enable f(o)llow mode (`o`) when moving monitors to keep them centered in the preview.

3. **Fullscreen Preview**: Use fullscreen mode (`F`) to get a better view when working with many monitors or complex layouts.

4. **Preview Before Applying**: All changes are shown in the preview immediately. You can experiment freely without affecting your actual configuration until you press `A` to apply **and confirm**.

5. **Profile Workflow**:
   - Arrange monitors in the Monitors view
   - Switch to Profile view with `Tab`
   - Create a new profile with `n` or update existing with `a`
   - The profile will be saved to your HyprDynamicMonitors config

6. **Panning**: Use panning mode (`p`) to explore large monitor layouts. The grid shows spacing dots that scale with zoom level.

---

## Visual Indicators

### Monitor List

- `>` - Currently selected item in the list
- `[EDITING]` - Monitor is in editing mode
- `[CHANGE MODE]` - Mode selection menu is open
- `[SCALE MODE]` - Scale selector is open
- `[MIRRORING]` - Mirror selection menu is open

### Monitor Preview

- Monitors are shown as colored rectangles on a grid
- The selected monitor has a different fill pattern
- Monitor labels show the name and an arrow indicating rotation
- Bottom edge is highlighted with a brighter color to show orientation
- Snap lines appear as vertical `|` and horizontal `-` lines when snapping occurs

### Status Messages

Success messages appear in the header and automatically clear after 2 seconds. Errors linger until cleared by any other success.

## See Also

- [CLI Commands - tui](../usage/commands#tui) - Command line options
- [FAQ - Using TUI](../faq#how-do-i-use-the-tui-to-create-or-edit-profiles) - Common questions about TUI usage
