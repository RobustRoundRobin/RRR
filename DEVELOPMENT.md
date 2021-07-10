# RRR Setup

Tooling in support of RRR development and the addition of RRR consensus
to quorum

This repository is currently pre-alpha.
It's "not supported" in any way and can change without notice.
If you do find a bug, do feel free to raise an issue, we will get back to you.

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

The tooling in this repository makes assumptions about the relative locations
for these repositories:

* https://github.com/RobustRoundRobin/quorum.git
* https://github.com/RobustRoundRobin/RRR.git

Pick any ROOT directory.

1. The quorum fork must be cloned to

   ROOT/gopath/src/github.com/ethereum/go-ethereum

2. RRR must be cloned directly under ROOT. If you want to use the vscode files
   as is, clone it to ROOT/rrr

If Visual Studio Code suites your needs, then create a symlink to the supplied
vscode config, (or derive your own.)

   ROOT/gopath/.vscode -> ROOT/rrr/vscode

Having done all of that open ROOT/gopath as a "folder" in vscode.

## tusk.yml

Uses [go-tusk](https://rliebz.github.io/tusk/) to provide a collection of runes
considered useful for developing rororo. Try `tusk -q -f ./tusk.yml -h`

* ./tusk-genesis.yaml commands for initialising nodes for use with docker
    compose
* ./tusk.yml a small number of generaly helpful commands - execuging js on a
    node, generating wallets and so on

# docker-compose nodes from scratch

The docker compose setup enables up to 12 nodes to be run in compose.

Pick a common root folder and peform the following steps


1. checkout https://github.com/RobustRoundRobin/quorum.git to

        gopath/src/github.com/ethereum/go-ethereum

1. checkout https://github.com/RobustRoundRobin/RRR.git to

        rrr

1. build the images

      tusk -q -f rrr/tusk-build.yml images

1. Generate the node keys, genesis and folders

        tusk -q -f rrr/tusk-genesis.yml all

1. update the rrr/compose/.env

   Set RRR_BOOTNODE_PUB to $(cat nodes/node0/enode)
   Set RRR_ETHERBASE to $(cat nodes/gensis-wallet.addr)

1. start some nodes

        cd /compose
        docker-compose up node0 node1 node2
        cd -
