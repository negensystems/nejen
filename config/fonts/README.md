# Fonts shipped with NEJEN

## NegenIcons.ttf

NEJEN's own icons, as a font, because the bar, notifications and the terminal
can only draw text. Family name `Negen Icons`.

The installer links it into `~/.local/share/fonts/`; the package installs it
under `/usr/share/fonts/nejen/`.

Every glyph is a character from U+100000 up, a Private Use plane no other
icon font touches, so anything that draws text finds it through ordinary font
fallback. Notifications and terminal output therefore carry just the
character. The bar also names the font, `<span font='Negen Icons'>…</span>`
around the character, so the icon never depends on fallback inside a line of
the bar's typeface.

Glyphs in use:

| Codepoint | Glyph | Where |
|---|---|---|
| U+100019 | ring with a small centre | unnamed workspace |
| U+100039 | arrow leaving a frame | tray expander |
| U+10003D | two chasing arrows | update available; "Update System" notice |
| U+100044 | bolt | battery while charging (bar, screensaver, `nejen battery status`); "Plug in soon" notice |
| U+10005B, U+10005C, U+10005D, U+10005E | battery: empty, one, two, three bars | battery level (bar, screensaver, `nejen battery status`) |
| U+10005F, U+100060 | play, pause | media state |
| U+100063 | sand timer | timer; voice typing, transcribing |
| U+100064, U+100065, U+100066 | bluetooth: on, connected, off | bluetooth |
| U+100067, U+100068, U+100069 | speaker: no wave, one wave, two waves | volume level |
| U+10006A | speaker with a cross | muted |
| U+10006B | headphones | volume, on headphones or a headset |
| U+10006C | two beamed notes | media, unknown player |
| U+10006D | microphone | voice typing, recording |
| U+10006E, U+10006F, U+100070 | wifi: one, two, three arcs | wifi signal; "Set Up Wi-Fi" notice |
| U+100071 | wifi struck through | offline |
| U+100072 | network port | wired link |
| U+100077 | ring around a solid centre | screen recording indicator |
| U+100059 | coffee cup | idle locking disabled: indicator and notice |
| U+100037 | padlock | "Idle locking enabled" notice |
| U+100074, U+100073 | bell struck through, bell | do not disturb: indicator and on/off notices |
| U+100076, U+100075 | keyboard struck through, keyboard | keyboard cleaning mode: indicator, overlay and notices |
| U+10000D, U+10000C | moon, sun | night light on/off notices |
| U+100055 | monitor | display scale notice |
| U+10002D | four tiles | workspace layout notice |
| U+10001B | box | square aspect notices |
| U+100078 | rocket | first-login welcome |
| U+100023 | clock | "Clock Not Syncing" notice |

Where they are set:

- `config/waybar/config.jsonc`: the bar modules.
- `cmd/nejen/icons.go`: named constants for everything the dispatcher prints,
  which is the bar indicators, the notifications, the screensaver's battery
  line and `nejen battery status`.
- `install/first-run/*.sh`: the first-login notices.

Brand logos on the bar (Arch, Spotify, Firefox, Chromium) are still Nerd Font
glyphs: a brand is drawn its owner's way.

The font holds the whole Negen icon set (121 glyphs). It is built outside
this repo, from the Negen icon set's `glyphs.json`, and copied here; a
codepoint is never reassigned between builds.
