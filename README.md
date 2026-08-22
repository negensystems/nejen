# NEJEN

NEJEN is an opinionated, minimal Hyprland desktop environment for Arch Linux. It is built as a daily driver, focusing on clean separation of system defaults and user overrides.

## Key Features

* **Minimalist Shell**: Installs a complete Hyprland desktop shell (Waybar, Walker, Mako, SwayOSD, Hypridle/Hyprlock) without bloatware or pre-installed web browsers.
* **Single Entry CLI**: All subcommands are routed through the `nejen` binary (e.g., `nejen theme set`, `nejen doctor`), keeping the global namespace clean.
* **Layered Configuration**: System configuration defaults are stored under the installation directory and imported/sourced by thin wrappers in `~/.config`. Updates will never overwrite your personal overrides.
* **Declarative Theming**: Define themes as simple TOML variables (colors, fonts, wallpaper). The theme engine renders them into native application configs on target locations.
* **Declarative Keymaps**: Keybindings are configured via TOML, which is rendered into Hyprland config files and used to generate a dynamic cheat sheet.

## Project Structure

* `cmd/nejen/`: The Go source code for the CLI dispatcher and system commands.
* `config/`: System defaults (Hyprland, Waybar, Walker, Alacritty, Neovim, etc.).
* `themes/` & `templates/`: Theme definition TOMLs and template files.
* `packages/`: Package manifests (`core.txt` for the core ISO; `full.txt` for user extras).
* `packaging/`: PKGBUILD script for Pacman installation.

## Installation

### Method A: Install as an Arch Package (Recommended)

Build and install via pacman:
```sh
cd packaging
makepkg -si
```

### Method B: Developer Mode

If you are developing or modifying the configurations directly, run:
```sh
./install.sh
```

After installation, verify the environment:
```sh
nejen doctor
```

## Commands

All desktop operations and tools route through the `nejen` binary. You can inspect keybindings anytime with `nejen keys` (or `Super+/`).

### Core & System

| Command | Description |
| --- | --- |
| `nejen doctor` | Verify system health, dependencies, and path configuration |
| `nejen update` | Pull repository updates and run package upgrades |
| `nejen keys` | Open searchable keybinding reference |
| `nejen hub` | Open the central action menu (`Super+N`) |
| `nejen hub session` | Open power and session menu (`Super+Shift+N`) |
| `nejen lock` | Lock the screen immediately (`Ctrl+Escape`) |

### Theming & Display

| Command | Description |
| --- | --- |
| `nejen theme list` | List available themes |
| `nejen theme set <name>` | Switch active theme and re-render application configs |
| `nejen theme current` | Print the active theme name |
| `nejen theme bg pick` | Open visual wallpaper picker (`Super+Shift+W`) |
| `nejen night toggle` | Toggle screen warmth / night light filter (`Super+W`) |
| `nejen bar toggle` | Toggle Waybar visibility (`Super+B`) |
| `nejen bar restart` | Restart Waybar |

### Window Management

| Command | Description |
| --- | --- |
| `nejen win detach` | Float and pin current window (`Super+U`) |
| `nejen win gaps` | Toggle window gaps (`Super+G`) |
| `nejen win layout` | Switch between dwindle and master layouts (`Super+Y`) |
| `nejen win square` | Toggle 1:1 aspect ratio constraint for single windows (`Super+O`) |
| `nejen win close-all` | Close all windows on active workspace |

### Hardware & Utilities

| Command | Description |
| --- | --- |
| `nejen screenshot` | Capture selected screen region (`Print`) |
| `nejen screenrecord` | Start or stop screen recording (`Super+R`) |
| `nejen keyboard clean` | Temporarily disable keyboard input for wiping (`Super+Alt+K`) |
| `nejen dnd toggle` | Toggle notification Do Not Disturb (`Super+Ctrl+D`) |
| `nejen idle toggle` | Toggle idle sleep inhibitor / caffeine (`Super+I`) |
| `nejen audio switch` | Switch active audio output sink |
| `nejen keymap render` | Re-render Hyprland keybindings from `keymap.toml` |

## Customization

To customize your setup, edit the corresponding override files. These files load last and take precedence:

