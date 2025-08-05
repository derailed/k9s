# -----------------------------------------------------------------------------
# Build Stage: Compile K9s
FROM --platform=$BUILDPLATFORM golang:1.24.5-alpine3.21 AS build

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV CGO_ENABLED=0
ENV GOFLAGS="-trimpath"

WORKDIR /k9s

COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd

RUN apk add --no-cache make git gcc libc-dev \
  && make build

# -----------------------------------------------------------------------------
# Final Image: Minimal runtime with kubectl
FROM --platform=$TARGETPLATFORM alpine:3.21.3

ARG KUBECTL_VERSION="v1.32.2"
LABEL org.opencontainers.image.title="K9s"
LABEL org.opencontainers.image.version="${KUBECTL_VERSION}"
LABEL org.opencontainers.image.source="https://github.com/derailed/k9s"
LABEL org.opencontainers.image.description="K9s is a terminal UI to interact with your Kubernetes clusters"

COPY --from=build /k9s/execs/k9s /usr/local/bin/k9s

RUN apk add --no-cache ca-certificates curl \
  && TARGET_ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) \
  && curl -fsSL https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGET_ARCH}/kubectl -o /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  && apk del curl

ENTRYPOINT ["/usr/local/bin/k9s"]
