{{ define "compose.yml" }}
---

services:
    nzbd:
        image: linuxserver/sabnzbd
        container_name: nzbd
        restart: unless-stopped
        networks:
            - docker-bridge
        environment:
            - PUID=0
            - PGID=0
        labels:
            - "traefik.enable=true"
            - "traefik.http.routers.nzbd.middlewares=sso_auth"
            - "traefik.http.routers.nzbd.rule=Host(`nzb.gtech.dev`)"
        volumes:
            - nzbd-config:/config
            - nzbd-incomplete-downloads:/incomplete-downloads
            - nzbd-movies:/complete-downloads/movies
            - nzbd-television:/complete-downloads/tv
    sonarr:
        image: linuxserver/sonarr
        container_name: sonarr
        restart: unless-stopped
        networks:
            - docker-bridge
        environment:
            - PUID=0
            - PGID=0
        labels:
            - "traefik.enable=true"
            - "traefik.http.routers.sonarr.middlewares=sso_auth"
            - "traefik.http.routers.sonarr.rule=Host(`sonarr.gtech.dev`)"
        volumes:
            - sonarr-config:/config
            - nzbd-television:/nzbs
            - television:/downloads
    radarr:
        image: linuxserver/radarr
        container_name: radarr
        restart: unless-stopped
        networks:
            - docker-bridge
        environment:
            - PUID=0
            - PGID=0
        labels:
            - "traefik.enable=true"
        volumes:
            - radarr-config:/config
            - nzbd-movies:/nzbs
            - movies:/downloads
    overseerr:
        image: linuxserver/overseerr
        container_name: overseerr
        restart: unless-stopped
        networks:
            - docker-bridge
        environment:
            - PUID=0
            - PGID=0
        labels:
            - "traefik.enable=true"
            - "traefik.http.routers.overseerr.rule=Host(`requests.gtech.dev`)"
        volumes:
            - overseerr-config:/config
    plex:
        image: plexinc/pms-docker
        container_name: plex
        runtime: nvidia
        restart: unless-stopped
        networks:
            - docker-bridge
        ports:
            - "1900:1900"
            - "8324:8324"
            - "32400:32400"
            - "32410:32410"
            - "32412:32412"
            - "32413:32413"
            - "32414:32414"
            - "32469:32469"
        environment:
            - TZ=Etc/UTC
            - NVIDIA_VISIBLE_DEVICES=all
            - NVIDIA_DRIVER_CAPABILITIES=all
            - PLEX_CLAIM={{ .PlexClaim }}
        volumes:
            - plex-config:/config
            - plex-transcode:/transcode
            - movies:/data/movies
            - television:/data/tv
    cloudflared:
        image: cloudflare/cloudflared
        container_name: cloudflared
        command: tunnel --no-autoupdate run --token {{ .CloudflaredToken }}
        restart: unless-stopped
        networks:
            - docker-bridge

networks:
    docker-bridge:
        driver: bridge

volumes:
    nzbd-config:
        external: true
    nzbd-incomplete-downloads:
        external: true
    nzbd-movies:
        external: true
    nzbd-television:
        external: true
    radarr-config:
        external: true
    sonarr-config:
        external: true
    overseerr-config:
        external: true
    plex-config:
        external: true
    plex-transcode:
        external: true
    movies:
        external: true
    television:
        external: true
{{ end }}