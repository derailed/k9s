// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
)

const (
	clipboardModeEnv = "K9S_CLIPBOARD"
	osc52MaxEnv      = "K9S_OSC52_MAX"

	clipboardModeAuto   = "auto"
	clipboardModeNative = "native"
	clipboardModeOSC52  = "osc52"

	defaultOSC52MaxEncodedLen = 74994
)

func clipboardWrite(text string) error {
	if text == "" {
		return nil
	}

	switch clipboardMode() {
	case clipboardModeNative:
		return clipboard.WriteAll(text)
	case clipboardModeOSC52:
		return writeOSC52(text)
	default:
		err := clipboard.WriteAll(text)
		if err == nil {
			return nil
		}
		if !canTryOSC52() {
			return err
		}
		if oscErr := writeOSC52(text); oscErr == nil {
			return nil
		}

		return err
	}
}

func clipboardMode() string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(clipboardModeEnv)))
	switch mode {
	case "", clipboardModeAuto:
		return clipboardModeAuto
	case clipboardModeNative:
		return clipboardModeNative
	case clipboardModeOSC52:
		return clipboardModeOSC52
	default:
		return clipboardModeAuto
	}
}

func canTryOSC52() bool {
	if !isTTY(os.Stdout) {
		return false
	}
	return strings.ToLower(strings.TrimSpace(os.Getenv("TERM"))) != "dumb"
}

func isTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}

func writeOSC52(text string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	maxLen := osc52MaxEncodedLen()
	if len(encoded) > maxLen {
		return fmt.Errorf("osc52 payload exceeds encoded size limit (%d > %d)", len(encoded), maxLen)
	}

	seq := osc52Sequence(encoded, os.Getenv("TMUX") != "", strings.HasPrefix(strings.ToLower(os.Getenv("TERM")), "screen"))
	_, err := io.WriteString(os.Stdout, seq)

	return err
}

func osc52MaxEncodedLen() int {
	v := strings.TrimSpace(os.Getenv(osc52MaxEnv))
	if v == "" {
		return defaultOSC52MaxEncodedLen
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultOSC52MaxEncodedLen
	}

	return n
}

func osc52Sequence(encoded string, tmux, screen bool) string {
	switch {
	case tmux:
		// tmux DCS passthrough — requires allow-passthrough in tmux.conf.
		return "\033Ptmux;\033\033]52;c;" + encoded + "\a\033\\"
	case screen:
		// GNU screen DCS passthrough.
		return "\033P\033]52;c;" + encoded + "\a\033\\"
	default:
		return "\033]52;c;" + encoded + "\a"
	}
}
