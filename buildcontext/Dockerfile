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
ADD quorum /build/quorum
ADD go-rrr /build/go-rrr

ENV GOBIN=/go/bin

WORKDIR /build/quorum
RUN go install github.com/ethereum/go-ethereum/cmd/geth

WORKDIR /build/go-rrr/tools
RUN go install github.com/RobustRoundRobin/go-rrr/tools/cmd/rrrctl

FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    python3 \
    py-pip \
    py-requests && \
    pip install dnspython

COPY --from=builder /go/bin/geth /usr/local/bin/
COPY --from=builder /go/bin/rrrctl /usr/local/bin/

ENTRYPOINT ["geth"]

