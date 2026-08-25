-- Options are automatically loaded before lazy.nvim startup
-- Default options that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/options.lua
-- Add any additional options here
vim.opt.relativenumber = true

-- Soft wrap: don't cut words in half, keep indent on wrapped lines
vim.opt.wrap = true
vim.opt.linebreak = true
vim.opt.breakindent = true
vim.opt.breakindentopt = "shift:2,min:20,sbr"
vim.opt.showbreak = "↳ " -- visual marker at wrap

-- Optional: horizontal context if you ever turn nowrap back on
-- vim.opt.sidescrolloff = 8
