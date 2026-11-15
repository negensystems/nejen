-- Third-party app rules, grouped by concern rather than one file per app.

local function resolveAppsDir()
    local src = debug and debug.getinfo(1, "S").source or ""
    if src:sub(1, 1) == "@" then
        local dir = src:sub(2):match("^(.*)/[^/]+$")
        if dir and dir ~= "" then
            return dir .. "/apps"
        end
    end
    local home = os.getenv("HOME") or ""
    local nejenPath = os.getenv("NEJEN_PATH")
    if not nejenPath or nejenPath == "" then
        nejenPath = home .. "/.local/share/nejen"
    end
    return nejenPath .. "/config/hypr/apps"
end

local appsDir = resolveAppsDir()

dofile(appsDir .. "/browsers.lua")
dofile(appsDir .. "/terminals.lua")
dofile(appsDir .. "/media.lua")
dofile(appsDir .. "/privacy.lua")
dofile(appsDir .. "/dialogs.lua")
