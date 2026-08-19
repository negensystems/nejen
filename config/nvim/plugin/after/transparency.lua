-- NEJEN transparency pass.
--
-- Strips the background color from UI highlight groups so the compositor's
-- blur and wallpaper show through, leaving every other attribute intact.
-- Lives in plugin/after/ so it runs once the colorscheme has loaded; the
-- nejen-theme-hotreload plugin re-sources it after every theme switch.

local groups = {
	-- editor chrome
	"Normal",
	"NormalNC",
	"EndOfBuffer",
	"SignColumn",
	"LineNr",
	"CursorLineNr",
	"FoldColumn",
	"Folded",
	-- floats and menus
	"NormalFloat",
	"FloatBorder",
	"FloatTitle",
	"Pmenu",
	-- pickers and side panels
	"TelescopeNormal",
	"TelescopeBorder",
	"TelescopePromptBorder",
	"TelescopePromptTitle",
	"WhichKeyFloat",
	"NeoTreeNormal",
	"NeoTreeNormalNC",
	"NeoTreeEndOfBuffer",
	"NeoTreeWinSeparator",
}

-- nvim-notify defines a group per severity and part; add them all.
for _, level in ipairs({ "INFO", "WARN", "ERROR", "DEBUG", "TRACE" }) do
	for _, part in ipairs({ "Body", "Title", "Border" }) do
		groups[#groups + 1] = "Notify" .. level .. part
	end
end

for _, name in ipairs(groups) do
	local ok, hl = pcall(vim.api.nvim_get_hl, 0, { name = name, link = false })
	if ok and hl.bg then
		hl.bg = nil
		vim.api.nvim_set_hl(0, name, hl)
	end
end
