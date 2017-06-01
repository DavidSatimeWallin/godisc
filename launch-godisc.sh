#!/bin/bash
SESSION=$USER

CONFDIR="/home/$SESSION/.config/godisc"
COMMFILE="$CONFDIR/tellChat.log"
XPFILE="$CONFDIR/xp.log"

if [[ ! -e $CONFDIR ]]; then
    mkdir $CONFDIR
elif [[ ! -d $CONFDIR ]]; then
    echo "$CONFDIR already exists but is not a directory" 1>&2
fi

if [ ! -e "$COMMFILE" ] ; then
    touch "$COMMFILE"
fi

if [ ! -w "$COMMFILE" ] ; then
    echo cannot write to $COMMFILE
    exit 1
fi

if [ ! -e "$XPFILE" ] ; then
    touch "$XPFILE"
fi

if [ ! -w "$XPFILE" ] ; then
    echo cannot write to $XPFILE
    exit 1
fi


tmux -2 new-session -d -s $SESSION

# Setup a window for tailing log files
tmux new-window -t $SESSION:1 -n 'GODISC'
tmux split-window -v
tmux select-pane -t 0
tmux send-keys "tail -f /home/$SESSION/.config/godisc/tellChat.log" C-m
tmux split-window -h
tmux select-pane -t 1
tmux resize-pane -t 1 -x 30
tmux send-keys "tail -f /home/$SESSION/.config/godisc/xp.log" C-m
tmux select-pane -t 2
tmux resize-pane -t 2 -y 42
tmux send-keys "/usr/local/bin/godisc discworld.starturtle.net 4242" C-m

# Set default window
tmux select-window -t $SESSION:1

# Attach to session
tmux -2 attach-session -t $SESSION
