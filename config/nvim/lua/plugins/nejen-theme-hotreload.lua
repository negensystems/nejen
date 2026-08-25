-- NEJEN theme hot-reload.
--
-- Contract with `nejen theme set`: switching themes re-renders the templates
-- into ~/.local/state/nejen/theme/, including neovim.lua (a lazy-style spec
-- of per-theme Neovim overrides). Nothing pings running editors, so this
-- plugin watches that file on disk and re-applies the colorscheme (plus the
-- transparency pass) whenever a fresh render lands.

return {
	{
		name = "nejen-theme-hotreload",
		dir = vim.fn.stdpath("config"),
		lazy = false,
		priority = 500,
		config = function()
			local uv = vim.uv or vim.loop
			-- stdpath("state") is $XDG_STATE_HOME/nvim; nejen renders next door.
			local theme_dir = vim.fn.fnamemodify(vim.fn.stdpath("state"), ":h") .. "/nejen/theme"
			local theme_file = theme_dir .. "/neovim.lua"
			local transparency = vim.fn.stdpath("config") .. "/plugin/after/transparency.lua"

			-- Pull a colorscheme name out of the rendered overrides, if any.
			-- Accepted shapes: { colorscheme = "x" }, or a spec list holding
			-- { ..., opts = { colorscheme = "x" } } (the plugins/theme.lua shape).
			local function rendered_colorscheme()
				local ok, spec = pcall(dofile, theme_file)
				if not ok or type(spec) ~= "table" then
					return nil
				end
				if type(spec.colorscheme) == "string" then
					return spec.colorscheme
				end
				for _, entry in ipairs(spec) do
					if type(entry) == "table" and type(entry.opts) == "table" and type(entry.opts.colorscheme) == "string" then
						return entry.opts.colorscheme
					end
				end
				return nil
			end

			local function apply()
				-- The shipped nejen theme renders an empty overrides file (it
				-- themes Neovim through plugins/theme.lua), so fall back to
				-- re-applying whatever scheme is already active.
				local scheme = rendered_colorscheme() or vim.g.colors_name or "default"
				pcall(vim.cmd.colorscheme, scheme)
				if vim.fn.filereadable(transparency) == 1 then
					pcall(vim.cmd.source, transparency)
				end
				vim.cmd("redraw!")
			end

			local watcher = uv.new_fs_event()
			local debounce = uv.new_timer()
			if not (watcher and debounce) then
				return
			end

			-- Watch the directory rather than the file: the renderer replaces
			-- the file, and a watch pinned to the old inode would go stale.
			local started = pcall(function()
				watcher:start(theme_dir, {}, function(_, filename)
					if filename ~= nil and filename ~= "neovim.lua" then
						return
					end
					-- A render touches the file more than once; collapse the burst.
					debounce:start(100, 0, function()
						debounce:stop()
						vim.schedule(apply)
					end)
				end)
			end)
			if not started then
				-- No theme rendered yet; the next nvim start will pick it up.
				return
			end

			vim.api.nvim_create_autocmd("VimLeavePre", {
				callback = function()
					watcher:stop()
				end,
			})
		end,
	},
}
