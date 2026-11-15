-- NEJEN input defaults. Machine-specific taste (layout variants, pointer
-- speed, per-device blocks) belongs in the user's overrides, which are
-- loaded after this file and win.

hl.config({
    input = {
        kb_layout = "us",
        kb_options = "compose:caps",
        repeat_rate = 35,
        repeat_delay = 500,

        follow_mouse = 1,
        sensitivity = 0,

        touchpad = {
            natural_scroll = true,
            clickfinger_behavior = true,
            disable_while_typing = true,
        },
    },
})

-- Three-finger swipe moves between workspaces
hl.gesture({
    fingers = 3,
    direction = "horizontal",
    action = "workspace",
})
