# Build...
FROM golang:1.12.6-alpine AS build

WORKDIR /k9s
COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd
RUN apk --no-cache add make git gcc libc-dev curl && make build

# -----------------------------------------------------------------------------
# Build Image...

FROM alpine:3.10.0
COPY --from=build /k9s/execs/k9s /bin/k9s
ENTRYPOINT [ "/bin/k9s" ]