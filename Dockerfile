# Build...
FROM golang:1.14.1-alpine3.11 AS build

WORKDIR /k9s
COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd
RUN apk --no-cache add make git gcc libc-dev curl && make build

# -----------------------------------------------------------------------------
# Build Image...

FROM alpine:3.10.0

COPY --from=build /k9s/execs/k9s /bin/k9s
ENV KUBE_LATEST_VERSION="v1.18.1"
RUN apk add --update ca-certificates \
  && apk add --update -t deps curl \
  && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  && apk del --purge deps \
  && rm /var/cache/apk/*

ENTRYPOINT [ "/bin/k9s" ]