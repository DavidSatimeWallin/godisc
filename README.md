# What is godisc
godisc is a MUD client used to play the [Discworld MUD](http://discworld.starturtle.net) game. It is still very much in early, possibly unstable, release. godisc is a cli tool written in go.

You need to have tmux and Go installed!

![godisc mud client in tmux](/godisc.png?raw=true "godisc mud client in tmux")

# How to install and use

## From Source

* `cd $GOPATH`
* `go get github.com/dvwallin/godisc`
* `go install github.com/dvwallin/godisc`
* `sudo cp $GOPATH/bin/godisc /usr/local/bin/`
* `$GOPATH/src/github.com/dvwallin/godisc/launch-godisc.sh`

## From binary

( Only amd64 supported on Linux so far )

* `git clone https://github.com/dvwallin/godisc.git ~/godisc`
* `cp ~/godisc/binaries/amd64-godisc /usr/local/bin/godisc`
* `~/godisc/launch-godisc.sh`

# License
As seen in the [LICENSE file](https://github.com/dvwallin/godisc/blob/master/LICENSE) godisc is licensed under the MIT License.

# Who built it
David V. Wallin wrote godisc as a tiny fun hobby project.

# Contact
David V. Wallin can be reached on [twitter](https://twitter.com/dvwallin) or on email < david@dwall.in >

# Issues and Bugs
Any issues, bugs and requests can be posted [here](https://github.com/dvwallin/godisc/issues)

# Contribute
If you would like to contribute to the project please do so by 
* reporting issues and bugs
* giving suggestions
* writing guides
* forking the project and giving pull requests
* creating new projects based on it
