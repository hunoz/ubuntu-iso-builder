from typing import TypedDict, Any

import yaml

from isobuilder.cloud_init.constants import FIRST_BOOT_COMMANDS, USERPROFILE_COMMANDS, DOCKER_COMMANDS, NVIDIA_COMMANDS
from isobuilder.utils.password import hash_password

class CustomDumper(yaml.Dumper):
    def increase_indent(self, flow=False, indentless=False):
        return super(CustomDumper, self).increase_indent(flow, False)

    def represent_str(self, data):
        if '\n' in data:
            # Strip leading whitespace but keep trailing newline for literal block style
            stripped = data.strip()
            if not stripped.endswith('\n'):
                stripped += '\n'
            return self.represent_scalar('tag:yaml.org,2002:str', stripped, style='|')
        return self.represent_scalar('tag:yaml.org,2002:str', data)

CustomDumper.add_representer(str, CustomDumper.represent_str)

class CloudInitContext(TypedDict):
    hostname: str
    admin_username: str
    admin_password: str
    ssh_keys: list[str]
    disk_serial: str
    plex_claim: str
    cloudflared_token: str

def render_cloudinit_config(config: dict[str, Any]) -> str:
    dump = yaml.dump(config, Dumper=CustomDumper, default_flow_style=False, sort_keys=False, width=float('inf'))
    dump = "#cloud-config\n" + dump
    return dump

def generate_cloudinit_config(context: CloudInitContext) -> dict[str, Any]:
    # TODO: Add NVIDIA drivers
    # TODO: Check if drives are raided and if not, raid them. Otherwise, mount the array
    # TODO: Stand up the Docker containers
    return {
        'autoinstall': {
            'version': 1,
            'timezone': 'Etc/UTC',
            'locale': 'en_US.UTF-8',
            'keyboard': {
                'layout': 'us',
            },
            "user-data": {
                "hostname": context['hostname'],
                "users": [
                    {
                        "name": "root",
                        "passwd": hash_password(context['admin_password']),
                        "lock_passwd": False,
                        "ssh_authorized_keys": context['ssh_keys'],
                    },
                    {
                        "name": context['admin_username'],
                        "primary_group": context['admin_username'],
                        "groups": ["sudo"],
                        "passwd": hash_password(context['admin_password']),
                        "lock_passwd": False,
                        "ssh_authorized_keys": context['ssh_keys'],
                        "sudo": "ALL=(ALL) NOPASSWD:ALL",
                        "shell": "/bin/bash",
                    },
                ],
            },
            'ssh': {
                'install-server': True,
                'allow-pw': True,
            } if len(context['ssh_keys']) == 0 else {
                'install-server': True,
                'allow-pw': False,
                'authorized-keys': context['ssh_keys'],
            },
            'storage': {
                'layout': {
                    'name': 'lvm',
                    'match': {
                        'serial': f"*{context['disk_serial']}*",
                    },
                },
            },
            'packages': [
                'vim',
                'curl',
                'git',
                'htop',
                'net-tools',
                'mdadm',
                'ca-certificates',
                'ubuntu-drivers-common',
                'build-essential',
                'dkms',
                'linux-headers-generic',
            ],
            'late-commands': [
                f"curtin in-target -- mkdir /opt/post-install",
                f"curtin in-target -- echo '{context['plex_claim']}' > /opt/post-install/plex-claim",
                f"curtin in-target -- echo '{context['cloudflared_token']}' > /opt/post-install/cloudflared-token",
                f"curtin in-target -- sed -i 's|GRUB_CMDLINE_LINUX_DEFAULT=|GRUB_CMDLINE_LINUX_DEFAULT=\"nosplash usb-storage.quirks=2109:0715:j\" /etc/default/grub",
                f"curtin in-target -- update-grub",
                *DOCKER_COMMANDS,
                *NVIDIA_COMMANDS,
                *[s.replace("{{ admin_username }}", context['admin_username']) for s in FIRST_BOOT_COMMANDS],
                *[s.replace("{{ admin_username }}", context['admin_username']) for s in USERPROFILE_COMMANDS],
            ],
            'shutdown': 'reboot',
        },
    }
