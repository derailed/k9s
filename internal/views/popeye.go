package views

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"os"
// 	"path/filepath"

// 	"github.com/derailed/k9s/internal/config"
// 	"github.com/derailed/popeye/pkg"
// 	cfg "github.com/derailed/popeye/pkg/config"
// 	"github.com/derailed/tview"
// 	"github.com/gdamore/tcell"
// 	"github.com/rs/zerolog/log"
// )

// type popeyeView struct {
// 	*detailsView

// 	current    igniter
// 	ansiWriter io.Writer
// }

// func newPopeyeView(app *appView) *popeyeView {
// 	v := popeyeView{}
// 	{
// 		v.detailsView = newDetailsView(app, v.backCmd)
// 		v.SetBorderPadding(0, 0, 1, 1)
// 		v.current = app.content.GetPrimitive("main").(igniter)
// 		v.SetDynamicColors(true)
// 		v.SetWrap(true)
// 		v.setTitle("Popeye")
// 		v.ansiWriter = tview.ANSIWriter(v)
// 	}
// 	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)

// 	return &v
// }

// func (v *popeyeView) init(ctx context.Context, ns string) {
// 	defer func() {
// 		if err := recover(); err != nil {
// 			v.app.flash(flashErr, fmt.Sprintf("%v", err))
// 		}
// 	}()

// 	c := cfg.New()

// 	spinach := filepath.Join(config.K9sHome, "spinach.yml")

// 	if _, err := os.Stat(spinach); err == nil {
// 		c.Spinach = spinach
// 	}

// 	if v.app.config.K9s.CurrentContext != "" {
// 		v.app.flags.Context = &v.app.config.K9s.CurrentContext
// 	}

// 	if err := c.Init(v.app.flags); err != nil {
// 		log.Error().Err(err).Msg("Unable to load spinach config")
// 	}

// 	p := pkg.NewPopeye(c, &log.Logger, v.ansiWriter)
// 	p.Sanitize(false)
// }

// func (v *popeyeView) getTitle() string {
// 	return "Popeye"
// }

// func (v *popeyeView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	v.app.command.previousCmd()
// 	v.app.inject(v.current)

// 	return nil
// }
