package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func luaQuote(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if c < 0x20 || c == 0x7f {
				fmt.Fprintf(&b, `\x%02x`, c)
			} else {
				b.WriteByte(c)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func luaNumberOrQuote(s string) string {
	trimmed := strings.TrimSpace(s)
	if _, err := strconv.Atoi(trimmed); err == nil {
		return trimmed
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil && !strings.ContainsAny(trimmed, "nNiI") {
		_ = f
		return trimmed
	}
	return luaQuote(s)
}

func hyprEval(code string) error {
	return exec.Command("hyprctl", "-q", "eval", code).Run()
}

func hyprDispatchLua(dspExpr string) error {
	return hyprEval("hl.dispatch(" + dspExpr + ")")
}

func hyprFocusWindowExec(addr string) error {
	hyprctlPath, err := exec.LookPath("hyprctl")
	if err != nil {
		return err
	}
	code := fmt.Sprintf("hl.dispatch(hl.dsp.focus({ window = %s }))", luaQuote("address:"+addr))
	return syscall.Exec(hyprctlPath, []string{"hyprctl", "eval", code}, os.Environ())
}
