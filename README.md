# RRR Spec

## Strategy

The aim is to get RRR into quorum. However, following the path of the IBFT
EIP seems like a good fit even though we don't strictly need this in ethereum.

There are enough similarities in the protocol. At a coarse level it does
similar things: leader selection, block sealing, rounds, communication between
validators. It just does them according to a different protocol. This leads us
to expect that the mechanical implementation choices for IBFT's integration
should at least be a good starting point for RRR. Following the path laid down
by IBFT will make our efforts _familiar_ to upstream. And the pull request
feedback for IBFT should prove helpful in avoiding mistakes.

To that end, we will maintain this EIP as we go. Who knows, if we are successful
with the implementation, this document's commit history should make for an
interesting read for future contributors.

## Robust Round Robin (RRR) Consensus

RRR: A consensus algorithm that can be used in both permissioned and permissionless
settings. It provides fairness, liveness and high throughput using a simple
round robin leader selection complimented with a lightweight endorsement mechanism.
In return for accepting the use of stable long term validator
identities, this approach scales to 1000's of nodes.

See the [paper](https://arxiv.org/pdf/1804.07391.pdf) for details.

## Abtstract

Proof-of-Stake systems randomly choose, on each round, one of the
participants as a consensus leader that extends the chain with the
next block such that the selection probability is proportional to the
owned stake. However, distributed random number generation is
notoriously difficult. Systems that derive randomness from the previous
blocks are completely insecure; solutions that provide secure
random selection are inefficient due to their high communication
complexity; and approaches that balance security and performance
exhibit selection bias. When block creation is rewarded with new
stake, even a minor bias can have a severe cumulative effect.

Here, we implement Robust Round Robin, a new consensus
scheme that addresses this selection problem. We create reliable
long-term identities by bootstrapping from an existing infrastructure.
For leader selection we use a deter-ministic approach. On each round,
we select a set of the previously created identities as consensus leader
candidates in round robin manner. Because simple round-robin alone is
vulnerable to attacks and offers poor liveness, we complement this deterministic
selection policy with a lightweight endorsement mechanism that is
an interactive protocol between the leader candidates and a small
subset of other system participants. Our solution has low good
efficiency as it requires no expensive distributed randomness generation
and it provides block creation fairness which is crucial in
deployments that reward it with new stake.

## Motivation

The main motivation for implementing RRR comes from wanting to enable entreprise
DApps to scale to 1000's of nodes. Existing production consensus algorithms
such as Raft only scale to a few dozen nodes. This severely restricts
the set of applications that are amenable to DLTs and forces application
developers to incorporate unnecessary delegation. A highly scalable consensus
protocol like RRR promises to help DApps become massively distributed and
consequently remain as decentralised as possible.

## Roadmap

* [x] Implement the 'round' in robust round robin
* [x] Implement the 'robust', at least dealing with idle leaders - no blocks
      for n round, don't include in candidates
      when selecting leader for new round. And idealy random endorser selection
      with 'agreed' seed in block: at least, sort endorser by public key and
      idx=mod something
* [x] Sort out the name change from RoRoRo to RRR
* [ ] Internal review of implementation & crypto, EIP -> RRR-spec, general tidyup
* [x] Open the repository
* [x] Put in intent phase timer, so that we can properly select the "oldest seen
    candidate" for confirmation, rather than "first seen"
* [x] VRF & seeding for the candidate and endorser selection
* [x] Membership changes - convenience here is essential for addoption over raft
* [x] Empirical testing to decide upon sync on block arrival vs synchronised clocks

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

### Close nodes

The original work found the majority of block dissemination delay to be due to
the initial block transmission by the leader candidate. Due to high out degree
and number of hops. In this implementation we rely on geth to do this.

### Multiple identity queues

This is a medium-term goal but not currently a high priority ticket item.

### Initial enrolment

In 4.1 the paper describes how to securely enrol the initial identities for
participants and agree an initial seed. Our implementation makes provision for
this in extraData of the genesis block but the attestation Quotes (Qi's)

   Block0 = (pk1, Q1, pk2, Q2,..., he, seed0, π0, id)

The pk's are the node public keys for the initial members

The paper specifies that a Verifiable Randome Function and associated Proof
be used. In the initial implementation the block minter generates the seed from
a source suitable for cryptographic operations. The seed proof is currently set to 0.
This is insecure for BFT networks but okay for CFT ones.


### Enrolment

In 4.1 the paper outlines the enrolment process. The candidate requests
enrolment after installing code in its secure enclave, generating and sealing a
key and providing the public key in the enrolment to *any* member. That member
then verifies the identity and, assuming its ok, broadcasts an enrolment
message to the network

  Enrol = (Qn, pkn,r,hb)

In our implementation, the candidate provides their public node key as pkn, and
the latest block hash as hb. And sets Qn to 0.

An out of band mechanism is used to supply acceptable pkn's to current members.
XXX: lets at least respect --permissioned and the nodes identified by
permissioned.json

### Re-enrolment

[Will be ]automatically performed by the leader for now.
### Consensus Rounds

From 3.3 Starting Point: Deterministic Slection

> Such a system is able to produce a new block on each round with high
> probability and proceed even from the rare scenarios where all selected
> candidates are unavailable to communicate, 

We run a ticker on all nodes which ticks according to the configured intent
and confirm phase times. Endorsed leader candiates will only broadcast their
block at the end of their configured round time (intent + confirm). When a
node sees a new confirmed block it resets its ticker. In this way all
participants should align on the same time window for a round. Yet if there
is outage of any kind, each node will independently seek to initiate a new
round.

For now the round timeout is configured on the command line. Later we can put
this in the genesis block and provide a mechanism for changing it post chain
creation

To achieve liveness we require a means for identities to 'skip' a round in
the event that no candidate produces a block within the alloted time
(intent+confirm phase). When this occurs the successful proposer can not, by
definition, be one of the oldest Nc identities for that particular round.

To allow for this we define a round as being however long it takes to produce
a block. round number == block number, and allow each node to independently
declare the round attempt it considers the network to be on when broadcasting
its intent. If the endorsers agree it is a leader candidate for the round +
the attempt adjustment, and if it is the *oldest* identity those endorsers
see, then they will endorse the intent. Individual honest nodes will reset
their attempt counts and intent/confirm timers on receipt of a new block.

Note that a malicious node can not produce a block simply by faking an
attempt number and contacting any endorser it choses. A node must include its
failedAttempt counter in the block it creates. The network will not accept
blocks where the endorsments included were not endorsers for that round +
failedAttempt.

The validity of a block proposer as a candiate in any round is dependent on
two things:

1. The age order of identities selected as active for Ta worth of blocks.
2. The number of attempts made by the by the *successful proposer*, to
   produce a block.
   
The current round is defined as the block number. Age is, as before, the
block the identity last minted or was enroled on.

1. Does not change unless a new block is accepted by the network.
2. Is determined by the *oldest* identity to be endorsed by a quorum of
   active identities.

In a healthy network, nodes will tend to be "losely synchronised" - both on
phase and on a failecount of 0.

In situations where a network is starting up from scratch or in small
networks that are temporaily disrupted, the phase and attempt counters on the
nodes may be different by aribtrary amounts. In such situations we want
liveness. And we want to avoid the rules for marking identities idle from
making so many identities idle that we can't advance the chain to re-enrol
them.

Each endorser, as before, will endorse the intent from the oldest identity it
sees within its *own* intent phase. The endorser allows the proposer to
declare itself a legitemate candidate due to failed attempts. It specifically
does not attempt to work out if it itself is an endorser for the attempt
indicated by the proposer. This would require it to play forward or back the
endorser selections to get the selection implied by its peer and it gains us
nothing as the peer can say what it likes for failcount.

Attempt camping, where a nodes simply declares itself a candidate in each
round by setting its intent failedcount appropriately, can only be successful
where all older honest nodes are not able to publish intents. (And we could
probably penalise this behaviour)

No attempt to co-ordinate the phase with other nodes is made.

All participants run the cycle of intent -> confirm phases described in the
paper. They do so independently. Each full cycle counts as an attempt. When a
node completes a cycle without seeing a new (verifiable) chain head it starts
a new attempt. The endorsers are resampled for every failed attempt. After Nc
failed attempts the node advances passed the current Nc oldest identities.
(Advancing by one favours the youngest of the oldest)

Leader candidates who are 'oldest' for Nc *rounds*, as specified by the
paper, *are* made idle - but this only occurs when blocks arrive. We do not
make identities idle at the end of attempts

On any given node, when there are Nc failed attempt cycles we advance the
candidate selection by Nc - but do NOT make them idle. (Nc is chosen to align
with the idles rule for rounds). So for a particular node the range of leader
candidates, is `[ floor(a/Nc).Nc - ceil(a/NC).Nc )` where `a = failedAttempts
% max(na, nc+ne)` and `na` is the number of active identities, `nc` the
leader candidates and `ne` the endorsers. Intuitively:

    [N0, N1, ... Nc, Nc+1, ..., 2.Nc, ....]

The sampling of endorsers on each node is made *around* that window. The
sampling is updated for every attempt.

## Implementation

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
queue. And worker threads pick them up.

RRR uses the same message processing model, but only has a single worker go
routine.

### Mining (update ticker)

IBFT has an explicit hook in eth/miner/worker.go which invokes an Istanbul
specific Start method on the consensus engine. RRR uses the same hook but
expects an RRR specific interface for the chain to be passed (and checks it
with a type assertion)

### Mining vanity (extraData)

RRR uses the extraData on blocks to convey its consensus parameters.

### Identity establishment

Initialisation and subsequent enrolment of identities.  Here we describe "Attested"
identities, but "attested" by the geth node private keys directly.

From 4.1 (in the paper) we have

1.
        Block0=(Q1,pk1,Q2,pk2,...,he,seed0,π0,id)
2.
        Enrolln=(Qn,pkn,r,hb)

The Q's are the attestations

For 1. We create all the Qn "attestations" for Block0 using the chain creators
private key to sign all the initial member public keys (including the chain
creators).

For 2. The contacted member signs the hash in the Q with its own private key.

### block header content

Appendix A describes the "SelectActive" algorithm. This describes a key
activity for the consensusm algorithm. Selecting active nodes is the basis of
both leader candidate and endorser selection. It also features in branch
verification. Everything needed for that algorithm is encoded in the extraData

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
    ii. Enroln=(Qn,pkn,r,hb, f)


And

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

    Blockr=(Intent,{Confirm},{tx},{Enrol},seedr,πr,siдc)

    -- 5.2 endorsment protocol

In the top -> genesis traversal of the branch identities are marked active or
not based on finding {Confirm} messages which are *sent* by them.

The 'age' of an identity is the number of rounds since it last created a block
or since its enrolment which ever is later.
The Intent portion identifies the creator of the block and the round the Intent
belongs too.

#### Intent

Intent(id, pkc, r, f, hp, htx, sc)

    id     : CHAIN_ID
    ni     : pkc above (and in paper). Candidate's node id
    r      : round (on candidate node when message sent)
    f      : failed attempts count - as counted by the proposer, 0 <= f
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

#### Overview 

a list of identities at covering at least HEAD - Ta (or to genesis) is
maintained at all times with entries sorted back -> front oldest -> youngest

#### accumulateActive

Record activity of all blocks until we reach the genesis block or until we reach
a block we have recorded already. We are traversing 'youngest' block to
'oldest'. We use youngestKnown (at the start of the traversal) as our marker and
add each identity we see just infront (younger) than it. Since each subsequent
identity is 'older' than the previous, this preserves the overall front -> back,
youngest -> oldest ordering in activeSelection.

For example: Given a genesis block with the enrolments `[ia, ib, ic]` and where
where `ib` minted `block 2`, and `ia` minted `block 1`, We start with `head ==
block 2`. The active selection is primed with the genesis identies in the
order of their enrolment in the genesis block.

    [ic, ib, ia]

We note the youngest known at start (Syk) at the start of the pass as the end of the
currently empty set:

        [ic, ib, ia]
    Syk -^

We see that `ib` minted `block 2` and place its identity at the position imediately younger than `Syk`:

        [ib, ic, ia]
    yk ---^
    Syk -----^

We advance to `block 1` and see that `ia` minted `block 1` and place its identity at the position imediately younger than `Syk`:

        [ib, ia, ic]
    yk --^
    Syk ---------^

Finally we reach the genesis block which was also minted by ia. We need to
notice that we have found a younger position for `ia` and not move it again:

        [ib, ia, ic]
    yk --^
    Syk ---------^

Now, the size of the active selection is fixed regardless of how many identities
are enroled and have been seen. This means the 'front' may be empty. So at the
end of the accumulation we take note of the youngest known. This is the first
occupied slot in the fixed storage allocation.

Now consider the more general case where each `block Bn`, introduces enrolments
`En [age, age, ..., age]` - so the sealer and the enrolments must be inserted
into the ordering:

    E0 = [0, 1, 2]
    E1 = [3, 4, 5]
    E2 = [6, 7, 8]

accumulateActive sees the blocks in the order `E2, E1, E0`. The sealer is listed
first followed by the identities it enrols.  enrolments are listed oldest ->
youngest (relative to each other). We need to pick a relative age order for
identities enroled in a block. As the active selection is youngest -> oldest, to
maintain the set in age order we need to reverse them as we add them.

Eg for `E2 6, 7, 8`, the insertion order is `[8, 7, 6]`. We can trivially ensure
the list is maintained in age order by enumerating the enrolments in reverse and
always inserting just before the youngest known at start (Syk) position.:

             [] Syk
    E2 =>    [8] Syk
             [8, 7] Syk
             [8, 7, 6] Syk
    E1 =>    [8, 7, 6, 5] Syk
             [8, 7, 6, 5, 4] Syk
             [8, 7, 6, 5, 4, 3] Syk
    E0 =>    [8, 7, 6, 5, 4, 3, 2] Syk
             [8, 7, 6, 5, 4, 3, 2, 1] Syk
             [8, 7, 6, 5, 4, 3, 2, 1, 0] Syk

   Youngest --^ (most recently active)

Finally, accumulateActive remembers the last block it saw. By assuming it always
sees blocs in consecutive descending order it can define the 'oldest' position
as the position *after* the last block it processed in the previous round. So we
end up with:

    E3 = [9, 10, 11]
    E3 => [8, 7, 6, 5, 4, 3, 2, 1, 0 ]
    Syk ---^
          [11, 8, 7, 6, 5, 4, 3, 2, 1, 0 ]
    Syk -------^
          [11, 10, 8, 7, 6, 5, 4, 3, 2, 1, 0 ]
    Syk -----------^
          [11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0 ]
    Syk --------------^

Syk is effectively a fence. And is always initialised to the youngest known at
the end of the previous round, which will be, by the above assumption, be older
than all the identities it is about to enrol. Where this assumption fails - a
condition we can cheaply detect - we can recover by completely (or partially)
re-processing HEAD - Ta worth of blocks.

#### refreshAge

refreshAge called to indicate that nodeID has minted a block or been
enrolled. Counter intuitively, we always insert the at the 'oldest' position.
Because accumulateActive works from the head (youngest) towards genesis. If
no fence is provided the entry is added at the back (oldest position). If a
fence is provided, the entry is added immediately after the fence - which is
the oldest position *after* the fence. accumulateActive uses the last block
it saw as the fence. enrolIdentities processes enrolments for a block in
reverse order of age. These two things combined give us an efficient way to
always have identities sorted in age order.

## Geth - relevant implementation nodes (this is going to get removed or appendixed)

In geth, when mining, the confirmation of a NewChainHead imediately
results a new 'work' request being issued. In rrr (as in IBFT) this is where
we hook in end of round/start of new round

### Geth IBFT 'new chain' / warmup

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
