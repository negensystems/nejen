#!/bin/bash
# Regenerate the PNG assets in ../nejen/ from the palette below.
#
# The splash is composed from pre-rendered art rather than Image.Text because
# inside the initramfs Plymouth draws labels with the freetype backend against
# one embedded font, so family, weight and letterspacing are ignored there.
# Baking the mark and the panel chrome into PNGs is the only way to get the
# waybar look onto the LUKS passphrase prompt.
#
# The mark is the block wordmark in ../../../logo.txt, the same art `nejen
# open about` and the screensaver draw, blown up with a nearest-neighbour
# scale so the cells stay square-edged at any size. Boot, lock and About
# therefore show one mark rather than three interpretations of it.
#
# The PNGs are committed, so this only needs to run when the mark or the
# palette (themes/nejen/theme.toml) changes.
#
# Needs: imagemagick.

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

MARK_SRC=../../../logo.txt

# Everything is authored at this width and scaled to a fraction of the screen
# by nejen.script, so the proportions hold from 1080p to 4K.
#
# The mark's cell is 40x64, the 1:1.6 ratio of a monospace terminal cell, so
# the blown-up art keeps the proportions the wordmark has in About.
CELL_W=40
CELL_H=64
FIELD_W=1360          # passphrase field
FIELD_H=180
FIELD_R=26
MARGIN=120            # transparent padding the outer glow bleeds into

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

# ---- mark -------------------------------------------------------------------
# logo.txt is a grid of full-block characters and spaces. Turn it into a 1-bit
# PBM at one pixel per cell, then blow that up with the point filter: nearest
# neighbour keeps every cell edge hard, so the mark stays the same shape it has
# in a terminal instead of acquiring a font's idea of the letters.
#
# In the PBM a set bit reads back as black, so the mask is negated before it
# becomes the alpha channel of a flat foreground fill.
blockmark() {
  local dst=$1 cols rows
  cols=$(awk 'NR==1 {print length($0)}' "$MARK_SRC")
  rows=$(grep -c . "$MARK_SRC")
  {
    echo P1
    echo "$cols $rows"
    sed 's/ /0 /g; s/\xe2\x96\x88/1 /g' "$MARK_SRC"
  } > "$tmp/mark.pbm"

  magick "$tmp/mark.pbm" \
    -filter point -resize "$((cols * CELL_W))x$((rows * CELL_H))!" \
    -negate -background "$FOREGROUND" -alpha shape "$dst"
}

# ---- logo: the bare mark ----------------------------------------------------
# No panel behind it. The mark carries its own accent bloom and sits straight
# on the aura, so the splash reads as the wordmark rather than as a badge.

blockmark "$tmp/ink.png"
MARK_W=$(magick identify -format %w "$tmp/ink.png")
MARK_H=$(magick identify -format %h "$tmp/ink.png")

bloom() {
  magick -size "$1x$2" xc:none \
    \( "$tmp/ink.png" -channel A -blur 0x22 -evaluate multiply 0.35 +channel \
       -fill "$ACCENT" -colorize 100 \) -gravity center -composite \
    "$tmp/ink.png" -gravity center -composite \
    "$3"
}

bloom $((MARK_W + 2 * MARGIN)) $((MARK_H + 2 * MARGIN)) "$out/logo.png"

# ---- mark for hyprlock ------------------------------------------------------
# The lock screen draws the same art through hyprlock's image widget, which
# lays an image out in a square box. Authoring it on a square canvas means the
# mark keeps its proportions there whatever the widget does with the box.
#
# The installer links this into ~/.config/nejen/branding/mark.png, beside the
# about.txt the same wordmark is drawn from.

bloom $((MARK_W + 2 * MARGIN)) $((MARK_W + 2 * MARGIN)) ../../../mark.png

# ---- passphrase field -------------------------------------------------------
# waybar's panel chrome, brighter border: this is the one element asking for
# input, and the only panel left in the scene.

panel "$FIELD_W" "$FIELD_H" "$FIELD_R" 0.70 "$out/field.png"

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

# Retired by earlier iterations of the mark; drop them so the initramfs
# stays small.
rm -f "$out/peak.png" "$out/wordmark.png"

# The splash lives in the initramfs, where 16-bit channels and metadata are pure
# weight there, so flatten every asset to plain 8-bit RGBA.
for f in "$out"/*.png ../../../mark.png; do
  magick "$f" -strip -depth 8 -define png:compression-level=9 "$f"
done

echo "assets rendered into $out/ (and mark.png at the repo root)"
