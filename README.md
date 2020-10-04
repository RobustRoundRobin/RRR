# devclutter

Tooling in support of rororo development and the addition of RoRoRo consensus
to quorum

As the name hopefuly implies, this repository is just a home for development
tooling that we don't want to clutter up the rororo package repository or the
quorum fork. It's "not supported" in any way and can change without notice.

## Go Versions etc

All current development happens on macos using go 1.14. Any posix platform
should work but there may be rough edges.

## Visual Studio Code for development and debuging

If a Visual Studio Code environment is useful to you, conform to the Layout
assumptions (below) and you can use ./vscode/launch.json 'as is' after opening
the quorum-rororo-gopath folder. All the usual code navigation features in
vscode should 'just work'

## Truffle for test support contracts

[eth-enabled-cli-tools-with-truffle](https://www.trufflesuite.com/tutorials/creating-a-cli-with-truffle-3)

Its probably the easiest path to generating transactions for development and
testing.


## Layout assumptions

To simplify life, the tooling in this repository makes assumptions about the
relative locations for these repositories:

* https://github.com/RobustRoundRobin/quorum.git
* https://github.com/RobustRoundRobin/rororo.git
* https://github.com/RobustRoundRobin/devcluter.git

And expects certain symlinks to be created.

Pick any ROOT directory.

1. The quorum fork must be cloned to ROOT/qorum-rororo-gopath/src/github.com/ethereum/go-ethereum
2. There must be a symlink ROOT/quorum-rororo -> ROOT/qorum-rororo-gopath/src/github.com/ethereum/go-ethereum
3. rororo must be cloned to ROOT/quorum-rororo-gopath/src/github.com/RobustRoundRobin/rororo
4. There must be a symlink ROOT/rororo -> ROOT/qorum-rororo-gopath/src/github.com/RobustRoundRobin/rororo
5. devclutter must be cloned directly under ROOT. Call it ROOT/rororo-devclutter if you want the vscode support

If Visual Studio Code suites your needs, then create a symlink to the supplied
vscode config, (or derive your own.)

   ROOT/quorum-rororo-gopath/.vscode -> ROOT/rororo-devclutter

Having done all of that open ROOT/quorum-rororo-gopath as a "folder" in vscode.

## tusk.yml

Uses [go-tusk](https://rliebz.github.io/tusk/) to provide a collection of runes
considered useful for developing rororo. Try `tusk -q -f ./tusk.yml -h`
