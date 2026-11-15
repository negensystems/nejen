package main

import "testing"

func TestLuaQuote(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"plain", `"plain"`},
		{`with "quotes" and \slash`, `"with \"quotes\" and \\slash"`},
		{"line1\nline2\t!", `"line1\nline2\t!"`},
		{"ctrl\x07byte", `"ctrl\x07byte"`},
	}
	for _, c := range cases {
		if got := luaQuote(c.in); got != c.want {
			t.Errorf("luaQuote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLuaNumberOrQuote(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1300", "1300"},
		{"-60", "-60"},
		{"1.5", "1.5"},
		{"50%", `"50%"`},
		{"NaN", `"NaN"`},
		{"inf", `"inf"`},
	}
	for _, c := range cases {
		if got := luaNumberOrQuote(c.in); got != c.want {
			t.Errorf("luaNumberOrQuote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
