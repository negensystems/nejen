-- NEJEN Wi-Fi quick-connect (elephant menu provider, surfaced in walker via
-- `nejen open wifi`, which is the waybar network module's click target and the
-- hub's Connect > Wi-Fi entry).
--
-- The inventory comes straight from `nejen wifi list`, so this menu can never
-- drift from what the CLI can actually join. Selecting a network connects it,
-- or disconnects it when it is the one already in use.
--
-- This covers the common case only -- joining a network you can see. Editing
-- profiles, static addressing and VPNs live in nmtui, one entry down at the
-- bottom of the list.
--
-- No entry carries an Icon field, and `nejen open wifi` passes --nosearch:
-- the rows are a status table, and a column of stock GTK icons would be the
-- one thing here not following the NEJEN palette. Matches
-- config/elephant/menus/nejen_bluetooth.lua and nejen_audio.lua.

Name = "nejenwifi"
NamePretty = "Nejen Wi-Fi"
HideFromProviderlist = true
SearchName = true

-- Signal strength and the active network change constantly; a cached menu
-- would lie about both.
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

-- MHz is what nmcli reports; the band is what anyone actually thinks in.
local function band(freq)
  local mhz = tonumber(freq) or 0
  if mhz >= 5925 then
    return "6 GHz"
  elseif mhz >= 4900 then
    return "5 GHz"
  elseif mhz > 0 then
    return "2.4 GHz"
  end
  return ""
end

-- Detail line, in the bar's idiom: dot-namespaced keys, same field order every
-- row so the columns line up when read down the list. LINK.UP first because
-- that is the one fact being acted on, then signal, which is what decides
-- whether joining is worth trying at all.
local function details(inuse, saved, signal, security, freq, rate, chan)
  local parts = {}

  -- Not "OPEN.*": in a Wi-Fi list "open" means unencrypted, and this field is
  -- about whether credentials are already stored. Pairing it with the security
  -- column two fields along would have read as a flat contradiction on every
  -- WPA2 network that has never been joined.
  if inuse == "yes" then
    table.insert(parts, "LINK.UP")
  elseif saved == "yes" then
    table.insert(parts, "SAVED")
  else
    table.insert(parts, "NEW")
  end

  if signal ~= "" then
    table.insert(parts, "SIG " .. signal .. "%")
  end

  -- An empty security field means an unencrypted network. That is worth
  -- calling out plainly rather than leaving as a blank column.
  if security ~= "" then
    table.insert(parts, security:upper())
  else
    table.insert(parts, "UNSECURED")
  end

  local b = band(freq)
  if b ~= "" then
    table.insert(parts, b:upper() .. " CH" .. chan)
  end
  if rate ~= "" then
    table.insert(parts, rate:upper())
  end

  return table.concat(parts, "  ·  ")
end

function GetEntries()
  local nejen = nejen_bin()
  local entries = {}

  for _, line in ipairs(lines_from(nejen .. " wifi list 2>/dev/null")) do
    local ssid, inuse, saved, signal, security, freq, rate, chan =
      line:match("^(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.-)\t(.*)$")
    if ssid then
      table.insert(entries, {
        Text = ssid,
        Subtext = details(inuse, saved, signal, security, freq, rate, chan),
        Value = ssid,
        Actions = {
          activate = "nejen wifi toggle " .. quoted(ssid),
        },
      })
    end
  end

  -- Walker sorts by Text with an empty query box and the menus provider has no
  -- ordering field, so the leading U+25B8 (the hub's own mark for an entry that
  -- opens another surface) is what lands these below every network name. Same
  -- reasoning as nejen_bluetooth.lua.
  table.insert(entries, {
    Text = "▸ Rescan",
    Subtext = "Look for networks again",
    Value = "rescan",
    Actions = {
      activate = "nejen wifi rescan",
    },
  })

  table.insert(entries, {
    Text = "▸ Toggle Wi-Fi",
    Subtext = "Power the radio on or off",
    Value = "power",
    Actions = {
      activate = "nejen wifi power toggle",
    },
  })

  table.insert(entries, {
    Text = "▸ Manage networks",
    Subtext = "Profiles, static addressing, VPN  ·  nmtui",
    Value = "manage",
    Actions = {
      activate = "nejen open wifi --manager",
    },
  })

  return entries
end
