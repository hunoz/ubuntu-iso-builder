import argparse

from isobuilder.cloud_init import render_template

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

    print(args)

    cloud_init_file = render_template({
        "hostname": args.hostname,
        "admin_username": args.admin_username,
        "admin_password": args.admin_password,
        "ssh_keys": args.ssh_key,
        "encryption_password": args.encryption_password,
    })
    print(cloud_init_file)



if __name__ == "__main__":
    main()
