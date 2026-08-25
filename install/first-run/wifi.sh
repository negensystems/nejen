# Nudge the fresh install toward its first system update - and, if the
# machine is offline, toward getting Wi-Fi connected first. Connectivity
# probe: a few short pings to Cloudflare's anycast resolver.
if ping -c3 -W1 1.1.1.1 >/dev/null 2>&1; then
  notify-send "    Update System" "Run your first update from the hub: Super + N, then Update." -u critical
else
  notify-send "    Update System" "Once you're online, update from the hub: Super + N, then Update." -u critical
  notify-send "󰖩    Set Up Wi-Fi" "Open the Wi-Fi manager from the hub (Super + N). Tab moves, Space selects, ? shows help." -u critical
fi
