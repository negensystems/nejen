# Elephant backs the launcher with data providers; register its user unit
# and bring it up right away so the first launcher invocation is warm.
elephant service enable
systemctl --user start elephant.service
