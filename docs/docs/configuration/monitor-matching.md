---
sidebar_position: 2
---

# Monitor Matching

Profiles match monitors based on their properties. Understanding how monitor matching works is essential for creating effective profiles.

:::tip
For fake outputs (headless displays) the description from `hyprctl` will be empty so use the `name` matcher.
:::

## Matching Properties

You can match monitors by:

### Name

The monitor's connector name (e.g., `eDP-1`, `DP-1`, `HDMI-A-1`)

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[[profiles.laptop_only.conditions.required_monitors]]
name = "eDP-1"  # Match by connector name
```

:::info
You can use regex matching for the monitor name:
```toml title="~/.config/hyprdynamicmonitors/config.toml"
[[profiles.laptop_only.conditions.required_monitors]]
name = "DP-.*"
match_name_using_regex = true
```
It is disabled by default.
:::

### Description

The monitor's model/manufacturer string as reported by Hyprland

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[[profiles.external_4k.conditions.required_monitors]]
description = "Dell U2720Q"  # Match by monitor model
```

:::info
You can use regex matching for the monitor description:
```toml title="~/.config/hyprdynamicmonitors/config.toml"
[[profiles.laptop_only.conditions.required_monitors]]
description = "Dell.*"
match_description_using_regex = true
```

It is disabled by default.
:::

## Tags (Optional)

Custom labels you assign to monitors for easier reference in templates

```toml title="~/.config/hyprdynamicmonitors/config.toml"
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

```toml title="~/.config/hyprdynamicmonitors/config.toml"
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

### Customizing Scoring Weights

You can customize the scoring weights to prioritize different matching criteria:

```toml title="~/.config/hyprdynamicmonitors/config.toml"
[scoring]
name_match = 10       # Points for exact monitor name match
description_match = 5 # Points for exact monitor description match
power_state_match = 3 # Bonus points for matching power state
lid_state_match = 2   # Bonus points for matching lid state
```

Higher values give more weight to specific criteria. For example, if you want power state matching to have more influence, increase the `power_state_match` value.

See [Configuration Overview](./overview#scoring) for more details.

## Regex matching for monitor name or description

You can use regexes to match similar monitor setups, e.g. you are in a library that has very similar monitors but with different `description` each -- you can write
a catch all profile to match these:
```toml title="~/.config/hyprdynamicmonitors/config.toml"
[[profiles.library.conditions.required_monitors]]
description = "Dell.*"
match_description_using_regex = true
```

It is important to note that one `required_monitors` will match exactly one connected output, e.g. imagine that you're connected to two `Dell1` and `Dell2` monitors:
```bash
❯ hyprctl monitors | grep desc
        description: Dell1
        description: Dell2
```

Writing a profile like this:
```toml title="~/.config/hyprdynamicmonitors/config.toml"
[[profiles.library_one.conditions.required_monitors]]
description = "Dell.*"
match_description_using_regex = true
```
This profile will match because all required monitors (just one) are present. However, it only matches one of the two Dell monitors.

To match both Dell monitors, define two required_monitor entries:
```toml title="~/.config/hyprdynamicmonitors/config.toml"
[[profiles.library_two.conditions.required_monitors]]
description = "Dell.*"
match_description_using_regex = true

[[profiles.library_two.conditions.required_monitors]]
description = "Dell.*"
match_description_using_regex = true
```

Since one rule matches to one output, this profile will also match the current setup -- but since it defines more constrained rules (see [scoring](#scoring-system)) it would be picked up as the current setup.

Moreover, if you do assign tags, they are deterministic -- so for exactly the same setup (`monitor id` matters), same template would be produced.

## Best Practices

1. **Order matters**: Define more generic profiles first, specific ones last
2. **Test with `--dry-run`**: Use `hyprdynamicmonitors run --dry-run` to see which profile would be selected
3. **Use tags for templates**: Tags make templates more readable and maintainable

## See Also

- [Profiles](./profiles) - Learn about profile configuration
- [Templates](../advanced/templates) - Use tags in dynamic templates
- [Examples - Scoring](https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples/scoring) - Detailed scoring examples
