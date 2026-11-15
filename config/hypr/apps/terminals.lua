-- Terminals share one tag so the theme can style them uniformly.

hl.window_rule({
    name = "nejen-terminal-tag",
    match = { class = "(Alacritty|kitty|com.mitchellh.ghostty)" },
    tag = "+terminal",
})
hl.window_rule({
    name = "nejen-terminal-untag-glass",
    match = { tag = "terminal" },
    tag = "-nejen-glass",
})
