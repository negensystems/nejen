return {
  {
    "folke/tokyonight.nvim",
    lazy = false,
    priority = 1000,
    opts = {
      style = "night",
      transparent = true,

      styles = {
        sidebars = "transparent",
        floats = "transparent",
      },

      on_colors = function(colors)
        colors.bg = "#08080a"
        colors.bg_dark = "#070406"
        colors.bg_float = "#0e0d12"
        colors.bg_highlight = "#1f2030"

        colors.fg = "#c7d0ff"
        colors.fg_dark = "#767ca0"

        colors.blue = "#4a4f8a"
        colors.cyan = "#6f8fbf"
        colors.purple = "#7c6aa6"
        colors.red = "#ff3030"

        colors.border = "#393a54"
      end,

      on_highlights = function(hl, c)
        local nejen = {
          bg0 = "#08080a",
          bg1 = "#0e0d12",
          bg2 = "#141625",
          panel = "#1f2030",
          panel2 = "#29293f",

          accent = "#4a4f8a",
          blue = "#6f8fbf",
          purple = "#7c6aa6",
          red = "#ff3030",

          text = "#c7d0ff",
          muted = "#767ca0",
          border = "#393a54",
        }

        hl.Normal = { fg = nejen.text, bg = nejen.bg0 }
        hl.NormalNC = { fg = nejen.text, bg = nejen.bg0 }

        hl.NormalFloat = { fg = nejen.text, bg = nejen.bg1 }
        hl.FloatBorder = { fg = nejen.accent, bg = nejen.bg1 }

        hl.CursorLine = { bg = nejen.panel }
        hl.CursorLineNr = { fg = nejen.blue, bold = true }
        hl.LineNr = { fg = nejen.muted }

        hl.Visual = { bg = nejen.accent }
        hl.Search = { fg = nejen.bg0, bg = nejen.blue }
        hl.IncSearch = { fg = nejen.bg0, bg = nejen.red }

        hl.WinSeparator = { fg = nejen.border }
        hl.StatusLine = { fg = nejen.text, bg = nejen.panel }
        hl.StatusLineNC = { fg = nejen.muted, bg = nejen.bg1 }

        hl.Pmenu = { fg = nejen.text, bg = nejen.bg1 }
        hl.PmenuSel = { fg = nejen.text, bg = nejen.accent, bold = true }
        hl.PmenuThumb = { bg = nejen.accent }

        hl.TelescopeNormal = { fg = nejen.text, bg = nejen.bg1 }
        hl.TelescopeBorder = { fg = nejen.accent, bg = nejen.bg1 }
        hl.TelescopeSelection = { fg = nejen.text, bg = nejen.panel2, bold = true }
        hl.TelescopePromptBorder = { fg = nejen.blue, bg = nejen.bg1 }
        hl.TelescopePromptTitle = { fg = nejen.bg0, bg = nejen.blue, bold = true }
        hl.TelescopeResultsTitle = { fg = nejen.bg0, bg = nejen.accent, bold = true }
        hl.TelescopePreviewTitle = { fg = nejen.bg0, bg = nejen.purple, bold = true }

        hl.WhichKeyFloat = { bg = nejen.bg1 }
        hl.WhichKeyBorder = { fg = nejen.accent, bg = nejen.bg1 }

        hl.NeoTreeNormal = { fg = nejen.text, bg = nejen.bg0 }
        hl.NeoTreeNormalNC = { fg = nejen.text, bg = nejen.bg0 }
        hl.NeoTreeWinSeparator = { fg = nejen.border, bg = nejen.bg0 }
        hl.NeoTreeDirectoryName = { fg = nejen.blue, bold = true }
        hl.NeoTreeGitModified = { fg = nejen.purple }

        hl.Comment = { fg = nejen.muted, italic = true }

        hl.DiagnosticError = { fg = nejen.red }
        hl.DiagnosticWarn = { fg = "#c08b5c" }
        hl.DiagnosticInfo = { fg = nejen.blue }
        hl.DiagnosticHint = { fg = nejen.purple }
        -- Softer markdown / inline-code blocks
        -- hl["@markup.raw.markdown_inline"] = { fg = nejen.blue, bg = nejen.bg1 }
        hl["@markup.raw.markdown"] = { fg = nejen.blue, bg = nejen.bg1 }
        hl.markdownCode = { fg = nejen.blue, bg = nejen.bg1 }
        -- hl.markdownCodeBlock = { fg = nejen.blue, bg = nejen.bg0 }
        -- hl.markdownCodeDelimiter = { fg = nejen.muted, bg = nejen.bg1 }
        hl["@markup.raw.markdown_inline"] = { fg = nejen.blue, bg = "NONE" }
        hl.markdownCode = { fg = nejen.blue, bg = "NONE" }
        hl.markdownCodeDelimiter = { fg = nejen.muted, bg = "NONE" }

        -- Softer headings, removes the heavy boxed look
        hl["@markup.heading.1.markdown"] = { fg = "#d59a5a", bold = true }
        hl["@markup.heading.2.markdown"] = { fg = "#d59a5a", bold = true }
        hl["@markup.heading.3.markdown"] = { fg = "#d59a5a", bold = true }
        hl.markdownH1 = { fg = "#d59a5a", bold = true }
        hl.markdownH2 = { fg = "#d59a5a", bold = true }
        hl.markdownH3 = { fg = "#d59a5a", bold = true }

        -- Softer search/current-word highlights
        hl.Search = { fg = nejen.text, bg = "#252946" }
        hl.IncSearch = { fg = nejen.bg0, bg = nejen.red }
        hl.CurSearch = { fg = nejen.bg0, bg = nejen.red }

        -- LSP reference highlights, prevents random word boxes from looking too loud
        hl.LspReferenceText = { bg = "#1b1d31" }
        hl.LspReferenceRead = { bg = "#1b1d31" }
        hl.LspReferenceWrite = { bg = "#252946" }

        -- Visual selection should be readable, not nuclear
        hl.Visual = { fg = nejen.text, bg = "#2b315f" }
      end,
    },
  },

  {
    "LazyVim/LazyVim",
    opts = {
      colorscheme = "tokyonight-night",
    },
  },
}
