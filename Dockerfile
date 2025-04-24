FROM python:3.12.2-bookworm

ARG AZURE_CLI_VERSION="2.70.0"
ARG K9S_VERSION="v0.50.3"
#https://github.com/derailed/k9s/releases/download/v0.26.7/k9s_Darwin_arm64.tar.gz 
ARG USERNAME=k9s
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ARG KUBECTL_VERSION="v1.32.2"
#ARG KUBELOGIN_VERSION="v0.0.27"
ARG LINODE_CLI_VERSION="5.56.3"
#ARG ARGO_CLI_VERSION="v3.5.2"
ARG HELM_VERSION="v3.17.1"


RUN apt-get update && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends sudo apt-transport-https ca-certificates neovim fish jq \
    && pip install azure-cli==${AZURE_CLI_VERSION} linode-cli==${LINODE_CLI_VERSION} \
    # Install k9s
    && ARCHITECTURE=$(dpkg --print-architecture) \
    && wget --output-document=/tmp/k9s.tar.gz https://github.com/derailed/k9s/releases/download/${K9S_VERSION}/k9s_Linux_${ARCHITECTURE}.tar.gz \
    && cd tmp && tar xzf k9s.tar.gz \
    && mv k9s /usr/local/bin/ && rm -rf /tmp/* \
    # Install kubectl
    && curl -LO https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCHITECTURE}/kubectl \
    && chmod +x ./kubectl \
    && mv ./kubectl /usr/bin/ \
    # Install helm 
    && curl -L -o /tmp/helm.tar.gz https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCHITECTURE}.tar.gz \
    && cd /tmp && tar xzf helm.tar.gz \
    && mv linux-${ARCHITECTURE}/helm /usr/local/bin/helm && chmod +x /usr/local/bin/helm \
    # Install azure kubelogin
    # && cd /tmp \
    # && curl -LO https://github.com/Azure/kubelogin/releases/download/${KUBELOGIN_VERSION}/kubelogin-linux-${ARCHITECTURE}.zip \
    # && unzip kubelogin-linux-${ARCHITECTURE}.zip \
    # && mv bin/linux_${ARCHITECTURE}/kubelogin /usr/local/bin/ \
    # Install yq
    && curl -L -o /tmp/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${ARCHITECTURE} \
    && chmod +x /tmp/yq \
    && mv /tmp/yq /usr/local/bin/ \
    # Install argo
    #&& cd /tmp \
    #&& echo "https://github.com/argoproj/argo-workflows/releases/download/${ARGO_CLI_VERSION}/argo-${ARGO_CLI_VERSION}-linux-${ARCHITECTURE}.tar.gz" \
    #&& curl -L -o /tmp/argo.tar.gz argo-linux-amd64.gz https://github.com/argoproj/argo-workflows/releases/download/${ARGO_CLI_VERSION}/argo-linux-${ARCHITECTURE}.gz \
    #&& tar xvf argo.tar.gz \
    #&& mv argo /usr/local/bin/ \
    # Create a non-root user to use if preferred - see https://aka.ms/vscode-remote/containers/non-root-user.
    && groupadd --gid $USER_GID $USERNAME \
    && useradd -s /bin/bash --uid $USER_UID --gid $USER_GID -m $USERNAME \
    # [Optional] Add sudo support for non-root user
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    # Clean up
    && rm -rf /tmp/* \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

COPY ./kubeconfig-gen.py /usr/local/bin/linode-kubeconfig
COPY ./in-cluster-config.sh /usr/local/bin/in-cluster-config

USER k9s
WORKDIR /home/k9s
