package ui_test

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestFlash(t *testing.T) {
	const delay = 1 * time.Millisecond
	uu := map[string]struct {
		l    model.FlashLevel
		i, e string
	}{
		"info": {l: model.FlashInfo, i: "hello", e: "😎 hello\n"},
		"warn": {l: model.FlashWarn, i: "hello", e: "😗 hello\n"},
		"err":  {l: model.FlashErr, i: "hello", e: "😡 hello\n"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := ui.NewApp("test")
	f := ui.NewFlash(a)
	f.SetTestMode(true)
	go f.Watch(ctx, a.Flash().Channel())

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			a.Flash().SetMessage(u.l, u.i)
			time.Sleep(delay)
			assert.Equal(t, u.e, f.GetText(false))
		})
	}
}
