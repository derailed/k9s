# K9s Emojis

This directory contains emoji configurations for K9s. You can use these emoji sets in your skins or create your own custom set.

## Usage

To use an emoji file in your skin, add a reference to it in your skin configuration file:

```yaml
k9s:
  # ... other skin settings
  emojis:
    file: "emojis/fun.yaml"  # Path relative to ~/.k9s or absolute path
```

You can also define emojis directly in your skin file:

```yaml
k9s:
  # ... other skin settings
  emojis:
    prompt:
      filter: "🔍"
      query: "❓"
    status:
      warn: "😗"
      info: "😎"
      error: "😡"
    # ... more emoji definitions
```

## Available Emoji Files

- `standard.yaml` - Default emoji set used by K9s
- `fun.yaml` - A more playful set of emojis

## Creating Custom Emoji Files

You can create your own emoji sets by copying one of the existing files and customizing the emojis. The file structure should follow this pattern:

```yaml
prompt:
  filter: "🔍"  # Used for filter prompts
  query: "❓"    # Used for query prompts

status:
  warn: "😗"     # Used for warning notifications
  info: "😎"     # Used for info notifications
  error: "😡"    # Used for error notifications

xray:
  namespace: "🗂 "           # Used for namespaces in xray view
  node: "🖥 "                # Used for nodes
  pod: "🚛"                  # Used for pods
  service: "💁‍♀️"             # Used for services
  serviceAccount: "💳"       # Used for service accounts
  persistentVolume: "📚"     # Used for persistent volumes
  pvc: "🎟 "                 # Used for persistent volume claims
  secret: "🔒"               # Used for secrets
  configMap: "🗺 "           # Used for config maps
  deployment: "🪂"           # Used for deployments
  statefulSet: "🎎"          # Used for stateful sets
  daemonSet: "😈"            # Used for daemon sets
  replicaSet: "👯‍♂️"          # Used for replica sets
  container: "🐳"            # Used for containers
  hpa: "♎️"                  # Used for horizontal pod autoscalers
  networkPolicy: "📕"        # Used for network policies
  podDisruptionBudget: "🏷 " # Used for pod disruption budgets
  clusterRole: "👩‍"          # Used for cluster roles
  clusterRoleBinding: "👩‍"  # Used for cluster role bindings
  role: "👨🏻‍"                # Used for roles
  roleBinding: "👨🏻‍"         # Used for role bindings
  default: "📎"              # Used as fallback

custom:
  example: "🦮"              # Custom user-defined emojis
  kubernetes: "☸️"           # You can add any custom emojis you want
```

Place your custom emoji file in the `~/.k9s/emojis/` directory to use it with your skins.