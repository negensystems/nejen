# Theme engine integration: required config repointing

The theme engine renders `templates/*.tmpl` with `themes/<name>/theme.toml`
into `~/.local/state/nejen/theme/` (`nejen theme render`), and `nejen theme
set <name>` additionally writes `~/.local/state/nejen/theme.name` and symlinks
the wallpaper to `~/.local/state/nejen/background`.

The shipped configs still point at the legacy `~/.config/nejen/current/*`
layout (or hardcode colors). This file lists every repoint needed. Do NOT
apply these by hand-editing rendered output; edit the listed config files.

## 1. Alacritty

`config/alacritty/alacritty.toml`, line 1:

    general.import = [ "~/.config/nejen/current/theme/alacritty.toml" ]

must become:

    general.import = [ "~/.local/state/nejen/theme/alacritty.toml" ]

## 2. btop

`config/btop/btop.conf` has `color_theme = "current"`, which btop resolves to
`~/.config/btop/themes/current.theme`. btop does not expand `~` inside config
values, so keep `color_theme = "current"` and make the installer step
create:

    ln -sfn ~/.local/state/nejen/theme/btop.theme ~/.config/btop/themes/current.theme

Also delete `config/btop/themes/current.theme` from the repo: it is a stale
hand-written copy of the old Kanagawa theme (not even the nejen palette)
and must never be shipped as a regular file, or it will shadow the symlink
mechanism. btop reloads on SIGUSR2; `nejen theme set` already sends it.

## 3. Hyprland (borders / decoration)

`config/hypr/hyprland.conf` must gain a source line for the rendered theme,
ordered so the theme overrides shipped defaults but user files still win:

    source = $NEJEN_PATH/config/hypr/core/rules.conf
    source = ~/.local/state/nejen/theme/hyprland.conf   # <- add this
    ...user files (monitors.conf, theme.conf, local.conf) stay after it

(Since the core/ config split, this ordering lives in
`config/hypr/hyprland.conf`; border colors come only from the rendered
theme.) Then remove any duplicated hardcoded colors:
- `config/hypr/theme.conf`: `col.active_border = rgba(4a4f8aff)` and
  `col.inactive_border = rgba(2a2a2aaa)` duplicate the rendered
  `hyprland.conf` exactly; remove them (the rest of the nejen look --
  gaps, blur, shadows -- can stay for now, or later move into the template).

`nejen theme set` runs `hyprctl reload`, which re-reads sourced files.

## 4. hyprlock

`config/hypr/hyprlock.conf`:

- line 1: `source = ~/.config/nejen/current/theme/hyprlock.conf` -- there is
  no hyprlock template yet. Either delete the line, or add
  `templates/hyprlock.conf.tmpl` (defining `$bg`, `$text`, `$accent`, ... from
  the palette) and repoint to `~/.local/state/nejen/theme/hyprlock.conf`.
- line 27: `path = ~/.config/nejen/current/background` must become
  `path = ~/.local/state/nejen/background`.

## 5. Wallpaper daemons

- `config/hypr/core/session.conf` (formerly defaults/autostart.conf):
  `exec-once = uwsm-app -- swaybg -i ~/.config/nejen/current/background -m fill`
  -> `-i ~/.local/state/nejen/background`.
- `config/hypr/hyprland.conf` also runs `exec-once = hyprpaper`, and
  `config/hypr/hyprpaper.conf` hardcodes
  `~/images/wallpapers/galcezon_hud.jpg`. Running both swaybg and
  hyprpaper is redundant; pick one. Recommendation: drop hyprpaper (swaybg is
  already restarted by `nejen theme set` against the state symlink). If
  hyprpaper is kept, repoint its `preload`/`wallpaper` lines to
  `~/.local/state/nejen/background` and add `hyprctl hyprpaper reload` to the
  reload step.

## 6. Waybar

GTK CSS has `@import` but does not expand `~`, so the import must live in the
user's `~/.config/waybar/style.css` using a relative path (relative to
`~/.config/waybar/`), theme first so its `@define-color`s take effect:

    @import "../../.local/state/nejen/theme/waybar.css";
    @import "<path-to-$NEJEN_PATH>/config/waybar/style.css";

Follow-up to make theming real: move the 8 `@define-color nejen_*` lines at
the top of `config/waybar/style.css` into `templates/waybar.css.tmpl`
(palette-driven, e.g. `@define-color nejen_green ${palette.accent};`). Until
that move, the rendered `waybar.css` is intentionally an empty override and
the defines stay hardcoded in `config/waybar/style.css`. Waybar restart is
handled by `nejen theme set` (`nejen bar restart`, SIGUSR2 fallback).

