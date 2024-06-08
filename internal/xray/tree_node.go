// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/fvbommel/sortorder"
	"github.com/rs/zerolog/log"
)

const (
	// KeyParent indicates a parent node context key.
	KeyParent TreeRef = "parent"

	// KeySAAutomount indicates whether an automount sa token is active or not.
	KeySAAutomount TreeRef = "automount"

	// PathSeparator represents a node path separator.
	PathSeparator = "::"

	// StatusKey status map key.
	StatusKey = "status"

	// InfoKey state map key.
	InfoKey = "info"

	// OkStatus stands for all is cool.
	OkStatus = "ok"

	// ToastStatus stands for a resource is not up to snuff
	// aka not running or incomplete.
	ToastStatus = "toast"

	// CompletedStatus stands for a completed resource.
	CompletedStatus = "completed"

	// MissingRefStatus stands for a non existing resource reference.
	MissingRefStatus = "noref"
)

// ----------------------------------------------------------------------------

// TreeRef namespaces tree context values.
type TreeRef string

// ----------------------------------------------------------------------------

// NodeSpec represents a node resource specification.
type NodeSpec struct {
	GVRs, Paths, Statuses []string
}

// ParentGVR returns the parent GVR.
func (s NodeSpec) ParentGVR() *string {
	if len(s.GVRs) > 1 {
		return &s.GVRs[1]
	}
	return nil
}

// ParentPath returns the parent path.
func (s NodeSpec) ParentPath() *string {
	if len(s.Paths) > 1 {
		return &s.Paths[1]
	}
	return nil
}

// GVR returns the current GVR.
func (s NodeSpec) GVR() string {
	return s.GVRs[0]
}

// Path returns the current path.
func (s NodeSpec) Path() string {
	return s.Paths[0]
}

// Status returns the current status.
func (s NodeSpec) Status() string {
	return s.Statuses[0]
}

// AsPath returns path hierarchy as string.
func (s NodeSpec) AsPath() string {
	return strings.Join(s.Paths, PathSeparator)
}

// AsGVR returns a gvr hierarchy as string.
func (s NodeSpec) AsGVR() string {
	return strings.Join(s.GVRs, PathSeparator)
}

// AsStatus returns a status hierarchy as string.
func (s NodeSpec) AsStatus() string {
	return strings.Join(s.Statuses, PathSeparator)
}

// ----------------------------------------------------------------------------

// ChildNodes represents a collection of children nodes.
type ChildNodes []*TreeNode

// Len returns the list size.
func (c ChildNodes) Len() int {
	return len(c)
}

// Swap swaps list values.
func (c ChildNodes) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less returns true if i < j.
func (c ChildNodes) Less(i, j int) bool {
	id1, id2 := c[i].ID, c[j].ID

	return sortorder.NaturalLess(id1, id2)
}

// ----------------------------------------------------------------------------

// TreeNode represents a resource tree node.
type TreeNode struct {
	GVR, ID  string
	Children ChildNodes
	Parent   *TreeNode
	Extras   map[string]string
}

// NewTreeNode returns a new instance.
func NewTreeNode(gvr, id string) *TreeNode {
	return &TreeNode{
		GVR:    gvr,
		ID:     id,
		Extras: map[string]string{StatusKey: OkStatus},
	}
}

// CountChildren returns the children count.
func (t *TreeNode) CountChildren() int {
	return len(t.Children)
}

// Count all the nodes from this node.
func (t *TreeNode) Count(gvr string) int {
	counter := 0
	if t.GVR == gvr || gvr == "" {
		counter++
	}
	for _, c := range t.Children {
		counter += c.Count(gvr)
	}
	return counter
}

// Diff computes a tree diff.
func (t *TreeNode) Diff(d *TreeNode) bool {
	if t == nil {
		return d != nil
	}

	if t.CountChildren() != d.CountChildren() {
		return true
	}

	if t.ID != d.ID || t.GVR != d.GVR || !reflect.DeepEqual(t.Extras, d.Extras) {
		return true
	}
	for i := 0; i < len(t.Children); i++ {
		if t.Children[i].Diff(d.Children[i]) {
			return true
		}
	}
	return false
}

// Sort sorts the tree nodes.
func (t *TreeNode) Sort() {
	sort.Sort(t.Children)
	for _, c := range t.Children {
		c.Sort()
	}
}

// Spec returns this node specification.
func (t *TreeNode) Spec() NodeSpec {
	var GVRs, Paths, Statuses []string
	for parent := t; parent != nil; parent = parent.Parent {
		GVRs = append(GVRs, parent.GVR)
		Paths = append(Paths, parent.ID)
		Statuses = append(Statuses, parent.Extras[StatusKey])
	}

	return NodeSpec{
		GVRs:     GVRs,
		Paths:    Paths,
		Statuses: Statuses,
	}
}

