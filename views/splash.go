package views

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

const (
	company = "imhotep.io"
	product = "Kubernetes CLI Island Style!"
)

var logo = []string{
	` ____  __.________       `,
	`|    |/ _/   __   \______`,
	`|      < \____    /  ___/`,
	`|    |  \   /    /\___ \ `,
	`|____|__ \ /____//____  >`,
	`         \/            \/`,
}

var co = []string{
	`.__        .__            __                    .__        `,
	`|__| _____ |  |__   _____/  |_  ____ ______     |__| ____  `,
	`|  |/     \|  |  \ /  _ \   ___/ __ \\____ \    |  |/  _ \ `,
	`|  |  Y Y  |   Y  (  <_> |  | \  ___/|  |_> >   |  (  <_> )`,
	`|__|__|_|  |___|  /\____/|__|  \___  |   __/ /\ |__|\____/ `,
	`         \/     \/                 \/|__|    \/            `,
}

// Splash screen definition
type Splash struct {
	*tview.Flex
}

// NewSplash instantiates a new splash screen with product and company info.
func NewSplash(rev string) *Splash {
	v := Splash{tview.NewFlex()}

	t1 := tview.NewTextView()
	t1.SetDynamicColors(true)
	t1.SetBackgroundColor(tcell.ColorDefault)
	t1.SetTextAlign(tview.AlignCenter)
	v.layoutLogo(t1)

	t2 := tview.NewTextView()
	t2.SetDynamicColors(true)
	t2.SetBackgroundColor(tcell.ColorDefault)
	t2.SetTextAlign(tview.AlignCenter)
	v.layoutCo(t2)

	t3 := tview.NewTextView()
	t3.SetDynamicColors(true)
	t3.SetBackgroundColor(tcell.ColorDefault)
	t3.SetTextAlign(tview.AlignCenter)
	v.layoutRev(t3, rev)

	v.SetDirection(tview.FlexRow)
	v.AddItem(t2, 0, 2, false)
	v.AddItem(t1, 0, 4, false)
	v.AddItem(t3, 2, 1, false)
	return &v
}

func (v *Splash) layoutLogo(t *tview.TextView) {
	logo := strings.Join(logo, "\n[orange::b]")
	fmt.Fprintf(t, "%s[orange::b]%s\n", strings.Repeat("\n", 2), logo)
}

func (v *Splash) layoutCo(t *tview.TextView) {
	cos := strings.Join(co, "\n[yellowgreen::b]")
	fmt.Fprintf(t, "[yellowgreen::b]%s\n", cos)
}

func (v *Splash) layoutRev(t *tview.TextView, rev string) {
	fmt.Fprintf(t, "[white::b]Revision [red::b]%s", rev)
}
