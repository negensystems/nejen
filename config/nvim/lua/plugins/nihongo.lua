return {
  {
    "takumiymd/nihongo.nvim",
    name = "nihongo.nvim",
    event = "VeryLazy",
    opts = {
      motion = {
        enabled = true,
        override_defaults = true,
      },
      ime = {
        enabled = true,
        method = "fcitx5", -- Options: "fcitx5", "ibus", "im-select", or "custom"
        -- default_ime = "1",
        -- japanese_ime = "2",
      },
    },
  },
}
