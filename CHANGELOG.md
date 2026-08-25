# Changelog

Notable changes to NEJEN, newest first. Versions follow [semantic versioning](https://semver.org/).

## 1.0.0

First public release.

### The desktop

- Curated Hyprland stack configured out of the box: Waybar, Walker (with the elephant backend), Mako, SwayOSD, Hypridle and Hyprlock, alongside Alacritty, Neovim, btop and fastfetch.
- Plymouth boot splash, greetd/tuigreet login, and a themed lock screen sharing one design, built from the ASCII block mark in `logo.txt`.
- `packages/core.txt` is the whole distribution: boot and base system, the Hyprland shell, and the tooling `nejen` needs. No browsers, no office suite, no consumer apps.

### The `nejen` command

- A single compiled Go binary dispatches every desktop action: themes, wallpapers, window management, capture, toggles, hardware, and session control. `nejen help` lists the command surface; `nejen doctor` checks that the system can run it.
- **The hub** (`Super+N`): one menu for everything without a dedicated key, defined declaratively in `config/hub.toml`. Branches open a level at a time behind a breadcrumb prompt, and `Escape` steps back up rather than closing.
- **Keymaps**: `config/keymap.toml` is the single source of truth. It renders into Hyprland bind lines and into the searchable cheat sheet behind `nejen keys` (`Super+/`), so the two can never drift.
- **Themes**: TOML definitions rendered through `templates/*.tmpl` into native configs for Alacritty, Waybar, SwayOSD, btop, Mako, hyprlock, bluetuith and more. `nejen theme set` swaps them and reloads the running apps in place.
- **Wallpapers**: a theme-independent collection in `~/.config/nejen/backgrounds/`, with a previewing picker, cycling, and random selection.
- **Cleaning mode** (`nejen keyboard clean`, `Super+Alt+K`): takes the keyboard offline so it can be wiped, with four independent ways back out.
- **Built-in bar modules**: `nejen weather` (Open-Meteo, no API key) and `nejen countdown`, so the status bar needs no external script or interpreter.

### Installation

- `install.sh` bootstraps `paru` from the AUR when neither it nor `yay` is on `PATH`, so a bare Arch system reaches a working desktop in one command. Without it the AUR half of the manifest, including `walker` and `elephant`, would be skipped, leaving the launcher and the hub dead.
- `--no-packages` skips the package step; `--packages <name|path>` selects a different manifest.

### Configuration

- Every app gets a dedicated override file under `~/.config` that loads last and is never touched by an update. The generated files it sits beside carry a `DO NOT EDIT` header.
- Pre-existing configuration the installer would replace is moved to `<file>.pre-nejen.bak` first, and a second install never overwrites that first backup.
