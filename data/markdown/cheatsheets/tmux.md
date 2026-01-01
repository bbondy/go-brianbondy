# tmux

## Info

Some of these shortcuts assume you have this [.tmux.conf file](https://github.com/bbondy/dotfiles/blob/master/.tmux.conf).


## Sessions

Creating a new session:  
`tmux new -s <session-name>`

Killing a session:  
`tmux kill-session -t <session-name>`

Attaching to a session:  
`tmux a -t <session-name>`

Getting a list of sessions:  
`Ctrl-b s`

Dettaching from a session:  
`Ctrl-b d`

##Pane management

Split vertically:  
`Ctrl-b %`

Split horizontally:  
`Ctrl-b "`

Killing the current pane:  
`Ctrl-b x`

## Pane navigation

Moving around to different panes:  
`Ctrl-b <arrow-key>`

## Reloading

Reloading config:  
`Ctrl-b r`

##Mouse

Mouse mode on/off:  
`Ctrl-b m`  
`Ctrl-b M`
