//go:build linux || darwin || windows

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	ansiReset   = "\x1b[0m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[1;31m"
	ansiGreen   = "\x1b[1;32m"
	ansiYellow  = "\x1b[1;33m"
	ansiBlue    = "\x1b[1;34m"
	ansiMagenta = "\x1b[1;35m"
	ansiCyan    = "\x1b[1;36m"
)

type printer struct {
	out            io.Writer
	color          bool
	restoreConsole func()
}

func newPrinter(out *os.File, mode string) (*printer, error) {
	var color bool
	var restore func()
	switch strings.ToLower(mode) {
	case "auto":
		_, noColor := os.LookupEnv("NO_COLOR")
		if !noColor && os.Getenv("TERM") != "dumb" {
			restore, color = prepareConsole(out)
		}
	case "always":
		restore, _ = prepareConsole(out)
		color = true
	case "never":
		color = false
	default:
		return nil, fmt.Errorf("invalid -color value %q; expected auto, always, or never", mode)
	}
	return &printer{out: out, color: color, restoreConsole: restore}, nil
}

func (p *printer) Close() {
	if p.restoreConsole != nil {
		p.restoreConsole()
	}
}

func (p *printer) lifecycle(at time.Time, action string, id string, title string) {
	style := ansiYellow
	switch action {
	case "START":
		style = ansiGreen
	case "ATTACH":
		style = ansiMagenta
	case "STOP":
		style = ansiRed
	}
	fmt.Fprintf(
		p.out,
		"%s  %s  %s  %q\n",
		p.paint(ansiDim, displayTime(at)),
		p.paint(style, fmt.Sprintf("%-6s", action)),
		p.paint(ansiCyan, id),
		sanitizeTitle(title),
	)
}

func (p *printer) tokens(at time.Time, id string, title string, count tokenCount) {
	reasoningStyle := ansiYellow
	if count.reasoning == 516 || count.reasoning == 1_034 || count.reasoning == 1_552 {
		reasoningStyle = ansiRed
	} else if count.reasoning > 1_000 {
		reasoningStyle = ansiGreen
	}
	fmt.Fprintf(
		p.out,
		"%s  %s  %s  %q  %s  %s\n",
		p.paint(ansiDim, displayTime(at)),
		p.paint(ansiBlue, "TOKENS"),
		p.paint(ansiCyan, id),
		sanitizeTitle(title),
		p.paint(ansiCyan, "output="+formatUint(count.output)),
		p.paint(reasoningStyle, "reasoning="+formatUint(count.reasoning)),
	)
}

func (p *printer) paint(style string, text string) string {
	if !p.color {
		return text
	}
	return style + text + ansiReset
}

func displayTime(at time.Time) string {
	if at.IsZero() {
		at = time.Now()
	}
	return at.Local().Format("2006-01-02 15:04:05.000")
}

func sanitizeTitle(title string) string {
	title = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, title))
	if title == "" {
		return "(untitled)"
	}
	runes := []rune(title)
	if len(runes) > 120 {
		title = string(runes[:119]) + "…"
	}
	return title
}

func formatUint(value uint64) string {
	digits := strconv.FormatUint(value, 10)
	if len(digits) <= 3 {
		return digits
	}
	first := len(digits) % 3
	if first == 0 {
		first = 3
	}
	var formatted strings.Builder
	formatted.Grow(len(digits) + len(digits)/3)
	formatted.WriteString(digits[:first])
	for offset := first; offset < len(digits); offset += 3 {
		formatted.WriteByte(',')
		formatted.WriteString(digits[offset : offset+3])
	}
	return formatted.String()
}
