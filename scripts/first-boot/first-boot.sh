#!/bin/bash

# Logging
LOG_FILE="/var/log/first-boot.log"
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

# First boot script
echo 'Running first boot configuration...'

# Fix ownership of user home directory
echo 'Fixing ownership of user home directory...'
chown -R {{ admin_username }}:{{ admin_username }} /home/{{ admin_username }}
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
curl https://github.com/hunoz/ubuntu-iso-builder/scripts/first-boot/raid/raid-setup.sh | bash
echo "RAID setup complete"

# Disable this service after first run
echo 'Disabling first-boot service...'
systemctl disable first-boot.service
echo 'Finished disabling first-boot service.'
echo "First boot completed at $(date)"
