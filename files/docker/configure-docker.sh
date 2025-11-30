systemctl stop docker

cat >> /etc/docker/daemon.json << 'DOCKER'
{
    "runtimes": {
        "data-root": "/docker",
        "nvidia": {
            "args": [],
            "path": "nvidia-container-runtime"
        }
    }
}
DOCKER

systemctl start docker

cat >> /etc/systemd/system/containers.service << 'SERVICE'
[Unit]
Description=Containers Compose Application
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/post-install
ExecStart=/usr/local/bin/docker compose up -f compose.yml -d
ExecStop=/usr/local/bin/docker compose -f compose.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
SERVICE

VOLUMES=(
"nzbd-config"
"nzbd-incomplete-downloads"
"nzbd-movies"
"nzbd-television"
"sonarr-config"
"radarr-config"
"overseerr-config"
"plex-config"
"plex-transcode"
"movies"
"television"
)

for volume in "${VOLUMES[@]}"; do
  if ! docker volume ls -q | grep -q "$volume"; then
    docker volume create "$volume"
  fi
done

curl https://raw.githubusercontent.com/hunoz/ubuntu-iso-builder/refs/heads/main/files/docker/compose.yml -o /opt/post-install/compose.yml
PLEX_CLAIM=$(cat /opt/post-install/plex-claim)
CLOUDFLARED_TOKEN=$(cat /opt/post-install/cloudflared-token)
sed -i "s#{{ plex_claim }}#${PLEX_CLAIM}#g" /opt/post-install/compose.yml
sed -i "s#{{ cloudflared_token }}#${CLOUDFLARED_TOKEN}#g" /opt/post-install/compose.yml
