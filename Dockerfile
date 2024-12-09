# -----------------------------------------------------------------------------
# The base image for building the k9s binary

FROM golang:1.23.3-alpine3.20 AS build

WORKDIR /k9s
COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd
RUN apk --no-cache add --update make libx11-dev git gcc libc-dev curl \
  && make build

# -----------------------------------------------------------------------------
# Build the final Docker image

FROM alpine:3.21.0
ARG KUBECTL_VERSION="v1.31.2"

COPY --from=build /k9s/execs/k9s /bin/k9s
RUN apk --no-cache add --update ca-certificates \
  && apk --no-cache add --update -t deps curl vim \
  && TARGET_ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) \
  && curl -f -L https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGET_ARCH}/kubectl -o /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  && apk del --purge deps

ENTRYPOINT [ "/bin/k9s" ]
