---
sidebar_position: 2
---

# Monitor Matching

Profiles match monitors based on their properties. Understanding how monitor matching works is essential for creating effective profiles.

## Matching Properties

You can match monitors by:

### Name

The monitor's connector name (e.g., `eDP-1`, `DP-1`, `HDMI-A-1`)

```toml
[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"  # Match by connector name
```

### Description

The monitor's model/manufacturer string as reported by Hyprland

```toml
[[profiles.external_4k.conditions.required_monitors]]
description = "Dell U2720Q"  # Match by monitor model
```

### Tags (Optional)

Custom labels you assign to monitors for easier reference in templates

```toml
[[profiles.dual_setup.conditions.required_monitors]]
name = "eDP-1"
monitor_tag = "laptop"  # Assign a tag for template use
```

Tags are particularly useful in Go templates where you can reference monitors by their tags:

```go
{{- $laptop := index .MonitorsByTag "laptop" -}}
monitor={{$laptop.Name}},preferred,0x0,1
```

## Finding Monitor Information

Use `hyprctl monitors` to see all available monitors and their properties:

```bash
$ hyprctl monitors
Monitor eDP-1 (ID 0):
        2880x1920@120.00000 at 0x0
        description: BOE 0x0C6B
        ...
```

From this output:
- **Name**: `eDP-1`
- **Description**: `BOE 0x0C6B`

## Profile Scoring and Selection

When multiple profiles could match the current setup, HyprDynamicMonitors uses a scoring system:

1. Each required monitor that's connected adds to the profile's score
2. The profile with the highest score is selected
3. If multiple profiles have equal scores, the **last profile** defined in the configuration file is selected

### Example: Profile Priority

```toml
# This profile will be selected for laptop-only setup
[profiles.laptop_basic]
config_file = "hyprconfigs/laptop-basic.conf"
[[profiles.laptop_basic.conditions.required_monitors]]
name = "eDP-1"

# This profile will be selected instead (defined last)
[profiles.laptop_optimized]
config_file = "hyprconfigs/laptop-optimized.conf"
[[profiles.laptop_optimized.conditions.required_monitors]]
name = "eDP-1"
```

When only `eDP-1` is connected, both profiles match, but `laptop_optimized` is selected because it's defined last.

## Best Practices

1. **Order matters**: Define more generic profiles first, specific ones last
2. **Test with `--dry-run`**: Use `hyprdynamicmonitors run --dry-run` to see which profile would be selected
3. **Use tags for templates**: Tags make templates more readable and maintainable

## See Also

- [Profiles](./profiles) - Learn about profile configuration
- [Templates](../advanced/templates) - Use tags in dynamic templates
- [Examples - Scoring](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/scoring) - Detailed scoring examples
