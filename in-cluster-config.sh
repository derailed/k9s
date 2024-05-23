#!/bin/bash

MASTER="https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT_HTTPS"
KUBECONFIG="${HOME}/.kube/config"
TOKEN="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
CA_CRT="$(base64 -w 0 /var/run/secrets/kubernetes.io/serviceaccount/ca.crt)"

mkdir -p "${HOME}/.kube"

cat <<EOF > "${KUBECONFIG}"
apiVersion: v1
kind: Config
clusters:
- name: default-cluster
  cluster:
    certificate-authority-data: ${CA_CRT}
    server: ${MASTER}
contexts:
- name: default-context
  context:
    cluster: default-cluster
    namespace: default
    user: default-user
current-context: default-context
users:
- name: default-user
  user:
    token: ${TOKEN}
EOF
