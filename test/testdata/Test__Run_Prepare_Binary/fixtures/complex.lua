-- <<<< TUI
-- whatever
hl.monitor({ output = "eDP-1", mode = "2880x1920@120.00000", position = "0x0", scale = 2.0, vrr = 1 })
hl.monitor({ output = "DP-11", disabled = true })
hl.workspace_rule({ workspace = "1", monitor = "eDP-1", default = true })
hl.monitor({output="DP-12",disabled=true})
hl.monitor({ output = "DP-13", disabled = true, })
hl.monitor({ disabled = true, output = "DP-14" })
  hl.monitor({ output = "DP-15", disabled  =  true })
hl.monitor({ output = "DP-16", disabled = false })

hl.config({
  decoration = {
    blur = {
      enabled = false, -- Disable blur for performance
    },
  },
  animations = {
    enabled = false, -- Disable animations for performance
  },
  misc = {
    vrr = 3,
  },
})
hl.exec_cmd("brightnessctl set 100")
-- comment
