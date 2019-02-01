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

var logoSmall = []string{

	` __     ______         `,
	`|  | __/  __  \  ______`,
	`|  |/ />      < /  ___/`,
	`|    </   --   \\___ \ `,
	`|__|_ \______  /____  >`,
	`     \/      \/     \/ `,
}

var logo = []string{
	` ____  __. ______        _________              .___`,
	`|    |/ _|/  __  \  _____\_   ___ \  _____    __| _/`,
	`|      <  >      < /  ___/    \  \/ /     \  / __ | `,
	`|    |  \/   --   \\___ \\     \___|  Y Y  \/ /_/ | `,
	`|____|__ \______  /____  >\______  /__|_|  /\____ | `,
	`        \/      \/     \/        \/      \/      \/ `,
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
