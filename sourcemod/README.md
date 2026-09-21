# SourceMod plugins

Plugins for the game server, built and shipped by this repository. Nothing here
has to be copied to the server by hand.

```
sourcemod/
├── scripting/   # sources (.sp), compiled by `make plugins`
└── plugins/     # compiled binaries (.smx), mounted into the game container
```

## How a plugin reaches the server

1. `nmrih_server/Dockerfile` downloads **Metamod:Source** and **SourceMod** at
   image build time, pinned to exact versions, and stages them in the image.
2. `nmrih_server/entrypoint.sh` copies them into the game directory on every
   start. They live under `addons/`, which steamcmd does not manage, so the
   install survives game updates.
3. `docker-compose.yaml` mounts `sourcemod/plugins` into the container read
   only, and the entrypoint copies every `.smx` from it into
   `addons/sourcemod/plugins/`.

So the whole loop is:

```bash
vim sourcemod/scripting/nmrih_hudbars.sp
make plugins          # compile, using the compiler inside the server image
make server-restart    # the new .smx is copied in on start
```

`make plugins` needs no local SourcePawn toolchain: it compiles with the
`spcomp64` that ships inside the game server image, which is by construction the
same SourceMod the server runs.

**Adding a plugin** is dropping a `.sp` into `scripting/`: `make plugins`
compiles everything in that directory, the entrypoint copies everything in
`plugins/`, and CI compiles every source on each pull request.

**Commit the `.smx` together with the `.sp`.** The container installs binaries,
it does not compile, and CI fails a pull request whose source has no committed
binary next to it.

## What a restart does and does not touch

| Path | On every start |
| --- | --- |
| `addons/metamod/bin`, `addons/sourcemod/{bin,extensions,gamedata,scripting,translations}` | replaced, so they always match the pinned version |
| `addons/sourcemod/configs`, `cfg/sourcemod` | written once, never overwritten — edits made on the server survive |
| `addons/sourcemod/plugins` | stock and repository plugins refreshed; anything else left alone |
| `addons/sourcemod/{data,logs}` | untouched |

## Versions

Both pins live in `nmrih_server/Dockerfile` and are the single source of truth —
the `Makefile` and CI read the SourceMod one from there.

```dockerfile
ARG METAMOD_VERSION=1.12.0-git1226
ARG SOURCEMOD_VERSION=1.12.0-git7253
```

Bump them, then `make plugins` (recompiles against the new SourceMod) and
`make up` (rebuilds the image). Current builds are listed at
[sourcemm.net](https://www.sourcemm.net/downloads.php) and
[sourcemod.net](https://www.sourcemod.net/downloads.php).

`NMRIH_SOURCEMOD=0` in `.env` skips the whole thing and leaves the server
vanilla. An existing install in `server_data/` is left in place, not removed.

---

# nmrih_hudbars

Compact health and stamina bars in a corner of the screen, with the current
value printed inside the bar, plus icons for bleeding and infection.

```
HP [█████72██░░░]
SP [███░░58░░░░░]
♦ BLEEDING   ▲ INFECTED
```

The health bar shifts from green through yellow to red as you lose health, the
status line blinks, and every player can hide the whole thing with `!hud`.

## Netprops find themselves

Health comes from `GetClientHealth` and works anywhere. Stamina, bleeding and
infection are netprops whose names differ between NMRiH builds, so the plugin
resolves them itself against the first player it draws for, and again after
every map change in case a game update moved something. There is nothing to
configure and nothing to run.

What it settled on is reported three ways:

- in the SourceMod log: `netprops: stamina=m_flStamina bleeding=none infection=none`
- as `sm_hudbars_detected`, which is `FCVAR_NOTIFY` and therefore part of the
  server's A2S rules — readable from outside without touching the server
- with `sm_hudbars_scan`, which additionally lists every candidate name that
  exists on this build with its current value

`sm_hudbars_scan` works from the server console and over rcon as well as in
game: with no calling player it reads the props off the first connected one.

The `sm_hudbars_prop_*` cvars are overrides, not requirements:

| Value | Meaning |
| --- | --- |
| empty (default) | detect automatically |
| a netprop name | use it — but if this build does not have it, detection takes over anyway |
| `none` | hide that element |

That middle row is deliberate: a config written for one build cannot silently
switch a bar off on another.

## Commands

| Command | Access | What it does |
| --- | --- | --- |
| `sm_hud` | everyone | Hides/shows the bars for yourself |
| `sm_hudbars_scan` | generic admin, or the server console | Lists the netprops this build exposes |

## ConVars

Generated on first run into `cfg/sourcemod/nmrih_hudbars.cfg`.

| ConVar | Default | What it does |
| --- | --- | --- |
| `sm_hudbars_enabled` | `1` | Master switch |
| `sm_hudbars_x` | `0.015` | Horizontal position, `0.0` left … `1.0` right, `-1` centered |
| `sm_hudbars_y` | `0.045` | Vertical position of the first line |
| `sm_hudbars_line_height` | `0.035` | Distance between the lines |
| `sm_hudbars_interval` | `0.2` | Refresh interval in seconds |
| `sm_hudbars_cells` | `12` | Width of a bar in characters |
| `sm_hudbars_char_full` | `█` | Filled cell |
| `sm_hudbars_char_empty` | `░` | Empty cell |
| `sm_hudbars_number_inside` | `1` | `1` = value inside the bar, `0` = after it |
| `sm_hudbars_health_max` | `100` | Health that fills the bar |
| `sm_hudbars_stamina_max` | `100` | Stamina that fills the bar |
| `sm_hudbars_stamina_color` | `90 190 255` | Colour of the stamina bar, `"R G B"` |
| `sm_hudbars_icon_bleeding` | `♦` | Bleeding icon |
| `sm_hudbars_icon_infected` | `▲` | Infection icon |
| `sm_hudbars_prop_*` | empty | Netprop overrides, see above |
| `sm_hudbars_detected` | — | Reports what detection settled on; setting it does nothing |

## Notes

- **Position.** The default is the top left corner. NMRiH draws its own health
  and stamina at the bottom left, so moving these bars down there will overlap
  it. Change `sm_hudbars_x` / `sm_hudbars_y` to taste; they apply on the next
  refresh, so the position can be dialled in live over rcon.
- **Font.** The bar is built out of block characters. They render in every font
  carrying the WGL4 set, which covers the usual Source HUD fonts, but if players
  see empty boxes, switch to ASCII:
  ```
  sm_hudbars_char_full "#"
  sm_hudbars_char_empty "-"
  sm_hudbars_icon_bleeding "!"
  sm_hudbars_icon_infected "*"
  ```
- **Precision.** With the value printed inside the bar the digits cover two or
  three cells, so the filled part cannot be read exactly in the middle of the
  range — the number next to it is the exact readout. Set
  `sm_hudbars_number_inside 0` for `[████████░░░░ 72]`, where the whole bar stays
  readable.
- **HUD channels.** The plugin uses 3 of the 6 `game_text` channels the engine
  offers (health, stamina, status). If another plugin has taken them all, this
  one falls back to hint text, which cannot be positioned or coloured, and logs
  a line saying so to `addons/sourcemod/logs/`.
- Bleeding and infection share one channel, so when both are active the line is
  red — bleeding is the more urgent of the two.
