#!/bin/bash
# Regenerate the PNG assets in ../nejen/ from the palette below.
#
# The splash is composed from pre-rendered art rather than Image.Text because
# inside the initramfs Plymouth draws labels with the freetype backend against
# one embedded font -- family, weight and letterspacing are ignored there.
# Baking the wordmark and the panel chrome into PNGs is the only way to get
# the waybar look onto the LUKS passphrase prompt.
#
# The PNGs are committed, so this only needs to run when the mark or the
# palette (themes/nejen/theme.toml) changes.
#
# Needs: imagemagick, noto-fonts.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"
out=../nejen
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$out"

# Palette (themes/nejen/theme.toml) and the waybar/walker chrome built from it:
#   panel  = background at 65%     border = border at 35-60%
#   glow   = border, blurred       text   = foreground
ACCENT='#7f87df'
BORDER='#4a4f8a'
FOREGROUND='#c7d0ff'
PANEL='#0a0a12'

WORDMARK_FONT=/usr/share/fonts/noto/NotoSans-Black.ttf

# Everything is authored at this width and scaled to a fraction of the screen
# by nejen.script, so the proportions hold from 1080p to 4K.
CARD_W=1360           # wordmark card, 2.34:1 like the mark it is drawn from
CARD_H=580
CARD_R=32
FIELD_H=180           # passphrase field, same width so the two stay aligned
FIELD_R=26
MARGIN=120            # transparent padding the outer glow bleeds into
TRACKING=44           # letterspacing of the wordmark, in authoring pixels

# ---- panel ------------------------------------------------------------------
# A waybar module blown up to poster size: 65% panel fill, a hairline slate
# border, and the soft bloom that stands in for the CSS box-shadow.
#
#   panel <width> <height> <radius> <border-alpha> <out>
panel() {
  local w=$1 h=$2 r=$3 ba=$4 dst=$5
  local cw=$((w + 2 * MARGIN)) ch=$((h + 2 * MARGIN))
  local rect="roundrectangle $MARGIN,$MARGIN $((MARGIN + w)),$((MARGIN + h)) $r,$r"

  magick -size "${cw}x${ch}" xc:none -fill "$BORDER" -draw "$rect" \
    -blur 0x44 -channel A -evaluate multiply 0.40 +channel MIFF:- |
  magick MIFF:- \
    \( -size "${cw}x${ch}" xc:none -fill "$PANEL" -draw "$rect" \
       -channel A -evaluate multiply 0.65 +channel \) -composite \
    \( -size "${cw}x${ch}" xc:none -fill none -stroke "$BORDER" -strokewidth 14 \
       -draw "$rect" -blur 0x12 -channel A -evaluate multiply 0.30 +channel \) -composite \
    \( -size "${cw}x${ch}" xc:none -fill none -stroke "$BORDER" -strokewidth 5 \
       -draw "$rect" -channel A -evaluate multiply "$ba" +channel \) -composite \
    "$dst"
}

# ---- wordmark ---------------------------------------------------------------
# Set letter by letter rather than as one label: Noto Sans Black gives the
# weight of the mark, but its J drops below the baseline and the mark's does
# not, so the J is normalised to cap height before the glyphs are spaced.
wordmark() {
  local dst=$1 i=0 g cap
  for g in N E J E N; do
    magick -background none -fill "$FOREGROUND" -font "$WORDMARK_FONT" \
      -pointsize 240 label:"$g" -trim +repage "$tmp/g$i.png"
    i=$((i + 1))
  done
  cap=$(magick identify -format %h "$tmp/g0.png")
  magick "$tmp/g2.png" -resize "x${cap}!" "$tmp/g2.png"
  magick "$tmp"/g0.png "$tmp"/g1.png "$tmp"/g2.png "$tmp"/g3.png "$tmp"/g4.png \
    -background none -gravity south +smush "$TRACKING" +repage "$dst"
}

# ---- logo: the wordmark card ------------------------------------------------

CANVAS_W=$((CARD_W + 2 * MARGIN))
CANVAS_H=$((CARD_H + 2 * MARGIN))

panel "$CARD_W" "$CARD_H" "$CARD_R" 0.55 "$tmp/card.png"
wordmark "$tmp/word-raw.png"
magick "$tmp/word-raw.png" -resize "$((CARD_W * 76 / 100))x" "$tmp/word.png"

# Faint dot field inside the card, clipped to the rounded silhouette.
magick -size 76x76 xc:none -fill "$ACCENT" -draw 'circle 6,6 6,9' "$tmp/dot.png"
magick -size "${CANVAS_W}x${CANVAS_H}" "tile:$tmp/dot.png" \
  -channel A -evaluate multiply 0.10 +channel "$tmp/dots.png"

magick "$tmp/card.png" \
  \( "$tmp/dots.png" "$tmp/card.png" -alpha extract -compose CopyOpacity -composite \) \
  -compose over -composite \
  \( "$tmp/word.png" -channel A -blur 0x14 -evaluate multiply 0.40 +channel \
     -fill "$ACCENT" -colorize 100 \) -gravity center -composite \
  "$tmp/word.png" -gravity center -composite \
  "$out/logo.png"

# ---- passphrase field -------------------------------------------------------
# Same chrome, brighter border: this is the one element asking for input.

panel "$CARD_W" "$FIELD_H" "$FIELD_R" 0.70 "$out/field.png"

# ---- progress line ----------------------------------------------------------

magick -size 8x8 xc:"$BORDER" "$out/track.png"
magick -size 8x8 xc:"$ACCENT" "$out/fill.png"
magick -size 128x128 radial-gradient:"$ACCENT"-none \
  -channel A -evaluate multiply 0.8 +channel "$out/head.png"

# ---- passphrase bullet and caret --------------------------------------------

magick -size 64x64 xc:none -fill "$ACCENT" -draw 'circle 32,32 32,43' \
  \( +clone -blur 0x9 -channel A -evaluate multiply 0.7 +channel \) \
  -reverse -background none -compose over -flatten "$out/bullet.png"

magick -size 32x160 xc:none -fill "$ACCENT" -draw 'rectangle 12,0 19,159' \
  \( +clone -blur 0x7 -channel A -evaluate multiply 0.6 +channel \) \
  -reverse -background none -compose over -flatten "$out/caret.png"

# ---- scene aura -------------------------------------------------------------
# waybar's box-shadow at room scale: lifts the middle of the screen just
# enough that the translucent panels read as panels rather than as flat ink.

magick -size 1024x1024 radial-gradient:"$BORDER"-none \
  -channel A -evaluate multiply 0.22 +channel "$out/aura.png"

# Retired by the wordmark card; drop them so the initramfs stays small.
rm -f "$out/peak.png" "$out/wordmark.png"

# The splash lives in the initramfs -- 16-bit channels and metadata are pure
# weight there, so flatten every asset to plain 8-bit RGBA.
for f in "$out"/*.png; do
  magick "$f" -strip -depth 8 -define png:compression-level=9 "$f"
done

echo "assets rendered into $out/"
