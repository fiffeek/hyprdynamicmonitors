Output preview/output/lid_tui.gif

Require hyprdynamicmonitors

Set Shell "bash"
Set FontFamily "Iosevka Nerd Font Propo"
Set FontSize 16
Set Width 1400
Set Height 1000

Hide Type@0ms "./dist/hdmtest tui --config $HDM_CONFIG $HDM_EXTRA_FLAGS" Sleep 500ms Enter
Show

Sleep 1s
Tab
Sleep 5s
