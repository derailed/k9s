package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

type logoView struct {
	*tview.Flex
	logo, status *tview.TextView
	app          *appView
}

func newLogoView(app *appView) *logoView {
	v := logoView{
		Flex:   tview.NewFlex(),
		logo:   logo(),
		status: status(),
		app:    app,
	}
	v.SetDirection(tview.FlexRow)
	v.AddItem(v.logo, 0, 6, false)
	v.AddItem(v.status, 0, 1, false)
	v.refreshLogo(app.styles.Style.LogoColor)

	return &v
}

func (v *logoView) reset() {
	v.status.Clear()
	v.status.SetBackgroundColor(v.app.styles.BgColor())
	v.refreshLogo(v.app.styles.Style.LogoColor)
}

func (v *logoView) warn(msg string) {
	v.update(msg, "red")
}

func (v *logoView) info(msg string) {
	v.update(msg, "green")
}

func (v *logoView) update(msg, c string) {
	v.refreshStatus(msg, c)
	v.refreshLogo(c)
}

func (v *logoView) refreshStatus(msg, c string) {
	v.status.SetBackgroundColor(config.AsColor(c))
	v.status.SetText(fmt.Sprintf("[white::b]%s", msg))
}

func (v *logoView) refreshLogo(c string) {
	v.logo.Clear()
	for i, s := range LogoSmall {
		fmt.Fprintf(v.logo, "[%s::b]%s", c, s)
		if i+1 < len(LogoSmall) {
			fmt.Fprintf(v.logo, "\n")
		}
	}
}

func logo() *tview.TextView {
	v := tview.NewTextView()
	v.SetWordWrap(false)
	v.SetWrap(false)
	v.SetTextAlign(tview.AlignLeft)
	v.SetDynamicColors(true)

	return v
}

func status() *tview.TextView {
	v := tview.NewTextView()
	v.SetWordWrap(false)
	v.SetWrap(false)
	v.SetTextAlign(tview.AlignCenter)
	v.SetDynamicColors(true)

	return v
}
