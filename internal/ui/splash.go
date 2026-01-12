// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/slogs"
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

func (*Splash) layoutLogo(t *tview.TextView, styles *config.Styles) {
	logo := strings.Join(LogoBig, fmt.Sprintf("\n[%s::b]", styles.Body().LogoColor))
	_, _ = fmt.Fprintf(t, "%s[%s::b]%s\n",
		strings.Repeat("\n", 2),
		styles.Body().LogoColor,
		logo)
}

func (*Splash) layoutRev(t *tview.TextView, rev string, styles *config.Styles) {
	_, _ = fmt.Fprintf(t, "[%s::b]Revision [red::b]%s", styles.Body().FgColor, rev)
}

// GetLogo function to get the logo []string from the LogoUrl
// by making a request to the LogoUrl
func GetLogo(logoUrl string) {
	slog.Debug("Fetching logo from URL", slogs.URL, logoUrl)
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
			slog.Error("Error reading logo from file", slogs.Path, filePath, slogs.Error, err)
			LogoSmall = defaultLogo
			return
		}
		slog.Debug("Successfully fetched logo from file", slogs.Path, filePath)

		logo := strings.Split(string(body), "\n")
		// last line is empty, remove it
		if len(logo) > 0 && logo[len(logo)-1] == "" {
			logo = logo[:len(logo)-1]
		}
		LogoSmall = logo
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// #nosec G107 - URL is validated and controlled by configuration
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logoUrl, http.NoBody)
	if err != nil {
		slog.Error("Error creating request for logo URL", slogs.URL, logoUrl, slogs.Error, err)
		LogoSmall = defaultLogo
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error fetching logo from URL", slogs.URL, logoUrl, slogs.Error, err)
		LogoSmall = defaultLogo
		return
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Non-OK HTTP status", slogs.Message, resp.Status)
		LogoSmall = defaultLogo
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading response body", slogs.Error, err)
		LogoSmall = defaultLogo
		return
	}

	slog.Debug("Successfully fetched logo from URL", slogs.URL, logoUrl)

	logo := strings.Split(string(body), "\n")
	// last line is empty, remove it
	if len(logo) > 0 && logo[len(logo)-1] == "" {
		logo = logo[:len(logo)-1]
	}
	LogoSmall = logo
}
