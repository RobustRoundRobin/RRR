# Written assuming that:
#
# our fork of ConsenSys is put in buildcontext/quorum
# go-rrr is put in buildcontext/go-rrr
#
# For local development, in our fork of ConsenSys/quorum, quorum/go.mod has a
# relative replace directive which require this layout.
#
# Our tusk.yml build step uses rsync to maintain a buildcontext in ${workdir}

FROM golang:1.15.5-alpine as builder
RUN apk add --no-cache make gcc musl-dev linux-headers git

RUN mkdir -p /build/quorum
WORKDIR /build/quorum

COPY quorum /build/quorum
COPY quorum/go.mod .
COPY quorum/go.sum .

# COPY go-rrr/consensus /build/go-rrr/consensus
COPY go-rrr/consensus/go.mod /build/go-rrr/consensus/
COPY go-rrr/consensus/go.sum /build/go-rrr/consensus/

RUN go mod download

ENV GOBIN=/go/bin

COPY go-rrr/consensus /build/go-rrr/consensus

RUN go install \
    github.com/ethereum/go-ethereum/common/hexutil \
    github.com/ethereum/go-ethereum/common \
    github.com/ethereum/go-ethereum/common/mclock \
    github.com/ethereum/go-ethereum/common/prque \
    github.com/ethereum/go-ethereum/common/math \
    github.com/ethereum/go-ethereum/crypto/secp256k1 \
    github.com/ethereum/go-ethereum/rlp \
    github.com/ethereum/go-ethereum/crypto \
    github.com/ethereum/go-ethereum/log \
    github.com/ethereum/go-ethereum/params \
    github.com/ethereum/go-ethereum/private/engine \
    github.com/ethereum/go-ethereum/core/types \
    github.com/ethereum/go-ethereum/ethdb \
    github.com/ethereum/go-ethereum/metrics \
    github.com/ethereum/go-ethereum/ethdb/leveldb \
    github.com/ethereum/go-ethereum/ethdb/memorydb \
    github.com/ethereum/go-ethereum/core/rawdb \
    github.com/ethereum/go-ethereum/trie \
    github.com/ethereum/go-ethereum/core/state/snapshot \
    github.com/ethereum/go-ethereum/core/state \
    github.com/ethereum/go-ethereum/common/bitutil \
    github.com/ethereum/go-ethereum/crypto/ecies \
    github.com/ethereum/go-ethereum/event \
    github.com/ethereum/go-ethereum/permission/core \
    github.com/ethereum/go-ethereum/internal/plugin \
    github.com/ethereum/go-ethereum/plugin/security

RUN go install \
    github.com/ethereum/go-ethereum/p2p/enr \
    github.com/ethereum/go-ethereum/p2p/netutil \
    github.com/ethereum/go-ethereum/p2p/enode \
    github.com/ethereum/go-ethereum/p2p/discover/v4wire \
    github.com/ethereum/go-ethereum/p2p/discover \
    github.com/ethereum/go-ethereum/p2p/discv5 \
    github.com/ethereum/go-ethereum/p2p/nat \
    github.com/ethereum/go-ethereum/p2p \
    github.com/ethereum/go-ethereum/rpc

RUN go install github.com/RobustRoundRobin/go-rrr/consensus/rrr
RUN go install github.com/ethereum/go-ethereum/consensus
RUN go install github.com/ethereum/go-ethereum/cmd/geth

FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    python3 \
    py-pip \
    py-requests && \
    pip install dnspython

COPY --from=builder /go/bin/geth /usr/local/bin/

ENTRYPOINT ["geth"]

