# NMRiH Server Dashboard

This repository runs a whole NMRiH deployment from one `docker compose`: the
game server itself (installed and kept up to date from Steam) plus a dashboard
that monitors it - a backend for log parsing and CSV management (log_api), a
responsive, dark-themed React frontend (log_frontend), and a Traefik reverse
proxy for secure routing and HTTPS.

The two halves meet in one directory: srcds writes its `L<MMDD><NNN>.log` files
into `logs/`, and log_api reads the same directory as `/logs`. Live player
counts come from an A2S query straight to the game container.

## Project Structure

```
nmrih/
├── docker-compose.yaml       # Orchestrates all Docker containers (nmrih_server, log_api, log_frontend, traefik, etc.)
├── docker-compose.shared-proxy.yaml  # Override for hosts where another Traefik already owns :80/:443
├── .env.example              # Every knob, from the server hostname to the ACME e-mail
├── nmrih_server/             # The game server itself
│   ├── Dockerfile            # steamcmd + the 32-bit libs srcds needs
│   ├── entrypoint.sh         # Installs/updates app 317670, renders the cfg, runs srcds
│   └── cfg/server.cfg.template  # server.cfg rendered from .env at container start
├── server_data/              # Steam install, ~9 GB (created on first run, gitignored)
├── logs/                     # srcds logs; written by the game, read by log_api (gitignored)
├── log_api/                  # Backend API for log parsing and CSV file management
│   ├── Dockerfile            # Dockerfile for building the log_api container
│   ├── cmd/servercheck/      # CLI: is the server reachable from outside and listed on Steam?
│   └── internal/             # Source code for log parsing, CSV generation, etc.
├── log_frontend/             # React-based dashboard application
│   ├── public/
│   │   └── index.html        # HTML template
│   ├── src/
│   │   ├── components/       # Reusable React components (e.g., TopTimeChart, CountryPieChart, PlayersInfo, Controls)
│   │   ├── hooks/            # Custom hooks (e.g., useTopTimeChartData, useWindowDimensions)
│   │   ├── App.js            # Main application component
│   │   └── App.css           # Global styles (dark theme, responsive design)
│   └── package.json          # Frontend dependencies and scripts
├── sourcemod/                # SourceMod plugins, compiled and shipped by this repo
│   ├── scripting/            # Plugin sources (.sp), built by `make plugins`
│   └── plugins/              # Compiled plugins (.smx), mounted into the game container
└── traefik/                  # Traefik configuration and SSL certificate storage (acme.json)
```

## Features

- **Game server (nmrih_server):**
  - Installs "No More Room in Hell Dedicated Server" (Steam app `317670`) with
    steamcmd on first start and re-validates it on every start, so a restart is
    also the update.
  - Renders `cfg/server.cfg` from `.env` (hostname, passwords, region, GSLT,
    SourceTV), so the running config is reproducible from the repo.
  - Writes its logs into the directory the dashboard parses.
  - `make server-console` attaches to the live srcds console.

- **Backend API (log_api):**
  - Parses server logs and saves data as CSV files.
  - Provides various endpoints for retrieving:
    - Top time-spent players
    - Countries statistics (for the pie chart)
    - Detailed player information (e.g., name, score, formatted session duration)
  
- **Responsive Frontend (log_frontend):**
  - **Top Time-Spent Players:**  
    Displays a bar chart of players sorted by their session durations.
  - **Top Countries:**  
    Shows a pie chart with connection percentages by country plus a legend.
  - **Player Info:**  
    Lists connected players with details such as score and duration, formatted (e.g., `50s`, `50m40s`, `24h59m59s`).
  - **Controls:**  
    Provides buttons for refreshing data and copying the server address to the clipboard (with temporary "Copied!" feedback).
  - Uses a dark theme and adapts its layout based on the device (horizontal on PC, vertical on mobile).

- **SourceMod Plugins (sourcemod):**
  - Metamod:Source and SourceMod are installed into the game server container
    automatically, pinned to exact versions.
  - Plugins in `sourcemod/plugins` are copied into the game on every start, so
    shipping a new build of one is `make plugins && make server-restart`.
  - **HUD Bars:** health and stamina bars in a corner of the screen with the
    value inside the bar, plus bleeding and infection icons.
  - See [sourcemod/README.md](sourcemod/README.md).

- **Traefik Reverse Proxy:**
  - Provides secure HTTPS support and manages SSL certificates.
  - Routes requests to the appropriate containers (backend or frontend).

- **Dockerized Environment:**
  - Easily build and run the complete project using Docker Compose.

## Getting started

