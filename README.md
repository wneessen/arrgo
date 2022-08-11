[![Go Reference](https://pkg.go.dev/badge/github.com/wneessen/arrgo.svg)](https://pkg.go.dev/github.com/wneessen/arrgo) [![Go Report Card](https://goreportcard.com/badge/github.com/wneessen/arrgo)](https://goreportcard.com/report/github.com/wneessen/arrgo) [![#sotbot on Discord](https://img.shields.io/badge/Discord-%23arrgo-blue.svg)](https://discord.gg/zZprPUSW)

# ArrGo - Your humble Sea of Thieves based Discord bot
ArrGo is a Sea of Thieves themed Discord bot written in Go (Golang) and makes heavy use of the
fantastic [discordgo](https://github.com/bwmarrin/discordgo) library.

## Support
Need support? Join [#arrgo](https://discord.gg/zZprPUSW) on our Discord server.

## Requirements
To run your own ArrGo instance, you require a Discord bot token. You can create one in the
[Discord developer portal](https://discord.com/developers/applications)

Also the bot uses PostgreSQL database as it's storage for user data. Therefore connectivity to a PgSQL
server is required for the bot to operate correctly.

To build the bot from the sources, you need to have Go installed, as well.

## Releases
ArrGo is released as Docker image only. You can find the different branches on its
[Github Packages](https://github.com/wneessen/arrgo/pkgs/container/arrgo) page.

### Docker
Running the Docker image/container requires the exposure of the `/arrgo/etc` configuration path from
your local storage. This path holds the `arrgo.toml` configuration file. You can also use the `ARRGO_TOKEN`
environment variable within your Docker environment to pass the discord authentication token.

An example execution could look like this:
```shell
$ sudo docker run -ti -e ARRGO_TOKEN=<your discord token> -v /local/path/to/arrgo/etc:/arrgo/etc \
    ghcr.io/wneessen/arrgo:main 
````

### docker-compose
We've also prepared an example `docker-compose.yml` file for you to integrate with your Docker environment to run the
bot without much hassle of configuration

```yaml
version: "3"
services:
  arrgo:
    image: ghcr.io/wneessen/arrgo:main
    container_name: arrgo
    network_mode: "host"
    restart: always
    environment:
      - ARRGO_TOKEN=<your token here>
    volumes:
      - /var/db/arrgo/etc:/arrgo/etc
      - /etc/localtime:/etc/localtime:ro
    logging:
      driver: local
```

## Configuration
The configuration of the bot is fairly simple. The config file is in the TOML format, so it's really easy
to understand.

