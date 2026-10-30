-- Example: hl.monitor({ output = "DP-1", disabled = true })
-- Note: monitor settings = all disabled configurations
hl.monitor({ output = "eDP-1", mode = "2880x1920@120.00000", position = "0x0", scale = 2.0, vrr = 1 })
-- This comment should be kept: normal comment
hl.monitor({ output = "DP-2", disabled = true })
-- Another hl.monitor({ output = "something", disabled = true }) comment that matches
hl.workspace_rule({ workspace = "1", monitor = "eDP-1", default = true })
