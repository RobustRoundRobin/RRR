---
eip: <to be assigned>
title: Robust Round Robin Consensus
author: Robin Bryce <robinbryce@gmail.com>, Mansoor Ahmed-Rengers <mansoor.ahmed@cl.cam.ac.uk>, Kari Kostianinen <kari.kostianinen@inf.ethz.ch>
discussions-to: <URL>
status: NOT READY FOR PEER REVIEW
type: <Standards Track, Meta, or Informational>
category (*only required for Standards Track): <Core, Networking, Interface, or ERC>
created: <date created on, in ISO 8601 (yyyy-mm-dd) format>
requires (*optional): <EIP number(s)>
replaces (*optional): <EIP number(s)>
---

# Strategy

The aim is to get RoRoRo into quorum. However, following the path of the IBFT
EIP seems like a good fit even though we don't strictly need this in ethereum.

There are enough similarities in the protocol. At a course level it does
similar things leader selection, block sealing, rounds, communication between
validators. It just does them according to a different protocol. This leads us
to expect that the mechanical implementation choices for IBFT's integration
should at least be a good starting point for RoRoRo. Following the trail blazed
by IBFT will make our efforts _familiar_ to upstream. And the pull request
feedback for IBFT should prove helpful in avoiding mistakes.

To that end, we can maintain this EIP as we go. Who knows, if we are successful
with the implementation, we could very well propose it.

# Robust Round Robin Consensus

A consensus algorithm adding fairness, liveness to simple round robin leader
selection. In return for accepting the use of stable long term validator
identities, this approach scales to 1000's of nodes.

# Abtstract

See the [paper](https://arxiv.org/pdf/1804.07391.pdf)

# Motivation

Enterprise need for large scale consortia style networks with high throughput.

# Specification

Be consistent with the IBFT implementation choices as far as makes sense. To
keep things familiar for upstream. And to ensure we don't make mistakes that
they have avoided.

https://github.com/ethereum/EIPs/issues/650 (IBFT's EIP for comparison)

## Long term identities

In our initial effort, we will not implement either of the papers proposed
methods for the attestation of long term identities. We will rely on public
node keys and out of band operational measures. This makes the implementation
only suitable for private networks. However, we do seek to make room for the
attestation to be added in future work.

## Configuration

### Nc, Ne - Option 1

Nc and Ne, respectively the number of leader candidates and endorsers, in each
round will initially be established in the genesis parameters. A mechanism for
re-configuring those post genesis can follow the model established by quorum
for changing things like the maximum transaction size and maximum contract code
size from a particular block number.

### Nc, Ne - Option 2

Nc and Ne are geth command line parameters and all participants must agree.
This will get us going, and likely be more convenient for early development.

### tr = 5 seconds

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

## Initialisation

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

## Enrolment

In 4.1 the paper outlines the enrolment process. The candidate requests
enrolment after installing code in its secure enclave, generating and sealing a
key and providing the public key in the enrolment to *any* member. That member
then verifies the identity and, assuming its ok, broadcasts an enrolment
message to the network

  Enrol = (Qn, pkn,r,hb)

In our implementation, the candidate provides their public node key as pkn, and
the latest block hash as hb. And sets Qn to 0.

An out of band mechanism is used to supply acceptable pkn's to current members

## Re-enrolment

Is as described in the paper

## IBFT (and other) issues we want to be careful of

### eth_getTransactionCount is relied on by application level tx singers

See [issue-comment](https://github.com/ethereum/EIPs/issues/650#issuecomment-360085474)

It's not clear to me that this can ever be reliable with concurrent application
signers, but certainly we should not change the behaviour of api's like
eth_getTransactionCount

# Implementation

# Security Considerations

