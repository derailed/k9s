#!/bin/bash


kubectl -n argocd run -it --rm --image=ghcr.io/ozlevka-work/k9s:v0.32.4-cluster \
    --overrides='{"apiVersion":"v1","spec":{"serviceAccountName":"events-executor"}}' \
    --image=ghcr.io/ozlevka-work/k9s:v0.32.4-cluster \
    pythjon-k9s -- bash
