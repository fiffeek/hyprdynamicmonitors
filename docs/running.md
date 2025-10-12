---
layout: default
title: Deploy
---

## Running with systemd

For production use, it's recommended to run HyprDynamicMonitors as a systemd user service. This ensures automatic restart on failures and proper integration with session management.

**Important**: Ensure you're properly [pushing environment variables to systemd](https://wiki.hypr.land/Nix/Hyprland-on-Home-Manager/#programs-dont-work-in-systemd-services-but-do-on-the-terminal).

### Hyprland under systemd
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
ExecStart=/usr/bin/hyprdynamicmonitors
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

### Run on boot and let restarts do the job
You can essentially just run it on boot and add restarts, e.g.:
```ini
[Unit]
Description=HyprDynamicMonitors - Dynamic monitor configuration for Hyprland
After=default.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```
It will keep failing until Hyprland is ready/launched and environment variables are propagated.

### Custom systemd target
You can also add [a custom systemd target that would be started by Hyprland](https://github.com/fiffeek/.dotfiles.v2/commit/2a0d400b81031e3786a2779c36f70c9771aee884), e.g.
```
exec-once = systemctl --user start hyprland-custom-session.target
bind = $mainMod, X, exec, systemctl --user stop hyprland-session.target
```

Then:
```bash
❯ cat ~/.config/systemd/user/hyprland-custom-session.target
[Unit]
Description=A target for other services when hyprland becomes ready
After=graphical-session-pre.target
Wants=graphical-session-pre.target
BindsTo=graphical-session.target
```
And:
```bash
❯ cat ~/.config/systemd/user/hyprdynamicmonitors.service
[Unit]
Description=Run hyprdynamicmonitors daemon
After=hyprland-custom-session.target
After=dbus.socket
Requires=dbus.socket
PartOf=hyprland-custom-session.target

[Service]
Type=exec
ExecStart=/usr/bin/hyprdynamicmonitors
Restart=on-failure
RestartSec=5


[Install]
WantedBy=hyprland-custom-session.target
```

### Alternative: Wrapper script

If you prefer a wrapper script approach, create a simple restart loop:

```bash
#!/bin/bash
while true; do
    /usr/bin/hyprdynamicmonitors
    echo "HyprDynamicMonitors exited with code $?, restarting in 5 seconds..."
    sleep 5
done
```
Then execute it from Hyprland:
```
exec-once = /path/to/the/script.sh
```