Same story applies to `config/walker/style.css` (identical hardcoded
`@define-color nejen_*` block); a `walker.css.tmpl` is an optional follow-up.

## 7. Mako (no include mechanism)

mako reads exactly one config file and has no include directive. Therefore
`~/.config/mako/config` must be GENERATED dynamically, never hand-edited in place. Proposed mechanism:

1. Split `config/mako/config` into `config/mako/config` (global options only,
   colors removed into `templates/mako.ini.tmpl`) and
   `config/mako/sections` (the `[mode=do-not-disturb]` block and any other
   `[criteria]` sections).
2. `nejen theme set` concatenates, in this exact order (mako requires global
   options before any `[criteria]` section):

       cat "$NEJEN_PATH/config/mako/config" \
           "$STATE_DIR/theme/mako.ini" \
           "$HOME/.config/nejen/mako-overrides" (optional, if present) \
           "$NEJEN_PATH/config/mako/sections" > ~/.config/mako/config

3. then `makoctl reload` (already done by `nejen theme set`).

The colors currently hardcoded in `config/mako/config` that belong in
`mako.ini.tmpl` when step 1 happens: `background-color=#0a0a12a6`,
`text-color=#c8cbff`, `border-color=#4a4f8a59`,
`progress-color=over #2c3366ff` (these use alpha suffixes; that will need the
`.nohash` transform, e.g. `border-color=#${palette.border.nohash}59`).

## 8. swayosd

`config/swayosd/style.css` hardcodes the nejen theme colors and there is no user-file
layer. Either move the colors into `templates/swayosd.css.tmpl` and point
swayosd at the rendered file (swayosd-server `--style
~/.local/state/nejen/theme/swayosd.css`, set in its systemd user unit /
autostart line), or add `@import "../../.local/state/nejen/theme/swayosd.css";`
at the top of a user `~/.config/swayosd/style.css`. `nejen theme set` already
does `systemctl --user try-restart swayosd.service`.

## 9. VS Code

The rendered `vscode.json` is a settings *fragment*
(`{"workbench.colorTheme": ...}`), not a standalone file VS Code can import.
Mechanism: merge it into `~/.config/Code/User/settings.json`, e.g.

    jq -s '.[0] * .[1]' ~/.config/Code/User/settings.json \
        ~/.local/state/nejen/theme/vscode.json > tmp && mv tmp .../settings.json

This belongs in `nejen theme set` (guarded on the settings file existing).

## 10. Neovim

`config/nvim/lua/plugins/theme.lua` currently hardcodes the whole nejen
colorscheme; the rendered `neovim.lua` is an empty spec (`return {}`).
Eventual mechanism: make
`config/nvim/lua/plugins/theme.lua` a symlink to
`~/.local/state/nejen/theme/neovim.lua` and move the colorscheme spec into
`templates/neovim.lua.tmpl` / per-theme overrides. The existing
`config/nvim/lua/plugins/nejen-theme-hotreload.lua` already reloads
`plugins.theme` on change. No repoint is required for nvim to keep working
today.

## 11. Icon theme

Rendered `icons.theme` contains the icon theme name. `nejen theme set`
applies it via
`gsettings set org.gnome.desktop.interface icon-theme "$(< .../icons.theme)"`.
No config file repoint needed.

## 12. install.sh / legacy state layout

`install.sh` `setup_theme_state()` (lines ~90-107) creates the legacy
`~/.config/nejen/current/{theme,theme.name,background}` layout and copies
theme files into it. Replace the whole function with:

    nejen theme set nejen

(after the dispatcher is linked), which creates the state layout under
`~/.local/state/nejen/`. Also: `~/.config/nejen/themes` should be a real
(initially empty) directory for user themes, not a symlink to
`$NEJEN_PATH/themes` as install.sh currently makes it -- the engine already
searches both layers.

## 13. Removed files

`themes/nejen/` now contains only `theme.toml` and `backgrounds/`. The old
hand-written `alacritty.toml`, `btop.theme`, `hyprland.conf`, `mako.ini`,
`neovim.lua`, `swayosd.css`, `vscode.json`, `waybar.css`, `icons.theme`, and
`colors.toml` were deleted; their data lives in `theme.toml` (the ANSI slot
mapping formerly in `colors.toml` is encoded in
`templates/alacritty.toml.tmpl`). Anything outside this repo still reading
`~/.config/nejen/current/theme/<file>` must switch to
`~/.local/state/nejen/theme/<file>`.
