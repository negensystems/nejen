-- Password managers never appear in screen shares and float like panels.

hl.window_rule({
    name = "nejen-1password-noscreenshare",
    match = { class = "^(1[pP]assword)$" },
    no_screen_share = true,
    tag = "+floating-window",
})

hl.window_rule({
    name = "nejen-bitwarden-noscreenshare",
    match = { class = "^(Bitwarden)$" },
    no_screen_share = true,
    tag = "+floating-window",
})

-- Bitwarden browser-extension popup (Chrome extension window class)
hl.window_rule({
    name = "nejen-bitwarden-ext-noscreenshare",
    match = { class = "chrome-nngceckbapebfimnlniiiahkandclblb-Default" },
    no_screen_share = true,
    tag = "+floating-window",
})
