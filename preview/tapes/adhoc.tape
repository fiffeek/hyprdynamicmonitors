Require hdmtest
Require nvim

Set Shell "bash"
Set FontFamily "Iosevka Nerd Font Propo"
Set FontSize 16
Set Width 1400
Set Height 1000

Hide Type "hdmtest tui --hypr-monitors-override ./preview/tapes/fixtures/two.json --config ./preview/tapes/configs/adhoc.toml" Sleep 500ms  Enter
Show

Output preview/output/ad-hoc.gif

Sleep 1s
Tab
Sleep 1s

Type "C"
Sleep 1s

Type "6G"
Sleep 500ms
Type "o"
Sleep 500ms
Enter
Type "[profiles.home.static_template_values]"
Escape
Sleep 500ms
Type "o"
Type `mode = "standard"`
Sleep 500ms
Escape
Sleep 500ms
Type ":wq"
Sleep 500ms
Enter
Sleep 1s

Type "e"
Sleep 500ms
Type "2G"
Sleep 500ms
Type "O"
Backspace
Backspace
Type `{{ if eq .mode "standard" }}`
Escape
Type "G"
Sleep 500ms
Type "o"
Sleep 500ms
Backspace
Backspace
Sleep 500ms
Type "{{ end }}"
Escape
Sleep 500ms
Type ":wq"
Sleep 500ms
Enter
Sleep 500ms
Type "R"
Sleep 1s

Type "e"
Sleep 1s
Type "2G"
Sleep 500ms
Type "yy"
Sleep 500ms
Type "G"
Sleep 500ms
Type "p"
Sleep 500ms
Type "7G"
Sleep 500ms
Type "yy"
Sleep 500ms
Type "G"
Sleep 500ms
Type "p"
Sleep 500ms
Type "3G"
Sleep 500ms
Type "dd"
Sleep 500ms
Type "7G"
Sleep 500ms
Type "p"
Sleep 500ms
Type "5G"
Sleep 500ms
Type "dd"
Sleep 500ms
Type "7G"
Sleep 500ms
Type "p"
Sleep 500ms
Type "6G"
Sleep 500ms
Type "0"
Type "w"
Type "w"
Type "w"
Type "w"
Type "w"
Type "w"
Sleep 500ms
Type "cw"
Sleep 500ms
Type "presentation"
Sleep 500ms
Escape
Sleep 500ms
Type ":wq"
Sleep 500ms
Enter
Sleep 1s
Type "R"
Sleep 1s

Tab
Sleep 1s
Enter
Sleep 500ms
Type "m"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Enter
Sleep 1s
Type "s"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Enter
Sleep 1s

Tab
Sleep 1s
Type "a"
Sleep 1s
Type "Y"
Sleep 1s
Type "R"
Sleep 1s

Type "C"
Sleep 500ms
Type "9G"
Sleep 500ms
Type "0"
Sleep 500ms
Type "w"
Type "w"
Type "w"
Type "cw"
Type "presentation"
Sleep 500ms
Escape
Sleep 500ms
Type ":wq"
Sleep 500ms
Enter
Sleep 1s
Type "R"
Sleep 3s
