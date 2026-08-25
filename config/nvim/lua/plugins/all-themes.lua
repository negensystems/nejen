-- Colorscheme plugins the NEJEN theme engine references.
--
-- The shipped "nejen" theme styles Neovim entirely through tokyonight
-- (configured in plugins/theme.lua); its rendered overrides file
-- (~/.local/state/nejen/theme/neovim.lua) is empty, so no other colorscheme
-- plugin is required. User themes whose neovim.lua names a different
-- colorscheme must install that plugin themselves, e.g. with a spec in
-- ~/.config/nvim/lua/plugins/.
return {
	{ "folke/tokyonight.nvim", lazy = true, priority = 1000 },
}
