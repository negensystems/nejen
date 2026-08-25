# NEJEN's default firewall stance: outbound traffic is unrestricted,
# inbound is dropped unless a rule below opens it.
sudo ufw default deny incoming
sudo ufw default allow outgoing

# LocalSend discovers and transfers over 53317 on both protocols.
sudo ufw allow 53317/tcp
sudo ufw allow 53317/udp

# Containers resolve names through the DNS stub on the docker0 gateway;
# without this rule the deny-incoming default silently breaks their DNS.
sudo ufw allow in proto udp from 172.16.0.0/12 to 172.17.0.1 port 53 comment 'nejen: docker container dns'

# Activate now and on every subsequent boot.
sudo ufw --force enable
sudo systemctl enable ufw

# Docker publishes ports by rewriting iptables behind ufw's back;
# ufw-docker installs the counter-rules that close that hole.
sudo ufw-docker install
sudo ufw reload
