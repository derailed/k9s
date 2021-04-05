FROM python:3-alpine3.12

ARG KUBECTL_VERSION="v1.18.2"
ARG K9S_VERSION="v0.24.7"


RUN apk add --update ca-certificates \
  && apk add --update -t deps curl vim tar \
  && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  && curl -L -o k9s.tar.gz https://github.com/derailed/k9s/releases/download/${K9S_VERSION}/k9s_Linux_x86_64.tar.gz \
  && tar xvfz k9s.tar.gz \
  && mv ./k9s /bin/k9s \
  && rm -f k9s.tar.gz \
  && apk del --purge deps \
  && rm /var/cache/apk/* \
  && pip install awscli