// Flatten returns a collection of node specs.
func (t *TreeNode) Flatten() []NodeSpec {
	refs := make([]NodeSpec, 0, len(t.Children))
	for _, c := range t.Children {
		if c.IsLeaf() {
			refs = append(refs, c.Spec())
			continue
		}
		refs = append(refs, c.Flatten()...)
	}
	return refs
}

// Blank returns true if this node is unset.
func (t *TreeNode) Blank() bool {
	return t.GVR == "" && t.ID == ""
}

// Hydrate hydrates a full tree bases on a collection of specifications.
func Hydrate(specs []NodeSpec) *TreeNode {
	root := NewTreeNode("", "")
	nav := root
	for _, spec := range specs {
		for i := len(spec.Paths) - 1; i >= 0; i-- {
			if nav.Blank() {
				nav.GVR, nav.ID, nav.Extras[StatusKey] = spec.GVRs[i], spec.Paths[i], spec.Statuses[i]
				continue
			}
			c := NewTreeNode(spec.GVRs[i], spec.Paths[i])
			c.Extras[StatusKey] = spec.Statuses[i]
			if n := nav.Find(spec.GVRs[i], spec.Paths[i]); n == nil {
				nav.Add(c)
				nav = c
			} else {
				nav = n
			}
		}
		nav = root
	}

	return root
}

// Level computes the current node level.
func (t *TreeNode) Level() int {
	var level int
	p := t
	for p != nil {
		p = p.Parent
		level++
	}
	return level - 1
}

// MaxDepth computes the max tree depth.
func (t *TreeNode) MaxDepth(depth int) int {
	max := depth
	for _, c := range t.Children {
		m := c.MaxDepth(depth + 1)
		if m > max {
			max = m
		}
	}
	return max
}

// Root returns the current tree root node.
func (t *TreeNode) Root() *TreeNode {
	for p := t; p != nil; p = p.Parent {
		if p.Parent == nil {
			return p
		}
	}
	return nil
}

// IsLeaf returns true if node has no children.
func (t *TreeNode) IsLeaf() bool {
	return t.CountChildren() == 0
}

// IsRoot returns true if node is top node.
func (t *TreeNode) IsRoot() bool {
	return t.Parent == nil
}

// ShallowClone performs a shallow node clone.
func (t *TreeNode) ShallowClone() *TreeNode {
	return &TreeNode{GVR: t.GVR, ID: t.ID, Extras: t.Extras}
}

// Filter filters the node based on query.
func (t *TreeNode) Filter(q string, filter func(q, path string) bool) *TreeNode {
	specs := t.Flatten()
	matches := make([]NodeSpec, 0, len(specs))
	for _, s := range specs {
		if filter(q, s.AsPath()+s.AsStatus()) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		return nil
	}
	return Hydrate(matches)
}

// Add adds a new child node.
func (t *TreeNode) Add(c *TreeNode) {
	c.Parent = t
	t.Children = append(t.Children, c)
}

// Clear delete all descendant nodes.
func (t *TreeNode) Clear() {
	t.Children = []*TreeNode{}
}

// Find locates a node given a gvr/id spec.
func (t *TreeNode) Find(gvr, id string) *TreeNode {
	if t.GVR == gvr && t.ID == id {
		return t
	}
	for _, c := range t.Children {
		if v := c.Find(gvr, id); v != nil {
			return v
		}
	}
	return nil
}

// Title computes the node title.
func (t *TreeNode) Title(noIcons bool) string {
	return t.computeTitle(noIcons)
}

// ----------------------------------------------------------------------------
// Helpers...

// Dump for debug...
func (t *TreeNode) Dump() {
	dump(t, 0)
}

func dump(n *TreeNode, level int) {
	if n == nil {
		log.Debug().Msgf("NO DATA!!")
		return
	}
	log.Debug().Msgf("%s%s::%s\n", strings.Repeat("  ", level), n.GVR, n.ID)
	for _, c := range n.Children {
		dump(c, level+1)
	}
}

// DumpStdOut to stdout for debug.
func (t *TreeNode) DumpStdOut() {
	dumpStdOut(t, 0)
}

func dumpStdOut(n *TreeNode, level int) {
	if n == nil {
		fmt.Println("NO DATA!!")
		return
	}
	fmt.Printf("%s%s::%s\n", strings.Repeat("  ", level), n.GVR, n.ID)
	for _, c := range n.Children {
		dumpStdOut(c, level+1)
	}
}

func category(gvr string) string {
	meta, err := dao.MetaAccess.MetaFor(client.NewGVR(gvr))
	if err != nil {
		return ""
	}

	return meta.SingularName
}

func (t TreeNode) computeTitle(noIcons bool) string {
	if !noIcons {
		return t.toEmojiTitle()
	}

	return t.toTitle()
}

const (
	titleFmt    = " [gray::-]%s/[white::b][%s::b]%s[::]"
	topTitleFmt = " [white::b][%s::b]%s[::]"
	toast       = "TOAST"
)

