# -----------------------------------------------------------------------------
# The base image for building the k9s binary

FROM golang:1.17-alpine3.14 AS build

WORKDIR /k9s
COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd
RUN apk --no-cache add make git gcc libc-dev curl && make build

# -----------------------------------------------------------------------------
# Build the final Docker image

FROM alpine:3.14.2
ARG KUBECTL_VERSION="v1.21.2"

COPY --from=build /k9s/execs/k9s /bin/k9s
RUN apk add --update ca-certificates \
  && apk add --update -t deps curl vim \
  && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  && apk del --purge deps \
  && rm /var/cache/apk/*

ENTRYPOINT [ "/bin/k9s" ]
