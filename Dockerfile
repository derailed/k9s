# -----------------------------------------------------------------------------
# The base image for building the k9s binary

FROM golang:1.23-alpine3.20 AS build

WORKDIR /k9s
COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd
# Install dependencies, build, and clean up
RUN apk --no-cache add make libx11-dev git gcc libc-dev curl \
    && make build \
    && apk del gcc libc-dev git make libx11-dev curl 
# Remove build-time dependencies

# -----------------------------------------------------------------------------
# Build the final Docker image

FROM alpine:3.20.3
ARG KUBECTL_VERSION="v1.29.0"

# Combine apk installs, kubectl download, and cleanup into a single layer
COPY --from=build /k9s/execs/k9s /bin/k9s
RUN apk --no-cache add ca-certificates curl vim \
    && TARGET_ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) \
    && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/${TARGET_ARCH}/kubectl -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && rm -rf /var/cache/apk/* /tmp/* /usr/local/bin/kubectl.sha256

ENTRYPOINT [ "/bin/k9s" ]
