# RRR Spec

## Strategy

The aim is to get RRR into quorum. However, following the path of the IBFT
EIP seems like a good fit even though we don't strictly need this in ethereum.

There are enough similarities in the protocol. At a course level it does
similar things leader selection, block sealing, rounds, communication between
validators. It just does them according to a different protocol. This leads us
to expect that the mechanical implementation choices for IBFT's integration
should at least be a good starting point for RRR. Following the trail blazed
by IBFT will make our efforts _familiar_ to upstream. And the pull request
feedback for IBFT should prove helpful in avoiding mistakes.

To that end, we can maintain this EIP as we go. Who knows, if we are successful
with the implementation, we could very well propose it.

## Robust Round Robin Consensus

A consensus algorithm adding fairness, liveness to simple round robin leader
selection. In return for accepting the use of stable long term validator
identities, this approach scales to 1000's of nodes.

## Abtstract

See the [paper](https://arxiv.org/pdf/1804.07391.pdf)

# Motivation

Enterprise need for large scale consortia style networks with high throughput.

# Roadmap

* [ ] Implement the 'round' in robust round robin
* [ ] Implement the 'robust', at least dealing with idle leaders - no blocks
      for n round, don't include in candidates
      when selecting leader for new round. And idealy random endorser selection
      with 'agreed' seed in block: at least, sort endorser by public key and
      idx=mod something
* [ ] Sort out the name change from RRR to RRR
* [ ] Internal review of implementation & crypto, EIP -> RRR-spec, general tidyup
* [ ] Open the repository
* [ ] Put in intent phase timer, so that we can properly select the "oldest seen
    candidate" for confirmation, rather than "first seen"
* [ ] VRF & seeding for the candidate and endorser selection
* [ ] Membership changes - convenience here is essential for addoption over raft

## Specification

Be consistent with the IBFT implementation choices as far as makes sense. To
keep things familiar for upstream. And to ensure we don't make mistakes that
they have avoided.

https://github.com/ethereum/EIPs/issues/650 (IBFT's EIP for comparison)

### Long term identities

In our initial effort, we will not implement either of the papers proposed
methods for the attestation of long term identities. We will rely on public
node keys and out of band operational measures. This makes the implementation
only suitable for private networks. However, we do seek to make room for the
attestation to be added in future work.

Here, where we say node id we mean  Keccak256 ( PublicKey X || Y ). Where
PublicKey is the nodes ecdsa public key.

### Configuration

#### Nc, Ne - Option 1

Nc and Ne, respectively the number of leader candidates and endorsers, in each
round will initially be established in the genesis parameters. A mechanism for
re-configuring those post genesis can follow the model established by quorum
for changing things like the maximum transaction size and maximum contract code
size from a particular block number.

#### Nc, Ne - Option 2

Nc and Ne are geth command line parameters and all participants must agree.
This will get us going, and likely be more convenient for early development.

#### tr = 5 seconds

But will eventually make this either configurable or derived in some way from
Nc, Ne

### Close nodes

The original work found the majority of block dissemination delay to be due to
the initial block transmission by the leader candidate. Due to high out degree
and number of hops. Having a way to configure 'network close' nodes to serve
has broadcast hubs for the initial block transmission is something we would
like in from the start. The paper makes clear this does not undermine the
security properties.

### Multiple identity queues

Ah, maybe not at first but definitely want a path to trying this.

### Initialisation

In 4.1 the paper describes how to securely enrol the initial identities for
participants and agree an initial seed. Our implementation makes provision for
this in extraData but the attestation Quotes (Qi's), and the seed proof π0 are
zeroed.

   Block0 = (pk1, Q1, pk2, Q2,..., he, seed0, π0, id)

The pk's are the node public keys for the initial members

1. ? Can VerifyBranch work at all without a seed proof
2. ? What happens if there are < Ne available endorsers. Especially during
   network establishment where, operationaly, it makes sense to have a small
   population in the genesis block and have the rest join in an early 'rush'

### Enrolment

In 4.1 the paper outlines the enrolment process. The candidate requests
enrolment after installing code in its secure enclave, generating and sealing a
key and providing the public key in the enrolment to *any* member. That member
then verifies the identity and, assuming its ok, broadcasts an enrolment
message to the network

  Enrol = (Qn, pkn,r,hb)

In our implementation, the candidate provides their public node key as pkn, and
the latest block hash as hb. And sets Qn to 0.

An out of band mechanism is used to supply acceptable pkn's to current members

### Re-enrolment

Automatically performed by the leader for now.

### IBFT (and other) issues we want to be careful of

### eth_getTransactionCount is relied on by application level tx singers

See [issue-comment](https://github.com/ethereum/EIPs/issues/650#issuecomment-360085474)

It's not clear to me that this can ever be reliable with concurrent application
signers, but certainly we should not change the behaviour of api's like
eth_getTransactionCount

## Implementation

### Consensus Rounds

From 3.3 Starting Point: Deterministic Slection

> Such a system is able to produce a new block on each round with high
> probability and proceed even from the rare scenarios where all selected
> candidates are unavailable to communicate, 

We run a ticker on all nodes which ticks according to the configured round
time. Endorsed leader candiates will only broadcast their block at the end of
their configured round time. When a node sees a new confirmed block it resets
its ticker. In this way all participants should align on the same time window
for a round. Yet if there is outage of any kind, each node will independently
seek to initiate a new round.

Note that in geth, when mining, the confirmation of a NewChainHead imediately
results a new 'work' request being issued. In rrr (as in IBFT) this is where
we hook in end of round

For now the round timeout is configured on the command line. Later we can
seek to put this in the genesis block and provide a mechanism for changing
it post chain creation

### Node p2p

We leverage the mechanism provided in the quorum go-ethereum for IBFT. For each
inbound message on a peer connection, the consensus engine is given a "first
refusal" on the message. If the engine returns an error, just as for normal
message handling, the connection is terminated. If the engine indicates it has
consumed the message, normal 'eth' protcol msg handling is skiped. The
following source references for the quorum 2.7.0 release indicate where these
arrangements are made.

IBFT allocates a single message code (0x11) by which it recognises its own
protocol messages. All IBFT consensum and admin messages are sub encoded in the
data.

* eth/handler.go ProtocolManager handle - does the connection life cycle for a
  single peer
* eth/handler.go ProtocolManager handleMsg - is invoked for each in bound
  messagage from a particular peer. And here, the consensus engine gets its
  "first refusal".
* istanbul/backend/handler.go HandleMsg is the IBFT "first refusal"
  implementation and this consumes all messages that have the msg.Code

The IBFT consensus implementation then posts messages it claims to an internal
queue. And worker threads pick them up. Depending on the actual message, and
the current state of the consensus protocol, those messages may be "gossiped"
on to other nodes.

RRR will, at least initialy, use the same message processing model.  This
means leaders and endorsers will be only losely connected. We rely on the
gossip protocol to diseminate our confirmations.

### Mining (update ticker)

IBFT has an explicit hook in eth/miner/worker.go which invokes an Istanbul
specific Start method on the consensus engine. RRR adds its own hook.

### Mining vanity (extraData)

The default extra data included in each block is set by the command line config
option

    --miner.extradata

Which eventually reaches

    go-ethereum/worker/worker.go setExtra

We will be adding to that data to carry the block material for RRR
### Identity establishment

Initialisation and subsequent enrolment of identities.  Here we describe "Attested"
identities, but "attested" by the geth node private keys directly.

From 4.1 (in the paper) we have

1.
        Block0=(Q1,pk1,Q2,pk2,...,he,seed0,π0,id)
2.
        Enrolln=(Qn,pkn,r,hb)

The Q's are the attestations

For i. We create all the Qn "attestations" for Block0 using the chain creators
private key to sign all the initial member public keys (including the chain
creators).

For ii. The contacted member uses its private key to attest the new member.

### block header content

Appendix A describes the "SelectActive" algorithm. This describes a key
activity for the consensusm algorithm. Selecting active nodes is the basis of
both leader candidate and endorser selection. It also features in branch
verification.

So structuring the data in the blocks (probably the block headers), so that
this can be done efficiently and robustly is correspondingly a key
implementation concern


#### extraData of Block0

1. Ck, Cp, CI  : Chain creator private and public keys, and public key derived identity
2. Mk, Mp, MI  : Chain member private and public keys, and pub key derived identity
3. he          : 0
4. seed0       : crypto/rand for now, mansoor looking at RandHound based seeding
5. π0          : Sig(Ck, seed0) for now, mansoor looking at VFR stuff and seed establishment
6. Q0          : Sig(Ck, CI)
7. Qi          : Sig(Ck, MIi) for all initial members
8. IDENT_INIT  : RLP([[Q0, Cp], [Q0, Cp], [Qi, Mpi]])
9. CHAIN_INIT  : RLP([IDENT_INIT, he, seed0, π0])
10.CHAIN_ID    : Keccak256(CHAIN_INIT)
11.extraData   : RLP([CHAIN_INIT, CHAIN_ID])

Note: the identities *are* keccak 256 hashes of the respective nodes public
key, hence we don't apply keccak to them before signing.

Note: for the purposes of 'age' tie breaks, for the first block, the order of
the identities is their 'order of enrolment'

#### Enrolment data

From 4.1 we have
i.  hash (id || pkn || r || hb )
ii. Enrolln=(Qn,pkn,r,hb, f)

1. Dk, Dp, Di  : Debutant (or re-enroling) private and public keys, and public key derived identity
           Di  : Keccak256(Dp[1:65])
2. Ak, Ap      : Active chain member private and public keys
3. r           : round number
4. f           : re-enrolment flag
5. hb          : hash of latest block
6. U           : Keccak256(CHAIN_ID || Di || r || hb )  created on the debutant
7. Usig        : Sig(Dk, U)
7. Qn          : Sig(Ak, U)
8. EMsg        : RLP([[Qn, Di], r, hb, f])


Debutant request enrolment:

    D -> request enrol -> A
                          A -> If ok, NETORK enrol

Debutant sends RLP([U, Usig]) to A (A already has Dp from connection handshake)

    D -> request enrol -> A
                          A -> verifies, then broadcasts Emsg -> NETWORK

LEADER -> includes EMsg on block
D is now established

#### extraData for Blockn

Needs to include

    Blockr=(Intent,{Confirm},{tx},{Enroll},seedr,πr,siдc)

    -- 5.2 endorsment protocol

In the top -> genesis traversal of the branch identities are marked active or
not based on finding {Confirm} messages which are *sent* by them.

The 'age' of an identity is the number of rounds since it last created a block
or since its enrolment which ever is later.
The Intent portion identifies the creator of the block and the round the Intent
belongs too.

#### Intent

Intent(id, pkc, r, hp, htx, sc)

id     : CHAIN_ID
ni     : pkc above (and in paper). Candidate's node id
r      : round (on candidate node when message sent)
hp     : previous block hash, parent of the intended new block
htx    : merkle root of the transactions to be put in new block
sc     : candidates signature over id || ni || r || hp || htx

#### Confirm

Confirm (id, hi, h(pkv), sige)

id     : CHAIN_ID (defined for extraData of Block0)
hi     : Keccak256(Intent)
ei     : h(pkv) above (and in paper). Endorser's node id.
se     : endorsers signature over id || hi || ei

### Deterimistic selection

SelectActive, SelectCandidates, SelectEndorsers

#### Key details

* a list of identities at covering at least HEAD - Ta (or to genesis) is
  maintained at all times with entries sorted back -> front oldest -> youngest

##### accumulateActive

Record activity of all blocks until we reach the genesis block or until we
reach a block we have recorded already. We are traversing 'youngest' block to
'oldest'. We use youngestKnown (at the start of the traversal) as our marker
and add each identity we see after it. This preserves the overall oldest ->
youngest ordering in activeSelection.

For example: Given a genesis block with `[ia, ib, ic]` and `head == block 2`,
where no new identities are enrolled in `block 2` or `block 1`, and where
`ib` minted `block 2`, and `ia` minted `block 1`. We start at block 2 and
traverse backwards:

    A=   [ia, ib, ic] - from genesis
    yk -----------^
    A=   [ia, ic, ib] - ib is moved 'after' ic, becoming the youngest identity
    yk -------^
    A=   [ic, ia, ib] - ia is moved 'after' ic, becoming the *second* youngest identity
    yk ---^

Now consider the more general case where each `block Bn`, introduces
enrolments `En [age, age, ..., age]`:

    E0 = [0, 1, 2]
    E1 = [3, 4, 5]
    E2 = [6, 7, 8]

accumulateActive sees the blocks in the order `E2, E1, E0`. By visiting each
enrolment in each block in reverse, eg for `E2 8, 7, 6`, we can trivially
ensure the list is maintained in age order by always inserting at the
'oldest' position:

          []
    E2 => [8]
          [7, 8]
          [6, 7, 8]
    E1 => [5, 6, 7, 8]
          [4, 5, 6, 7, 8]
          [3, 4, 5, 6, 7, 8]
    E0 => [0, 1, 2, 3, 4, 5, 6, 7, 8]


Finally, accumulateActive remembers the last block it saw. By assuming it
always sees blocs in consecutive ascending order it can define the 'oldest'
possition as the position *after* the last block it processed in the previous
round. So we end up with:

    E0 => [ ..., fence, 0, 1, 2, 3, 4, 5, 6, 7, 8]

And fence is set to the youngest known at the end of the previous round,
which will be, by the above assumption, be older than all the identities it
is about to enrol. Where this assumption fails - a condition we can cheaply
detect - we can recover by completely (or partially) re-processing HEAD - Ta
worth of blocks.

##### refreshAge

refreshAge called to indicate that nodeID has minted a block or been
enrolled. Counter intuitively, we always insert the at the 'oldest' position.
Because accumulateActive works from the head (youngest) towards genesis. If
no fence is provided the entry is added at the back (oldest position). If a
fence is provided, the entry is added immediately after the fence - which is
the oldest position *after* the fence. accumulateActive uses the last block
it saw as the fence. enrolIdentities processes enrolments for a block in
reverse order of age. These two things combined give us an efficient way to
always have identities sorted in age order.

## Security Considerations

# Geth - relevant implementation nodes (this is going to get removed or appendixed)

## static nodes

Initial nodes need to be included in the genesis block extraData. We provide a
utility subcommand for geth which creates the appropriate extraData for the
genesis.json config file. After genesis, a new node is introduced by making a
request to an existing node and then waiting for that request to be confirmed
by the consensus engine.

If boot nodes are being used instead, the initial boot nodes will need to be
similarly in the genesis block. The easiest way to do that is to list them in a
static-nodes.json and generate the extraData for genesis as above. The
--bootnodes options at startup can then be used as normal. In this scenario the
static-nodes.json is just a way to configure the genesis extraData and can be
discared after that is done.

On startup geths p2p engine populates its view of 'localnodes' to include those
declared in static
## Geth IBFT 'new chain' / warmup

For orientation with the IBFT implementation it helps to follow what happens
when starting up on a new chain.

In summary, startup and catch up, verifying any blocks as we find them once
at 'head', start mining

1. VerifyHeader for the genesis block (block 0)

    Called via geth/main/startNode

        go-ethereum/node/node.go .Start
        go-ethereum/core.NewBlockChain
        consensus -> VerifyHeader 

2. Prepare for the first block (block 1)

    Called via miner
        miner/worker.go newWorker
        go mainLoop()
        miner/miner.go mainLoop -> commitNewWork

    This block arrives with the --miner.extradata set on the geth command line.
    We can update the block.extraData inline at this point

3. FinalizeAndAssemble post-transaction state mods (block rewards & other
   consensus rules)

    Called via miner
        miner/worker.go newWorker
        go mainLoop()

        miner/miner.go mainLoop -> commitNewWork

    Last chance ? to update block.extraData

4. engine.Start

    This is the indication we are ready to start mining (running consensus),
    there is a corresponding Stop. This is the point at which IBFT starts its
    first consensus round

    The consensu engine Start hook is called Called via miner Start

        miner/miner.go New -> go miner.update()

        miner/miner.go Miner.Start
        consensus engine Start


    The CurrentHeader at this point is the genesis block


The Istanbul specific Start method takes a ChainReader. And further requires a
callback which is used to get the CurrentBlock. It's not clear why it does this
in preference to using 'CurrentHash' followed by "GetBlock".