# Enable systemd-timesyncd for time synchronization.
# Always exit 0 to prevent aborting the first-run sequence.

if ! command -v timedatectl >/dev/null 2>&1; then
  echo "nejen: timedatectl not available; skipping time sync"
  exit 0
fi

# Don't retune a machine that already runs a full NTP daemon.
for unit in chronyd.service ntpd.service openntpd.service; do
  if systemctl is-enabled --quiet "$unit" 2>/dev/null ||
    systemctl is-active --quiet "$unit" 2>/dev/null; then
    echo "nejen: ${unit%.service} already manages the clock; leaving time sync alone"
    exit 0
  fi
done

# set-ntp both enables the unit and flips the flag timedatectl reports.
if sudo timedatectl set-ntp true; then
  echo "nejen: system clock now synchronized over NTP"
else
  # Send a notification if enabling NTP fails.
  echo "nejen: could not enable NTP"
  notify-send "􀀣  Clock Not Syncing" "The system clock is not synchronized and will drift. Once you are online, run: sudo timedatectl set-ntp true" -u critical
fi

# We don't touch the RTC to avoid dual-boot issues.
