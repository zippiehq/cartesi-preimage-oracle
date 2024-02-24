FROM --platform=$BUILDPLATFORM golang:1.21.3-alpine3.18 as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash

# We copy the go.mod/sum first, so the `go mod download` does not have to re-run if dependencies do not change.
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app

RUN echo "go mod cache: $(go env GOMODCACHE)"
RUN echo "go build cache: $(go env GOCACHE)"

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go mod download

# NOTE: the Dockerfile.dockerignore file effectively describes all dependencies
COPY . /app

# We avoid copying the full .git dir into the build for just some metadata.
# Instead, specify:
# --build-arg GIT_COMMIT=$(git rev-parse HEAD)
# --build-arg GIT_DATE=$(git show -s --format='%ct')
ARG GIT_COMMIT
ARG GIT_DATE

ARG CANNON_VERSION=v0.0.0
ARG OP_PROGRAM_VERSION=v0.0.0

ARG OP_HEARTBEAT_VERSION=v0.0.0
ARG OP_WHEEL_VERSION=v0.0.0

ARG OP_NODE_VERSION=v0.0.0
ARG OP_CHALLENGER_VERSION=v0.0.0
ARG OP_DISPUTE_MON_VERSION=v0.0.0
ARG OP_BATCHER_VERSION=v0.0.0
ARG OP_PROPOSER_VERSION=v0.0.0
ARG OP_CONDUCTOR_VERSION=v0.0.0


# separate docker-builds:
# - op-ufm
# - proxyd
# - any JS/TS/smart-contract builds

ARG TARGETOS TARGETARCH

# Build the Go services, utilizing caches and share the many common packages.
# The "id" defaults to the value of "target", the cache will thus be reused during this build.
# "sharing" defaults to "shared", the cache will thus be available to other concurrent docker builds.

# note: we only build the host, that's all the user needs. No Go MIPS cross-build in docker
RUN --mount=type=cache,target=/root/.cache/go-build cd op-program && make op-program-host  \
    GOOS=$TARGETOS GOARCH=$TARGETARCH GITCOMMIT=$GIT_COMMIT GITDATE=$GIT_DATE VERSION="$OP_PROGRAM_VERSION"

FROM alpine:3.18

COPY --from=builder /app/op-program/bin/op-program /usr/local/bin/
