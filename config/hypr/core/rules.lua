-- NEJEN window rules. NEJEN's own windows all carry an org.nejen.*
-- app-id, so one prefix rule handles the whole toolkit; per-app tweaks
-- for third-party software live under apps/.

local function resolveHyprDir()
    local src = debug and debug.getinfo(1, "S").source or ""
    if src:sub(1, 1) == "@" then
        local dir = src:sub(2):match("^(.*)/core/[^/]+$")
        if dir and dir ~= "" then
            return dir
        end
    end
    local home = os.getenv("HOME") or ""
    local nejenPath = os.getenv("NEJEN_PATH")
    if not nejenPath or nejenPath == "" then
        nejenPath = home .. "/.local/share/nejen"
    end
    return nejenPath .. "/config/hypr"
end

-- Never let clients maximize themselves; tiling owns geometry
hl.window_rule({
    name = "nejen-suppress-maximize",
    match = { class = ".*" },
    suppress_event = "maximize",
})

-- Every NEJEN tool window (TUI panels, about, screensaver picker) floats
-- centered at a comfortable size. `match` is a FULL match, so the prefix
-- needs an explicit trailing `.*`; a bare `^org\.nejen\.` matches nothing.
hl.window_rule({
    name = "nejen-panel-tag",
    match = { class = "^org\\.nejen\\..*$" },
    tag = "+nejen-panel",
})
hl.window_rule({
    name = "nejen-panel-float",
    match = { tag = "nejen-panel" },
    float = true,
    center = true,
})

-- Size is carried by its own tag so a panel can opt out: two `size` rules
-- matching the same window do not override each other, the tag one wins.
-- bluetuith draws a flat device list with its keybind bar pinned to the
-- bottom of the terminal, so the panel default left ~30 blank rows between
-- the two. It gets a height sized to its content instead.
hl.window_rule({
    name = "nejen-panel-tall-tag",
    match = { tag = "nejen-panel" },
    tag = "+nejen-panel-tall",
})
hl.window_rule({
    name = "nejen-panel-bluetuith-untag",
    match = { class = "^org\\.nejen\\.bluetuith$" },
    tag = "-nejen-panel-tall",
})
hl.window_rule({
    name = "nejen-panel-tall-size",
    match = { tag = "nejen-panel-tall" },
    size = "900 620",
})
hl.window_rule({
    name = "nejen-panel-bluetuith-size",
    match = { class = "^org\\.nejen\\.bluetuith$" },
    size = "900 340",
})

-- The screensaver covers everything
hl.window_rule({
    name = "nejen-screensaver",
    match = { class = "^org\\.nejen\\.screensaver$" },
    fullscreen = true,
    animation = "fade",
})

-- The cleaning overlay (`nejen keyboard clean`) covers the desktop for a
-- second reason on top of being readable: covering it is what makes the
-- overlay catch the stray pointer taps a cloth produces, instead of letting
-- them land in whatever was on screen.
--
-- Maximized rather than fullscreen, deliberately. A fullscreen window renders
-- above the `top` layer, which is where the bar lives, so fullscreen hid
-- the cleaning indicator and made the "click the bar to finish" way out
-- unclickable. Maximize fills everything the bar does not reserve instead.
hl.window_rule({
    name = "nejen-cleaning",
    match = { class = "^org\\.nejen\\.cleaning$" },
    maximize = true,
    animation = "fade",
})

-- Detached windows (nejen win detach) get softer corners
hl.window_rule({
    name = "nejen-pop-rounding",
    match = { tag = "pop" },
    rounding = 8,
})

-- Focus-less ghost windows from XWayland drag helpers
hl.window_rule({
    name = "nejen-xwayland-ghost",
    match = {
        class = "^$",
        title = "^$",
        xwayland = true,
        float = true,
        fullscreen = false,
        pin = false,
    },
    no_focus = true,
})

-- Slight translucency for terminals over the wallpaper HUD
hl.window_rule({
    name = "nejen-terminal-opacity",
    match = { tag = "terminal" },
    opacity = "0.98 0.95",
})

-- Default translucency: tag everything, let per-app files opt out below,
-- then apply to whatever kept the tag
hl.window_rule({
    name = "nejen-glass-tag-all",
    match = { class = ".*" },
    tag = "+nejen-glass",
})
dofile(resolveHyprDir() .. "/apps.lua")
hl.window_rule({
    name = "nejen-glass-opacity",
    match = { tag = "nejen-glass" },
    opacity = "0.97 0.9",
})
