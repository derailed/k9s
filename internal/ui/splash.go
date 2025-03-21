// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

// LogoSmall K9s small log.
var LogoSmall []string


// LogoBig K9s big logo for splash page.
var LogoBig = []string{
	` ____  __ ________        _______  ____     ___ `,
	`|    |/  /   __   \______/   ___ \|    |   |   |`,
	`|       /\____    /  ___/    \  \/|    |   |   |`,
	`|    \   \  /    /\___  \     \___|    |___|   |`,
	`|____|\__ \/____//____  /\______  /_______ \___|`,
	`         \/           \/        \/        \/    `,
}

// Splash represents a splash screen.
type Splash struct {
	*tview.Flex
}

// NewSplash instantiates a new splash screen with product and company info.
func NewSplash(styles *config.Styles, version string) *Splash {
	s := Splash{Flex: tview.NewFlex()}
	s.SetBackgroundColor(styles.BgColor())

	logo := tview.NewTextView()
	logo.SetDynamicColors(true)
	logo.SetTextAlign(tview.AlignCenter)
	s.layoutLogo(logo, styles)

	vers := tview.NewTextView()
	vers.SetDynamicColors(true)
	vers.SetTextAlign(tview.AlignCenter)
	s.layoutRev(vers, version, styles)

	s.SetDirection(tview.FlexRow)
	s.AddItem(logo, 10, 1, false)
	s.AddItem(vers, 1, 1, false)

	return &s
}

func (s *Splash) layoutLogo(t *tview.TextView, styles *config.Styles) {
	logo := strings.Join(LogoBig, fmt.Sprintf("\n[%s::b]", styles.Body().LogoColor))
	fmt.Fprintf(t, "%s[%s::b]%s\n",
		strings.Repeat("\n", 2),
		styles.Body().LogoColor,
		logo)
}

func (s *Splash) layoutRev(t *tview.TextView, rev string, styles *config.Styles) {
	fmt.Fprintf(t, "[%s::b]Revision [red::b]%s", styles.Body().FgColor, rev)
}

// function to get the logo []string from the LogoUrl
// by making a request to the LogoUrl
func GetLogo(logoUrl string) {
	slog.Debug("Fetching logo from URL", slog.String("url", logoUrl))
    defaultLogo := []string{
        ` ____  __ ________       `,
        `|    |/  /   __   \______`,
        `|       /\____    /  ___/`,
        `|    \   \  /    /\___  \`,
        `|____|\__ \/____//____  /`,
        `         \/           \/ `,
    }

    if logoUrl == "" {
        LogoSmall = defaultLogo
        return
    }

    if strings.HasPrefix(logoUrl, "file:") {
        filePath := strings.TrimPrefix(logoUrl, "file:")
        body, err := os.ReadFile(filePath)
        if err != nil {
            slog.Error("Error reading logo from file", slog.String("file", filePath), slog.Any("error", err))
            LogoSmall = defaultLogo
            return
        }
        logo := strings.Split(string(body), "\n")
        // last line is empty, remove it
        if len(logo) > 0 && logo[len(logo)-1] == "" {
            logo = logo[:len(logo)-1]
        }
        slog.Debug("Successfully fetched logo from file", slog.String("file", filePath))
        LogoSmall = logo
        return
    }

    resp, err := http.Get(logoUrl)
    if err != nil {
        slog.Error("Error fetching logo from URL", slog.String("url", logoUrl), slog.Any("error", err))
        LogoSmall = defaultLogo
        return
    }
    defer func() {
        if resp != nil {
            resp.Body.Close()
        }
    }()

    if resp.StatusCode != http.StatusOK {
        slog.Error("Non-OK HTTP status", slog.String("status", resp.Status))
        LogoSmall = defaultLogo
        return
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        slog.Error("Error reading response body", slog.Any("error", err))
        LogoSmall = defaultLogo
        return
    }

    logo := strings.Split(string(body), "\n")
    // last line is empty, remove it
    if len(logo) > 0 && logo[len(logo)-1] == "" {
        logo = logo[:len(logo)-1]
    }
    slog.Debug("Successfully fetched logo from URL", slog.String("url", logoUrl))
    LogoSmall = logo
}

