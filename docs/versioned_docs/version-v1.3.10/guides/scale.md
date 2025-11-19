---
sidebar_position: 3
---

# Adjusting Monitor Scale in the TUI

## What is scale snapping?

Scale snapping is a feature in the TUI that automatically adjusts your monitor scale to valid values that Hyprland can render without fractional pixels. This prevents rendering artifacts and ensures crisp display output.

When you adjust the scale with the up/down arrow keys, the TUI automatically "snaps" to the nearest scale value that:
- Produces whole-number logical pixel dimensions (no fractional pixels)
- Is closest to your requested scale value
- Matches Hyprland's internal scale validation logic

## Why does this matter?

Hyprland requires that the logical size of your display (physical pixels divided by scale) results in whole numbers. For example:

**Valid scale examples** (for 1920x1080):
- Scale 1.0 → 1920÷1.0 = 1920, 1080÷1.0 = 1080 ✓
- Scale 1.5 → 1920÷1.5 = 1280, 1080÷1.5 = 720 ✓
- Scale 2.0 → 1920÷2.0 = 960, 1080÷2.0 = 540 ✓

**Invalid scale examples** (for 1920x1080):
- Scale 1.3 → 1920÷1.3 = 1476.92..., 1080÷1.3 = 830.77... ✗ (fractional pixels)
- Scale 1.7 → 1920÷1.7 = 1129.41..., 1080÷1.7 = 635.29... ✗ (fractional pixels)

Invalid scales can cause:
- Blurry text and UI elements
- Rendering artifacts
- Incorrect pixel alignment
- Visual glitches in applications

## Using the scale selector

### Access the scale selector

In the TUI, navigate to the Monitors view and:
1. Select a monitor with arrow keys
2. Press `Enter` to edit
3. Press `s` to access the scale selector

### Adjusting scale

**Basic controls:**
- `↑` or `k` - Increase scale
- `↓` or `j` - Decrease scale
- `Enter` - Apply the scale and return to monitor editor
- `Esc` - Cancel and return to monitor editor

**Acceleration:**
Hold the up/down keys to increase the increment speed:
- Single press: Fine adjustment (0.005 increments)
- Rapid presses (2-3): 2x speed (0.01 increments)
- Rapid presses (4-5): 5x speed (0.025 increments)
- Rapid presses (6+): 10x speed (0.05 increments)

### Toggle scale snapping

Press `e` to toggle scale snapping on/off:

**With snapping enabled (default):**
- Title shows: "Adjust scale (snapping)"
- Step size: 0.005
- Automatically finds valid scales based on your monitor resolution
- Recommended for most users

**With snapping disabled:**
- Title shows: "Adjust scale"
- Step size: 0.0001 (very fine control)
- Allows any scale value (may produce fractional pixels)
- Useful for testing or when you know you need a specific scale

:::tip
Keep scale snapping enabled unless you have a specific reason to disable it. Hyprland will still apply its own validation, so invalid scales may be adjusted when applied.
:::

## How scale validation works

