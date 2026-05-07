# amd-report

Static DayZ ADM event reporting tools.

This repo contains three CLIs:

1. `init`
   - scaffolds an event folder
   - writes `config.yaml`
   - writes `teams.yaml` for team events
2. `eventbuild`
   - reads `config.yaml`
   - parses the ADM log and reviewed team data
   - writes the JSON event bundle
3. `renderhtml`
   - reads `config.yaml`
   - renders a static HTML report from the JSON bundle

## Build and run

```bash
go run ./cmd/init
go run ./cmd/eventbuild
go run ./cmd/renderhtml
```

## Workflow

1. Run `init` to create an event folder.
2. Review `teams.yaml` if `teams_event: true`.
3. Run `eventbuild --config ...`.
4. Run `renderhtml --config ...`.

## 1. init

`init` is interactive by default. It prompts for:

- event folder
- ADM log path
- start time, defaulting to the earliest timestamp found in the log
- end time, defaulting to the latest timestamp found in the log
- event name
- assist window
- whether the event is a team event

At the end of interactive init, if the log file is outside the event folder, it offers to move the log file into that folder. The default answer is yes.

Example:

```bash
go run ./cmd/init
```

You can also prefill or fully bypass prompts:

```bash
go run ./cmd/init \
  --dir ./battle-at-blackjack \
  --log-file ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 15:33:25 \
  --end 20:52:21 \
  --event-name "Battle at Blackjack Valley" \
  --assist-window-seconds 30 \
  --teams-event=true \
  --no-input
```

Flags:

- `--dir`: event folder to create
- `--log-file`: ADM log path
- `--start`: event start time, `HH:MM:SS`
- `--end`: event end time, `HH:MM:SS`
- `--event-name`: event name
- `--assist-window-seconds`: assist window in seconds
- `--teams-event`: `true` or `false`
- `--no-input`: disable prompts and require missing values from flags
- `--force`: allow overwrite of existing `config.yaml` or `teams.yaml`

`init` writes:

- `config.yaml`
- `teams.yaml` only when `teams_event: true`

It does not pre-create `data/` or `html/`.

## 2. eventbuild

Generate the JSON event bundle from `config.yaml`.

```bash
go run ./cmd/eventbuild \
  --config ./battle-at-blackjack/config.yaml
```

Optional overrides:

```bash
go run ./cmd/eventbuild \
  --config ./battle-at-blackjack/config.yaml \
  --teams ./custom-teams.yaml \
  --out ./custom-data
```

Flags:

- `--config`: path to `config.yaml`
- `--teams`: optional override for `teams.yaml`
- `--out`: optional override for `data_dir`

`eventbuild` writes:

- `metadata.json`
- `roster.json`
- `events.json`
- `players.json`
- `teams.json`
- `summary.json`

If `teams_event: true`, `eventbuild` requires sibling `teams.yaml` unless `--teams` is set.

If `teams_event: false`, `eventbuild` runs without team config unless `--teams` is set explicitly.

`eventbuild` still scans the full ADM for roster discovery. Any non-ignored player seen anywhere in the file is included in the JSON bundle and HTML site, even if they never take part in combat during the event window.

## 3. renderhtml

Generate the static HTML site from the bundle described by `config.yaml`.

```bash
go run ./cmd/renderhtml \
  --config ./battle-at-blackjack/config.yaml
```

Optional overrides:

```bash
go run ./cmd/renderhtml \
  --config ./battle-at-blackjack/config.yaml \
  --data-dir ./custom-data \
  --out ./custom-html
```

Flags:

- `--config`: path to `config.yaml`
- `--data-dir`: optional override for `data_dir`
- `--out`: optional override for `html_dir`

## config.yaml

Example team event config:

```yaml
log_file: ../DayZServer_x64_2026-05-02_15-33-25.ADM
start: "15:33:25"
end: "20:52:21"
event_name: Battle at Blackjack Valley
assist_window_seconds: 30
teams_event: true
data_dir: data
html_dir: html
```

Example singles config:

```yaml
log_file: ../DayZServer_x64_2026-05-02_15-33-25.ADM
start: "17:40:00"
end: "18:30:00"
event_name: Sunday Singles Event
assist_window_seconds: 30
teams_event: false
data_dir: data
html_dir: html
```

Notes:

- `log_file`, `data_dir`, and `html_dir` resolve relative to `config.yaml`
- absolute paths still work
- `start` and `end` use `HH:MM:SS`
- `event_name` is required
- `assist_window_seconds` is always explicit

## teams.yaml

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

Ignored players are dropped from the roster, events, stats, summaries, and rendered output.

## Example: teams event

1. Scaffold the event:

```bash
go run ./cmd/init \
  --dir ./battle-at-blackjack \
  --log-file ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 15:33:25 \
  --end 20:52:21 \
  --event-name "Battle at Blackjack Valley" \
  --assist-window-seconds 30 \
  --teams-event=true \
  --no-input
```

2. Review `./battle-at-blackjack/teams.yaml`.

3. Build data:

```bash
go run ./cmd/eventbuild \
  --config ./battle-at-blackjack/config.yaml
```

4. Render HTML:

```bash
go run ./cmd/renderhtml \
  --config ./battle-at-blackjack/config.yaml
```

5. Open `./battle-at-blackjack/html/index.html`.

## Example: singles event

1. Scaffold the event:

```bash
go run ./cmd/init \
  --dir ./sunday-singles \
  --log-file ./DayZServer_x64_2026-05-02_15-33-25.ADM \
  --start 17:40:00 \
  --end 18:30:00 \
  --event-name "Sunday Singles Event" \
  --assist-window-seconds 30 \
  --teams-event=false \
  --no-input
```

2. Build data:

```bash
go run ./cmd/eventbuild \
  --config ./sunday-singles/config.yaml
```

3. Render HTML:

```bash
go run ./cmd/renderhtml \
  --config ./sunday-singles/config.yaml
```

4. Open `./sunday-singles/html/index.html`.
