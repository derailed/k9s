package ui

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
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

// SplashView represents a splash screen.
type SplashView struct {
	*tview.Flex
}

// NewSplash instantiates a new splash screen with product and company info.
func NewSplash(styles *config.Styles, version string) *SplashView {
	v := SplashView{Flex: tview.NewFlex()}

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

func (v *SplashView) layoutLogo(t *tview.TextView, styles *config.Styles) {
	logo := strings.Join(Logo, fmt.Sprintf("\n[%s::b]", styles.Body().LogoColor))
	fmt.Fprintf(t, "%s[%s::b]%s\n",
		strings.Repeat("\n", 2),
		styles.Body().LogoColor,
		logo)
}

func (v *SplashView) layoutRev(t *tview.TextView, rev string, styles *config.Styles) {
	fmt.Fprintf(t, "[%s::b]Revision [red::b]%s", styles.Body().FgColor, rev)
}
