package xray

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/popeye/pkg/config"
)

// Section represents an xray renderer.
type Section struct{}

// Render renders an xray node.
func (s *Section) Render(ctx context.Context, ns string, o interface{}) error {
	section, ok := o.(render.Section)
	if !ok {
		return fmt.Errorf("Expected Section, but got %T", o)
	}
	root := NewTreeNode(section.Title, section.Title)
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}
	s.outcomeRefs(root, section)
	parent.Add(root)

	return nil
}

func cleanse(s string) string {
	s = strings.Replace(s, "[", "(", -1)
	s = strings.Replace(s, "]", ")", -1)
	s = strings.Replace(s, "/", "::", -1)
	return s
}

func (c *Section) outcomeRefs(parent *TreeNode, section render.Section) {
	for k, issues := range section.Outcome {
		p := NewTreeNode(section.Title, cleanse(k))
		parent.Add(p)
		for _, i := range issues {
			msg := colorize(cleanse(i.Message), i.Level)
			c := NewTreeNode(fmt.Sprintf("issue_%d", i.Level), msg)
			if i.Group == "__root__" {
				p.Add(c)
				continue
			}
			if pa := p.Find(childOf(section.Title), i.Group); pa != nil {
				pa.Add(c)
				continue
			}
			pa := NewTreeNode(childOf(section.Title), i.Group)
			pa.Add(c)
			p.Add(pa)
		}
	}
}

func childOf(s string) string {
	switch s {
	case "deployment", "statefulset", "daemonset":
		return "v1/pods"
	case "pod":
		return "containers"
	default:
		return ""
	}
}

func colorize(s string, l config.Level) string {
	c := "green"
	switch l {
	case config.ErrorLevel:
		c = "red"
	case config.WarnLevel:
		c = "orange"
	case config.InfoLevel:
		c = "blue"
	}
	return fmt.Sprintf("[%s::]%s", c, s)
}
