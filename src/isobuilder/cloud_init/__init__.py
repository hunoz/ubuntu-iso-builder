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
                    'password': context['encryption_password'],
                },
                'config': [
                    {
                        'type': 'disk',
                        'id': 'disk0',
                        'ptable': 'gpt',
                        'wipe': 'superblock',
                        'grub_device': True,
                        'match': {
                            'serial': context['disk_serial'],
                        },
                    },
                    {
                        'type': 'partition',
                        'id': 'boot-partition',
                        'device': 'disk0',
                        'size': '512M',
                        'flag': 'boot',
                    },
                    {
                        'type': 'partition',
                        'id': 'lvm-partition',
                        'device': 'disk0',
                        'size': -1,
                    },
                    {
                        'type': 'format',
                        'id': 'boot-fs',
                        'volume': 'boot-partition',
                        'fstype': 'fat32',
                    },
                    {
                        'type': 'dm_crypt',
                        'id': 'dmcrypt0',
                        'volume': 'lvm-partition',
                        'key': context['encryption_password'],
                    },
                    {
                        'type': 'lvm_volgroup',
                        'id': 'vg0',
                        'name': 'ubuntu-vg',
                        'devices': ['dmcrypt0'],
                    },
                    {
                        'type': 'lvm_partition',
                        'id': 'lv-swap',
                        'volgroup': 'vg0',
                        'name': 'swap',
                        'size': '8G',
                    },
                    {
                        'type': 'lvm_partition',
                        'id': 'lv-root',
                        'volgroup': 'vg0',
                        'name': 'root',
                        'size': -1,
                    },
                    {
                        'type': 'format',
                        'id': 'swap-fs',
                        'volume': 'lv-swap',
                        'fstype': 'swap',
                    },
                    {
                        'type': 'format',
                        'id': 'root-fs',
                        'volume': 'lv-root',
                        'fstype': 'ext4',
                    },
                    {
                        'type': 'mount',
                        'id': 'mount-boot',
                        'device': 'boot-fs',
                        'path': '/boot/efi',
                    },
                    {
                        'type': 'mount',
                        'id': 'mount-swap',
                        'device': 'swap-fs',
                        'path': 'none',
                    },
                    {
                        'type': 'mount',
                        'id': 'mount-root',
                        'device': 'root-fs',
                        'path': '/',
                    },
                ],
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
                'curtin in-target -- dd if=/dev/urandom of=/root/.luks/boot_os.keyfile bs=4096 count=1',
                'curtin in-target -- chmod 000 /root/.luks/boot_os.keyfile',
                'curtin in-target -- chmod -R 700 /root/.luks',
                f"curtin in-target -- sh -c 'echo \"{context['encryption_password']}\" | cryptsetup luksAddKey $(blkid -t TYPE=crypto_LUKS -o device | head -n1) /root/.luks/boot_os.keyfile'",
                "curtin in-target -- sh -c 'CRYPT_UUID=$(blkid -s UUID -o value /dev/disk/by-id/dm-uuid-*); sed -i \"s|none|/root/.luks/boot_os.keyfile|\" /etc/crypttab",
                "curtin in-target -- sh -c 'echo \"KEYFILE_PATTERN=/root/.luks/*.keyfile\" >> /etc/cryptsetup-initramfs/conf-hook'",
                "curtin in-target -- sh -c 'echo \"UMASK=0077\" >> /etc/initramfs-tools/initramfs.conf'",
                'curtin in-target -- update-initramfs -u -k all',
                *FIRST_BOOT_COMMANDS,
                *USERPROFILE_COMMANDS,
            ],
            'shutdown': 'reboot',
        },
    }
