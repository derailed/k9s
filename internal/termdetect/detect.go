// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package termdetect

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/term"
)

const osc11QueryTimeout = 200 * time.Millisecond

func IsBackgroundLight() (bool, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false, fmt.Errorf("open /dev/tty: %w", err)
	}
	defer tty.Close()

	fd := int(tty.Fd())
	oldState, err := term.GetState(fd)
	if err != nil {
		return false, fmt.Errorf("get terminal state: %w", err)
	}
	defer term.Restore(fd, oldState)

	if _, err := term.MakeRaw(fd); err != nil {
		return false, fmt.Errorf("make raw: %w", err)
	}

	if _, err := tty.Write([]byte("\033]11;?\033\\")); err != nil {
		return false, fmt.Errorf("write osc 11 query: %w", err)
	}

	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		buf := make([]byte, 128)
		n, err := tty.Read(buf)
		ch <- result{buf[:n], err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			return false, fmt.Errorf("read osc 11 response: %w", r.err)
		}
		col, err := parseOSC11(r.data)
		if err != nil {
			return false, err
		}
		return luminance(col) > 0.5, nil
	case <-time.After(osc11QueryTimeout):
		return false, fmt.Errorf("osc 11 query timed out")
	}
}

func parseOSC11(data []byte) (colorful.Color, error) {
	s := string(data)
	idx := strings.Index(s, "rgb:")
	if idx < 0 {
		return colorful.Color{}, fmt.Errorf("no rgb: prefix in osc 11 response")
	}
	rgb := s[idx+4:]
	for i, c := range rgb {
		if c == '\033' || c == '\a' {
			rgb = rgb[:i]
			break
		}
	}
	parts := strings.Split(rgb, "/")
	if len(parts) != 3 {
		return colorful.Color{}, fmt.Errorf("expected 3 rgb components, got %d", len(parts))
	}
	r, err := parseComp(parts[0])
	if err != nil {
		return colorful.Color{}, fmt.Errorf("parse red: %w", err)
	}
	g, err := parseComp(parts[1])
	if err != nil {
		return colorful.Color{}, fmt.Errorf("parse green: %w", err)
	}
	b, err := parseComp(parts[2])
	if err != nil {
		return colorful.Color{}, fmt.Errorf("parse blue: %w", err)
	}
	return colorful.Color{R: r, G: g, B: b}, nil
}

func parseComp(s string) (float64, error) {
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hex %q: %w", s, err)
	}
	maxVal := float64(uint64(1)<<(4*uint(len(s))) - 1)
	return float64(v) / maxVal, nil
}

func luminance(c colorful.Color) float64 {
	return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
}
