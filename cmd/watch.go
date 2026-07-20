// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// nativeWorkloadGVRs lists the common native workload resources, in display
// order, that the wizard suggests when present in the target cluster.
var nativeWorkloadGVRs = []string{
	"v1/pods",
	"v1/services",
	"apps/v1/deployments",
	"apps/v1/statefulsets",
	"apps/v1/daemonsets",
	"apps/v1/replicasets",
	"batch/v1/jobs",
	"batch/v1/cronjobs",
}

// standardGroups mirrors the built-in k8s api groups so the wizard can tell
// native resources apart from CRDs.
var standardGroups = map[string]struct{}{
	"apps/v1":            {},
	"autoscaling/v1":     {},
	"autoscaling/v2":     {},
	"batch/v1":           {},
	"batch/v1beta1":      {},
	"extensions/v1beta1": {},
	"policy/v1":          {},
	"policy/v1beta1":     {},
	"v1":                 {},
}

// resourceEntry represents a discoverable cluster resource the wizard can add
// to the workload aggregation view.
type resourceEntry struct {
	gvr   string
	kind  string
	isCRD bool
}

func watchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Interactive wizard to build a workload (:wk) aggregation view",
		Long:  "Interactively pick cluster resources and namespaces, then write them to the workload (:wk) aggregation view in views.yaml.",
		RunE:  runWatch,
	}
	// Reuse the same connection flags as the main command (context, kubeconfig,
	// insecure-skip-tls-verify, ...) so the wizard can target any cluster.
	k8sFlags.AddFlags(cmd.Flags())

	return cmd
}

func runWatch(*cobra.Command, []string) error {
	if err := config.InitLocs(); err != nil {
		return err
	}

	k8sCfg := client.NewConfig(k8sFlags)
	conn, err := client.InitConnection(k8sCfg, slog.Default())
	if err != nil {
		return fmt.Errorf("k8s connection failed: %w", err)
	}
	if !conn.CheckConnectivity() {
		return fmt.Errorf("unable to connect to the cluster, check your kubeconfig/context")
	}

	resources, err := discoverResources(conn)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
		return fmt.Errorf("no resources discovered in the cluster")
	}
	namespaces := discoverNamespaces(conn)

	entries, err := runWatchWizard(os.Stdin, out, resources, namespaces)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(out, "No resources selected, nothing written.")
		return nil
	}

	if err := writeWorkloadsView(config.AppViewsFile, entries); err != nil {
		return err
	}
	fmt.Fprintf(out, "\n✅ Wrote %s, run `k9s` then `:wk` to view.\n", config.AppViewsFile)

	return nil
}

// discoverResources lists the native workloads present in the cluster plus all
// discovered CRDs.
func discoverResources(conn client.Connection) ([]resourceEntry, error) {
	dial, err := conn.CachedDiscovery()
	if err != nil {
		return nil, err
	}
	rr, err := dial.ServerPreferredResources()
	if err != nil {
		// Partial discovery results are still useful (e.g. a flaky aggregated
		// api group), so only fail when nothing came back at all.
		slog.Warn("Partial resource discovery", "error", err)
		if len(rr) == 0 {
			return nil, fmt.Errorf("resource discovery failed: %w", err)
		}
	}

	known := make(map[string]resourceEntry)
	var crds []resourceEntry
	for _, r := range rr {
		for i := range r.APIResources {
			res := r.APIResources[i]
			if strings.Contains(res.Name, "/") || !slices.Contains(res.Verbs, "list") {
				continue
			}
			gvr := client.FromGVAndR(r.GroupVersion, res.Name).String()
			entry := resourceEntry{gvr: gvr, kind: res.Kind, isCRD: !isStandardGroup(r.GroupVersion)}
			known[gvr] = entry
			if entry.isCRD {
				crds = append(crds, entry)
			}
		}
	}

	out := make([]resourceEntry, 0, len(nativeWorkloadGVRs)+len(crds))
	for _, g := range nativeWorkloadGVRs {
		if e, ok := known[g]; ok {
			out = append(out, e)
		}
	}
	slices.SortFunc(crds, func(a, b resourceEntry) int { return strings.Compare(a.gvr, b.gvr) })
	out = append(out, crds...)

	return out, nil
}

func isStandardGroup(gv string) bool {
	if _, ok := standardGroups[gv]; ok {
		return true
	}
	return strings.Contains(gv, ".k8s.io")
}