The scale validation algorithm is based on [Hyprland's internal implementation](https://github.com/hyprwm/Hyprland/blob/8e9add2afda58d233a75e4c5ce8503b24fa59ceb/src/helpers/Monitor.cpp#L895) and works as follows:

1. **Check if already valid**: If your requested scale produces whole-number logical dimensions, it's used as-is

2. **Try common scales**: Tests a list of well-known scales:
   - 0.50, 0.75, 0.90, 1.00, 1.10, 1.125, 1.25, 1.3333, 1.50, 1.6667, 1.75, 2.00, 2.125, 2.25, 2.50, 2.6667, 2.75, 3.00

3. **Search in increments**: If no common scale works, searches in fine increments:
   - 1/120th increments (0.00833...)
   - 1/1000th increments (0.001)
   - 1/600th increments (0.00166...)

4. **Find closest match**: Selects the valid scale with the smallest difference from your requested value

## Examples

### Example 1: 1920x1080 monitor

Starting from scale 1.0, pressing up/down will snap to these valid scales:

**Scales below 1.0:**
- 0.90 → 1920÷0.9 = 2133.33... ✗ (skipped)
- 0.75 → 1920÷0.75 = 2560, 1080÷0.75 = 1440 ✓
- 0.50 → 1920÷0.5 = 3840, 1080÷0.5 = 2160 ✓

**Scales above 1.0:**
- 1.125 → 1920÷1.125 = 1706.67... ✗ (skipped)
- 1.2 → 1920÷1.2 = 1600, 1080÷1.2 = 900 ✓
- 1.25 → 1920÷1.25 = 1536, 1080÷1.25 = 864 ✓
- 1.3333... → 1920÷1.3333 = 1440, 1080÷1.3333 = 810 ✓
- 1.5 → 1920÷1.5 = 1280, 1080÷1.5 = 720 ✓
- 2.0 → 1920÷2.0 = 960, 1080÷2.0 = 540 ✓

### Example 2: 2560x1440 monitor (1440p)

Valid scales include:
- 1.0 → 2560x1440 logical
- 1.25 → 2048x1152 logical ✓
- 1.6667 → 1536x864 logical ✓
- 2.0 → 1280x720 logical ✓

### Example 3: 3840x2160 monitor (4K)

Valid scales include:
- 1.0 → 3840x2160 logical
- 1.5 → 2560x1440 logical ✓
- 2.0 → 1920x1080 logical ✓
- 2.25 → 1706.67x960 ✗ (skipped, fractional)
- 3.0 → 1280x720 logical ✓

## Troubleshooting

### Scale jumps to unexpected values

This is normal when scale snapping is enabled. The TUI is finding the nearest valid scale for your monitor resolution. Try:
- Checking which scales are valid for your monitor's resolution
- Using smaller increments by tapping up/down instead of holding
- Disabling snapping with `e` if you need a specific scale value

### Scale doesn't change when pressing up/down

Possible causes:
- You're at the minimum (0.1) or maximum (10.0) scale limit
- The next valid scale is the same as your current scale
- Try pressing multiple times or holding the key

### Applied scale looks different in Hyprland

If you disable scale snapping and apply an invalid scale:
- Hyprland will apply its own validation
- The scale may be adjusted to the nearest valid value
- Some scales may be rejected entirely
- Always verify the scale after applying in Hyprland

:::warning
If you disable scale snapping and apply an invalid scale, Hyprland may override it with a different value or display rendering issues. It's recommended to keep scale snapping enabled.
:::

## Advanced usage

### Testing multiple scales quickly

1. Press `e` to disable snapping
2. Use fine adjustments (0.0001 steps) to test exact values
3. Note which scales produce whole-number logical dimensions
4. Re-enable snapping with `e`
5. Adjust to the valid scales you identified

### Finding all valid scales for your monitor

To see which scales work for your resolution:

```bash
# For a 1920x1080 monitor
python3 -c "
res_w, res_h = 1920, 1080
for i in range(10, 300):
    s = i / 100
    if res_w % 1 == 0 and res_h % 1 == 0:
        if (res_w / s).is_integer() and (res_h / s).is_integer():
            print(f'Scale {s:.2f}: {int(res_w/s)}x{int(res_h/s)}')
"
```

This will show you all valid scales between 0.10 and 3.00 for your resolution.

### Validating scales in configuration files

When using the daemon with manually configured scales, ensure they produce whole-number logical dimensions:

```toml
[profiles.high_dpi]
config_file = "hyprconfigs/high_dpi.conf"

# In hyprconfigs/high_dpi.conf:
# monitor=eDP-1,3840x2160@60,0x0,2.0  # 2.0 is valid (1920x1080 logical)
# monitor=eDP-1,3840x2160@60,0x0,1.5  # 1.5 is valid (2560x1440 logical)
# monitor=eDP-1,3840x2160@60,0x0,1.7  # 1.7 is INVALID (2258.82x1270.59 logical)
```

## Related documentation

- [TUI Guide](../quickstart/tui) - Complete TUI usage guide
- [Monitor Configuration](../configuration/profiles) - Configure monitor profiles
- [Hyprland Monitor Scaling](https://wiki.hyprland.org/Configuring/Monitors/#scaling) - Official Hyprland scaling documentation
