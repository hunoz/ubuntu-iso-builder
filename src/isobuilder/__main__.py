import argparse
from typing import Literal

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
    required=False,
    default="localadmin",
    help="Username that the default admin user will have"
)

parser.add_argument(
    "--admin-password", "-p",
    required=False,
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
    "--type", "-t",
    required=False,
    default="server",
    choices=["server", "desktop"],
    help="The type of ISO to generate",
)

parser.add_argument(
    "--disk-serial", "-s",
    required=True,
    help="The serial number of the disk to use for the OS",
)

parser.add_argument(
    "--plex-claim", "-c",
    required=True,
    help="The claim token for the Plex Media Server",
)

parser.add_argument(
    "--cloudflared-token", "-ct",
    required=True,
    help="The cloudflared token for the ISO",
)

def main():
    args = parser.parse_args()

    cloudinit_config = generate_cloudinit_config({
        "hostname": args.hostname,
        "admin_username": args.admin_username,
        "admin_password": args.admin_password,
        "ssh_keys": args.ssh_key,
        "disk_serial": args.disk_serial,
        "plex_claim": args.plex_claim,
        "cloudflared_token": args.cloudclared_token,
    })

    BUILD_DIR.mkdir(parents=True, exist_ok=True)

    build_iso(
        cloudinit_config,
        ubuntu_type=args.type,
        work_dir=str(BUILD_DIR.absolute())
    )





if __name__ == "__main__":
    main()
