# synthwaves

A terminal client for [synthwaves.fm](https://synthwaves.fm) -- browse your music library, manage playlists, and stream audio from the command line.

## Features

- **Interactive TUI** -- full-screen terminal UI with keyboard navigation, search, and audio playback
- **Audio streaming** -- play tracks and live radio directly in your terminal
- **Live radio** -- stream public radio stations with inline album art (sixel)
- **Library management** -- browse and manage tracks, albums, artists, playlists, favorites, tags, and play history
- **Multiple output formats** -- table, JSON, or plain text for scripting

## Install

### Homebrew

First, add the tap (one-time setup):

```
brew tap leopolicastro/tap
```

Then install:

```
brew install synthwaves
```

### Go

Requires Go 1.26+.

```
go install github.com/leopolicastro/synthwaves.fm-cli@latest
```

### From source

```
git clone https://github.com/leopolicastro/synthwaves.fm-cli.git
cd synthwaves.fm-cli
make build
```

## Setup

1. Deploy your own [synthwaves.fm](https://github.com/leopolicastro/synthwaves.fm) instance
2. Create an API key from within your synthwaves.fm dashboard
3. Run the login command:

```
synthwaves auth login
```

This prompts for your server URL (the domain where you deployed synthwaves.fm) and API credentials, tests the connection, and saves the config to `~/.config/synthwaves/config.toml`.

## Usage

Run `synthwaves` with no arguments to launch the interactive TUI.

### Commands

```
synthwaves tracks list              List tracks
synthwaves tracks show <id>         Show track details
synthwaves albums list              List albums
synthwaves albums show <id>         Show album details
synthwaves artists list             List artists
synthwaves artists show <id>        Show artist details
synthwaves playlists list           List playlists
synthwaves playlists create         Create a playlist
synthwaves favorites list           List favorites
synthwaves favorites add            Add a favorite
synthwaves search <query>           Search your library
synthwaves history list             List play history
synthwaves stats show               Show listening stats
synthwaves radio list               List radio stations
synthwaves radio start <id>         Start a radio station
synthwaves radio stop <id>          Stop a radio station
synthwaves live list                List public live radio stations
synthwaves live play <slug>         Stream a live radio station
synthwaves tags list                List tags
synthwaves profile show             Show your profile
synthwaves auth login               Set up API credentials
synthwaves config show              Show current config
synthwaves config set <key> <val>   Update a config value
```

### Flags

```
-f, --format <fmt>    Output format: table, json, text (default: table)
-v, --verbose         Verbose output
    --config <path>   Config file path
    --page <n>        Page number (default: 1)
    --per-page <n>    Items per page (default: 24, max: 100)
    --sort <col>      Sort column
    --direction <dir> Sort direction: asc, desc
```

## Configuration

Config is stored at `~/.config/synthwaves/config.toml` (respects `XDG_CONFIG_HOME`):

```toml
base_url = "https://your-instance.example.com"
client_id = "bc_..."
secret_key = "your-secret-key"
```

## Built with

- [Cobra](https://github.com/spf13/cobra) -- CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) -- TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) -- Terminal styling
- [Huh](https://github.com/charmbracelet/huh) -- Interactive forms
- [Beep](https://github.com/gopxl/beep) -- Audio playback

## Releasing a new version

Releases are automated with GitHub Actions and GoReleaser.

### One-time setup

Add a `HOMEBREW_TAP_TOKEN` repository secret in GitHub with write access to `leopolicastro/homebrew-tap`.

### Release flow

1. Tag the new version and push it:

```
git tag v0.X.Y
git push github v0.X.Y
```

2. GitHub Actions will:

- build release archives for macOS and Linux
- publish the GitHub release
- update `Formula/synthwaves.rb` in `leopolicastro/homebrew-tap`

3. Upgrade locally:

```
brew upgrade synthwaves
```

## License

[MIT](LICENSE)