* **Keybindings**: `~/.config/nejen/keymap.toml`
* **Hub Actions**: `~/.config/nejen/hub.toml`
* **Hyprland Settings**: `~/.config/hypr/overrides.conf`
* **Waybar Styling**: `~/.config/waybar/overrides.css`
* **Alacritty Configuration**: `~/.config/alacritty/overrides.toml`
* **Custom Themes**: `~/.config/nejen/themes/<name>/theme.toml`
* **Wallpapers**: `~/.config/nejen/backgrounds/`

## Wallpapers

Drop images into `~/.config/nejen/backgrounds/` and they join the picker and
the cycle. That folder is theme-independent: switching themes will not
replace a wallpaper you chose yourself, and your collection survives
`nejen theme set`. If you keep wallpapers elsewhere, symlink the whole
directory into place instead of copying:

```bash
ln -s ~/Pictures/wallpapers ~/.config/nejen/backgrounds
```

| Action | Key | Command |
| --- | --- | --- |
| Browse with previews | `Super+Shift+W` | `nejen theme bg pick` |
| Next wallpaper | `Super+Ctrl+W` | `nejen theme bg next` |
| Previous wallpaper | `Super+Ctrl+Shift+W` | `nejen theme bg prev` |
| Random wallpaper | | `nejen theme bg random` |
| Add images to the collection | | `nejen theme bg add <file>...` |
| List what is available | | `nejen theme bg list [--names]` |
| Show the current wallpaper | | `nejen theme bg current [--name]` |
| Print the wallpaper folder | | `nejen theme bg dir` |

The same actions are under **Theme > Wallpaper** in the hub (`Super+N`).

`nejen theme bg set` takes a path, a filename, or the display name shown in
the menus, so `nejen theme bg set "Misty Ridge"` and
`nejen theme bg set ~/Pictures/ridge.jpg` both work. Wallpapers are searched
in `~/.config/nejen/backgrounds/`, then the active theme's `backgrounds/`
directory (user override first, then the shipped one); a filename found
earlier shadows the same name later, so you can replace a shipped wallpaper
without renaming yours. Add `--link` to `nejen theme bg add` to symlink
large files instead of copying them.

## Cleaning Mode

`nejen keyboard clean` turns the keyboard off so you can wipe it, then turns
it back on by itself. While it runs, every key press -- with or without
modifiers -- is swallowed by a Hyprland submap, so a cloth dragged across the
keys types nothing, opens nothing, and cannot answer a dialog. A full-screen
overlay counts down the time left and catches the stray pointer taps a cloth
produces.

| Action | Key | Command |
| --- | --- | --- |
| Clean for 30 seconds | `Super+Alt+K` | `nejen keyboard clean` |
| Clean for another duration | | `nejen keyboard clean --for 2m` |
| Finish early | middle click | `nejen keyboard clean --stop` |

It is also under **Toggles > Cleaning mode** in the hub (`Super+N`), and the
bar grows a `󰌐` indicator you can click to finish.

Durations accept `45`, `45s`, `1m30s`, and are clamped to between 5 seconds
and 5 minutes. Because the pointer is deliberately left alive, there are
always four ways back to a working keyboard -- the deadline, `--stop`, a
middle click, and the bar indicator, which reaps the session if the process
holding it open ever dies.

## Built-in Bar Modules

Weather and countdown ship inside the `nejen` binary itself -- no interpreter, no external script:

* **Weather** (`nejen weather`): location-aware current conditions from Open-Meteo, no API key. Set fixed coordinates in `~/.config/nejen/weather.toml` (recommended on a VPN); auto-detects from your IP otherwise.
* **Countdown** (`nejen countdown`): day counters toward dated events, managed from the bar via `nejen countdown menu`. Events live in `~/.config/nejen/countdown.json`; an existing `~/.config/waybar-countdown/events.json` is picked up automatically.

## Recommended Companion Widgets

NEJEN also includes pre-configured Waybar blocks that integrate with optional external widgets:

* **[waybar-nordvpn](https://github.com/takumiymd/waybar-nordvpn)**: Interactive status monitor and controller for NordVPN connections.
* **[waybar-tradingview](https://github.com/takumiymd/waybar-tradingview)**: Live stock and cryptocurrency ticker for your status bar.

If you don't install these, the corresponding modules will fail silently without breaking the bar.

## Why Go?

NEJEN's core logic is powered by a single compiled Go binary. See [docs/why-go.md](docs/why-go.md) for the full architectural and performance motivations behind this design choice.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
