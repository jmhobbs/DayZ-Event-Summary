# amd-report

Static DayZ ADM event reporting tools.

This repo contains three CLIs:

1. `teamguess`
   - Scans one ADM file inside a time window
   - Suggests team groupings in YAML
   - Adds `display_name` for grouped players
2. `eventbuild`
   - Reads one ADM file, event settings, and reviewed team YAML
   - Produces normalized JSON plus precomputed aggregates
3. `renderhtml`
   - Reads the JSON bundle from `eventbuild`
   - Produces a self-contained static HTML report

## Build and run

Run each tool with `go run`:

```bash
go run ./cmd/teamguess
go run ./cmd/eventbuild
go run ./cmd/renderhtml
```

## Tool usage

### 1. teamguess

Generate provisional team suggestions for one event window.

```bash
go run ./cmd/teamguess \
  --adm ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 15:33:25 \
  --end 20:52:21 \
  --out ./team-suggestions.yaml
```

Flags:

- `--adm`: ADM file path
- `--start`: event start time, `HH:MM:SS` or RFC3339
- `--end`: event end time, `HH:MM:SS` or RFC3339
- `--out`: output YAML path, optional, defaults to stdout

### 2. eventbuild

Generate the JSON event bundle.

```bash
go run ./cmd/eventbuild \
  --adm ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 15:33:25 \
  --end 20:52:21 \
  --event-config ./event-settings.yaml \
  --teams ./team-suggestions.yaml \
  --out ./event-data
```

Flags:

- `--adm`: ADM file path
- `--start`: event start time, `HH:MM:SS` or RFC3339
- `--end`: event end time, `HH:MM:SS` or RFC3339
- `--event-config`: YAML event settings
- `--teams`: reviewed team YAML
- `--out`: output directory for the JSON bundle

`eventbuild` writes:

- `metadata.json`
- `roster.json`
- `events.json`
- `players.json`
- `teams.json`
- `summary.json`

### 3. renderhtml

Generate the static HTML site from the JSON bundle.

```bash
go run ./cmd/renderhtml \
  --data-dir ./event-data \
  --out ./site
```

Flags:

- `--data-dir`: directory created by `eventbuild`
- `--out`: output directory for the rendered site

## Config files

### Event settings

Example `event-settings.yaml`:

```yaml
event_name: Saturday Teams Event
assist_window_seconds: 30
```

### Team config

`teamguess` writes the same shape that `eventbuild` consumes.

Grouped team members can include:

- `preferred_name`: full display name
- `display_name`: display-first name with the team prefix removed

Ungrouped players keep only `preferred_name`.

You can also add an ignored list:

```yaml
ignored:
  - player_id: some-player-id
    preferred_name: Some Player
```

`eventbuild` will drop ignored players from the roster, events, stats, summaries, and rendered output.

`eventbuild` still scans the full ADM for roster discovery. Any non-ignored player seen anywhere in the file is included in the JSON bundle and HTML site, even if they never take part in combat during the event window.

## Example: teams event

1. Generate suggestions:

```bash
go run ./cmd/teamguess \
  --adm ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 15:33:25 \
  --end 20:52:21 \
  --out ./team-suggestions.yaml
```

2. Review and edit `team-suggestions.yaml`.

3. Create `event-settings.yaml`:

```yaml
event_name: Saturday Teams Event
assist_window_seconds: 30
```

4. Build the event bundle:

```bash
go run ./cmd/eventbuild \
  --adm ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 15:33:25 \
  --end 20:52:21 \
  --event-config ./event-settings.yaml \
  --teams ./team-suggestions.yaml \
  --out ./event-data
```

5. Render the HTML:

```bash
go run ./cmd/renderhtml \
  --data-dir ./event-data \
  --out ./site
```

Open `./site/index.html`.

## Example: singles event

For a singles event, skip team guessing and provide an empty team file.

Create `singles-settings.yaml`:

```yaml
event_name: Sunday Singles Event
assist_window_seconds: 30
```

Create `singles-teams.yaml`:

```yaml
teams: []
ungrouped: []
```

Build the event bundle:

```bash
go run ./cmd/eventbuild \
  --adm ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 17:40:00 \
  --end 18:30:00 \
  --event-config ./singles-settings.yaml \
  --teams ./singles-teams.yaml \
  --out ./singles-data
```

Render the HTML:

```bash
go run ./cmd/renderhtml \
  --data-dir ./singles-data \
  --out ./singles-site
```

Open `./singles-site/index.html`.
