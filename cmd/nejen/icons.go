package main

// NEJEN's icons are characters in its own font, Negen Icons
// (config/fonts/NegenIcons.ttf), from U+100000 up: a Private Use plane no
// other icon font touches. Anything that draws text finds them through font fallback, so a
// notification or a terminal line needs only the character. Waybar mixes them
// into a line of the bar's typeface, so there the font is named outright with
// barIcon. config/fonts/README.md says what each one looks like.
const (
	iconSun         = "\U0010000c"
	iconMoon        = "\U0010000d"
	iconSquare      = "\U0010001b"
	iconLayout      = "\U0010002d"
	iconLock        = "\U00100037"
	iconBolt        = "\U00100044"
	iconMonitor     = "\U00100055"
	iconCoffee      = "\U00100059"
	iconBell        = "\U00100073"
	iconBellOff     = "\U00100074"
	iconKeyboard    = "\U00100075"
	iconKeyboardOff = "\U00100076"
	iconRecord      = "\U00100077"
)

// batteryIcon is the battery glyph for a charge level: empty below 10%, then
// one, two and three bars from 10%, 40% and 70%, the same steps the bar uses.
func batteryIcon(percent int) string {
	switch {
	case percent < 10:
		return "\U0010005b"
	case percent < 40:
		return "\U0010005c"
	case percent < 70:
		return "\U0010005d"
	}
	return "\U0010005e"
}

// barIcon wraps an icon in the Pango markup waybar needs to draw it from
// Negen Icons rather than from the bar's typeface.
func barIcon(glyph string) string {
	return "<span font='Negen Icons'>" + glyph + "</span>"
}
