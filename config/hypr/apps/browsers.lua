-- Browser handling.
--
-- Browsers manage their own surfaces, so they get near-full opacity and
-- are grouped by engine family with tags for anything else that needs to
-- tell them apart.

hl.window_rule({
    name = "nejen-browser-blink-tag",
    match = { class = "((google-)?[cC]hrom(e|ium)|[bB]rave-browser|[mM]icrosoft-edge|Vivaldi-stable|helium)" },
    tag = "+browser-blink",
})
hl.window_rule({
    name = "nejen-browser-gecko-tag",
    match = { class = "([fF]irefox|zen|librewolf)" },
    tag = "+browser-gecko",
})

hl.window_rule({
    name = "nejen-browser-blink-untag-glass",
    match = { tag = "browser-blink" },
    tag = "-nejen-glass",
})
hl.window_rule({
    name = "nejen-browser-gecko-untag-glass",
    match = { tag = "browser-gecko" },
    tag = "-nejen-glass",
})
hl.window_rule({
    name = "nejen-browser-blink-opacity",
    match = { tag = "browser-blink" },
    opacity = "1.0 0.97",
})
hl.window_rule({
    name = "nejen-browser-gecko-opacity",
    match = { tag = "browser-gecko" },
    opacity = "1.0 0.97",
})

-- Chromium --app windows sometimes spawn floating; force them to tile
hl.window_rule({
    name = "nejen-browser-blink-tile",
    match = { tag = "browser-blink" },
    tile = true,
})

-- Web video apps: fully opaque, no browser treatment
hl.window_rule({
    name = "nejen-web-video-untag-browser",
    match = { class = "(chrome-youtube.com__-Default|chrome-app.zoom.us__wc_home-Default)" },
    tag = "-browser-blink",
})
hl.window_rule({
    name = "nejen-web-video-untag-glass",
    match = { class = "(chrome-youtube.com__-Default|chrome-app.zoom.us__wc_home-Default)" },
    tag = "-nejen-glass",
})
