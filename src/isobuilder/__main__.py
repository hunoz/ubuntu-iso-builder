import argparse
from pathlib import Path

from isobuilder.cloud_init import render_template
from isobuilder.iso_build import build_iso
from isobuilder.utils.constants import BUILD_DIR

parser = argparse.ArgumentParser(description = "CLI for generating a complete cloud-init file and building the ISO")
parser.add_argument(
    "--hostname", "-s",
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

def main():
    args = parser.parse_args()

    cloud_init_file = render_template({
        "hostname": args.hostname,
        "admin_username": args.admin_username,
        "admin_password": args.admin_password,
        "ssh_keys": args.ssh_key,
        "encryption_password": args.encryption_password,
    })

    BUILD_DIR.mkdir(parents=True, exist_ok=True)

    cloud_init_filepath = Path(f"{BUILD_DIR}/cloud-init.yaml")

    with open(cloud_init_filepath, "w+") as f:
        f.write(cloud_init_file)

    build_iso(
        str(cloud_init_filepath.absolute()),
        work_dir=str(BUILD_DIR.absolute())
    )





if __name__ == "__main__":
    main()
