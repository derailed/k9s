package views

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/derailed/tview"
)

const (
	company = "imhotep.io"
	product = "Kubernetes CLI Island Style!"
)

var logoSmall = []string{
	` ____  __.________       `,
	`|    |/ _/   __   \______`,
	`|      < \____    /  ___/`,
	`|    |  \   /    /\___ \ `,
	`|____|__ \ /____//____  >`,
	`        \/            \/ `,
}

var logo = []string{
	` ____  __.________      _________ .____    .___ `,
	`|    |/ _/   __   \_____\_   ___ \|    |   |   |`,
	`|      < \____    /  ___/    \  \/|    |   |   |`,
	`|    |  \   /    /\___ \\     \___|    |___|   |`,
	`|____|__ \ /____//____  >\______  /_______ \___|`,
	`        \/            \/        \/        \/    `,
}

// Splash screen definition
type Splash struct {
	*tview.Flex
}

// NewSplash instantiates a new splash screen with product and company info.
func NewSplash(rev string) *Splash {
	v := Splash{tview.NewFlex()}

	logo := tview.NewTextView()
	{
		logo.SetDynamicColors(true)
		logo.SetBackgroundColor(tcell.ColorDefault)
		logo.SetTextAlign(tview.AlignCenter)
	}
	v.layoutLogo(logo)

	vers := tview.NewTextView()
	{
		vers.SetDynamicColors(true)
		vers.SetBackgroundColor(tcell.ColorDefault)
		vers.SetTextAlign(tview.AlignCenter)
	}
	v.layoutRev(vers, rev)

	v.SetDirection(tview.FlexRow)
	v.AddItem(logo, 10, 1, false)
	v.AddItem(vers, 1, 1, false)
	return &v
}

func (v *Splash) layoutLogo(t *tview.TextView) {
	logo := strings.Join(logo, "\n[orange::b]")
	fmt.Fprintf(t, "%s[orange::b]%s\n", strings.Repeat("\n", 2), logo)
}

func (v *Splash) layoutRev(t *tview.TextView, rev string) {
	fmt.Fprintf(t, "[white::b]Revision [red::b]%s", rev)
}
