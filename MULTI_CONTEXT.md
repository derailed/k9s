# Multi-Context Support in K9s

This document describes the new multi-context feature that allows you to monitor and manage multiple Kubernetes clusters simultaneously within a single K9s session.

## Overview

The multi-context feature enables K9s to connect to and aggregate resources from multiple Kubernetes contexts at once, presenting them as if they were part of a single unified cluster. This is particularly useful for:

- Monitoring resources across multiple clusters (dev, staging, production)
- Managing multi-cluster deployments
- Comparing configurations across environments
- Reducing context-switching overhead

## How It Works

### Architecture

The feature introduces:

1. **MultiConnection**: A wrapper that manages multiple Kubernetes API connections simultaneously
2. **Extended Context View**: Enhanced with multi-selection capabilities (Space, M, U keys)
3. **Aggregated Resource Views**: Resources from all selected contexts are merged and displayed together
4. **Context Indicators**: Visual feedback in the header showing active contexts

### Key Components

- `internal/client/multi_connection.go`: Core multi-connection management
- `internal/view/context_multi.go`: Extended context view with selection
- `internal/view/registrar.go`: Registration of the new context viewer

## Usage

### Switching to Multi-Context Mode

1. **Open the contexts view**:
   ```
   :ctx
   ```
   Or press `Ctrl+K` from any view

2. **Select multiple contexts**:
   - Use `Space` to toggle mark on the currently selected context
   - Use `M` to mark all contexts
   - Use `U` to unmark all contexts
   - Navigate with arrow keys

3. **Activate multi-context mode**:
   - Press `Enter` to switch to the selected contexts
   - A confirmation dialog will appear
   - Confirm to enter multi-context mode

### Visual Indicators

When multiple contexts are active:

- **Header shows**: `multi:N` where N is the number of active contexts
  - Example: `multi:3` means 3 contexts are active

- **Single context shows**: `context-name` (normal behavior)

### Viewing Resources

Once in multi-context mode:

1. **Navigate to any resource view**:
   ```
   :pods
   :deployments
   :services
   ```

2. **Resources from all contexts are shown**:
   - Each resource is prefixed with its context (if needed)
   - All standard K9s operations work normally
   - Filtering, sorting, and searching work across all contexts

### Returning to Single Context

1. **Go back to contexts view**:
   ```
   :ctx
   ```

2. **Select a single context**:
   - Navigate to desired context
   - Press `Enter` (without marking multiple)

3. **K9s switches to single-context mode**

## Key Bindings

### In Context View

| Key       | Action                        |
|-----------|-------------------------------|
| `Space`   | Toggle mark on selected context |
| `M`       | Mark all contexts             |
| `U`       | Unmark all contexts           |
| `Enter`   | Use selected context(s)       |
| `Ctrl+D`  | Delete context (if not read-only) |
| `R`       | Rename context (if not read-only) |

## Examples

### Example 1: Monitor Pods Across Dev, Staging, and Production

```bash
# Start K9s
k9s

# Navigate to contexts
:ctx

# Mark dev, staging, and production contexts
# Use arrow keys and Space to select:
#   ▪ dev-cluster
#   ▪ staging-cluster
#   ▪ prod-cluster

# Press Enter to activate multi-context mode
# Header now shows: multi:3

# View pods from all three clusters
:pods

# All pods from dev, staging, and prod are shown
# You can filter, sort, and interact as normal
```

### Example 2: Compare Deployments Across Regions

```bash
# Navigate to contexts
:ctx

# Mark all regional contexts
#   ▪ us-east-1
#   ▪ us-west-2
#   ▪ eu-central-1

# Activate multi-context mode
Enter

# View deployments
:deployments

# Compare deployment states across all regions
```

### Example 3: Return to Single Context

```bash
# From any view, go to contexts
:ctx

# Select desired context (without marking)
# Navigate to "prod-cluster"

# Press Enter
# K9s switches back to single-context mode
# Header shows: prod-cluster
```

## Technical Details

### Authorization

- Authorization checks (`CanI`) require permission in **all** selected contexts
- If you lack permission in any context, the operation is denied
- This ensures consistent security across all contexts

### Resource Aggregation

Resources are aggregated by:
1. Querying each context's API server
2. Merging results into a unified view
3. Preserving context information for each resource
4. Applying filters and sorting across the aggregate