func (t TreeNode) toTitle() (title string) {
	_, n := client.Namespaced(t.ID)
	color, status := "white", "OK"
	if v, ok := t.Extras[StatusKey]; ok {
		switch v {
		case ToastStatus:
			color, status = "orangered", toast
		case MissingRefStatus:
			color, status = "orange", toast+"_REF"
		}
	}
	defer func() {
		if status != "OK" {
			title += fmt.Sprintf("  [gray::-][yellow:%s:b]%s[gray::-]", color, status)
		}
	}()

	categ := category(t.GVR)
	if categ == "" {
		title = fmt.Sprintf(topTitleFmt, color, n)
	} else {
		title = fmt.Sprintf(titleFmt, categ, color, n)
	}

	if !t.IsLeaf() {
		title += fmt.Sprintf("[white::d](%d[-::d])[-::-]", t.CountChildren())
	}

	info, ok := t.Extras[InfoKey]
	if !ok {
		return
	}
	title += fmt.Sprintf(" [antiquewhite::][%s][::]", info)

	return
}

const colorFmt = "%s [%s::b]%s[::]"

func (t TreeNode) toEmojiTitle() (title string) {
	_, n := client.Namespaced(t.ID)
	color, status := "white", "OK"
	if v, ok := t.Extras[StatusKey]; ok {
		switch v {
		case ToastStatus:
			color, status = "orangered", toast
		case MissingRefStatus:
			color, status = "orange", toast+"_REF"
		}
	}
	defer func() {
		if status != "OK" {
			title += fmt.Sprintf(" [gray::-][yellow:%s:b]%s[gray::-]", color, status)
		}
	}()

	title = fmt.Sprintf(colorFmt, toEmoji(t.GVR), color, n)
	if !t.IsLeaf() {
		title += fmt.Sprintf("[white::d](%d[-::d])[-::-]", t.CountChildren())
	}

	info, ok := t.Extras[InfoKey]
	if !ok {
		return
	}
	title += fmt.Sprintf(" [antiquewhite::][%s][::]", info)

	return
}

func toEmoji(gvr string) string {
	if e := v1Emoji(gvr); e != "" {
		return e
	}
	if e := appsEmoji(gvr); e != "" {
		return e
	}
	if e := issueEmoji(gvr); e != "" {
		return e
	}
	switch gvr {
	case "autoscaling/v1/horizontalpodautoscalers":
		return "♎️"
	case "rbac.authorization.k8s.io/v1/clusterrolebindings", "rbac.authorization.k8s.io/v1/clusterroles":
		return "👩‍"
	case "rbac.authorization.k8s.io/v1/rolebindings", "rbac.authorization.k8s.io/v1/roles":
		return "👨🏻‍"
	case "networking.k8s.io/v1/networkpolicies":
		return "📕"
	case "policy/v1/poddisruptionbudgets":
		return "🏷 "
	case "policy/v1beta1/podsecuritypolicies":
		return "👮‍♂️"
	case "containers":
		return "🐳"
	case "report":
		return "🧼"
	default:
		return "📎"
	}
}

func issueEmoji(gvr string) string {
	switch gvr {
	case "issue_0":
		return "👍"
	case "issue_1":
		return "🔊"
	case "issue_2":
		return "☣️ "
	case "issue_3":
		return "🧨"
	default:
		return ""
	}
}

func v1Emoji(gvr string) string {
	switch gvr {
	case "v1/namespaces":
		return "🗂 "
	case "v1/nodes":
		return "🖥 "
	case "v1/pods":
		return "🚛"
	case "v1/services":
		return "💁‍♀️"
	case "v1/serviceaccounts":
		return "💳"
	case "v1/persistentvolumes":
		return "📚"
	case "v1/persistentvolumeclaims":
		return "🎟 "
	case "v1/secrets":
		return "🔒"
	case "v1/configmaps":
		return "🗺 "
	default:
		return ""
	}
}

func appsEmoji(gvr string) string {
	switch gvr {
	case "apps/v1/deployments":
		return "🪂"
	case "apps/v1/statefulsets":
		return "🎎"
	case "apps/v1/daemonsets":
		return "😈"
	case "apps/v1/replicasets":
		return "👯‍♂️"
	default:
		return ""
	}
}

// EmojiInfo returns emoji help.
func EmojiInfo() map[string]string {
	GVRs := []string{
		"containers",
		"v1/namespaces",
		"v1/pods",
		"v1/services",
		"v1/serviceaccounts",
		"v1/persistentvolumes",
		"v1/persistentvolumeclaims",
		"v1/secrets",
		"v1/configmaps",
		"apps/v1/deployments",
		"apps/v1/statefulsets",
		"apps/v1/daemonsets",
	}

	m := make(map[string]string, len(GVRs))
	for _, g := range GVRs {
		m[client.NewGVR(g).R()] = toEmoji(g)
	}

	return m
}
