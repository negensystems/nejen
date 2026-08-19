# During installation NEJEN drops a temporary sudoers rule so the machine
# can reboot into the desktop without prompting for a password. That grant
# must not outlive the first login, so retire it here.
sudo rm -f /etc/sudoers.d/99-nejen-installer-reboot
