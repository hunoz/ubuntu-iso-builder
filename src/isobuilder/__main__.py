import argparse

from isobuilder.cloud_init import generate_cloudinit_config
from isobuilder.iso_build import build_iso
from isobuilder.utils.constants import BUILD_DIR

parser = argparse.ArgumentParser(description = "CLI for generating a complete cloud-init file and building the ISO")
parser.add_argument(
    "--hostname", "-n",
    required=True,
    help="Hostname that the machine using the ISO will have"
)

parser.add_argument(
    "--admin-username", "-a",
    required=True,
    default="localadmin",
    help="Username that the default admin user will have"
)

parser.add_argument(
    "--admin-password", "-p",
    required=True,
    default="password",
    help="Password that the default admin user will have"
)

parser.add_argument(
    "--ssh-key", "-k",
    required=False,
    action="append",
    default=[],
    help="The SSH keys to allow access onto the server with"
)

parser.add_argument(
    "--encryption-password", "-e",
    required=True,
    help="The password used to encrypt the OS volume",
)

parser.add_argument(
    "--disk-serial", "-s",
    required=True,
    help="The serial number of the disk to use for the OS",
)

def main():
    args = parser.parse_args()

    cloudinit_config = generate_cloudinit_config({
        "hostname": args.hostname,
        "admin_username": args.admin_username,
        "admin_password": args.admin_password,
        "ssh_keys": args.ssh_key,
        "encryption_password": args.encryption_password,
        "disk_serial": args.disk_serial
    })

    BUILD_DIR.mkdir(parents=True, exist_ok=True)

    build_iso(
        cloudinit_config,
        work_dir=str(BUILD_DIR.absolute())
    )





if __name__ == "__main__":
    main()
