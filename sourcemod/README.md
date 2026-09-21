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
git push              # the deploy installs it
```

A changed `.smx` alters no image, so `docker compose up --build` has no reason to
recreate the game server and the entrypoint never gets to copy the new file in.
`make server-plugins-sync` compares what is mounted with what is installed and
restarts the game server only when they differ; the deploy runs it after every
`make docker-re-run`, and it is also how you install a plugin by hand.

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
HP [▓▓▓▓▓▓░░░░] 62
SP [▓▓▓▓░░░░░░] 41
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
| `sm_hudbars_cells` | `10` | Width of a bar in characters |
| `sm_hudbars_style` | `shaded` | `shaded`, `blocks`, `squares`, `dots`, `ascii` |
| `sm_hudbars_glyphs` | empty | Overrides the style with `"<filled> <empty>"` |
| `sm_hudbars_alpha` | `220` | Opacity, 0-255 |
| `sm_hudbars_number_inside` | `0` | `1` = value inside the bar, `0` = after it |
| `sm_hudbars_labels` | `1` | `1` = `HP`/`SP` before the bars, `0` = no labels, which lines the bars up exactly |
| `sm_hudbars_health_max` | `100` | Health that fills the bar |
| `sm_hudbars_stamina_max` | `0` | Stamina that fills the bar, `0` = learn it |
| `sm_hudbars_stamina_color` | `120 175 220` | Colour of the stamina bar, `"R G B"` |
| `sm_hudbars_icon_bleeding` | `♦` | Bleeding icon |
| `sm_hudbars_icon_infected` | `▲` | Infection icon |
| `sm_hudbars_prop_*` | empty | Netprop overrides, see above |
| `sm_hudbars_detected` | — | Reports what detection settled on; setting it does nothing |

## Picking a look

All of these apply on the next refresh, a fifth of a second later, so a look can
be dialled in live over rcon without restarting anything:

```
sm_hudbars_style shaded     HP [▓▓▓▓▓▓░░░░] 62
sm_hudbars_style blocks     HP [██████░░░░] 62
sm_hudbars_style squares    HP [■■■■■■□□□□] 62
sm_hudbars_style dots       HP [●●●●●●○○○○] 62
sm_hudbars_style ascii      HP [######----] 62
```

`shaded` is the default because a full block fills the entire line box, so a row
of them reads as one solid slab rather than a bar. `ascii` is the fallback if a
font turns the others into empty squares. Anything else goes through
`sm_hudbars_glyphs "◆ ◇"`.

The rest of the knobs worth trying live: `sm_hudbars_cells` for length,
`sm_hudbars_alpha` for how much it fights with the game, `sm_hudbars_y` and
`sm_hudbars_line_height` for position and spacing, `sm_hudbars_number_inside 1`
to put the number back inside the bar.

## The bars do not line up exactly

The HUD font is proportional, so `HP` and `SP` are not the same width, and a
label before a bar means the two bars start a few pixels apart. Nothing can pad
that away from the server side.

`sm_hudbars_labels 0` is the fix: every line then starts with the same bracket,
so the bars line up perfectly, and the colours still say which is which.

```
HP [▓▓▓▓▓▓▓▓▓▓] 100      [▓▓▓▓▓▓▓▓▓▓] 100
SP [▓▓▓▓▓▓░░░░] 81       [▓▓▓▓▓▓░░░░] 81
```

The bar itself does not have this problem: the block glyphs share one width, so
a bar keeps its length as it fills. `ascii` is the exception — `#` is wider than
`-`, so that style's bar breathes. It is a fallback for broken fonts, not a look.

## Stamina has no fixed maximum

NMRiH does not report one, and it is not 100 on every build — this server reads
130 at full. So by default the plugin takes the highest value it has ever seen
as the maximum, which is right within seconds of play and needs no
configuration. Set `sm_hudbars_stamina_max` to a number to override that.

## Changing the defaults reaches an existing server

`AutoExecConfig` never rewrites a config that already exists, so new defaults
would be silently overridden by the file written for the previous version. The
plugin keeps a `cfg/sourcemod/nmrih_hudbars.version` stamp next to the config
and regenerates the config when it moves, which is the only time hand-made edits
to that file are lost.

## Notes

- **Position.** The default is the top left corner. NMRiH draws its own health
  and stamina at the bottom left, so moving these bars down there will overlap
  it. Change `sm_hudbars_x` / `sm_hudbars_y` to taste; they apply on the next
  refresh, so the position can be dialled in live over rcon.
- **Font.** If the bar renders as empty squares, the font is missing the block
  glyphs: `sm_hudbars_style ascii`, plus `sm_hudbars_icon_bleeding "!"` and
  `sm_hudbars_icon_infected "*"`.
- **Precision.** `sm_hudbars_number_inside 1` puts the value inside the bar,
  where its digits cover two or three cells — so the filled part cannot be read
  exactly in the middle of the range. The default keeps the number after the bar
  for that reason.
- **HUD channels.** The plugin uses 3 of the 6 `game_text` channels the engine
  offers (health, stamina, status). If another plugin has taken them all, this
  one falls back to hint text, which cannot be positioned or coloured, and logs
  a line saying so to `addons/sourcemod/logs/`.
- Bleeding and infection share one channel, so when both are active the line is
  red — bleeding is the more urgent of the two.
