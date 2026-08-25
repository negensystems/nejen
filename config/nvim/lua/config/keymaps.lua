-- Keymaps are automatically loaded on the VeryLazy event
-- Default keymaps that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/keymaps.lua
-- Add any additional keymaps here
--
-- Toggle soft wrap quickly
vim.keymap.set("n", "<leader>tw", function()
  local new = not vim.wo.wrap
  vim.wo.wrap = new
  vim.wo.linebreak = new
  vim.wo.breakindent = new
  vim.notify("wrap: " .. (new and "ON" or "OFF"))
end, { desc = "Toggle Wrap" })

-- Make j/k move by *screen* lines when wrap is ON (normal behavior when OFF)
-- This feels natural in wrapped files.
local function set_wrap_motions(buf)
  vim.keymap.set("n", "j", function()
    return vim.wo.wrap and "gj" or "j"
  end, { expr = true, buffer = buf, desc = "Down (wrap-aware)" })
  vim.keymap.set("n", "k", function()
    return vim.wo.wrap and "gk" or "k"
  end, { expr = true, buffer = buf, desc = "Up (wrap-aware)" })
end

-- Apply wrap-aware motions whenever a buffer is shown
vim.api.nvim_create_autocmd("BufWinEnter", {
  callback = function(args)
    set_wrap_motions(args.buf)
  end,
})

-- delete without yanking
vim.keymap.set("n", "dd", '"_dd', { noremap = true, silent = true })
vim.keymap.set("v", "d", '"_d', { noremap = true, silent = true })