Everything in this repository comes up with one command. This is the whole path
from a fresh clone to a server players can join and a dashboard you can open.

### What starts

| Container | What it is |
| --- | --- |
| `nmrih_server` | The game itself. Installs/updates NMRiH through steamcmd, installs Metamod:Source + SourceMod and the plugins from `sourcemod/plugins`, runs srcds. |
| `log_api` | Go backend: parses the server logs into CSV and serves the dashboard API. |
| `log_frontend` | The dashboard itself. |
| `redis` | Caches the dashboard's graph responses. |
| `traefik` | Reverse proxy and HTTPS certificates. Left out when `PROXY_NETWORK` is set. |

The game server writes its logs into `logs/`, which `log_api` reads. That shared
directory is the whole integration between the two halves.

### Prerequisites

- Docker and Docker Compose.
- ~15 GB of free disk space for the Steam install.
- Forwarded to this host, if players connect over the internet:
  `27015/udp` and `27015/tcp` for the game, `27020/udp` when SourceTV is on,
  and `80`/`443` for the dashboard.

### 1. Configure

```bash
git clone <repository_url>
cd nmrih
cp .env.example .env
```

Every value has a default, but these are the ones worth setting before the first
start:

| Variable | Why |
| --- | --- |
| `NMRIH_HOSTNAME` | The name players see in the browser. |
| `NMRIH_RCON_PASSWORD` | Empty leaves rcon disabled. |
| `NMRIH_GSLT` | Without it the server is not listed publicly — see below. |
| `REDIS_PASSWORD` | The dashboard cache. |
| `IP_INFO_API_TOKEN` | Resolves player IPs to countries for the pie chart. |
| `TRAEFIK_ACME_EMAIL` | Certificate registration. |

`make prepare`, which every `make up` runs for you, creates `logs/`,
`server_data/` and the Traefik ACME storage with the right ownership — docker
would otherwise create them owned by root and srcds could not write.

__Don't forget to change the host names in `docker-compose.yaml`!__

### 2. Start everything

```bash
make up
```

On a first run this downloads ~6 GB of game files, so it takes a while;
`make server-logs` shows the progress. The same command is how you apply changes
later — it rebuilds what changed and restarts it.

What happens inside `nmrih_server` on every start, in order: steamcmd checks for
a game update, Metamod:Source and SourceMod are copied into the game, plugins
from `sourcemod/plugins` are installed, `cfg/server.cfg` is rendered from `.env`,
and srcds takes over.

### 3. Check it came up

```bash
make ps                # every container up, log_api healthy
make server-logs       # ends with "starting: srcds_linux ..." then a map load
make server-check      # answers from the internet? listed on Steam?
```

SourceMod, from the game server console (`make server-console`, detach with
ctrl-p ctrl-q):

```
meta list              # Metamod loaded, SourceMod among its plugins
sm plugins list        # nmrih_hudbars among them
```

Then open the dashboard at your domain, and join the server with
`connect <host>:27015` in the game console.

### Running only one half

```bash
make server-up       # just the game server
make monitoring-up   # just the dashboard stack
```

### Day to day

| Command | What it does |
| --- | --- |
| `make logs` / `make ps` | Tail everything / status |
| `make server-restart` | Restart the game server |
| `make server-update` | Pull the latest game build from Steam and restart |
| `make server-console` | Attach to the srcds console |
| `make server-check` | Is the server reachable and listed on Steam? |
| `make plugins` | Compile `sourcemod/scripting/*.sp` |
| `make down` | Stop everything |

`make` on its own lists every target.

### Getting the server into the public server browser

Without a Game Server Login Token the server only accepts direct connects.
To get one: log in at <https://steamcommunity.com/dev/managegameservers> with a
Steam account that is not limited (it needs at least one purchase), create a
token for app id `224260` (the game — not `317670`, which is the server tool),
then put it in `.env` as `NMRIH_GSLT=` and run `make server-restart`.

`make server-check` confirms whether it worked.

### Another Traefik already on :80/:443?

Set `PROXY_NETWORK` in `.env` to that proxy's docker network (for example
`PROXY_NETWORK=dealscout_default`). The Makefile then adds
`docker-compose.shared-proxy.yaml`, which leaves the bundled Traefik out and
attaches log_api/log_frontend to that network instead — the existing Traefik
picks them up through their labels.

### When something is wrong

| Symptom | Where to look |
| --- | --- |
| Server missing from the browser | `make server-check` — it separates "does not answer" from "not listed by Steam" |
| `srcds_linux missing` in the logs | The steamcmd install did not finish; `make server-update` runs it again with validation |
| Dashboard empty | `make logs` for `log_api`; it needs `logs/` to contain server logs, which appear after the first round |
| Plugin not loaded | `sm plugins list` in the console; the HUD bars plugin needs its netprops checked with `sm_hudbars_scan` |

