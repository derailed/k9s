# k9s + kagent Integration

This integration brings AI-powered Kubernetes management to k9s through kagent.

## Features

### Native Resource Views

When kagent is installed in your cluster, k9s automatically discovers and displays:

- **Agents** (`agents`, `ka`) - AI agents with custom views showing type, model, tools, and status
- **ModelConfigs** (`modelconfigs`, `mc`) - LLM provider configurations with provider details
- **ToolServers** (`toolservers`, `ts`) - MCP tool servers with discovered tools count

### Custom Keybindings

When viewing kagent Agents:

| Key | Action | Description |
|-----|--------|-------------|
| `i` | Invoke | Run the agent with a message |
| `c` | Chat | Open embedded chat interface |
| `t` | Tools | List all tools available to the agent |
| `m` | ModelConfig | Navigate to the agent's model configuration |
| `l` | Logs | View logs for the agent's pods |
| `Ctrl+i` | Quick Invoke | Quick invoke with prompt |

When viewing ToolServers:

| Key | Action | Description |
|-----|--------|-------------|
| `t` | List Tools | Show all discovered tools |
| `a` | Show Agents | Find agents using this ToolServer |
| `Ctrl+t` | Test | Test the connection |

When viewing ModelConfigs:

| Key | Action | Description |
|-----|--------|-------------|
| `a` | Show Agents | Find agents using this ModelConfig |
| `s` | Show Secret | Navigate to the API key secret |
| `Ctrl+t` | Test Model | Show model configuration details |

### Plugins

Install the kagent plugins by copying `kagent.yaml` to your k9s plugins directory:

```bash
cp plugins/kagent.yaml ~/.config/k9s/plugins.yaml
# Or merge with existing plugins
```

Plugin shortcuts (work on any resource):

| Key | Action | Description |
|-----|--------|-------------|
| `Shift+K` | Investigate | Ask kagent to investigate the resource |
| `Shift+C` | Chat | Start interactive chat with context |
| `Shift+A` | Ask | Custom question about the resource |
| `Shift+L` | Analyze Logs | Get AI analysis of logs |
| `Shift+E` | Explain | Explain the resource configuration |
| `Shift+F` | Fix | Get suggestions to fix issues |

### Aliases

Add kagent aliases for quick navigation. Copy `kagent-aliases.yaml` to your aliases file:

```bash
# Merge with existing aliases
cat plugins/kagent-aliases.yaml >> ~/.config/k9s/aliases.yaml
```

Then you can use:
- `:ka` or `:agents` - View agents
- `:mc` or `:models` - View model configs  
- `:ts` or `:tools` - View tool servers
- `:mcp` - View MCP servers (alias for toolservers)

## Requirements

1. **kagent installed in your cluster**
   ```bash
   helm install kagent kagent/kagent -n kagent-system --create-namespace
   ```

2. **kagent CLI** (for plugins and chat)
   ```bash
   curl -sSL https://kagent.dev/install.sh | bash
   ```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KAGENT_K8S_AGENT` | Default agent for k8s operations | `k8s-agent` |
| `KAGENT_AGENT` | Default agent for chat | `k8s-agent` |

## Troubleshooting

### Agents not showing up

Make sure kagent CRDs are installed:
```bash
kubectl get crd agents.kagent.dev
```

### Chat not working

Ensure the kagent CLI is in your PATH:
```bash
which kagent
```

### Plugins not loading

Check your k9s plugins directory:
```bash
ls ~/.config/k9s/plugins.yaml
```

## Contributing

This integration is part of the k9s-kagent-integration project. Contributions welcome!
