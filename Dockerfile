FROM golang:alpine AS builder

ADD ./ /k9s
WORKDIR /k9s
RUN apk add make git gcc libc-dev curl
RUN curl -o aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/1.13.7/2019-06-11/bin/linux/amd64/aws-iam-authenticator
RUN make build

FROM alpine:latest

RUN mkdir /k9s /root/.kube /root/.aws
COPY --from=builder /k9s/execs/k9s /bin/k9s
COPY --from=builder /k9s/aws-iam-authenticator /bin/aws-iam-authenticator
RUN chmod +x /bin/aws-iam-authenticator
