-- Some existing config
require("monitors")

-- <<<<< TUI AUTO START
hl.monitor({ output = "desc:Old Monitor", mode = "1920x1080@60.00000", position = "0x0", scale = 1.00, transform = 0 })
-- Missing end marker

hl.bind("SUPER + Q", hl.dsp.exec_cmd("kitty"))
