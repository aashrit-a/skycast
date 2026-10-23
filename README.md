# skycast

*Weather in your terminal, animated.*

![skycast demo](assets/demo.gif)

[![CI](https://github.com/aashrit-a/skycast/actions/workflows/ci.yml/badge.svg)](https://github.com/aashrit-a/skycast/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/aashrit-a/skycast)](go.mod)
[![Release](https://img.shields.io/github/v/release/aashrit-a/skycast)](https://github.com/aashrit-a/skycast/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

skycast shows the current weather as a small, living ASCII scene beside a
weather panel. Rain falls and splashes, snow drifts and piles up, the sun's
rays turn, clouds scroll, fog breathes and lightning flashes. Tools like
wttr.in give you a report. skycast gives you a window.

## Features

- **11 animated scenes**: clear day, clear night (twinkling stars, the odd
  shooting star), partly cloudy, overcast, fog, drizzle, rain, freezing rain,
  snow (with accumulation), thunderstorm and hail (bouncing hailstones)
- **Real weather** from [Open-Meteo](https://open-meteo.com), with **no API key** and no sign-up
- **Day and night**: the sun gives way to the moon
- **Wind and intensity**: drizzle vs. downpour, rain that tilts in the wind, wind streaks in a gale
- **Current conditions plus a 3-day forecast** in a panel that adapts to your terminal size
- **Ambient mode**: the scene fills the whole terminal
- **Works offline** from a local cache (with an "offline · updated 2h ago" note)
- `--once` for shell startup, `--ascii` and `--no-color` for limited terminals

## Install

```sh
go install github.com/aashrit-a/skycast@latest
```

Homebrew (once the tap exists):

```sh
brew install aashrit-a/tap/skycast
```

Or download a binary for Linux, macOS or Windows from the
[Releases](https://github.com/aashrit-a/skycast/releases) page.

## Usage

```
skycast [city] [flags]
```

```sh
skycast                       # auto-detect location
skycast tokyo
skycast "san jose, ca" -u imperial
skycast --lat 48.85 --lon 2.35
skycast --once                # print one frame and exit
skycast --demo                # cycle through every scene, no network needed
skycast --scene snow          # force a scene, real weather in the panel
```

| Flag | Description |
|---|---|
| `-u, --units` | `metric` or `imperial` (default: based on the country) |
| `--lat`, `--lon` | coordinates (use both) |
| `--once` | print a single frame and exit |
| `--width`, `--height` | scene size for `--once` (default 40×12) |
| `--demo` | cycle through all scenes with fake data |
| `--ambient` | start in fullscreen ambient mode |
| `--scene` | `clear` `night` `cloudy` `overcast` `fog` `drizzle` `rain` `freezing` `snow` `storm` `hail` |
| `--intensity` | `light` `moderate` `heavy` (with `--scene`) |
| `--fps` | animation frames per second (default 12, range 4–30) |
| `--no-color` | disable colors (also honors `NO_COLOR`) |
| `--ascii` | ASCII-only glyphs, no Unicode or emoji |
| `--no-cache` | don't read cached weather |
| `-v, --version` | print the version |

Multiple words are joined, so `skycast new york` works like `skycast "new york"`.
Add a region or country after a comma to pick the right place: `skycast "paris, us"`.

### Keys

| Key | Action |
|---|---|
| `q`, `esc`, `ctrl+c` | quit |
| `r` | refresh now (bypasses the cache) |
| `u` | toggle °C / °F |
| `p` | toggle the info panel |
| `a` | toggle ambient mode |
| `→` / `n`, `←` / `b` | next / previous scene (`--demo` only) |
| `?` | toggle the key help in the footer |

skycast also refreshes on its own every 15 minutes.

## Scenes

| Scene | Shown for | Details |
|---|---|---|
| `clear` | clear / mainly clear (day) | pulsing sun with rotating rays that "breathe" |
| `night` | clear / mainly clear (night) | crescent moon, twinkling stars, a rare shooting star |
| `cloudy` | partly cloudy | sun or moon behind drifting clouds |
| `overcast` | overcast | layered clouds with parallax |
| `fog` | fog, rime fog | drifting, breathing mist bands; rime sparkles |
| `drizzle` | drizzle | fine, slow droplets |
| `rain` | rain, showers | depth layers, splashes, tilts with the wind |
| `freezing` | freezing rain / drizzle | icy drops, sparkling ground |
| `snow` | snow, snow showers, grains | swaying flakes that pile up |
| `storm` | thunderstorm | heavy rain, roiling clouds, branching lightning |
| `hail` | thunderstorm with hail | all of the above, plus bouncing hailstones |

These samples come from `skycast --once --demo --scene <name>` (scene only;
in the terminal they're animated and in color):

<details><summary><code>--scene clear</code></summary>

```
         ||     //
          |    //
    \\\   ~▓▓▓~~
         ▓▓███▓▓  ----
        ▓▓█████▓▓
   ---- ▓▓▓███▓▓
          ▓▓▓▓   \\\\
        //    |
       //     |
```

</details>

<details><summary><code>--scene night</code></summary>

```
    ·
                   ▓▓▓▓
                  ▓▓▓
·       ·        ▓▓▓▓
                 ▓▓▓▓
      ✧           ▓▓▓▓
                    ▓▓▓▓▓

 ·   ·                      ✧
```

</details>

<details><summary><code>--scene cloudy</code></summary>

```
         ||     //
.-~-.     |    //        .-~~~-.
_____)\    ▓▓▓       .-~~       ~-
         ▓▓███▓▓  --(_____________
        ▓▓█████▓▓
   ---- ▓▓▓███▓▓
          ▓▓▓▓   \\\\
        //    |
       //     |
```

</details>

<details><summary><code>--scene overcast</code></summary>

```
█████▓▓▓▓▓▓▓▓██████▓▓▓▓▓.--~~--.▒▒
)--.░░       ░░░▒▒▒░.--(        )-
--~~--.            (       .-~~~-.
       )--.         `--.-~~
           )          (___________
_______.--'
```

</details>

<details><summary><code>--scene fog</code></summary>

```
░░░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░░░░░░░░░░░
░░░░░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░░░░░░░░░░▒▒
░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░▒▒▒▒▒░░░░░░░░░░
░░░░       ░░░░░░░░░░░░░░░░░░░░░░
    ░░░░░░░░░░░░░░▒▒▒▒▒▒▒▒▒▒▒▒▒░░░
▒▒▒░░░░░░░░░░░░░     ░░░░░░░
░░░░░░░░░░░░  ░░░░░░░░░░░░░░
░░░░░░░░░░░    ░░░░░░░░░░░░      ░
```

</details>

<details><summary><code>--scene drizzle</code></summary>

```
████████████████████▓▓▓▓▓▓▓▓▓▓▓███
▒▒▒▒▒▓▓▒▒▓▓██▓▒░░░░░░░░     ░░  ░░

                          ..  '

                               .
                               '

▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
```

</details>

<details><summary><code>--scene rain</code></summary>

```
██████████▓▓▒▒▒▒▒▒▒▒▓▓▓▓███▓▓▓▓▓▓▓
▒▒▒▒▒▒▒▒▒▒░          ░░░░


            |     |       |
||            |           |    |
                       |   |
   ·        ··   ·       ·  ·
▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
```

</details>

<details><summary><code>--scene freezing</code></summary>

```
██████████████████████████████████
▒▒░░░▒▓▓▓▓▓▒▒░░    ░░░░░░░
           |
 |          |      |
   |
|
||                   |
 *            *  *   | *   *
^▁▁▁▁^▁▁▁▁^▁▁▁▁^▁▁▁▁^▁▁▁▁^▁▁▁▁^▁▁▁
```

</details>

<details><summary><code>--scene snow</code></summary>

```
░░░░▒▒▒▒▒▒░░░░░▒▒▒▒▒▒▒▓▓▓▓▒▒▒▒▒▒▒▒

                       ·        ❄
                      ❅
    ❅                           ❄

          ·                 ·
       •    •    ·
▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▄▁▄▁▁▄▁▄▁▁▄▁▁▁
```

</details>

<details><summary><code>--scene storm</code></summary>

```
██████████████████████████▓▓▓▓▓▓██
█████▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▒▒▒░░    ░░
▒░ ░░▒▒▒░░░░░░░░▒▒░░░░░░░░
 \\\   \ \      \     \     \
  \\   \\\        \  /      \
   \\\  \            |
      \              |           \
°°·\   ° ·    °· ·   |  \ ··  ° °·
▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
```

</details>

<details><summary><code>--scene hail</code></summary>

```
▓▓▓▓▓▓█████████████▓▓███████████▓▓
░░░░   ░░░░░░░▒▒▒▒░░░░  ░░▒▒▒▒▒▒▒▒

          \ \       \\           \
 \ \                \\       \  •
                \\\       \
                \         \  •
· ·  ··° \··  ·° ··   ·   °··
▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
```

</details>

## Add it to your shell startup

To see a weather frame every time you open a terminal, add this to
`~/.zshrc` or `~/.bashrc`:

```sh
skycast --once
```

The forecast is cached for 10 minutes, so new shells open instantly. When
the output is piped, colors are turned off automatically.

## Terminal support

skycast works in any modern terminal on Linux, macOS and Windows (Windows
Terminal, PowerShell). On the legacy Windows `cmd.exe` console, or with
fonts that lack block characters, use `--ascii`.

## Privacy

- Weather and place-name lookups go to [Open-Meteo](https://open-meteo.com)
  (`api.open-meteo.com` and `geocoding-api.open-meteo.com`).
- If you give neither a city nor coordinates, your approximate (city-level)
  location comes from [ip-api.com](https://ip-api.com). Their free tier is
  **HTTP only**, so this request is unencrypted. Pass a city or `--lat/--lon`
  to skip it.
- Nothing else is sent anywhere. The only thing stored is a forecast cache in
  your user cache directory (for example `~/.cache/skycast` on Linux), which
  you can delete at any time.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md). New scenes
make great first issues:

- aurora on clear, cold nights
- sandstorm
- a rainbow after the rain
- falling leaves in autumn

## Roadmap

Planned, each to be tracked as a GitHub issue:

- Hourly forecast strip with a temperature sparkline
- Config file `~/.config/skycast/config.toml` (default city, units, fps)
- Seasonal scenes: autumn leaves, cherry blossoms, aurora on clear cold nights
- Optional ambient rain sound (`--sound`)
- Several cities in a split view
- Weather alerts banner
- One-glyph mode for tmux and status bars (`skycast --glyph` → `🌧 18°`)

## Credits

- Weather data by [Open-Meteo.com](https://open-meteo.com/), licensed under
  [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and
  [Lip Gloss](https://github.com/charmbracelet/lipgloss) from
  [Charm](https://charm.sh), plus [pflag](https://github.com/spf13/pflag).

## License

[MIT](LICENSE)
