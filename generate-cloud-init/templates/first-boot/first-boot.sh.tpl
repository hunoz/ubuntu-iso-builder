#!/bin/bash

# Logging
LOG_FILE="/var/log/first-boot.log"
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

# First boot script
echo 'Running first boot configuration...'

# Fix ownership of user home directory
echo 'Fixing ownership of user home directory...'
chown -R {{ .AdminUsername }}:{{ .AdminUsername }} /home/{{ .AdminUsername }}
echo 'Finished fixing ownership of user home directory.'
echo 'Disabling motd...'
chmod -x /etc/update-motd.d/*
echo 'Finished disabling motd.'
echo 'Enabling landscape motd...'
chmod +x /etc/update-motd.d/50-landscape-sysinfo
echo 'Finished enabling landscape motd.'

# Update packages, but sleep first to give time for drivers to configure
echo 'Updating packages...'
sleep 30
apt-get update
apt-get upgrade -y
echo 'Finished updating packages.'

echo "Setting up RAID"
cat > /opt/post-install/raid-setup.sh << 'EOF'
{{ template "raid-setup" . }}
EOF
chmod +x /opt/post-install/raid-setup.sh
/opt/post-install/raid-setup.sh
echo "RAID setup complete"

echo "Setting up Docker"
cat > /opt/post-install/configure-docker.sh << 'EOF'
{{ template "configure-docker" . }}
EOF
chmod +x /opt/post-install/configure-docker.sh
/opt/post-install/configure-docker.sh
echo "Docker setup complete"

# Disable this service after first run
echo 'Disabling first-boot service...'
systemctl disable first-boot.service.tpl
echo 'Finished disabling first-boot service.'
echo "First boot completed at $(date)"
