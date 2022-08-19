# This dockerfile is used to build Mattermost ops-tool will all plugins.
# See makefile to configure plugins
# A multi stage build, with golang used as a builder
# and dockerhub.io/debian:bullseye-20220801 as runner
ARG GO_IMAGE=golang:1.18@sha256:90c06f42c1aa2b6b96441c0e6192aff48815cf5e7950cd661ed316fdbfb06ed4
# hadolint ignore=DL3006
FROM ${GO_IMAGE} as builder

#GO_BUILD_PLATFORMS holds the platforms that we will build the docker image against
ARG GO_BUILD_PLATFORMS=linux-amd64

# Setup directories structure and compile
COPY . /src
WORKDIR /src
RUN make go-build

# Cause of bash plugins we need to execute bash and shell commands.
FROM docker.io/debian:bullseye-20220801@sha256:82bab30ed448b8e2509aabe21f40f0607d905b7fd0dec72802627a20274eba55 as runner

# Istall dependencies and shell scripts requirements
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        curl=7.74.0-1.3+deb11u2 \
        jq=1.6-2.1 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/dist/ops_tool-linux-amd64 /opt/ops-tool/bin/ops-tool
COPY --from=builder /src/dist/plugins /opt/ops-tool/plugins
COPY scripts /opt/ops-tool/scripts

# We should refrain from running as privileged user
# Run as UID for nobody
USER 65534

WORKDIR /opt/ops-tool
ENTRYPOINT ["./bin/ops-tool"]
