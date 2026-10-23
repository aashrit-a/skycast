# Contributing to skycast

Thanks for helping make the terminal a nicer place to check the weather.

## Development setup

You need Go 1.23 or newer (older Go builds binaries that current macOS refuses to run).

```sh
make run ARGS="tokyo"   # run from source; skycast flags go in ARGS
make demo               # cycle through every scene, no network needed
make test               # go test ./... -race
make lint               # golangci-lint run
make build              # builds bin/skycast
make gif                # records assets/demo.gif (needs charmbracelet/vhs)
```

The tests never touch the network. The API clients are tested against
`httptest` servers that serve the fixtures in `testdata/`.

### Debugging aids

- `skycast <city> --debug-fetch` (a hidden flag) resolves the location,
  fetches the forecast and prints the parsed report as plain text instead of
  starting the UI. It uses the cache the same way the app does; add
  `--no-cache` to force a network request.
- `skycast --once --demo --scene <name>` prints any scene without the network.
- To look at raw scene frames without the UI, use the preview test:

  ```sh
  SKYCAST_PREVIEW=6 SKYCAST_IN=2 SKYCAST_WIND=0.5 SKYCAST_W=60 SKYCAST_H=16 \
    SKYCAST_STEPS=40 go test ./internal/scene -run TestPreview -v
  ```

  `SKYCAST_PREVIEW` is the `weather.Condition` number (0 = clear day …
  10 = hail) and `SKYCAST_IN` is the intensity (0–2). Set `SKYCAST_NIGHT=1`
  for night and `SKYCAST_ASCII=1` for ASCII glyphs.

### Golden files

`internal/ui/once_test.go` compares `--once` output with
`testdata/once_rain.golden`. If you change rendering on purpose, regenerate
the file and review the diff:

```sh
go test ./internal/ui -run TestOnceGolden -update
git diff testdata/
```

## Project layout

| Package | What it does |
|---|---|
| `internal/config` | flag parsing and validation |
| `internal/location` | geocoding and IP lookup |
| `internal/weather` | Open-Meteo client, WMO code mapping, cache, units |
| `internal/scene` | canvas, palette, particle helpers and every scene |
| `internal/ui` | Bubble Tea model, layout, panel, `--once`, `--demo` |

## Adding a new scene

1. Create `internal/scene/<name>.go` with a type that implements `Scene`
   (`Resize`, `Step`, `Draw`, `Name`). Take a `*rand.Rand` and use only that
   for randomness, so tests stay deterministic. Position everything relative
   to `w`/`h`. Scenes should look good from 24×10 up to 120×40 and must not
   panic at any size, including 1×1.
2. Give every non-ASCII glyph an ASCII fallback (`Params.ASCIIOnly`).
3. Register the scene in `scene.New`, adding a `weather.Condition` if needed.
4. Add a `--scene` name to `config.SceneNames` and `scene.ByName`.
5. Add an entry to `demoScenes` in `internal/ui/demo.go`.
6. Add your condition to `allParams()` in `internal/scene/scene_test.go`,
   which runs every scene through the panic, resize and determinism checks.
   If the scene has particles, implement `particles()` and `targetCount()` so
   the density check covers it too.
7. Add the scene to the table in the README.

## Code style

- Code must be `gofmt`-clean and pass `make lint`.
- Exported identifiers get doc comments.
- New logic comes with tests.
- Keep files small and readable (roughly 300 lines at most).
- We keep dependencies to a minimum, so please open an issue before adding one.

## Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org):

```
feat: aurora scene for clear cold nights
fix: rain splashes drawn off-screen on resize
docs: explain --once in the README
chore: bump bubbletea
```
