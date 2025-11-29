from typing import TypedDict
from jinja2.environment import Environment
from jinja2.loaders import PackageLoader

from isobuilder.utils.password.hasher import hash_password


class CloudInitContext(TypedDict):
    hostname: str
    admin_username: str
    admin_password: str
    ssh_keys: list[str]
    encryption_password: str

def render_template(context: CloudInitContext) -> str:
    loader = PackageLoader("isobuilder.cloud_init")
    env = Environment(loader=loader)
    template = env.get_template("cloud-init.yaml.jinja")

    context["admin_password"] = hash_password(context["admin_password"])

    return template.render(context)
