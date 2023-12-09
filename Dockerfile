FROM python:3.10

ARG AZURE_CLI_VERSION=2.47.0
ARG K9S_VERSION="v0.28.0"
#https://github.com/derailed/k9s/releases/download/v0.26.7/k9s_Darwin_arm64.tar.gz 
ARG USERNAME=k9s
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ARG KUBECTL_VERSION="v1.27.5"
#ARG KUBELOGIN_VERSION="v0.0.27"
ARG LINODE_CLI_VERSION="5.45.1"


RUN apt-get update \
    && apt-get install -y --no-install-recommends sudo apt-transport-https ca-certificates vim fish jq \
    && pip install azure-cli==${AZURE_CLI_VERSION} linode-cli==${LINODE_CLI_VERSION} \
    # Install k9s
    && ARCHITECTURE=$(dpkg --print-architecture) \
    && wget --output-document=/tmp/k9s.tar.gz https://github.com/derailed/k9s/releases/download/${K9S_VERSION}/k9s_Linux_${ARCHITECTURE}.tar.gz \
    && cd tmp && tar xzf k9s.tar.gz \
    && mv k9s /usr/local/bin/ && rm -rf /tmp/* \
    # Install kubectl
    && curl -LO https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/${ARCHITECTURE}/kubectl \
    && chmod +x ./kubectl \
    && mv ./kubectl /usr/bin/ \
    # Install azure kubelogin
    # && cd /tmp \
    # && curl -LO https://github.com/Azure/kubelogin/releases/download/${KUBELOGIN_VERSION}/kubelogin-linux-${ARCHITECTURE}.zip \
    # && unzip kubelogin-linux-${ARCHITECTURE}.zip \
    # && mv bin/linux_${ARCHITECTURE}/kubelogin /usr/local/bin/ \
    # Install yq
    && curl -L -o /tmp/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${ARCHITECTURE} \
    && chmod +x /tmp/yq \
    && mv /tmp/yq /usr/local/bin/ \
    # Create a non-root user to use if preferred - see https://aka.ms/vscode-remote/containers/non-root-user.
    && groupadd --gid $USER_GID $USERNAME \
    && useradd -s /bin/bash --uid $USER_UID --gid $USER_GID -m $USERNAME \
    # [Optional] Add sudo support for non-root user
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

COPY ./kubeconfig-gen.py /usr/local/bin/linode-kubeconfig

USER k9s
WORKDIR /home/k9s
