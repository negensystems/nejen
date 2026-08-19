return {
  "stevearc/conform.nvim",
  opts = {
    formatters_by_ft = {
      lua = {
        "stylua",
        "luaformatter",
      },

      c = { "clang-format" },
      cpp = { "clang-format" },

      javascript = { "prettier" },
      javascriptreact = { "prettier" },
      typescript = { "prettier" },
      typescriptreact = { "prettier" },
      html = { "prettier" },
      css = { "prettier" },

      json = { "fixjson" },
      jsonc = { "fixjson" },

      yaml = { "yamlfmt", "yamlfix" },

      sh = { "shfumpt" },
      python = { "black" },

      go = { "gofumpt", "goimports" },

      -- SQL
      sql = { "sql-formatter" },

      --arduino
      arduino = { "clang-format" },
    },

    format_on_save = {
      timeout_ms = 800,
      lsp_fallback = true,
    },
  },
}