## The dashboard

Open it at the domain configured in `docker-compose.yaml` (for example
`https://rulat-bot.duckdns.org`); serving it on a plain localhost port instead
needs a port mapping added to the compose file.

- **Dashboard Features:**
  - **Top Time-Spent Players:**  
    View a bar chart that details players’ session durations.
  - **Top Countries:**  
    See a pie chart reflecting connection percentages by country.
  - **Player Info:**  
    Inspect the list of connected players along with their score and session duration.
  - **Online Statistics:**  
    Allows to get insights about what is an average count of concurrent sessions at any hour in a day.
  - **Controls:**  
    Refresh the data or copy the server address using the provided buttons.

## Plugins

The game server container installs Metamod:Source and SourceMod by itself and
picks up every `.smx` in `sourcemod/plugins` on start. Nothing is copied to the
server by hand.

```bash
make plugins          # compile sourcemod/scripting/*.sp
make server-restart   # the new binaries are installed on start
```

`make plugins` compiles with the `spcomp64` inside the game server image, so no
local SourcePawn toolchain is needed and plugins are always built against
exactly the SourceMod the server runs. Both versions are pinned as build args in
`nmrih_server/Dockerfile`; the Makefile and CI read the SourceMod pin from there,
so there is one place to bump.

Configs edited on the server (`addons/sourcemod/configs`, `cfg/sourcemod`) and
plugins added there by hand survive restarts; the version-coupled parts of the
framework are refreshed on every start. `NMRIH_SOURCEMOD=0` in `.env` leaves the
server vanilla.

[sourcemod/README.md](sourcemod/README.md) covers the layout, what a restart
touches, and the bundled HUD bars plugin.

## Diagnostics

When nobody can find the server in the browser, there are two separate questions:
does it answer queries from the public internet, and does Steam list it at all.
`servercheck` answers both.

```bash
make server-check                                  # public address of this stack
make server-check ARGS="-addr=example.com -json"   # somewhere else, as JSON
```

```
A2S query from the public internet
  name      : Krich Server
  map       : nmo_cabin
  game      : NMRiH: Classic (nmrih, appid 224260)
  players   : 0/8 (0 bots)
  password  : no
  VAC       : yes

Steam global server list
  listed as : <ip>:27015
  game      : appid 224260 (nmrih)
  region    : 3 (Europe)
  sv_lan    : no

Verdict: reachable and listed globally.
```

It exits with `0` when the server is reachable and listed, `1` when it does not
answer at all and `2` when it answers but Steam does not know it, so it drops
straight into a cron job or a monitoring check.

Three things worth knowing when reading the output:

- **Run it from outside this host.** From inside the same network the query can
  be answered without ever leaving the LAN, which proves nothing about the path
  players take. The Steam half of the check is location independent.
- **An empty server is not a missing server.** Both server browsers push empty
  servers to the bottom of the list or hide them behind a filter, so a server
  that is listed can still look absent. Search it by name or connect by address.
- **The A2S packet stores the AppID in 16 bits**, so NMRiH reports `0` there.
  The tool reads the real AppID out of the 64-bit `GameID` field and says so.

## Customization

- **Game server settings:**
  Everything that belongs in `server.cfg` lives in
  `nmrih_server/cfg/server.cfg.template` (rendered on each start - editing the
  copy under `server_data/` is pointless), and everything that belongs on the
  srcds command line comes from `.env` (`NMRIH_MAP`, `NMRIH_MAXPLAYERS`,
  `NMRIH_EXTRA_ARGS`, ...). Custom maps and addons go into
  `server_data/nmrih/` and survive rebuilds.
- **API Endpoints:**  
  Adjust the API URLs in the frontend source code (`log_frontend/src/App.js` and related components) if your backend endpoints or proxy settings change.
- **Styling:**  
  Modify the dark theme and responsive layout via `log_frontend/src/App.css`.
- **Extending Functionality:**  
  Add new components (e.g., additional charts or logs) in the `components/` directory of the frontend.

## Contributing

Contributions, issues, and feature requests are welcome!  
Please check the [Issues](https://github.com/dmitriitimoshenko/nmrih/issues) page or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

For questions or further assistance, please contact via:
- Email: [dmitrii.timoshenko16@gmail.com](mailto:dmitrii.timoshenko16@gmail.com).
- Telegram: [@Kritcz](https://t.me/Kritcz)
