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
* https://github.com/RobustRoundRobin/devcluter.git

Pick any ROOT directory.

1. The quorum fork must be cloned to ROOT/qorum-rororo-gopath/src/github.com/ethereum/go-ethereum
2. devclutter must be cloned directly under ROOT. Call it ROOT/rororo-devclutter if you want the vscode support

If Visual Studio Code suites your needs, then create a symlink to the supplied
vscode config, (or derive your own.)

   ROOT/quorum-rororo-gopath/.vscode -> ROOT/rororo-devclutter

Having done all of that open ROOT/quorum-rororo-gopath as a "folder" in vscode.

Note: You MUST NOT set GO111MODULE=on in your environment, as go-ethereum is
not go.mod compatible.

## tusk.yml

Uses [go-tusk](https://rliebz.github.io/tusk/) to provide a collection of runes
considered useful for developing rororo. Try `tusk -q -f ./tusk.yml -h`

* ./tusk-genesis.yaml commands for initialising nodes for use with docker
    compose
* ./tusk.yml a small number of generaly helpful commands - execuging js on a
    node, generating wallets and so on

# docker-compose nodes from scratch

The docker compose setup enables up to 12 nodes to be run in compose

1. checkout https://github.com/RobustRoundRobin/quorum.git to

    ~/jitsuin/quorum-rororo-gopath/src/github.com/ethereum/go-ethereum

1. checkout https://github.com/RobustRoundRobin/devclutter.git to

    ~/jitsuin/rororo-devclutter

1. Generate a wallet for inclusion in the `alloc' in the genesis document

    tusk -q -f ./tusk.yml wallet

This creates

    rororo-nodes/node0/genesis-wallet.[key,pub,addr]

1. Generate the node keys and folders

    tusk -q -f ./tusk-genesis.yml keys

This creates

    node[0-11]/enode
    node[0-11]/key

1. before continuing ensure GO111MODULE is *not* set in your environment

1. Generate the node enrolments for the genesis extra data

    tusk -q -f ./tusk-genesis.yaml extra

1. Copy genesis.json to rororo-nodes/node0/gensis.json and copy the big hex
   string from the end of the output of the previous command in to the
   extraData value, replacing "<RORORO EXTRADATA to enrol validators". Be sure
   to prefix the string with '0x'

1. Copy the genesis wallet address from rororo-nodes/node0/genesis-wallet.addr
   into the "alloc", replacing "GENSIS-WALLET"

1. update the rororo-devclutter/compose/docker-compose.yml x-node-env-defaults
   to set --miner.etherbase to the genesis wallet address

1. Run genesis and initialise all the nodes (geth init)

   task -q -f ./tusk-genesis.yaml init-all

This prints the extraData needed to enrol all the nodes in the genesis block
(via the extraData field in the genesis.json). It also creates a
static-nodes.json with each of those nodes in which is appropriate for the
compose setup. This static-nodes.json is copied in to each node's director.

1. check that static-nodes.json has been copied into each of the nodes data
   directories by the previous step. For each of [n] nodes look for it here:

   rororo-nodes/node[n]/data/geth/static-nodes.json

1. build the base docker image for hosting the nodes. Only need to do this once
   as our compose file mounts the host source rather than building the code
   into images.

        cd rororo-devclutter/compose
        docker-compose build debug

1. start some nodes

    cd rororo-devcluter/compose
    docker-compose up node0 node1 node2

1. create a symlink from ~/jitsuin/quorum-rororo-gopath/.vscode to

    ~/jitsuin/rororo-devclutter/vscode

