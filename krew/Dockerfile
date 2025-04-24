FROM  ghcr.io/ozlevka-work/k9s:v0.50.3

ARG KREW_VERSION="v0.4.5"

# Deploy krew
RUN cd /tmp && curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/download/${KREW_VERSION}/krew-linux_amd64.tar.gz" \
    && tar zxvf krew-linux_amd64.tar.gz \
    && KREW=./krew-linux_amd64 && $KREW install krew 

ENV PATH=$PATH:/home/k9s/.krew/bin
