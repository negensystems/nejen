-- NEJEN Hyprland entry point.
--
-- Loaded through the thin ~/.config/hypr/hyprland.lua written by
-- install.sh; the user's ~/.config/hypr/overrides.lua is loaded after
-- this whole file and wins over everything here. $NEJEN_PATH is exported
-- at session launch by uwsm (config/uwsm/env).
--
-- Precedence, lowest to highest: core defaults -> rendered keymap ->
-- rendered theme -> this machine's setup -> user overrides.

local home = os.getenv("HOME") or ""
local stateHome = os.getenv("XDG_STATE_HOME")
if not stateHome or stateHome == "" then
    stateHome = home .. "/.local/state"
end

local function resolveHyprDir()
    local src = debug and debug.getinfo(1, "S").source or ""
    if src:sub(1, 1) == "@" then
        local dir = src:sub(2):match("^(.*)/[^/]+$")
        if dir and dir ~= "" then
            return dir
        end
    end
    local nejenPath = os.getenv("NEJEN_PATH")
    if not nejenPath or nejenPath == "" then
        nejenPath = home .. "/.local/share/nejen"
    end
    return nejenPath .. "/config/hypr"
end

local function loadOptional(path)
    local f = io.open(path, "r")
    if not f then
        return
    end
    f:close()
    dofile(path)
end

local hyprDir = resolveHyprDir()

dofile(hyprDir .. "/core/session.lua")
dofile(hyprDir .. "/core/surface.lua")
dofile(hyprDir .. "/core/devices.lua")
dofile(hyprDir .. "/core/cleaning.lua")
dofile(hyprDir .. "/core/rules.lua")

-- Keybindings, rendered from keymap.toml by `nejen keymap render`.
-- Rebind in ~/.config/nejen/keymap.toml, never here.
loadOptional(stateHome .. "/nejen/keymap/hyprland.lua")

-- Active theme (border colors etc), written by `nejen theme set`.
loadOptional(stateHome .. "/nejen/theme/hyprland.lua")

-- This machine's opinionated setup.
dofile(hyprDir .. "/monitors.lua")
dofile(hyprDir .. "/input.lua")
dofile(hyprDir .. "/envs.lua")
dofile(hyprDir .. "/theme.lua")
dofile(hyprDir .. "/autostart.lua")
dofile(hyprDir .. "/local.lua")
