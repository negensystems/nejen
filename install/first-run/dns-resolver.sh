# Hand DNS lookups to systemd-resolved: /etc/resolv.conf becomes a symlink
# to the stub resolver, which is the setup the Arch wiki recommends for
# systemd-resolved (see Systemd-resolved on the wiki).
echo "Pointing /etc/resolv.conf at the systemd-resolved stub"

sudo ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf
