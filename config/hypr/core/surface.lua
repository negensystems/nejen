-- NEJEN surface: how windows look and move. Colors are deliberately
-- absent; they come from the rendered theme, loaded later.

hl.config({
    general = {
        gaps_in = 4,
        gaps_out = 8,
        border_size = 2,
        layout = "dwindle",
        resize_on_border = true,
    },

    decoration = {
        rounding = 6,
        active_opacity = 1.0,
        inactive_opacity = 0.94,

        shadow = {
            enabled = true,
            range = 14,
            render_power = 3,
            color = "rgba(00000033)",
            color_inactive = "rgba(00000018)",
        },

        blur = {
            enabled = true,
            size = 4,
            passes = 2,
            ignore_opacity = true,
        },
    },

    animations = {
        enabled = true,
    },

    dwindle = {
        preserve_split = true,
        force_split = 2,
    },

    master = {
        mfact = 0.55,
        new_status = "slave",
    },

    misc = {
        disable_hyprland_logo = true,
        disable_splash_rendering = true,
        focus_on_activate = true,
        -- Allow hyprlock relaunch on locker crash instead of lockdead.
        allow_session_lock_restore = true,
        key_press_enables_dpms = true,
        mouse_move_enables_dpms = true,
    },
})

-- Motion: quick to settle, no bounce. One curve for movement, a faster
-- one for fades, so the desktop feels crisp rather than springy.
hl.curve("nejenSettle", { type = "bezier", points = { {0.22, 0.9}, {0.28, 1.0} } })
hl.curve("nejenFade",   { type = "bezier", points = { {0.4,  0.0}, {0.6,  1.0} } })

hl.animation({ leaf = "windows",    enabled = true, speed = 3, bezier = "nejenSettle", style = "popin 90%" })
hl.animation({ leaf = "windowsOut", enabled = true, speed = 3, bezier = "nejenFade",   style = "popin 90%" })
hl.animation({ leaf = "fade",       enabled = true, speed = 2, bezier = "nejenFade" })
hl.animation({ leaf = "workspaces", enabled = true, speed = 3, bezier = "nejenSettle", style = "slide" })
hl.animation({ leaf = "layers",     enabled = true, speed = 2, bezier = "nejenFade",   style = "fade" })
hl.animation({ leaf = "border",     enabled = true, speed = 4, bezier = "default" })
