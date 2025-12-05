{{ define "configure-docker" }}

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
{{ end }}