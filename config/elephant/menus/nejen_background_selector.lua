-- NEJEN wallpaper picker (elephant menu provider, surfaced in walker via
-- `nejen theme bg pick`, which is Super+Shift+W and the hub's
-- Theme > Wallpaper > Browse entry).
--
-- Both the inventory and the display names come straight from
-- `nejen theme bg list`, so this menu can never drift from what the CLI
-- cycles through: the user's own folder (~/.config/nejen/backgrounds), then
-- the active theme's backgrounds, with the user layer shadowing same-named
-- files. Selecting an entry runs `nejen theme bg set`.

Name = "nejenbackgrounds"
NamePretty = "Nejen Wallpapers"
Cache = false
HideFromProviderlist = true
SearchName = true

local function quoted(s)
  return "'" .. s:gsub("'", "'\\''") .. "'"
end

local function exists(path)
  local f = io.open(path, "r")
  if not f then
    return false
  end
  f:close()
  return true
end

-- elephant runs as a user service whose PATH does not include ~/.local/bin,
-- so entry generation has to reach the dispatcher by absolute path. Actions
-- are executed in a login context and can use the bare name.
local function nejen_bin()
  local home = os.getenv("HOME")
  local nejen_path = os.getenv("NEJEN_PATH") or (home .. "/.local/share/nejen")
  local candidate = nejen_path .. "/bin/nejen"
  if exists(candidate) then
    return quoted(candidate)
  end
  return "nejen"
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

local function basename(path)
  return path:match("([^/]+)$")
end

function GetEntries()
  local nejen = nejen_bin()

  -- Paths and display names come from the same command in the same order,
  -- so they pair up by index. Deriving the labels here instead would mean
  -- keeping a second copy of the CLI's naming rules in Lua.
  local paths = lines_from(nejen .. " theme bg list 2>/dev/null")
  local names = lines_from(nejen .. " theme bg list --names 2>/dev/null")

  -- `theme bg current` reports a fully resolved path while `theme bg list`
  -- reports it as reached through the layer it came from, so the two only
  -- line up on the filename. Filenames are unique across the list: the CLI
  -- shadows same-named files between layers.
  local current = lines_from(nejen .. " theme bg current 2>/dev/null")[1]
  local current_name = current and basename(current) or nil

  local entries = {}
  for i, path in ipairs(paths) do
    -- Falls back to the filename if the two listings disagree, which can
    -- only happen if the folder changed between the two calls.
    local label = names[i] or basename(path)
    if current_name and basename(path) == current_name then
      label = label .. "  (current)"
    end
    table.insert(entries, {
      Text = label,
      Value = path,
      Preview = path,
      PreviewType = "file",
      Actions = {
        activate = "nejen theme bg set " .. quoted(path),
      },
    })
  end

  return entries
end
