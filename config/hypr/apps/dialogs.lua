-- Dialogs, pickers, and app quirks.

-- Third-party windows that behave like panels: float centered at panel size.
-- (NEJEN's own org.nejen.* windows get this from core/rules.lua.)
hl.window_rule({
    name = "nejen-floating-window",
    match = { tag = "floating-window" },
    float = true,
    center = true,
    size = "900 620",
})
hl.window_rule({
    name = "nejen-floating-window-media-tag",
    match = { class = "(org.gnome.NautilusPreviewer|org.gnome.Evince|com.gabm.satty|imv|mpv)" },
    tag = "+floating-window",
})

-- File open/save pickers
hl.window_rule({
    name = "nejen-picker-gtk-tag",
    match = { class = "xdg-desktop-portal-gtk" },
    tag = "+floating-window",
})
hl.window_rule({
    name = "nejen-picker-dialogs-tag",
    match = {
        class = "(sublime_text|DesktopEditors|org.gnome.Nautilus)",
        title = "^(Open.*Files?|Open [Ff]older.*|Save.*Files?|Save.*As|Save|All Files|.*wants to (open|save).*|[Cc]hoose.*)",
    },
    tag = "+floating-window",
})

-- Small utility apps
hl.window_rule({
    name = "nejen-calculator",
    match = { class = "org.gnome.Calculator" },
    float = true,
})
hl.window_rule({
    name = "nejen-share-localsend",
    match = { class = "(Share|localsend)" },
    float = true,
    center = true,
})
hl.window_rule({
    name = "nejen-localsend-size",
    match = { class = "localsend" },
    size = "1080 720",
})

-- Focus behavior fixes
hl.window_rule({
    name = "nejen-telegram-focus",
    match = { class = "org.telegram.desktop" },
    focus_on_activate = false,
})
hl.window_rule({
    name = "nejen-resolve-stay-focused",
    match = { class = ".*[Rr]esolve.*", float = true },
    stay_focused = true,
})

-- JetBrains IDEs: splash, popups, and tooltips all misbehave under tiling
hl.window_rule({
    name = "jetbrains-splash",
    match = { class = "^(jetbrains-.*)$", title = "^(splash)$", float = true },
    center = true,
    no_focus = true,
    border_size = 0,
})
hl.window_rule({
    name = "jetbrains-popup",
    match = { class = "^(jetbrains-.*)", title = "^(| )$", float = true },
    center = true,
    stay_focused = true,
    border_size = 0,
    min_size = "(monitor_w*0.5) (monitor_h*0.5)",
})
hl.window_rule({
    name = "jetbrains-tooltip",
    match = { class = "^(jetbrains-.*)$", title = "^(win.*)$", float = true },
    no_initial_focus = true,
})

-- Overlay layers that must not animate
hl.layer_rule({
    name = "nejen-layer-walker",
    match = { namespace = "walker" },
    no_anim = true,
})
hl.layer_rule({
    name = "nejen-layer-selection",
    match = { namespace = "selection" },
    no_anim = true,
})
