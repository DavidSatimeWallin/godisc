# What is godisc
godisc is a MUD client used to play the [Discworld MUD](http://discworld.starturtle.net) game. It is still very much in early, possibly unstable, release. godisc is a cli tool written in go.

![godisc mud client in tmux](/godisc.png?raw=true "godisc mud client in tmux")

# How to use godisc
godisc is best used together with [tmux](https://tmux.github.io/).
* Start a tmux session
* press `CTRL+%` once
* move to upper pane
* press `CTRL+"`
* move to lower pane
* run `godisc discworld.starturtle.net`
* move to upper left pane
* run `tail -f ~/.config/godisc/talkerChat.log`
* move to upper right pane
* run `tail -f ~/.config/godisc/tellChat.log`
* move back to big lower pane and enjoy the game

# How to install
* `cd $GOPATH`
* `go get github.com/dvwallin/godisc`
* `go install github.com/dvwallin/godisc`

# License
As seen in the (LICENSE file)[https://github.com/dvwallin/godisc/blob/master/LICENSE] godisc is licensed under the MIT License.

# Who built it
David V. Wallin wrote godisc as a tiny fun hobby project.

# Contact
David V. Wallin can be reached on (twitter)[https://twitter.com/dvwallin] or on email < david@dwall.in >

# Issues and Bugs
Any issues, bugs and requests can be posted (here)[https://github.com/dvwallin/godisc/issues]

# Contribute
If you would like to contribute to the project please do so by 
* reporting issues and bugs
* giving suggestions
* writing guides
* forking the project and giving pull requests
* creating new projects based on it
