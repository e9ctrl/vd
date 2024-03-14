Require tmux

Output readme.gif
Set PlaybackSpeed 0.5
Set Width 2500
Set Height 1200
Set Margin 20
Set MarginFill "#5468FF"
Set BorderRadius 10
Hide
Type "tmux -f /dev/null -L test new-session -- bash" Enter
Type "tmux split-window -d -h -- bash && \" Enter
Type "tmux split-window -d -h -- bash && \" Enter
Type "tmux select-layout even-horizontal && \" Enter
Type "tmux set status && \" Enter
Type 'tmux setw pane-border-style "fg=0" && \' Enter
Type 'tmux setw pane-active-border-style "fg=0"' Enter
Ctrl+B
Right
Sleep 0.1
Ctrl+B
Right
Sleep 0.1
Type "tmux split-window -d -v -- bash" Enter
Ctrl+L
Sleep 0.5
Sleep 1

Ctrl+B
Left
Ctrl+B
Left
Ctrl+L

Show
Type "#Configuration file" Enter
Type "cat example" Enter
Sleep 5
Ctrl+B
Right
Type "#Simulator" Enter
Type "./vd example" Enter
Sleep 5
Ctrl+B
Right
Type "#Client" Enter
Type "telnet 127.0.0.1 9999" Enter
Type "CUR?" Enter
Sleep 5

Ctrl+B
Down
Type "#Built in HTTP API client" Enter

Type "./vd get current" Enter
Sleep 5

Ctrl+B
Up

Type "CUR 30" Enter 
Sleep 5

Ctrl+B
Down

Type "./vd get current" Enter
Sleep 5

Type "./vd set current 50" Enter
Sleep 5

Ctrl+B
Up

Type "CUR?" Enter
Sleep 20

Ctrl+B
Left
Ctrl+C
