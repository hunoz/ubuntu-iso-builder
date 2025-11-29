from typing import Any, TypedDict

FIRST_BOOT_COMMANDS = [
    '''cat > /target/opt/post-install/first-boot.sh << 'FIRSTBOOT'
#!/bin/bash
# First boot script
echo "Running first boot configuration..."

# Example: Update system
apt-get update
apt-get upgrade -y

# Add your custom commands here
echo "First boot completed at $(date)" >> /var/log/first-boot.log

# Disable this service after first run
systemctl disable first-boot.service
FIRSTBOOT
    ''',
    'curtin in-target -- chmod +x /opt/post-install/first-boot.sh',
    '''cat > /target/etc/systemd/system/first-boot.service << 'SERVICE'
[Unit]
Description=First Boot Configuration
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/opt/post-install/first-boot.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
SERVICE
    ''',
    'curtin in-target -- systemctl enable first-boot.service',
]

USERPROFILE_COMMANDS = [
    '''cat >> /target/home/{{ admin_username }}/.bashrc << 'BASHRC'
# Custom bash configuration
export EDITOR=vim
export VISUAL=vim

# Custom aliases
alias ll='ls -lah'
alias update='sudo apt-get update && sudo apt-get upgrade'

# Custom prompt
PS1='\[\\033[01;32m\]\\u@\h\[\\033[00m\]:\[\\033[01;34m\]\w\[\\033[00m\]\$ '
BASHRC
        ''',
    '''cat > /target/home/{{ admin_username }}/.vimrc << 'VIMRC'
set number
set expandtab
set tabstop=4
set shiftwidth=4
syntax on
set background=dark
VIMRC
        ''',
    '''cat > /target/etc/profile.d/custom-settings.sh << 'PROFILE'
#!/bin/bash
# System-wide custom settings
export HISTSIZE=10000
export HISTFILESIZE=20000
PROFILE
        ''',
    'curtin in-target -- chown -R {{ admin_username }}:{{ admin_username }} /home/{{ admin_username }}',
]
