-- Monitor layout for this machine.
--
-- `hyprctl monitors` lists connected outputs and their modes.
-- Syntax: hl.monitor({ output = <port>, mode = <resolution>, position = <position>, scale = <scale> })
-- `nejen display scale` cycles the focused monitor's scale at runtime;
-- put permanent choices here and relaunch Hyprland to apply.

-- Default: every display runs its preferred mode at 2x (HiDPI).
hl.env("GDK_SCALE", "2")
hl.monitor({
    output = "",
    mode = "preferred",
    position = "auto",
    scale = "auto",
})

-- Per-output entries pin a specific display, e.g.:
-- hl.monitor({ output = "DP-3", mode = "3840x2160@120", position = "0x0", scale = 1.5 })
-- hl.monitor({ output = "eDP-1", mode = "preferred", position = "auto", scale = 2 })
