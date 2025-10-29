Source ./preview/tapes/base.tape
Output preview/output/demo.gif

Sleep 2s

Sleep 500ms
Enter
Sleep 500ms

# preview rotate
Type "r"
Sleep 500ms
Type "r"
Sleep 500ms

# preview scale
Type "s"
Sleep 1s
Down 10
Sleep 500ms
Enter
Sleep 500ms

# preview move
Right 4
Sleep 200ms
Up 4
Sleep 200ms
Left 4
Sleep 200ms
Down 4
Sleep 500ms
Right 2
Up 2
Sleep 500ms

# preview enable/disable
Type "e"
Sleep 500ms
Type "e"
Sleep 500ms

# preview vrr
Type "v"
Sleep 500ms

# go back to select another monitor
Enter
Sleep 200ms
Down
Sleep 500ms
Enter
Sleep 500ms

# change mode
Type "m"
Sleep 500ms
Down
Sleep 1s
Enter
Sleep 500ms

# fit to screen
Type "T"
Sleep 500ms

# change color profile to wide
Type "C"
Sleep 1s
Down 2
Sleep 1s
Enter
Sleep 500ms

# move and follow monitor
Type "o"
Sleep 500ms
Right 9
Sleep 500ms

# mirror
Sleep 500ms
Type "i"
Sleep 1s
Down
Sleep 500ms
Enter

# fullscreen move
Sleep 500ms
Type "F"
Right
Sleep 200ms
Up
Sleep 200ms
Left
Sleep 200ms
Down
Sleep 500ms

# pan
Type "p"
Right 10
Sleep 500ms
Left 10
Sleep 500ms
Down 10
Sleep 500ms

# mock apply
Type "A"
Sleep 1s
Type "n"
Sleep 500ms

# tab to configs
Sleep 1s
Tab
Sleep 1s
Type "e"
Sleep 500ms
Type "G"
Sleep 500ms
Type "o"
Type "another line can be added, like: exec = notify-send 'hello'"
Sleep 1s
Escape
Type ":wq"
Enter
Sleep 2s

# create a new profile
Type "n"
Sleep 500ms
Type "new_profile"
Enter
Sleep 2s

# see config
Type "C"
Sleep 1s
Type "G"
Sleep 1s
Type ":q"
Enter
Sleep 1s

# render config
Type "R"
Sleep 2s

Sleep 1s
