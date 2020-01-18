package xray

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"vbom.ml/util/sortorder"
)

// TreeRef namespaces tree context values.
type TreeRef string

const (
	// KeyParent indicates a parent node context key.
	KeyParent TreeRef = "parent"

	// PathSeparator represents a node path separator.
	PathSeparator = "::"

	// StatusKey status map key.
	StatusKey = "status"

	// StateKey state map key.
	StateKey = "state"

	// OkStatus stands for all is cool.
	OkStatus = "ok"

	// ToastStatus stands for a resource is not up to snuff
	// aka not running or imcomplete.
	ToastStatus = "toast"

	// CompletedStatus stands for a completed resource.
	CompletedStatus = "completed"

	// MissingRefStatus stands for a non existing resource reference.
	MissingRefStatus = "noref"
)

type Childrens []*TreeNode

// Len returns the list size.
func (c Childrens) Len() int {
	return len(c)
}

// Swap swaps list values.
func (c Childrens) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less returns true if i < j.
func (c Childrens) Less(i, j int) bool {
	id1, id2 := c[i].ID, c[j].ID

	return sortorder.NaturalLess(id1, id2)
}

type TreeNode struct {
	GVR, ID  string
	Children Childrens
	Parent   *TreeNode
	Extras   map[string]string
}

func NewTreeNode(gvr, id string) *TreeNode {
	return &TreeNode{
		GVR:    gvr,
		ID:     id,
		Extras: make(map[string]string),
	}
}

func (t *TreeNode) Size() int {
	return len(t.Children)
}

func count(t *TreeNode, counter int) int {
	for _, c := range t.Children {
		counter += count(c, counter)
	}
	return counter
}

func (t *TreeNode) Diff(d *TreeNode) bool {
	if t == nil {
		return d != nil
	}

	if t.Size() != d.Size() {
		log.Debug().Msgf("SIZE-DIFF")
		return true
	}

	if t.ID != d.ID || t.GVR != d.GVR || !reflect.DeepEqual(t.Extras, d.Extras) {
		log.Debug().Msgf("ID DIFF")
		return true
	}
	for i := 0; i < len(t.Children); i++ {
		if t.Children[i].Diff(d.Children[i]) {
			log.Debug().Msgf("CHILD-DIFF")
			return true
		}
	}
	return false
}

func (t *TreeNode) Sort() {
	sortChildren(t)
}

func sortChildren(t *TreeNode) {
	sort.Sort(t.Children)
	for _, c := range t.Children {
		sortChildren(c)
	}
}

type NodeSpec struct {
	GVR, Path string
}

func (t *TreeNode) Spec() NodeSpec {
	parent := t
	var gvr, path []string
	for parent != nil {
		gvr = append(gvr, parent.GVR)
		path = append(path, parent.ID)
		parent = parent.Parent
	}

	return NodeSpec{
		GVR:  strings.Join(gvr, PathSeparator),
		Path: strings.Join(path, PathSeparator),
	}
}

func (t *TreeNode) Flatten() []NodeSpec {
	var refs []NodeSpec
	for _, c := range t.Children {
		if c.IsLeaf() {
			refs = append(refs, c.Spec())
			continue
		}
		refs = append(refs, c.Flatten()...)
	}
	return refs
}

func (t *TreeNode) Blank() bool {
	return t.GVR == "" && t.ID == ""
}

func Hydrate(refs []NodeSpec) *TreeNode {
	root := NewTreeNode("", "")
	nav := root
	for _, ref := range refs {
		ids := strings.Split(ref.Path, PathSeparator)
		gvrs := strings.Split(ref.GVR, PathSeparator)
		for i := len(ids) - 1; i >= 0; i-- {
			if nav.Blank() {
				nav.GVR, nav.ID = gvrs[i], ids[i]
				continue
			}
			c := NewTreeNode(gvrs[i], ids[i])
			if n := nav.Find(gvrs[i], ids[i]); n == nil {
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

func (t *TreeNode) Level() int {
	var level int
	p := t
	for p != nil {
		p = p.Parent
		level++
	}
	return level - 1
}

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

func makeSpacer(d int) string {
	return strings.Repeat("   ", d)
}

func (t *TreeNode) Root() *TreeNode {
	for p := t; p != nil; p = p.Parent {
		if p.Parent == nil {
			return p
		}
	}
	return nil
}

func (r *TreeNode) IsLeaf() bool {
	return r.Empty()
}

func (r *TreeNode) IsRoot() bool {
	return r.Parent == nil
}

func (r *TreeNode) ShallowClone() *TreeNode {
	return &TreeNode{GVR: r.GVR, ID: r.ID, Extras: r.Extras}
}

func (r *TreeNode) Filter(q string, filter func(q, path string) bool) *TreeNode {
	specs := r.Flatten()
	matches := make([]NodeSpec, 0, len(specs))
	for _, s := range specs {
		if filter(q, s.Path) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		return nil
	}
	return Hydrate(matches)
}

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

func (t *TreeNode) Title() string {
	const withNS = "[white::b]%s[-::d]"

	title := fmt.Sprintf(withNS, t.colorize())

	if t.Size() > 0 {
		title += fmt.Sprintf("([white::d]%d[-::d])[-::-]", t.Size())
	}

	return title
}

func (t *TreeNode) Empty() bool {
	return len(t.Children) == 0
}

func (t *TreeNode) Clear() {
	t.Children = []*TreeNode{}
}

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

func (t *TreeNode) Add(c *TreeNode) {
	c.Parent = t
	t.Children = append(t.Children, c)
}

// Helpers...

func statusEmoji(s string) string {
	switch s {
	case "ok":
		return "[green::b]✔︎"
	case "done":
		return "[gray::b]🏁"
	case "bad":
		return "[red::b]𐄂"
	default:
		return ""
	}
}

// 😡👎💥🧨💣🎭 🟥🟩✅✔︎☑️✔️✓
func toEmoji(gvr string) string {
	switch gvr {
	case "v1/pods":
		return "🚛"
	case "apps/v1/deployments":
		return "🪂"
	case "apps/v1/statefulset":
		return "🎎"
	case "apps/v1/daemonsets":
		return "😈"
	case "containers":
		return "🐳"
	case "v1/serviceaccounts":
		return "🛎"
	case "v1/persistentvolumes":
		return "📚"
	case "v1/persistentvolumeclaims":
		return "🎟"
	case "v1/secrets":
		return "🔒"
	case "v1/configmaps":
		return "🗄"
	default:
		return "📎"
	}
}

func (t TreeNode) colorize() string {
	const colorFmt = "%s %s [%s::b]%s[::]"

	_, n := client.Namespaced(t.ID)
	color, flag := "white", "[green::b]OK"
	if v, ok := t.Extras[StatusKey]; ok {
		switch v {
		case ToastStatus:
			color, flag = "orangered", "[red::b]TOAST"
		case MissingRefStatus:
			color, flag = "orange", "[orange::b]MISSING_REF"
		}
	}

	return fmt.Sprintf(colorFmt, toEmoji(t.GVR), flag, color, n)
}
