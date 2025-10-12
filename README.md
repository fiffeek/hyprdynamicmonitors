<img src="https://github.com/user-attachments/assets/0effc242-3d3d-4d39-a183-0a567c4da3a9" width="90" style="margin-right:10px" align=left alt="hyprdynamicmonitors logo">
<H1>HyprDynamicMonitors</H1><br>


An event-driven service that automatically manages Hyprland monitor configurations based on connected displays, power and lid state.

## Preview
![demo](./preview/output/demo.gif)

## Documentation

Documentation can be viewed in individual files or on the [fiffeek.github.io/hyprdynamicmonitors](https://fiffeek.github.io/hyprdynamicmonitors) website.

<!--ts-->
<!--te-->

## Runtime requirements

- Hyprland with IPC support
- UPower (optional, for power/lid state monitoring)
- Read-only access to system D-Bus (optional for power state monitoring; should already be your default)
- Write access to system D-Bus for notifications (optional; should already be your default)

## Alternative software

Most similar tools are more generic, working with any Wayland compositor. In contrast, `hyprdynamicmonitors` is specifically designed for Hyprland (using its IPC) but provides several advantages:

**Advantages of HyprDynamicMonitors:**
- **Interactive TUI**: Built-in terminal interface for visual monitor configuration and profile management
- **Full configuration control**: Instead of introducing another configuration format, you work directly with Hyprland's native config syntax
- **Template system**: Dynamic configuration generation based on connected monitors and power state
- **Power state awareness**: Built-in AC/battery detection for laptop users
- **Event-driven automation**: Automatically responds to monitor connect/disconnect events

**Trade-offs:**
- Hyprland-specific (not generic Wayland)
- Requires systemd or wrapper script for production use (fail-fast design)

**Similar Tools:**
- [kanshi](https://sr.ht/~emersion/kanshi/) - Generic Wayland output management
- [shikane](https://github.com/hw0lff/shikane) - Another Wayland output manager
- [nwg-displays](https://github.com/nwg-piotr/nwg-displays) - GUI-based display configuration tool for Sway/Hyprland
- [hyprmon](https://github.com/erans/hyprmon) - TUI-based display configuration tool for Hyprland

