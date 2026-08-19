package main

import "testing"

func themeData() map[string]interface{} {
	return map[string]interface{}{
		"colors": map[string]interface{}{
			"bg":      "#0a0a12",
			"indigo":  "#4a4f8a",
			"nested":  map[string]interface{}{"deep": "#ffffff"},
			"notacol": "plain",
		},
		"font":    map[string]interface{}{"size": int64(11), "name": "CaskaydiaMono"},
		"opacity": map[string]interface{}{"bar": 0.65},
	}
}

func TestResolveValues(t *testing.T) {
	d := themeData()
	cases := []struct{ expr, want string }{
		{"colors.bg", "#0a0a12"},
		{"colors.nested.deep", "#ffffff"},
		{"font.size", "11"},
		{"font.name", "CaskaydiaMono"},
		{"opacity.bar", "0.65"},
		// A transform consumes the trailing segment.
		{"colors.bg.nohash", "0a0a12"},
	}
	for _, c := range cases {
		got, err := resolve(d, c.expr, nil)
		if err != nil {
			t.Errorf("resolve(%q) error: %v", c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolve(%q) = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestResolveErrors(t *testing.T) {
	d := themeData()
	cases := []struct{ name, expr string }{
		{"missing key", "colors.nope"},
		{"missing root", "nope.bg"},
		{"table not value", "colors.nested"},
		// nohash on a table must fail rather than stringify the map.
		{"transform on table", "colors.nohash"},
		{"two unknown segments", "colors.a.b"},
		// nohash requires a #rrggbb value.
		{"nohash on non-color", "colors.notacol.nohash"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, err := resolve(d, c.expr, nil); err == nil {
				t.Errorf("resolve(%q) = %q, want error", c.expr, got)
			}
		})
	}
}

func TestResolveRejectsUnwantedArgument(t *testing.T) {
	d := themeData()
	arg := "x"
	if _, err := resolve(d, "colors.bg", &arg); err == nil {
		t.Error("plain value should reject an argument")
	}
	if _, err := resolve(d, "colors.bg.nohash", &arg); err == nil {
		t.Error("nohash should reject an argument")
	}
}

func TestRequireHex(t *testing.T) {
	for _, ok := range []string{"#0a0a12", "#FFFFFF", "#4a4f8a"} {
		if _, err := requireHex(ok, "x"); err != nil {
			t.Errorf("requireHex(%q) unexpected error: %v", ok, err)
		}
	}
	// Short form, missing hash, alpha channel and non-hex must all be rejected:
	// they would reach the rendered config and break the consuming program.
	for _, bad := range []string{"#fff", "0a0a12", "#0a0a12ff", "#gggggg", "", "red"} {
		if _, err := requireHex(bad, "x"); err == nil {
			t.Errorf("requireHex(%q) should have failed", bad)
		}
	}
}

func TestPlaceholderRegex(t *testing.T) {
	// A single segment is deliberately not a placeholder, so templates can
	// carry ${VAR} for other tools untouched.
	for _, s := range []string{"${bg}", "${PATH}", "$bg", "${}"} {
		if placeholderRegex.MatchString(s) {
			t.Errorf("%q should not be treated as a placeholder", s)
		}
	}
	for _, s := range []string{"${colors.bg}", "${colors.bg.nohash}", "${a.b:arg}", "${a.b:}"} {
		if !placeholderRegex.MatchString(s) {
			t.Errorf("%q should be a placeholder", s)
		}
	}
}
