-- External Monitor Only Configuration
-- Static configuration for using only external display

-- Disable laptop screen, use external only
hl.monitor({ output = "eDP-1", disabled = true })
hl.monitor({ output = "desc:LG Electronics LG SDQHD 301NTBKDU037", mode = "2560x2880@59.96700", position = "0x0", scale = 1.6, transform = 3 })

-- Workspace assignment - all on external
hl.workspace_rule({ workspace = "1", monitor = "desc:LG Electronics LG SDQHD 301NTBKDU037", default = true })
hl.workspace_rule({ workspace = "2", monitor = "desc:LG Electronics LG SDQHD 301NTBKDU037" })
