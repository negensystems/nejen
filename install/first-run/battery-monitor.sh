# Pick a power-profile default that suits the hardware: laptops get a
# balanced profile plus the low-battery watchdog timer, desktops (no
# battery) simply run at full tilt. powerprofilesctl failures are
# non-fatal - not every machine ships a profile daemon.
if nejen battery present; then
  powerprofilesctl set balanced || true

  # Poll charge level every 30s and warn when it runs low.
  systemctl --user enable --now nejen-battery-monitor.timer
else
  powerprofilesctl set performance || true
fi
