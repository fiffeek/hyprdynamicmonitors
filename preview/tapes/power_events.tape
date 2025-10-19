Output preview/output/power_events.gif

Require grep
Require pkill
Require cat
Require sleep
Require hyprdynamicmonitors

Set Shell "bash"
Set FontFamily "Iosevka Nerd Font Propo"
Set FontSize 16
Set Width 1400
Set Height 1000
Hide

Type "rm -f /tmp/power.log" Sleep 500ms Enter
Type "set -x" Sleep 500ms Enter
Show

Type@0ms "hyprdynamicmonitors run --config $HDM_CONFIG $HDM_EXTRA_FLAGS >/tmp/power.log 2>&1 &" Sleep 500ms Enter

Hide
Type@0ms "pid=$!" Sleep 500ms Enter
Type "set +x" Sleep 500ms Enter
Show

Type "cat $HDM_DESTINATION" Sleep 500ms Enter
Type@0ms "echo 'Wait for power change event'; sleep 5" Sleep 500ms Enter
Sleep 5
Type "grep profile_name /tmp/power.log" Sleep 500ms Enter
Sleep 1
Type "cat $HDM_DESTINATION" Sleep 500ms Enter
Sleep 3

Hide
Type "kill $pid || true"
Show
