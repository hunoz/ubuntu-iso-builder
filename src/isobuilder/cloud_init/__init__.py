from typing import TypedDict, Any

import yaml

from isobuilder.cloud_init.constants import FIRST_BOOT_COMMANDS, USERPROFILE_COMMANDS
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
    encryption_password: str
    disk_serial: str

def render_cloudinit_config(config: dict[str, Any]) -> str:
    dump = yaml.dump(config, Dumper=CustomDumper, default_flow_style=False, sort_keys=False, width=float('inf'))
    dump = "#cloud-config\n" + dump
    return dump

def generate_cloudinit_config(context: CloudInitContext) -> dict[str, Any]:
    # TODO: Add NVIDIA drivers
    # TODO: Check if drives are raided and if not, raid them. Otherwise, mount the array
    # TODO: Stand up the Docker containers
    encryption_password = context['encryption_password']
    return {
        'autoinstall': {
            'version': 1,
            'timezone': 'Etc/UTC',
            'locale': 'en_US.UTF-8',
            'keyboard': {
                'layout': 'us',
            },
            'identity': {
                'hostname': context['hostname'],
                'username': context['admin_username'],
                'password': hash_password(context['admin_password']),
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
                    'password': encryption_password,
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
            ],
            'late-commands': [
                'curtin in-target -- mkdir -p /root/.luks',
                'curtin in-target -- dd if=/dev/urandom of=/root/.luks/boot_os.keyfile bs=1024 count=4',
                'curtin in-target -- chmod 0400 /root/.luks/boot_os.keyfile',
                f"curtin in-target -- sh -c 'echo \"{encryption_password}\" > /tmp/keyfile'",
                f"curtin in-target -- sh -c 'cryptsetup luksAddKey --key-file=/tmp/keyfile $(blkid -t TYPE=crypto_LUKS -o device | head -n1) /root/.luks/boot_os.keyfile'"
                "curtin in-target -- rm -f /tmp/keyfile",
                "curtin in-target -- sed -i 's|none|/root/.luks/boot_os.keyfile|' /etc/crypttab"
                'curtin in-target -- chmod -R 700 /root/.luks',
                'curtin in-target -- update-initramfs -u -k all',
                *[s.replace("{{ admin_username }}", context['admin_username']) for s in FIRST_BOOT_COMMANDS],
                *[s.replace("{{ admin_username }}", context['admin_username']) for s in USERPROFILE_COMMANDS],
            ],
            'shutdown': 'reboot',
        },
    }
