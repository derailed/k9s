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
	"github.com/derailed/tview"
)

// Context key types to avoid collisions
type contextKey string

const (
	urlKey    contextKey = "url"
	fileKey   contextKey = "file"
	errorKey  contextKey = "error"
	statusKey contextKey = "status"
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
	ctx := context.WithValue(context.Background(), urlKey, logoUrl)
	slog.DebugContext(ctx, "Fetching logo from URL")
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
			ctx := context.WithValue(context.Background(), fileKey, filePath)
			ctx = context.WithValue(ctx, errorKey, err)
			slog.ErrorContext(ctx, "Error reading logo from file")
			LogoSmall = defaultLogo
			return
		}
		ctx := context.WithValue(context.Background(), fileKey, filePath)
		slog.DebugContext(ctx, "Successfully fetched logo from file")

		logo := strings.Split(string(body), "\n")
		// last line is empty, remove it
		if len(logo) > 0 && logo[len(logo)-1] == "" {
			logo = logo[:len(logo)-1]
		}
		LogoSmall = logo
		return
	}

	// #nosec G107 - URL is validated and controlled by configuration
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(logoUrl)
	if err != nil {
		ctx := context.WithValue(context.Background(), urlKey, logoUrl)
		ctx = context.WithValue(ctx, errorKey, err)
		slog.ErrorContext(ctx, "Error fetching logo from URL")
		LogoSmall = defaultLogo
		return
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		ctx := context.WithValue(context.Background(), statusKey, resp.Status)
		slog.ErrorContext(ctx, "Non-OK HTTP status")
		LogoSmall = defaultLogo
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx := context.WithValue(context.Background(), errorKey, err)
		slog.ErrorContext(ctx, "Error reading response body")
		LogoSmall = defaultLogo
		return
	}

	ctx = context.WithValue(context.Background(), urlKey, logoUrl)
	slog.DebugContext(ctx, "Successfully fetched logo from URL")

	logo := strings.Split(string(body), "\n")
	// last line is empty, remove it
	if len(logo) > 0 && logo[len(logo)-1] == "" {
		logo = logo[:len(logo)-1]
	}
	LogoSmall = logo
}
