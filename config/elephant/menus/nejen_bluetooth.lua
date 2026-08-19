-- NEJEN bluetooth quick-connect (elephant menu provider, surfaced in walker
-- via `nejen open bluetooth`, which is the waybar bluetooth module's click
-- target and the hub's Connect > Bluetooth entry).
--
-- The inventory comes straight from `nejen bluetooth list`, so this menu can
-- never drift from what the CLI can actually connect to. Selecting a device
-- toggles it: connected devices disconnect, disconnected ones connect.
--
-- This covers the common case only -- reconnecting something already paired.
-- Pairing, scanning, renaming and file transfer live in bluetuith, one entry
-- down at the bottom of the list.

Name = "nejenbluetooth"
NamePretty = "Nejen Bluetooth"
HideFromProviderlist = true
SearchName = true

-- Connection state changes constantly; a cached menu would lie about it.
Cache = false


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

-- The freedesktop icon name BlueZ reports doubles as a readable device kind.
-- No entry here carries an Icon field: the rows read as a status table, and a
-- column of stock colour icons pulled from the GTK icon theme is the one thing
-- in the menu that would not be following the NEJEN palette.
local kinds = {
  ["audio-headset"] = "HEADSET",
  ["audio-headphones"] = "HEADPHONES",
  ["audio-card"] = "AUDIO",
  ["input-mouse"] = "MOUSE",
  ["input-keyboard"] = "KEYBOARD",
  ["input-gaming"] = "CONTROLLER",
  ["input-tablet"] = "TABLET",
  ["phone"] = "PHONE",
  ["computer"] = "COMPUTER",
  ["video-display"] = "DISPLAY",
  ["printer"] = "PRINTER",
}

-- Detail line, built in the bar's own idiom: dot-namespaced keys, values in
-- the same order every row so the columns line up when read down the list.
-- LINK.UP / LINK.DOWN first because that is the one fact being acted on.
local function details(connected, icon, battery, flags, profiles, mac)
  local parts = {}

  if connected == "yes" then
    table.insert(parts, "LINK.UP")
  else
    table.insert(parts, "LINK.DOWN")
  end

  table.insert(parts, kinds[icon] or "DEVICE")

  if battery ~= "" then
    table.insert(parts, "BAT " .. battery .. "%")
  end
  if profiles ~= "" then
    table.insert(parts, (profiles:gsub("%+", " ")))
  end
  if flags ~= "" then
    table.insert(parts, (flags:gsub("%+", " ")))
  end

  table.insert(parts, mac)

  return table.concat(parts, "  ·  ")
end

function GetEntries()
  local nejen = nejen_bin()
  local entries = {}

  for _, line in ipairs(lines_from(nejen .. " bluetooth list 2>/dev/null")) do
    local mac, alias, connected, icon, battery, flags, profiles =
      line:match("^(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.*)$")
    if mac then
      table.insert(entries, {
        Text = alias,
        Subtext = details(connected, icon, battery, flags, profiles, mac),
        Value = mac,
        Actions = {
          activate = "nejen bluetooth toggle " .. quoted(mac),
        },
      })
    end
  end

  -- Walker sorts by Text with an empty query box, and neither a per-entry
  -- Score nor a provider-level Sort flag overrides that -- both were tried,
  -- and the menus provider carries no such field. The sort is by codepoint
  -- though, so the leading U+25B8 (which is what the hub already uses to
  -- mark an entry that opens another surface) both reads correctly and
  -- lands these two below every device name.
  table.insert(entries, {
    Text = "▸ Manage devices",
    Subtext = "Pair, scan, rename, send files  ·  bluetuith",
    Value = "manage",
    Actions = {
      activate = "nejen open bluetooth --manager",
    },
  })

  table.insert(entries, {
    Text = "▸ Toggle bluetooth",
    Subtext = "Power the adapter on or off",
    Value = "power",
    Actions = {
      activate = "nejen bluetooth power toggle",
    },
  })

  return entries
end
