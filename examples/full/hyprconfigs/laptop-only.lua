-- Laptop Only Configuration
-- Simple static configuration for single laptop display

-- Monitor configuration
hl.monitor({ output = "eDP-1", mode = "2880x1920@120.00000", position = "0x0", scale = 2.0 })

-- Workspace assignment
hl.workspace_rule({ workspace = "1", monitor = "eDP-1", default = true })
hl.workspace_rule({ workspace = "2", monitor = "eDP-1" })

-- Input configuration optimized for laptop
hl.config({
    input = {
        touchpad = {
            natural_scroll = true,
            tap_to_click = true,
            middle_button_emulation = true,
        },
    },
})