### Connection Management

- Each context maintains its own API connection
- Connections are established when entering multi-context mode
- Failed connections are logged but don't prevent other contexts
- At least one successful connection is required

### Namespace Handling

- Namespaces from all contexts are combined
- The active namespace applies to all contexts
- You can switch namespaces normally in multi-context mode

### Metrics

- Metrics server availability is checked per-context
- Primary context's metrics server is used for aggregated views
- Missing metrics in some contexts won't break the view

## Limitations and Considerations

1. **Performance**: More contexts = more API calls
   - Start with 2-3 contexts for best performance
   - Consider cluster locations and network latency

2. **Authorization**: Must have permissions in all contexts
   - RBAC rules apply independently per context
   - Missing permissions in one context may limit operations

3. **Resource Conflicts**: Resources with identical names across contexts
   - Context prefix helps distinguish them
   - Use filters to narrow down to specific contexts

4. **Configuration**: Context-specific settings apply to primary context
   - Custom views, skins, and hotkeys use primary context config
   - Other contexts use their own kubeconfig settings

5. **Feature Compatibility**:
   - Port-forwarding: Uses primary context
   - Shell access: Uses primary context
   - Logs: Work across all contexts
   - Exec: Uses resource's originating context

## Best Practices

1. **Group Related Contexts**:
   - Combine contexts that logically belong together
   - Example: All staging environments, all production regions

2. **Use Meaningful Context Names**:
   - Clear names help identify resources quickly
   - Example: `prod-us-east`, `prod-eu-west`

3. **Monitor Performance**:
   - Watch for slow responses with many contexts
   - Reduce contexts if experiencing lag

4. **Leverage Filters**:
   - Use filters to focus on specific contexts or resources
   - Example: `/prod` to show only production resources

5. **Combine with Views**:
   - Custom views work great with multi-context
   - Create views tailored for multi-cluster monitoring

## Troubleshooting

### Problem: "Failed to create multi-connection"

**Cause**: One or more contexts don't exist or are misconfigured

**Solution**:
- Verify contexts exist: `kubectl config get-contexts`
- Check kubeconfig is valid
- Ensure you have access to all selected contexts

### Problem: No resources showing

**Cause**: Authorization issues or connectivity problems

**Solution**:
- Check K9s logs: `tail -f ~/.local/state/k9s/k9s.log`
- Verify RBAC permissions in each context
- Test connectivity: `kubectl --context <context-name> get pods`

### Problem: Slow performance

**Cause**: Too many contexts or high network latency

**Solution**:
- Reduce number of active contexts
- Use contexts with good network connectivity
- Consider filtering to reduce data volume

### Problem: Can't switch back to single context

**Cause**: UI state issue

**Solution**:
- Go to `:ctx` view
- Select single context (ensure no marks)
- Press Enter
- If issue persists, restart K9s

## Configuration

### Per-Context Settings

Each context can have its own configuration at:
```
~/.local/share/k9s/clusters/<cluster>/<context>/config.yaml
```

These settings apply when the context is primary in multi-context mode.

### Global Settings

Global multi-context behavior can be configured in:
```
~/.config/k9s/config.yaml
```

Currently, multi-context uses existing settings and doesn't require special configuration.

## Future Enhancements

Potential improvements for future versions:

- **Saved Multi-Context Groups**: Save and recall context combinations
- **Context-Specific Colors**: Visual distinction per context
- **Parallel Operations**: Bulk operations across contexts
- **Resource Comparison**: Side-by-side diff view
- **Context-Aware Filters**: Advanced filtering by context attributes
- **Multi-Cluster Logs**: Interleaved log streaming from multiple pods
- **Cross-Context Port Forwarding**: Forward ports from multiple contexts

## Contributing

To contribute to multi-context support:

1. Review the architecture in `k9s/CLAUDE.md`
2. Check existing issues labeled `multi-context`
3. Submit PRs to the `feature/multi-context-support` branch
4. Include tests for new functionality

## Support

For issues, questions, or feature requests:

- GitHub Issues: https://github.com/derailed/k9s/issues
- Slack: #k9s in K9sers workspace
- Documentation: https://k9scli.io

## License

This feature is part of K9s and licensed under Apache 2.0.

---

**Note**: This is an experimental feature. While it's been tested thoroughly, please report any issues you encounter. Your feedback helps make K9s better!
