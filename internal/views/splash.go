package views

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const (
	company = "imhotep.io"
	product = "Kubernetes CLI Island Style!"
)

// LogoSmall K9s small log.
var LogoSmall = []string{
	` ____  __.________       `,
	`|    |/ _/   __   \______`,
	`|      < \____    /  ___/`,
	`|    |  \   /    /\___ \ `,
	`|____|__ \ /____//____  >`,
	`        \/            \/ `,
}

// Logo K9s big logo for splash page.
var Logo = []string{
	` ____  __.________      _________ .____    .___ `,
	`|    |/ _/   __   \_____\_   ___ \|    |   |   |`,
	`|      < \____    /  ___/    \  \/|    |   |   |`,
	`|    |  \   /    /\___ \\     \___|    |___|   |`,
	`|____|__ \ /____//____  >\______  /_______ \___|`,
	`        \/            \/        \/        \/    `,
}

// Splash screen definition
type splashView struct {
	*tview.Flex
}

// NewSplash instantiates a new splash screen with product and company info.
func newSplash(styles *config.Styles, version string) *splashView {
	v := splashView{Flex: tview.NewFlex()}

	logo := tview.NewTextView()
	logo.SetDynamicColors(true)
	logo.SetBackgroundColor(tcell.ColorDefault)
	logo.SetTextAlign(tview.AlignCenter)
	v.layoutLogo(logo, styles)

	vers := tview.NewTextView()
	vers.SetDynamicColors(true)
	vers.SetBackgroundColor(tcell.ColorDefault)
	vers.SetTextAlign(tview.AlignCenter)
	v.layoutRev(vers, version, styles)

	v.SetDirection(tview.FlexRow)
	v.AddItem(logo, 10, 1, false)
	v.AddItem(vers, 1, 1, false)

	return &v
}

func (v *splashView) layoutLogo(t *tview.TextView, styles *config.Styles) {
	logo := strings.Join(Logo, fmt.Sprintf("\n[%s::b]", styles.Style.LogoColor))
	fmt.Fprintf(t, "%s[%s::b]%s\n",
		strings.Repeat("\n", 2),
		styles.Style.LogoColor,
		logo)
}

func (v *splashView) layoutRev(t *tview.TextView, rev string, styles *config.Styles) {
	fmt.Fprintf(t, "[%s::b]Revision [red::b]%s", styles.Style.FgColor, rev)
}
