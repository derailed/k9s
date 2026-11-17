# K9s kcp workspace plugin

This plugin enables the user to navigate between kcp workspaces in k9s.
[About kcp](https://docs.kcp.io/)

## Keyboard shortcuts

`Shift-W` will start the interactive workspace tree navigator. (same as `kubectl ws -i`)
`Shift-P` will navigate to the parent workspace if available. (same as `kubectl kcp workspace use -`)

## Aliases

`:ws` will show the list of workspaces

## Pre-requisites

The kubectl kcp plugin is installed. [Link](https://docs.kcp.io/kcp/v0.33/setup/kubectl-plugin/)

## Installation

1. add `ws: tenancy.kcp.io/v1alpha1/workspaces` to your k9s aliases (`.config/k9s/aliases.yaml`)
2. copy the kcp directory into `.config/k9s/`
3. copy the plugins.yaml into `.config/k9s/` (or append to your existing plugins.yaml)
4. add `alias k9s="$HOME/.config/k9s/kcp/k9s-kcp"` to your `.bashrc` / `.zshrc`
