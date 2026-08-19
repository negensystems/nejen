# Bring GTK/GNOME apps in line with NEJEN's dark-by-default look and give
# them the Yaru blue icon set. Themed switching later is handled by the
# theme engine; this just sets a sane baseline.
gsettings set org.gnome.desktop.interface color-scheme "prefer-dark"
gsettings set org.gnome.desktop.interface gtk-theme "Adwaita-dark"
gsettings set org.gnome.desktop.interface icon-theme "Yaru-blue"

# Rebuild the icon cache so freshly installed Yaru icons resolve immediately.
sudo gtk-update-icon-cache /usr/share/icons/Yaru
