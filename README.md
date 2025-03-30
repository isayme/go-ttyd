# ttyd: A SSH Client over Web Browser

[![Docker Image Version (latest semver)](https://img.shields.io/docker/v/isayme/ttyd?sort=semver&style=flat-square)](https://hub.docker.com/r/isayme/ttyd)
![Docker Image Size (latest semver)](https://img.shields.io/docker/image-size/isayme/ttyd?sort=semver&style=flat-square)
![Docker Pulls](https://img.shields.io/docker/pulls/isayme/ttyd?style=flat-square)

![](./doc/screenshoot.png)

# Usage

## docker

`docker run --rm -p 1323:1323 -e TTYD_CMD='ssh root@192.168.68.8' isayme/ttyd`

## docker compose

```
services:
  ttyd:
    container_name: ttyd
    image: isayme/ttyd
    port:
      - 1323:1323
    environment:
      # specify ssh connect cmd
      - TTYD_CMD=ssh root@192.168.68.8
    restart: unless-stopped
```
