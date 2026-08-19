#!/bin/bash
# Render what the splash will look like, without rebooting.
#
# Plymouth can only be seen on a real boot, so this mirrors the layout maths in
# ../nejen/nejen.script with ImageMagick and writes two frames: the boot state
# (wordmark plus progress line) and the passphrase state (wordmark plus input
# panel). Keep it in step with nejen.script when the layout changes -- it is a
# preview, not the source of truth.
#
#   ./preview.sh [width] [height] [outdir]

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"
A=../nejen
W=${1:-3840}
H=${2:-2400}
OUT=${3:-/tmp/nejen-splash-preview}
mkdir -p "$OUT"

PROMPT_FONT=$(fc-match -f %{file} "CaskaydiaMono Nerd Font")
MUTED='#767ca0'

CX=$((W / 2))
CY=$((H / 2))
i() { magick identify -format "$2" "$1"; }

# --- geometry, straight out of nejen.script ---------------------------------
AURA=$((W * 62 / 100))
LOGO_W=$((W * 34 / 100))
LOGO_H=$((LOGO_W * $(i $A/logo.png %h) / $(i $A/logo.png %w)))
LOGO_X=$((CX - LOGO_W / 2))
LOGO_Y=$((CY - LOGO_H + H * 5 / 100))
CARD_BOTTOM=$((LOGO_Y + LOGO_H - LOGO_H * 145 / 1000))

LINE_W=$((W * 16 / 100))
LINE_H=$((H * 22 / 10000)); [ "$LINE_H" -lt 2 ] && LINE_H=2
LINE_X=$((CX - LINE_W / 2))
LINE_Y=$((CARD_BOTTOM + H * 75 / 1000))
HEAD=$((H * 18 / 1000))

FIELD_W=$((W * 30 / 100))
FIELD_H=$((FIELD_W * $(i $A/field.png %h) / $(i $A/field.png %w)))
FIELD_X=$((CX - FIELD_W / 2))
FIELD_Y=$((CARD_BOTTOM + H * 55 / 1000))
FIELD_CY=$((FIELD_Y + FIELD_H / 2))

BULLET=$((FIELD_H * 22 / 100))
GAP=$((BULLET * 165 / 100))
CARET_H=$((FIELD_H * 34 / 100))
FS=$((H * 185 / 10000))

base() {
  magick -size "${W}x${H}" gradient:'#08080a-#0a0a12' \
    \( "$A/aura.png" -resize "${AURA}x${AURA}!" \) \
      -geometry "+$((CX - AURA / 2))+$((CY - AURA / 2))" -composite \
    \( "$A/logo.png" -resize "${LOGO_W}x" \) \
      -geometry "+${LOGO_X}+${LOGO_Y}" -composite \
    "$1"
}

# --- boot frame: progress line at 62% ---------------------------------------
FILLED=$((LINE_W * 62 / 100))
base "$OUT/base.png"
magick "$OUT/base.png" \
  \( "$A/track.png" -resize "${LINE_W}x${LINE_H}!" -channel A -evaluate multiply 0.45 +channel \) \
    -geometry "+${LINE_X}+${LINE_Y}" -composite \
  \( "$A/fill.png" -resize "${FILLED}x${LINE_H}!" \) \
    -geometry "+${LINE_X}+${LINE_Y}" -composite \
  \( "$A/head.png" -resize "${HEAD}x${HEAD}!" \) \
    -geometry "+$((LINE_X + FILLED - HEAD / 2))+$((LINE_Y + LINE_H / 2 - HEAD / 2))" -composite \
  -resize 1600x "$OUT/boot.png"

# --- passphrase frame: five bullets and a caret ------------------------------
DOTS=5
ROW=$((DOTS * GAP - (GAP - BULLET)))
START=$((CX - ROW / 2))
CMD=(magick "$OUT/base.png"
  \( "$A/field.png" -resize "${FIELD_W}x" \) -geometry "+${FIELD_X}+${FIELD_Y}" -composite)
for n in $(seq 0 $((DOTS - 1))); do
  CMD+=(\( "$A/bullet.png" -resize "${BULLET}x${BULLET}!" \)
        -geometry "+$((START + n * GAP))+$((FIELD_CY - BULLET / 2))" -composite)
done
CMD+=(\( "$A/caret.png" -resize "x${CARET_H}" \)
      -geometry "+$((START + ROW + BULLET * 6 / 10))+$((FIELD_CY - CARET_H / 2))" -composite)
CMD+=(\( -background none -fill "$MUTED" -font "$PROMPT_FONT" -pointsize "$FS"
         label:'Please enter passphrase for disk root:' \)
      -gravity north -geometry "+0+$((FIELD_Y - FS * 3 / 2))" -composite)
CMD+=(-resize 1600x "$OUT/passphrase.png")
"${CMD[@]}"

rm -f "$OUT/base.png"
echo "$OUT/boot.png"
echo "$OUT/passphrase.png"
