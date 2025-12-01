---
sidebar_position: 8
---

# Theming

<video
  src="/previews/themes-showcase.mp4"
  controls
  autoPlay={true}
  muted
  loop
  style={{ width: "100%", borderRadius: "8px" }}
/>

`hyprdynamicmonitors` supports fully customizable theming for the TUI. You can use built-in themes, create your own, or generate themes dynamically from your wallpaper.

## Quick Start

The easiest way to use theming is to reference one of the built-in themes.

**For AUR installations:**

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "/usr/share/hyprdynamicmonitors/themes/static/rose-pine/theme.toml"
```

**For manual installations:** Copy the theme from the [repository](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/themes/static) to `~/.config/hyprdynamicmonitors/themes/` first, then:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "~/.config/hyprdynamicmonitors/themes/rose-pine.toml"
```

Changes to themes are detected automatically - no need to restart the TUI.

## Configuration Methods

### Using External Theme Files (Recommended)

Point to a theme file in your configuration:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "/path/to/theme.toml"
```

Theme files contain color definitions like:

```toml title="/path/to/theme.toml"
# Example theme file
active_pane_color = "#7aa2f7"
inactive_pane_color = "#565f89"
header_color = "#bb9af7"
# ... more colors
```

### Inline Configuration

You can also define colors directly in your config file:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
active_pane_color = "62"
inactive_pane_color = "240"
header_color = "205"
# ... more colors
```

**Note:** Colors can be specified as:
- **Hex colors**: `"#7aa2f7"` (recommended)
- **ANSI codes**: `"62"` (0-255)
- **Background**: Cannot be set - uses your terminal emulator's background

## Built-in Themes

`hyprdynamicmonitors` comes with 11 carefully crafted themes ready to use.

### Dark Themes

| Theme | Description |
|-------|-------------|
| `tokyo-night` | Popular purple/blue theme with vibrant colors |
| `nord` | Cool-toned minimalist theme with muted blues |
| `one-dark` | Balanced professional theme from Atom editor |
| `dracula` | Vibrant purple/pink/cyan aesthetic |
| `catppuccin-mocha` | Warm pastel dark theme with cozy vibes |
| `gruvbox-dark` | Retro-inspired warm color palette |
| `kanagawa` | Elegant theme inspired by "The Great Wave off Kanagawa" |
| `rose-pine` | Muted theme with natural pine and rose tones |

### Light Themes

| Theme | Description |
|-------|-------------|
| `gruvbox-light` | Light variant of the warm Gruvbox palette |
| `solarized-light` | Classic, carefully designed light theme |
| `alabaster` | Nearly monochrome minimalist theme |

### Using Built-in Themes

**AUR installations:** Themes are automatically installed to `/usr/share/hyprdynamicmonitors/themes/static/`

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "/usr/share/hyprdynamicmonitors/themes/static/rose-pine/theme.toml"
```

**Manual installations:** Copy themes from the [repository](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/themes/static) to `~/.config/hyprdynamicmonitors/themes/`

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "~/.config/hyprdynamicmonitors/themes/my_theme.toml"
```

## Dynamic Theming with Generators

For automatic color scheme generation based on your wallpaper, `hyprdynamicmonitors` provides templates for popular generators.

### Wallust

[Wallust](https://codeberg.org/explosion-mental/wallust) generates color schemes from your wallpaper using image analysis.

**1. Configure your hyprdynamicmonitors config to use the generated theme:**

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "~/.config/hyprdynamicmonitors/themes/wallust.toml"
```

**2. Configure wallust** to generate the theme file at `~/.config/hyprdynamicmonitors/themes/wallust.toml` using the provided template at `/usr/share/hyprdynamicmonitors/themes/templates/wallust/theme.toml` (or copy it from the [repository](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/themes/templates/wallust)), e.g.:
```toml title="~/.config/wallust/wallust.toml"
hyprdynamicmonitors.template = '/usr/share/hyprdynamicmonitors/themes/templates/wallust/theme.toml'
hyprdynamicmonitors.target = '~/.config/hyprdynamicmonitors/themes/theme.toml'
```

**3. Run wallust** - it will generate the theme file with colors extracted from your wallpaper

### Matugen

[Matugen](https://github.com/InioX/matugen) generates Material You color schemes with proper color theory.

**1. Configure your hyprdynamicmonitors config to use the generated theme:**

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "~/.config/hyprdynamicmonitors/themes/matugen.toml"
```

**2. Configure matugen** to generate the theme file at `~/.config/hyprdynamicmonitors/themes/matugen.toml` using the provided template at `/usr/share/hyprdynamicmonitors/themes/templates/matugen/theme.toml` (or copy it from the [repository](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/themes/templates/matugen)), e.g.:
```toml title="~/.config/matugen/config.toml"
[templates.hyprdynamicmonitors]
input_path = "/usr/share/hyprdynamicmonitors/themes/templates/matugen/theme.toml"
output_path = "~/.config/hyprdynamicmonitors/themes/matugen.toml"
```

**3. Run matugen** - it will generate Material You colors based on your wallpaper

**Tip:** Both generators can run automatically when your wallpaper changes. The TUI detects theme file changes instantly without requiring a restart.

### Pywal

[Pywal](https://github.com/dylanaraps/pywal) generates color schemes from wallpapers and applies them system-wide.

**1. Configure your hyprdynamicmonitors config to use the generated theme:**

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "~/.cache/wal/hyprdynamicmonitors.toml"
```

**2. Configure pywal to use the template** by copying or linking the template from `/usr/share/hyprdynamicmonitors/themes/templates/pywal/theme.toml` (or copy it from the [repository](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/themes/templates/pywal)) to pywal's templates directory:

```bash
ln -s /usr/share/hyprdynamicmonitors/themes/templates/pywal/theme.toml ~/.config/wal/templates/hyprdynamicmonitors.toml
```

**3. Run pywal** - it will generate colors from your wallpaper and populate the template:

```bash
wal -i /path/to/wallpaper.png
```

Pywal will extract a 16-color palette from your wallpaper and generate `~/.cache/wal/hyprdynamicmonitors.toml` with the theme applied.

## Creating Custom Themes

To create your own theme:

1. Copy an existing theme file as a starting point
2. Modify the color values (use hex or ANSI codes)
3. Save to `~/.config/hyprdynamicmonitors/themes/my-theme.toml`
4. Reference it in your config:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[tui.colors]
source = "~/.config/hyprdynamicmonitors/themes/my-theme.toml"
```

See the [built-in themes](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/themes/static) for complete examples of all available color options.

:::info
If you're adding support for a well-known theme do not hesitate to [make a pull request](https://github.com/fiffeek/hyprdynamicmonitors/pulls) to the repository!
:::
