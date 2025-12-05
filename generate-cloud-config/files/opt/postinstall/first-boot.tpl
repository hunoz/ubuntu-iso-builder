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

echo "Configuring docker"
/usr/local/bin/setup-raid
echo "Docker has been configured"

echo "Setting up raid"
/usr/local/bin/setup-raid
echo "Raid setup is complete"

# Disable this service after first run
echo "Creating systemd signal file to not run again"
echo "yes" > /var/lib/first-boot-complete
echo "Signal file created"
echo "First boot completed at $(date)"