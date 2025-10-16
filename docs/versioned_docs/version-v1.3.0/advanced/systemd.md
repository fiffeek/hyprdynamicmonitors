---
sidebar_position: 3
---

# Running with systemd

For production use, it's recommended to run HyprDynamicMonitors as a systemd user service. This ensures automatic restart on failures and proper integration with session management.

:::tip
Ensure you're properly [pushing environment variables to systemd](https://wiki.hypr.land/Nix/Hyprland-on-Home-Manager/#programs-dont-work-in-systemd-services-but-do-on-the-terminal) for Hyprland integration.
:::

## Hyprland under systemd

If you run [Hyprland under systemd](https://wiki.hypr.land/Useful-Utilities/Systemd-start/), setup is straightforward.

Create `~/.config/systemd/user/hyprdynamicmonitors.service`:

```ini
[Unit]
Description=HyprDynamicMonitors - Dynamic monitor configuration for Hyprland
After=graphical-session.target
Wants=graphical-session.target
PartOf=hyprland-session.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors run
Restart=on-failure
RestartSec=5

[Install]
WantedBy=hyprland-session.target
```

Enable and start the service:

```bash
systemctl --user daemon-reload
systemctl --user enable hyprdynamicmonitors
systemctl --user start hyprdynamicmonitors
```

## Run on Boot with Restarts

You can run the service on boot and let restarts handle initialization:

```ini
[Unit]
Description=HyprDynamicMonitors - Dynamic monitor configuration for Hyprland
After=default.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors run
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

This approach:
- Keeps failing until Hyprland is ready/launched
- Waits for environment variables to be propagated
- Automatically recovers once the system is ready

## Custom systemd Target

You can create [a custom systemd target started by Hyprland](https://github.com/fiffeek/.dotfiles.v2/commit/2a0d400b81031e3786a2779c36f70c9771aee884).

In `~/.config/hypr/hyprland.conf`:

```conf
exec-once = systemctl --user start hyprland-custom-session.target
bind = $mainMod, X, exec, systemctl --user stop hyprland-session.target
```

Create `~/.config/systemd/user/hyprland-custom-session.target`:

```ini
[Unit]
Description=A target for other services when hyprland becomes ready
After=graphical-session-pre.target
Wants=graphical-session-pre.target
BindsTo=graphical-session.target
```

Then create `~/.config/systemd/user/hyprdynamicmonitors.service`:

```ini
[Unit]
Description=Run hyprdynamicmonitors daemon
After=hyprland-custom-session.target
After=dbus.socket
Requires=dbus.socket
PartOf=hyprland-custom-session.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors run
Restart=on-failure
RestartSec=5

[Install]
WantedBy=hyprland-custom-session.target
```

## Alternative: Wrapper Script

If you prefer a wrapper script approach, create a simple restart loop:

```bash
#!/bin/bash
while true; do
    /usr/bin/hyprdynamicmonitors run
    echo "HyprDynamicMonitors exited with code $?, restarting in 5 seconds..."
    sleep 5
done
```

Then execute it from Hyprland:

```conf
exec-once = /path/to/the/script.sh
```

## Service Management

Once set up as a systemd service, you can manage it with standard systemd commands:

```bash
# Check status
systemctl --user status hyprdynamicmonitors

# View logs
journalctl --user -u hyprdynamicmonitors -f

# Restart the service
systemctl --user restart hyprdynamicmonitors

# Reload configuration (sends SIGHUP)
systemctl --user reload hyprdynamicmonitors

# Stop the service
systemctl --user stop hyprdynamicmonitors

# Disable automatic start
systemctl --user disable hyprdynamicmonitors
```

## Troubleshooting

### Service fails to start

Check the logs:

```bash
journalctl --user -u hyprdynamicmonitors -n 50
```

Common issues:
- Hyprland not running yet (expected with fail-fast design)
- Missing environment variables
- Invalid configuration file

### Configuration changes not applying

Reload the service:

```bash
systemctl --user reload hyprdynamicmonitors
```

Or rely on automatic hot reload (enabled by default).

### Environment variables not available

Ensure environment variables are properly exported to systemd. See [Hyprland's systemd integration guide](https://wiki.hypr.land/Nix/Hyprland-on-Home-Manager/#programs-dont-work-in-systemd-services-but-do-on-the-terminal).

## See Also

- [Usage - Signals](../usage/signals) - Signal handling and hot reload
- [Quick Start](../category/quick-start) - Initial setup guide
