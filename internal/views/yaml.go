package views

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

const (
	keyColor     = "#00FF00"
	valColor     = "#ADFF2F"
	yamlTitleFmt = " [aqua::-] YAML [orange::-]%s"
)

type yamlView struct {
	*tview.TextView

	app     *appView
	actions keyActions
}

func newYamlView(app *appView) *yamlView {
	v := yamlView{app: app, TextView: tview.NewTextView()}
	{
		v.SetBorder(true)
		v.SetDynamicColors(true)
		v.SetWrap(false)
		v.SetTitleColor(tcell.ColorAqua)
		v.SetInputCapture(v.keyboard)
	}
	return &v
}

func (v *yamlView) setTitle(t string) {
	v.TextView.SetTitle(fmt.Sprintf(yamlTitleFmt, t))
}

func (v *yamlView) clear() {
	v.TextView.Clear()
}

func (v *yamlView) blur() {
}

func (v *yamlView) init(_ context.Context) {
}

// SetActions to handle keyboard inputs
func (v *yamlView) setActions(aa keyActions) {
	v.actions = aa
}

func (v *yamlView) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}
	return nil
}

func (v *yamlView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if evt.Key() == tcell.KeyRune {
		if a, ok := v.actions[evt.Key()]; ok {
			a.action(evt)
			evt = nil
		}
	} else {
		if a, ok := v.actions[evt.Key()]; ok {
			a.action(evt)
			evt = nil
		}
	}
	return evt
}

func (v *yamlView) isEmpty(c interface{}) bool {
	switch c.(type) {
	case []string:
		return len(c.([]string)) == 0
	case map[string]string:
		return len(c.(map[string]string)) == 0
	case map[string]interface{}:
		return len(c.(map[string]interface{})) == 0
	}
	return false
}

func (v *yamlView) update(pp resource.Properties) {
	v.Clear()

	kk := make([]string, 0, len(pp))
	for k := range pp {
		kk = append(kk, k)
	}
	sort.Strings(kk)

	var level int
	for _, k := range kk {
		if v.isEmpty(pp[k]) {
			continue
		}
		fmt.Fprintf(v, "[%s::b]%-20s", keyColor, k+":")
		v.textFor(level, pp[k])
		fmt.Fprintf(v, "\n")
	}
	v.ScrollToBeginning()
	v.app.Draw()
}

func (v *yamlView) textFor(level int, p interface{}) string {
	var indent string
	for i := 0; i < level; i++ {
		if level < 1 {
			indent += strings.Repeat(" ", 20)
		} else {
			indent += strings.Repeat(" ", 2)
		}
	}
	switch p.(type) {
	case string:
		fmt.Fprintf(v, "[%s::-]%s", valColor, p.(string))
	case []string:
		aa := p.([]string)
		for i, s := range aa {
			if i == 0 {
				fmt.Fprintf(v, "[%s::-]%s", valColor, s)
			} else {
				if level < 1 {
					fmt.Fprintf(v, "%s[%s::-]%-20s%s", indent, valColor, "", s)
				} else {
					fmt.Fprintf(v, "%s[%s::-]%-13s%s", indent, valColor, "", s)
				}
			}
			if i+1 < len(aa) {
				fmt.Fprintf(v, "\n")
			}
		}
	case map[string]interface{}:
		m := p.(map[string]interface{})
		indent += strings.Repeat(" ", 2)
		fmt.Fprintf(v, "\n")
		var i int
		for key, val := range m {
			fmt.Fprintf(v, "%s[%s::b]%-12s ", indent, keyColor, key+":")
			v.textFor(level+1, val)
			i++
			if i < len(m) {
				fmt.Fprintf(v, "\n")
			}
		}
	}
	return indent
}
