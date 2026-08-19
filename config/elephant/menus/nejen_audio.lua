-- NEJEN audio quick-select (elephant menu provider, surfaced in walker via
-- `nejen open audio`, which is the waybar volume module's click target and the
-- hub's Connect > Audio entry).
--
-- The inventory comes straight from `nejen audio list`, so this menu can never
-- drift from what the CLI can actually switch to. Selecting a device makes it
-- the default for its kind, outputs and inputs alike.
--
-- This covers the common case only -- moving playback to another device, or
-- silencing the mic. Per-stream routing, profiles and levels live in wiremix,
-- one entry down at the bottom of the list.
--
-- No entry carries an Icon field, and `nejen open audio` passes --nosearch:
-- the rows are a status table of the handful of devices attached to this
-- machine, and a column of stock GTK icons would be the one thing here not
-- following the NEJEN palette. Matches config/elephant/menus/nejen_bluetooth.lua.

Name = "nejenaudio"
NamePretty = "Nejen Audio"
HideFromProviderlist = true
SearchName = true

-- Volume and default-device state change constantly; a cached menu would lie.
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

-- A card's own description is the same string for its output and its input
-- ("Ryzen HD Audio Controller Analog Stereo" for both), so titling rows with
-- it renders the two halves of this menu as a pair of identical-looking
-- duplicates. The active port is the physical thing being selected -- Speakers,
-- Headphones, Internal Microphone -- and is what the title should say. Devices
-- with no port at all (USB interfaces, some Bluetooth sinks) keep the
-- description, which is their only identity.
local function title(description, port)
  if port ~= "" then
    return port
  end
  return description
end

-- Detail line, in the bar's idiom: dot-namespaced keys, same field order every
-- row so the columns line up when read down the list. OUT/IN first because
-- which direction a device carries is the thing that separates the two halves
-- of this menu; DEFAULT vs READY is what the row acts on.
local function details(kind, isdefault, volume, muted, description, port, spec)
  local parts = {}

  local direction = kind == "sink" and "OUT" or "IN"
  if isdefault == "yes" then
    table.insert(parts, direction .. ".DEFAULT")
  else
    table.insert(parts, direction .. ".READY")
  end

  if muted == "yes" then
    table.insert(parts, "MUTED")
  elseif volume ~= "" then
    table.insert(parts, "VOL " .. volume .. "%")
  end

  -- Carried here rather than in the title, so the card stays identifiable
  -- when the port is what names the row. Skipped when the title already is
  -- the description, which would otherwise print it twice on one row.
  if port ~= "" and description ~= "" then
    table.insert(parts, description:upper())
  end
  if spec ~= "" then
    table.insert(parts, spec:upper())
  end

  return table.concat(parts, "  ·  ")
end

function GetEntries()
  local nejen = nejen_bin()
  local entries = {}

  for _, line in ipairs(lines_from(nejen .. " audio list 2>/dev/null")) do
    local kind, name, description, isdefault, volume, muted, port, spec =
      line:match("^(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.*)$")
    if kind then
      table.insert(entries, {
        Text = title(description, port),
        Subtext = details(kind, isdefault, volume, muted, description, port, spec),
        Value = name,
        Actions = {
          activate = "nejen audio default " .. kind .. " " .. quoted(name),
        },
      })
    end
  end

  -- Walker sorts by Text with an empty query box and the menus provider has no
  -- ordering field, so the leading U+25B8 (the hub's own mark for an entry that
  -- opens another surface) is what lands these below every device name. Same
  -- reasoning as nejen_bluetooth.lua.
  table.insert(entries, {
    Text = "▸ Mute output",
    Subtext = "Silence the default output device",
    Value = "mute-sink",
    Actions = {
      activate = "nejen audio mute sink",
    },
  })

  table.insert(entries, {
    Text = "▸ Mute input",
    Subtext = "Silence the default microphone",
    Value = "mute-source",
    Actions = {
      activate = "nejen audio mute source",
    },
  })

  table.insert(entries, {
    Text = "▸ Mixer",
    Subtext = "Per-stream routing, levels, profiles  ·  wiremix",
    Value = "mixer",
    Actions = {
      activate = "nejen open audio --mixer",
    },
  })

  return entries
end
