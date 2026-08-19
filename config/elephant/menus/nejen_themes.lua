-- NEJEN theme switcher (elephant menu provider, surfaced in walker via
-- `nejen open launcher -m menus:nejenthemes`).
--
-- The theme inventory comes straight from `nejen theme list` (which already
-- merges shipped and user themes), so this menu can never drift from the
-- CLI's idea of what is installed. Selecting an entry runs `nejen theme set`.

Name = "nejenthemes"
NamePretty = "Nejen Themes"
HideFromProviderlist = true

local function exists(path)
  local f = io.open(path, "r")
  if not f then
    return false
  end
  f:close()
  return true
end

local function lines_from(cmd)
  local collected = {}
  local pipe = io.popen(cmd)
  if not pipe then
    return collected
  end
  for line in pipe:lines() do
    if line ~= "" then
      table.insert(collected, line)
    end
  end
  pipe:close()
  return collected
end

-- "rose-pine" -> "Rose Pine"
local function prettify(name)
  local words = name:gsub("[-_]+", " ")
  return (words:gsub("(%w)(%S*)", function(head, tail)
    return head:upper() .. tail:lower()
  end))
end

function GetEntries()
  local home = os.getenv("HOME")
  local nejen_path = os.getenv("NEJEN_PATH") or (home .. "/.local/share/nejen")
  -- A dev-mode tree keeps the dispatcher under NEJEN_PATH/bin; a packaged
  -- install ships it at /usr/bin/nejen and has no bin/ of its own, so fall
  -- back to the bare name on PATH. The other menus already guard this way.
  local nejen = nejen_path .. "/bin/nejen"
  if not exists(nejen) then
    nejen = "nejen"
  end

  -- User layer first so its preview art wins for same-named themes.
  local roots = {
    home .. "/.config/nejen/themes",
    nejen_path .. "/themes",
  }

  -- Dedicated preview image if the theme ships one, else its first wallpaper.
  local function preview_for(theme)
    for _, root in ipairs(roots) do
      local dir = root .. "/" .. theme
      for _, candidate in ipairs({ dir .. "/preview.png", dir .. "/preview.jpg" }) do
        if exists(candidate) then
          return candidate
        end
      end
      local first = lines_from(
        "find -L '" .. dir .. "/backgrounds' -maxdepth 1 -type f 2>/dev/null | sort | head -n1"
      )[1]
      if first then
        return first
      end
    end
    return nil
  end

  local entries = {}

  for _, theme in ipairs(lines_from("'" .. nejen .. "' theme list 2>/dev/null")) do
    local entry = {
      Text = prettify(theme),
      Value = theme,
      Actions = {
        activate = "nejen theme set '" .. theme .. "'",
      },
    }
    local preview = preview_for(theme)
    if preview then
      entry.Preview = preview
      entry.PreviewType = "file"
    end
    table.insert(entries, entry)
  end

  return entries
end
