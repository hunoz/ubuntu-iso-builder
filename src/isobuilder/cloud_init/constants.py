DOCKER_COMMANDS = [
    'curtin in-target -- apt update',
    'curtin in-target -- install -m 0755 -d /etc/apt/keyrings',
    'curtin in-target -- curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc',
    'curtin in-target -- chmod a+r /etc/apt/keyrings/docker.asc',
    '''curtin in-target -- sh -c "cat > /etc/apt/sources.list.d/docker.sources << 'DOCKER'
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
DOCKER
    "
    ''',
    'curtin in-target -- apt update',
    'curtin in-target -- bash -c \'DEBIAN_FRONTEND=noninteractive apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin\'',
    'curtin in-target -- systemctl enable docker',
]

NVIDIA_COMMANDS = [
    '''curtin in-target -- sh -c "cat > /etc/modprobe.d/blacklist-nouveau.conf << 'EOF'
blacklist nouveau
options nouveau modeset=0
EOF
    "
    ''',
    'curtin in-target -- update-initramfs -u',
    'curtin in-target -- apt-get update',
    'curtin in-target -- bash -c \'DEBIAN_FRONTEND=noninteractive ubuntu-drivers install --gpgpu\'',
    'curtin in-target -- bash -c \'curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg\'',
    'curtin in-target -- bash -c \'curl -fsSL https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list -o /etc/apt/sources.list.d/nvidia-container-toolkit.list\'',
    'curtin in-target -- bash -c \'sed -i "s|^deb |deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] |" /etc/apt/sources.list.d/nvidia-container-toolkit.list\'',
    'curtin in-target -- apt-get update',
    'curtin in-target -- bash -c \'DEBIAN_FRONTEND=noninteractive apt-get install -y nvidia-container-toolkit\'',
    'curtin in-target -- nvidia-ctk runtime configure --runtime=docker',
]

FIRST_BOOT_COMMANDS = [
    'curtin in-target -- mkdir -p /opt/post-install',
    '''curtin in-target -- sh -c "cat > /opt/post-install/first-boot.sh << 'FIRSTBOOT'
#!/bin/bash
# First boot script
echo 'Running first boot configuration...' >> /var/log/first-boot.log

# Fix ownership of user home directory
echo 'Fixing ownership of user home directory...' >> /var/log/first-boot.log
chown -R {{ admin_username }}:{{ admin_username }} /home/{{ admin_username }}
echo 'Finished fixing ownership of user home directory.' >> /var/log/first-boot.log
echo 'Disabling motd...' >> /var/log/first-boot.log
chmod -x /etc/update-motd.d/*
echo 'Finished disabling motd.' >> /var/log/first-boot.log
echo 'Enabling landscape motd...' >> /var/log/first-boot.log
chmod +x /etc/update-motd.d/50-landscape-sysinfo
echo 'Finished enabling landscape motd.' >> /var/log/first-boot.log

# Update packages, but sleep first to give time for drivers to configure
echo 'Updating packages...' >> /var/log/first-boot.log
sleep 30
apt-get update
apt-get upgrade -y
echo 'Finished updating packages.' >> /var/log/first-boot.log

# Disable this service after first run
echo 'Disabling first-boot service...' >> /var/log/first-boot.log
systemctl disable first-boot.service
echo 'Finished disabling first-boot service.' >> /var/log/first-boot.log
echo \\"First boot completed at \$(date)\\" >> /var/log/first-boot.log
FIRSTBOOT
    "
    ''',
    'curtin in-target -- chmod +x /opt/post-install/first-boot.sh',
    '''curtin in-target -- sh -c "cat > /etc/systemd/system/first-boot.service << 'SERVICE'
[Unit]
Description=First Boot Configuration
After=multi-user.target
Wants=multi-user.target

[Service]
Type=oneshot
ExecStart=/opt/post-install/first-boot.sh
RemainAfterExit=yes
StandardOutput=journal+console
StandardError=journal+console

[Install]
WantedBy=default.target
SERVICE
    "
    ''',
    'curtin in-target -- systemctl enable first-boot.service',
]

USERPROFILE_COMMANDS = [
    'curtin in-target -- mkdir -p /home/{{ admin_username }}',
    '''curtin in-target -- sh -c "cat >> /home/{{ admin_username }}/.bashrc << 'BASHRC'
# Custom bash configuration
export EDITOR=vim
export VISUAL=vim

# Custom aliases
alias ll='ls -lah'
alias update='sudo apt-get update && sudo apt-get upgrade'

# Custom prompt
PS1='\[\\033[01;32m\]\\u@\h\[\\033[00m\]:\[\\033[01;34m\]\w\[\\033[00m\]\$ '
BASHRC
    "
    ''',
    '''curtin in-target -- sh -c "cat > /home/{{ admin_username }}/.vimrc << 'VIMRC'
set number
set expandtab
set tabstop=4
set shiftwidth=4
syntax on
set background=dark
VIMRC
    "
    ''',
    '''curtin in-target -- sh -c "cat > /etc/profile.d/custom-settings.sh << 'PROFILE'
#!/bin/bash
# System-wide custom settings
export HISTSIZE=10000
export HISTFILESIZE=20000
PROFILE
    "
    ''',
]
