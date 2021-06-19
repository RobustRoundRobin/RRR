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

# Cache the go mod download
COPY quorum/go.mod .
COPY quorum/go.sum .
COPY quorum/crypto/secp256k1/go.mod crypto/secp256k1/
COPY go-rrr/consensus/go.mod /build/go-rrr/consensus/
COPY go-rrr/consensus/go.sum /build/go-rrr/consensus/

RUN go mod download

ENV GOBIN=/go/bin

COPY quorum /build/quorum
COPY go-rrr/consensus /build/go-rrr/consensus

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

