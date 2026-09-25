## Description

This PR addresses Issue #4121 - "Make it clearer if there are no resources, or if they just haven't been loaded"

### Problem

When switching views in k9s, it was unclear whether:
- There are no resources in the current view
- Resources are still being loaded (especially noticeable in Secrets view which loads slowly)

Users would see a blank table with no indication of the loading state.

### Solution

Added a loading spinner indicator that displays "Loading [resource]..." when:
1. Switching to a new view
2. Resources are being synchronized from the cluster
3. Informer cache hasn't synced yet

The spinner uses Flash messages to display loading status, which provides clear visual feedback to users.

## Changes

### New Files
- `internal/ui/spinner.go` - Spinner component with animated braille characters
- `internal/ui/spinner_test.go` - Unit tests for Spinner component

### Modified Files
- `internal/view/browser.go` - Integrated spinner into Browser's loading lifecycle:
  - Added `spinner *ui.Spinner` field to Browser struct
  - `NewBrowser()` initializes spinner
  - `Start()` shows spinner when loading resources
  - `TableNoData()` shows spinner when informer not synced
  - `TableDataChanged()` hides spinner when data loads
  - `TableLoadFailed()` hides spinner on error
  - Added `showSpinner()` and `hideSpinner()` helper methods

## Testing

### Manual Testing Steps
1. Build k9s: `go build -o k9s.exe .`
2. Run k9s: `.\k9s.exe`
3. Switch between different resource views (pods, secrets, deployments, etc.)
4. Observe loading message appears briefly when switching views
5. Verify message disappears when resources are loaded

### Test Environment
- Windows with Docker Desktop Kubernetes
- Multiple namespaces with resources
- Different resource types (secrets, pods, deployments, etc.)

## Screenshots

When switching views, users will see:
```
Loading secrets...
```
or
```
Loading pods in "default" namespace...
```

This message appears at the bottom status bar and clears automatically when resources are loaded.

## Related Issues

Fixes #4121
