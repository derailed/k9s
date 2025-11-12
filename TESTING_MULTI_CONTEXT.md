# Testing Multi-Context Support

This guide will help you test the new multi-context feature in K9s.

## Prerequisites

1. **Multiple Kubernetes contexts configured**:
   ```bash
   kubectl config get-contexts
   ```
   You should see at least 2 contexts listed.

2. **Built K9s binary**:
   ```bash
   cd /path/to/k9s
   make build
   ./execs/k9s version
   ```

## Quick Test Guide

### Step 1: Launch K9s

```bash
./execs/k9s
```

### Step 2: Navigate to Contexts View

Press `:` and type `ctx`, then press Enter.

Or use the keyboard shortcut `Ctrl+K` (if configured).

You should see a list of all your Kubernetes contexts.

### Step 3: Mark Multiple Contexts

1. Use **arrow keys** to navigate to a context
2. Press **Space** to mark it (you'll see it highlighted/marked)
3. Navigate to another context
4. Press **Space** again to mark it
5. Repeat for as many contexts as you want

**Quick tip**: Press **M** to mark all contexts at once!

### Step 4: Activate Multi-Context Mode

With multiple contexts marked, press **Enter**.

You should see:
- A confirmation dialog asking if you want to switch to multi-context mode
- The number of contexts listed

Select "OK" or press Enter to confirm.

### Step 5: Verify Multi-Context Mode

After switching, you should see:

1. **Header indicator**: Top-left should show `multi:N` where N is the number of contexts
   - Example: `multi:3` means 3 contexts are active

2. **Success message**: A flash message confirming the switch

### Step 6: View Resources Across Contexts

Now try viewing resources:

```
:pods      # View pods from all contexts
:deploy    # View deployments from all contexts
:svc       # View services from all contexts
```

You should see resources from **all selected contexts** combined in a single view!

### Step 7: Return to Single Context

1. Press `:ctx` to go back to contexts view
2. Navigate to a single context (don't mark any)
3. Press **Enter**

You should:
- See the header change from `multi:N` to the context name
- Return to normal single-context mode

## Detailed Testing Scenarios

### Scenario 1: Test with 2 Contexts

**Purpose**: Verify basic multi-context functionality

1. Mark exactly 2 contexts
2. Activate multi-context mode
3. Navigate to pods view
4. Verify pods from both contexts are visible
5. Try filtering with `/` - should work across both contexts
6. Try sorting - should work across aggregated data

**Expected**: All operations work seamlessly with 2 contexts.

### Scenario 2: Test with All Contexts

**Purpose**: Test performance with many contexts

1. Press `M` in contexts view to mark all
2. Activate multi-context mode
3. Navigate to various resource views
4. Monitor performance and responsiveness

**Expected**: Slight delay proportional to number of contexts, but stable operation.

### Scenario 3: Test Permissions

**Purpose**: Verify authorization across contexts

1. Use contexts where you have different permissions
2. Activate multi-context mode
3. Try operations like delete, edit
4. Observe permission checks

**Expected**: Operations require permission in all contexts.

### Scenario 4: Test Namespace Switching

**Purpose**: Verify namespace handling

1. Activate multi-context with 2+ contexts
2. Press `n` to change namespace
3. Select a namespace
4. Verify resources update across all contexts

**Expected**: Namespace switch applies to all contexts.

### Scenario 5: Test Context Switching

**Purpose**: Verify returning to single context

1. Start in multi-context mode
2. Go to contexts view
3. Select single context without marking
4. Press Enter
5. Verify switch to single-context

**Expected**: Clean transition without errors.

## Keyboard Shortcuts Summary

### In Context View

| Key     | Action |
|---------|--------|
| Space   | Toggle mark on selected context |
| M       | Mark all contexts |
| U       | Unmark all contexts |
| Enter   | Use selected context(s) |
| Esc     | Go back |

### In Any View

| Key     | Action |
|---------|--------|
| :ctx    | Go to contexts view |
| :pods   | View pods |
| /       | Filter |
| n       | Change namespace |
| ?       | Show help |

## Troubleshooting Tests

### Test: Invalid Context Handling

1. Manually edit `~/.kube/config` to add an invalid context
2. Try selecting it with other valid contexts
3. Activate multi-context

**Expected**: Error message, but valid contexts should still work.

### Test: Connection Loss

1. Activate multi-context mode
2. Disconnect network to one cluster (if possible)
3. Navigate to resources

**Expected**: Resources from connected contexts still visible.

### Test: Permission Denied

1. Select contexts where you lack some permissions
2. Activate multi-context
3. Try restricted operations

**Expected**: Clear permission denied message.

## Debug Mode Testing

For detailed logs:

```bash
./execs/k9s -l debug
```

Then in another terminal:
```bash
tail -f ~/.local/state/k9s/k9s.log
```

Look for log entries containing:
- `multi-connection`
- `context switch`
- `Creating multi-context connection`

## Performance Testing

### Quick Performance Check

1. Activate multi-context with 3+ contexts
2. Navigate to pods view
3. Note the refresh time (shown in UI)
4. Compare with single-context mode

**Acceptable**: < 2x slower than single context for 3 contexts

### Stress Test

1. Activate multi-context with 5+ contexts (if available)
2. Navigate through different resource views
3. Apply filters
4. Check memory usage: `ps aux | grep k9s`

**Acceptable**: Stable operation, no memory leaks

## Visual Verification Checklist

- [ ] Multi-context indicator shows in header (`multi:N`)
- [ ] Marked contexts are visually distinct in context list
- [ ] Flash messages appear for context switches
- [ ] Resources from multiple contexts appear in views
- [ ] Filters work across all contexts
- [ ] Sorting works on aggregated data
- [ ] Context switching back to single mode works
- [ ] No UI glitches or rendering issues

## Integration Testing

### With Other K9s Features

1. **Logs**: View logs from pods in different contexts
2. **Describe**: Describe resources from different contexts
3. **Filter**: Filter resources across contexts
4. **Sort**: Sort aggregated resources
5. **Search**: Search across all contexts

## Reporting Issues

If you find issues, please report with:

1. **K9s version**: `./execs/k9s version`
2. **Number of contexts**: `kubectl config get-contexts | wc -l`
3. **Steps to reproduce**: Detailed steps
4. **Logs**: From `~/.local/state/k9s/k9s.log`
5. **Screenshots**: If UI-related

## Success Criteria

✅ Multi-context mode activates successfully
✅ Resources from all contexts are visible
✅ All standard K9s operations work
✅ Performance is acceptable
✅ Can switch back to single-context mode
✅ No crashes or hangs
✅ Logs show no errors (when using valid contexts)

## Next Steps

After testing:

1. **Try real-world scenarios**: Monitor your actual dev/staging/prod clusters
2. **Explore custom views**: Create views tailored for multi-context
3. **Test with teammates**: Get feedback from other users
4. **Report findings**: Share your experience!

---

**Happy Testing! 🚀**

If everything works well, you now have a powerful tool for managing multiple Kubernetes clusters simultaneously!
