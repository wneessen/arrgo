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
to understand. An example congfiguration [arrgo.toml.examle](./arrgo.toml.example) is provided with the bot.
Most settings are pre-configured with sane defaults, but some configuration (i. e. database settings) needs to
be provided by the user.

The configuration is seperated within different sections. The following documentation describes the different 
sections.

### Discord specific confguration
Within the `[discord]` section there is currently only one optional settings. The `token` setting specifies the
Discord API token for your bot. Instead of providing it via the config file, you can also use the `ARRGO_TOKEN`
environment variable to provide your token. The environment variable has higher importance than the config and
therefore will override the token provided in the `arrgo.toml`

**Example:**
```toml
[discord]
token = "<your discord token>"
```

### Log settings
The `[log]` section lets you configure the log level the bot is supposed to operrate on. Via the `level` setting
you can choose between the following levels:
 * debug
 * info
 * warn
 * error

The bot logs everything in JSON format. The log level defaults to `info`. 

**Example:**
```toml
[log]
level = "info"
```

### Database configuration
The `[db]` section is mandatory and requires to be filled by the user before running the bot. As described in the
requirement section, the bot operates on a PostgreSQL database. The following configuration settings can be
provided:

 * **host**: Specifies the hostname or IP address of the PostgreSQL database
 * **user**: Specifies the PostgreSQL username for authentication
 * **pass**: Specifies the PostgreSQL password for authentication
 * **db**: Specifies the PostgreSQL database to connect to
 * **use_tls**: Specifies if the connection to the database should force TLS encryption

**Example:**
```toml
[db]
user = "arrgo"
pass = "superS3cureP4ssw0rd"
db = "arrgo"
host = "pgsql.mynetwork.tld"
use_tls = true
```

### Data specific settings
The `[data]` section holds settings that are specific to the data processing of the bot. Currently it only holds
one setting - the `enc_key`. All sensitive data in the bot is encrypted on a per-user or per-guild basis. Since 
the user- and guild encryption keys need to be stored in the database as well, they are encrypted using a global
data encryption key.

The `enc_key` needs to be a 32 character long random string. If none is provided in the configuration, the bot will
generate one during startup and provide it to the user, before it then shuts down again. **The bot will not start 
without a valid global encryption key set in the config file.**

**Example:**
```toml
[data]
enc_key = "XbM,,I!23BO4AWr6T&@O?F{4gK@%RN!f"
```

### Timer settings
ArrGo performs a couple of background tasks. These are controlled by timers, which can be configured in the
`[timer]` section of the configuration. The bot uses sane defaults, but if you prefer to override some of the 
settings, you can do so here.

The following timer configurations are currently available:
 * `flameheart_spam (int)`: Sets a minimum amount of minutes for the random number generation of the Flameheard SPAM feature
 * `traderoutes_update (time.Duration)`: Sets the duration how often the bot should check the traderoutes API for updates
 * `userstats_update (time.Duration)`: Specifies the duration that the user should updates the user stats history
 * `ratcookie_check (time.Duration)`: The duration how often the bot checks the provided RAT cookies for validity

**Example (with default values):**
```toml
[timer]
flameheart_spam = "60"
traderoutes_update = "12h"
userstats_update = "30m"
ratcookie_check = "5m"
```

## Sea of Thieves specific commands
Any Sea of Thieves related bot command is only available to registered users, as it requires the bot to access 
the private SoT API with a user specific remote access token (`RAT`). This token has to be stored in the bot's 
database and has to be renewed approx. every 14 days (the bot will DM you when your cookie expired).

### Getting access to the API
Unfortunately the SoT API does not offer any kind of OAuth2 for authentication - at least not publicly available.
Hence we have to use a kind of hackish way to get your access cookie from the Microsoft Live login.

You can either use the [SoTBot-RAT-Chrome-Extension](https://github.com/wneessen/sotbot-rat-extension/tree/main/Chrome)
or the [SoTBot-Token-Extractor](https://github.com/wneessen/sotbot-token-extrator) (please read the notes in the 
project before using it) to get a current cookie and store it in the database of the bot. Unfortunately the cookie 
is only valid for approx. 14 days, so you'll have to renew it every now and then.

### The `RAT` cookie
The RAT cookie grants full access to your account on the Sea of Thieves website. Therefore, even though the cookie 
is encrypted at rest, before storing your cookie in the bot's DB, please make sure that you know what you 
are doing. Maybe at some time, RARE decides to offer a apublic API, which offers OAuth2, so we can allow the 
bot having access to the API data without having to store/renew the cookie.

## Automatic user balance tracking
The bot is able to track the users presence state. If a registered user with a valid RAT cookie has their 
"currently playing" feature activated with Discord and starts playing "Sea of Thieves", the bot will 
automagically fetch the users current balance and once the user stops playing, repeat the process. 
The before and after values are then compared and announced in either the guild's official system channel 
or alternatively in a channel that has been overriden by the guild administrator using the
`/override announce-channel` slash command. By default the summary announcing is disable per guild and has
to enabled by the guild administrator using the `/config announce-sot-summary enable` slash command.