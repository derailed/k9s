// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/popeye/pkg/config"
)

// Section represents an xray renderer.
type Section struct {
	render.Base
}

// Render renders an xray node.
func (s *Section) Render(ctx context.Context, ns string, o interface{}) error {
	section, ok := o.(render.Section)
	if !ok {
		return fmt.Errorf("expected Section, but got %T", o)
	}
	root := NewTreeNode(section.GVR, section.Title)
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}
	s.outcomeRefs(root, section)
	parent.Add(root)

	return nil
}

func (*Section) outcomeRefs(parent *TreeNode, section render.Section) {
	for k, issues := range section.Outcome {
		p := NewTreeNode(section.GVR, cleanse(k))
		parent.Add(p)
		for _, issue := range issues {
			msg := colorize(cleanse(issue.Message), issue.Level)
			c := NewTreeNode(fmt.Sprintf("issue_%d", issue.Level), msg)
			if issue.Group == "__root__" {
				p.Add(c)
				continue
			}
			if pa := p.Find(issue.GVR, issue.Group); pa != nil {
				pa.Add(c)
				continue
			}
			pa := NewTreeNode(issue.GVR, issue.Group)
			pa.Add(c)
			p.Add(pa)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func colorize(s string, l config.Level) string {
	c := "green"
	// nolint:exhaustive
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

func cleanse(s string) string {
	s = strings.Replace(s, "[", "(", -1)
	s = strings.Replace(s, "]", ")", -1)
	s = strings.Replace(s, "/", "::", -1)
	return s
}
