Output preview/output/lid_tui.gif

Require hyprdynamicmonitors

Set Shell "bash"
Set FontFamily "Iosevka Nerd Font Propo"
Set FontSize 16
Set Width 1400
Set Height 1000

Type "rm -f /tmp/lid.log" Sleep 500ms Enter
Type "set -x" Sleep 500ms Enter
Show

Type@0ms "./dist/hdmtest run --config $HDM_CONFIG $HDM_EXTRA_FLAGS >/tmp/lid.log 2>&1 &" Sleep 500ms Enter

Hide
Type@0ms "pid=$!" Sleep 500ms Enter
Type "set +x" Sleep 500ms Enter
Show

Hide Type@0ms "./dist/hdmtest tui --config $HDM_CONFIG $HDM_EXTRA_FLAGS" Sleep 500ms Enter
Show

Sleep 1s
Tab
Sleep 10s

Type "q"

Type "grep lid /tmp/lid.log" Sleep 500ms Enter
Sleep 1

Hide
Type "kill $pid || true"
Show
