-- Media and games: full opacity always (transparency ruins video), and
-- idle inhibition where a fullscreen session means "watching/playing".

-- Fully opaque media surfaces
local mediaClass = "^(zoom|vlc|mpv|org.kde.kdenlive|com.obsproject.Studio|com.github.PintaProject.Pinta|imv|org.gnome.NautilusPreviewer|qemu)$"
hl.window_rule({
    name = "nejen-media-untag-glass",
    match = { class = mediaClass },
    tag = "-nejen-glass",
})
hl.window_rule({
    name = "nejen-media-opacity",
    match = { class = mediaClass },
    opacity = "1 1",
})

-- Picture-in-picture: small pinned overlay, bottom of nothing, top-right
hl.window_rule({
    name = "nejen-pip-tag",
    match = { title = ".*Picture.?in.?[Pp]icture.*" },
    tag = "+pip",
})
hl.window_rule({
    name = "nejen-pip-untag-glass",
    match = { tag = "pip" },
    tag = "-nejen-glass",
})
hl.window_rule({
    name = "nejen-pip-style",
    match = { tag = "pip" },
    float = true,
    pin = true,
    size = "640 360",
    keep_aspect_ratio = true,
    border_size = 0,
    opacity = "1 1",
    move = "(monitor_w-window_w-32) (monitor_h*0.05)",
})

-- Webcam overlay drawn by nejen screenrecord
hl.window_rule({
    name = "nejen-webcam-overlay",
    match = { title = "nejen-webcam" },
    float = true,
    pin = true,
    no_initial_focus = true,
    no_dim = true,
    move = "(monitor_w-window_w-40) (monitor_h-window_h-40)",
})

-- Steam: floating client windows, opaque, no idle lock mid-game
hl.window_rule({
    name = "nejen-steam-float",
    match = { class = "steam" },
    float = true,
    idle_inhibit = "fullscreen",
})
hl.window_rule({
    name = "nejen-steam-main",
    match = { class = "steam", title = "Steam" },
    center = true,
    size = "1120 720",
})
hl.window_rule({
    name = "nejen-steam-friends",
    match = { class = "steam", title = "Friends List" },
    size = "480 780",
})
hl.window_rule({
    name = "nejen-steam-untag-glass",
    match = { class = "steam.*" },
    tag = "-nejen-glass",
})
hl.window_rule({
    name = "nejen-steam-opacity",
    match = { class = "steam.*" },
    opacity = "1 1",
})

-- Emulation / game streaming
hl.window_rule({
    name = "nejen-retroarch",
    match = { class = "com.libretro.RetroArch" },
    fullscreen = true,
    tag = "-nejen-glass",
    opacity = "1 1",
    idle_inhibit = "fullscreen",
})
hl.window_rule({
    name = "nejen-geforcenow-idle",
    match = { class = "GeForceNOW" },
    idle_inhibit = "fullscreen",
})

-- Anything tagged noidle keeps the machine awake while open
hl.window_rule({
    name = "nejen-noidle",
    match = { tag = "noidle" },
    idle_inhibit = "always",
})
