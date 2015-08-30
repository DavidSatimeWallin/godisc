#!/bin/bash
SESSION=$USER

tmux -2 new-session -d -s $SESSION

# Setup a window for tailing log files
tmux new-window -t $SESSION:1 -n 'GODISC'
tmux split-window -v
tmux select-pane -t 0
tmux send-keys "tail -f /home/$SESSION/.config/godisc/talkerChat.log" C-m
tmux split-window -h
tmux select-pane -t 1
tmux send-keys "tail -f /home/$SESSION/.config/godisc/tellChat.log" C-m
tmux select-pane -t 2
tmux resize-pane -t 2 -y 32
tmux send-keys "/home/$SESSION/gocode/bin/godisc discworld.starturtle.net 4242" C-m

# Set default window
tmux select-window -t $SESSION:1

# Attach to session
tmux -2 attach-session -t $SESSION