// discoverNamespaces returns sorted namespace names, best-effort. An empty
// result simply means the namespace prompt falls back to free-form input.
func discoverNamespaces(conn client.Connection) []string {
	nn, err := conn.ValidNamespaceNames()
	if err != nil {
		slog.Warn("Unable to list namespaces", "error", err)
		return nil
	}
	nss := make([]string, 0, len(nn))
	for ns := range nn {
		nss = append(nss, ns)
	}
	slices.Sort(nss)

	return nss
}

// runWatchWizard drives the interactive prompt and returns the assembled
// workload entries ("<gvr>" or "<gvr> <namespace>").
func runWatchWizard(in io.Reader, w io.Writer, resources []resourceEntry, namespaces []string) ([]string, error) {
	scanner := bufio.NewScanner(in)

	fmt.Fprintln(w, "Discovered resources:")
	var lastNative, lastCRD bool
	for i, r := range resources {
		if !r.isCRD && !lastNative {
			fmt.Fprintln(w, "\n  Native workloads:")
			lastNative = true
		}
		if r.isCRD && !lastCRD {
			fmt.Fprintln(w, "\n  Custom resources (CRDs):")
			lastCRD = true
		}
		fmt.Fprintf(w, "    %d) %s\n", i+1, r.gvr)
	}

	fmt.Fprint(w, "\nSelect resources to aggregate (enter numbers, comma-separated, e.g. 1,3,5): ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	indices, err := parseSelection(scanner.Text(), len(resources))
	if err != nil {
		return nil, err
	}

	if len(namespaces) > 0 {
		fmt.Fprintln(w, "\nCluster namespaces:")
		for i, ns := range namespaces {
			fmt.Fprintf(w, "    %d) %s\n", i+1, ns)
		}
	}

	entries := make([]string, 0, len(indices))
	for _, idx := range indices {
		r := resources[idx]
		fmt.Fprintf(w, "\nNamespace for %q? (number or name, enter = follow view/all): ", r.gvr)
		if !scanner.Scan() {
			return nil, scanner.Err()
		}
		ns := resolveNamespace(scanner.Text(), namespaces)
		entries = append(entries, workloadEntry(r.gvr, ns))
	}

	return entries, nil
}

// resolveNamespace turns a namespace answer (a list number or a raw name) into
// a namespace string. An empty answer follows the view's active namespace.
func resolveNamespace(answer string, namespaces []string) string {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return ""
	}
	if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(namespaces) {
		return namespaces[n-1]
	}

	return answer
}

// parseSelection parses a comma-separated 1-based selection (e.g. "1,3,5")
// into deduplicated 0-based indices, validating against the count n.
func parseSelection(input string, n int) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	seen := make(map[int]struct{})
	indices := make([]int, 0)
	for _, tok := range strings.Split(input, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		num, err := strconv.Atoi(tok)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q: not a number", tok)
		}
		if num < 1 || num > n {
			return nil, fmt.Errorf("selection %d out of range (1-%d)", num, n)
		}
		idx := num - 1
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		indices = append(indices, idx)
	}

	return indices, nil
}

// workloadEntry assembles a workload list entry. An empty namespace yields just
// the gvr (follow the view's active namespace); otherwise "<gvr> <namespace>".
func workloadEntry(gvr, ns string) string {
	gvr = strings.TrimSpace(gvr)
	ns = strings.TrimSpace(ns)
	if ns == "" || ns == client.NamespaceAll {
		return gvr
	}

	return gvr + " " + ns
}

// writeWorkloadsView merges the selected entries into the workloads.default
// list of the views.yaml at path, preserving any existing content.
func writeWorkloadsView(path string, entries []string) error {
	var existing []byte
	if path != "" {
		if bb, err := os.ReadFile(path); err == nil {
			existing = bb
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	merged, err := mergeWorkloads(existing, entries)
	if err != nil {
		return err
	}
	if err := data.EnsureDirPath(path, data.DefaultDirMod); err != nil {
		return err
	}

	return os.WriteFile(path, merged, data.DefaultFileMod)
}

// mergeWorkloads sets workloads.default to entries inside an existing views.yaml
// document, preserving the views section and any other workload sets.
func mergeWorkloads(existing []byte, entries []string) ([]byte, error) {
	doc := map[string]any{}
	if len(existing) > 0 {
		if err := yaml.Unmarshal(existing, &doc); err != nil {
			return nil, fmt.Errorf("parse existing views config: %w", err)
		}
		if doc == nil {
			doc = map[string]any{}
		}
	}

	workloads, _ := doc["workloads"].(map[string]any)
	if workloads == nil {
		workloads = map[string]any{}
	}
	workloads[config.DefaultWorkloadGVRs] = entries
	doc["workloads"] = workloads

	return yaml.Marshal(doc)
}
