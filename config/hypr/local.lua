-- Machine-local additions kept out of the shipped shell defaults.

-- NVIDIA
hl.env("NVD_BACKEND", "direct")
hl.env("LIBVA_DRIVER_NAME", "nvidia")
hl.env("__GLX_VENDOR_LIBRARY_NAME", "nvidia")

-- This machine's extras
hl.on("hyprland.start", function()
    hl.exec_cmd("nejen lock")
    hl.exec_cmd("env GDK_BACKEND=x11 solaar --window=hide")
end)
