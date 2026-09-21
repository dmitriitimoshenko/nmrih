# NMRiH HUD Bars

SourceMod plugin that draws a compact health and stamina bar in a corner of the
screen, with the current value printed inside the bar, plus icons for bleeding
and infection.

```
HP [█████72██░░░]
SP [███░░58░░░░░]
♦ BLEEDING   ▲ INFECTED
```

The health bar shifts from green through yellow to red as you lose health, the
status line blinks, and every player can hide the whole thing with `!hud`.

## Install

1. Copy the files onto the server:
   ```
   sourcemod/plugins/nmrih_hudbars.smx   -> <server>/nmrih/addons/sourcemod/plugins/
   ```
2. `sm plugins load nmrih_hudbars` (or change the map).
3. The config is generated on first run at
   `addons/sourcemod/configs/../../cfg/sourcemod/nmrih_hudbars.cfg`.

`nmrih_hudbars.smx` is built from `scripting/nmrih_hudbars.sp` with SourceMod
1.12.0-git7253. To rebuild it with the compiler that ships with the server:

```bash
cd addons/sourcemod/scripting
./compile.sh nmrih_hudbars.sp
```

## Check this first

The bars for health work on any build. **Stamina, bleeding and infection are
read from netprops, and the names differ between NMRiH builds.** If a bar or an
icon never shows up, connect to the server and run:

```
sm_hudbars_scan
```

It prints your player's network class and which of the known netprop names
actually exist on this build, with their current values. Point the cvars at
whatever it finds:

```
sm_hudbars_prop_stamina  "m_flStamina"
sm_hudbars_prop_bleeding "m_bIsBleeding"
sm_hudbars_prop_infected "m_bIsInfected"
```

An empty value hides that element. If the scan finds nothing, run
`sm_dump_netprops_xml props.xml` and search the dump for the class name the
scan printed.

## Commands

| Command | Access | What it does |
| --- | --- | --- |
| `sm_hud` | everyone | Hides/shows the bars for yourself |
| `sm_hudbars_scan` | generic admin | Lists the netprops this build exposes |

## ConVars

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
| `sm_hudbars_prop_*` | see above | Netprop names |

## Notes

- **Position.** The default is the top left corner. NMRiH draws its own health
  and stamina at the bottom left, so moving these bars down there will overlap
  it. Change `sm_hudbars_x` / `sm_hudbars_y` to taste.
- **Font.** The bar is built out of block characters. They render in every font
  that carries the WGL4 set, which covers the usual Source HUD fonts, but if
  players see empty boxes, switch to ASCII:
  ```
  sm_hudbars_char_full "#"
  sm_hudbars_char_empty "-"
  sm_hudbars_icon_bleeding "!"
  sm_hudbars_icon_infected "*"
  ```
- **Precision.** With the value printed inside the bar, the digits cover two or
  three cells, so the filled part cannot be read exactly in the middle of the
  range — the number next to it is the exact readout. Set
  `sm_hudbars_number_inside 0` to get `[████████░░░░ 72]` instead, where the
  whole bar stays readable.
- **HUD channels.** The plugin uses 3 of the 6 game_text channels the engine
  offers (health, stamina, status). If another plugin has taken them all, this
  one falls back to hint text, which cannot be positioned or coloured, and logs
  a line saying so.
- Bleeding and infection share one channel, so when both are active the line is
  coloured red — bleeding is the more urgent of the two.
