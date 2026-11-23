# dotm

[![CI](https://github.com/fkzys/dotm/actions/workflows/ci.yml/badge.svg)](https://github.com/fkzys/dotm/actions/workflows/ci.yml)
![License](https://img.shields.io/github/license/fkzys/dotm)
[![Spec](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/fkzys/specs/refs/heads/main/version.json&maxAge=300)](https://github.com/fkzys/specs)

Declarative dotfiles, packages, and services manager with plain file paths, delegated encryption, proper permission management, and first-class support for `dest = "/"`.

## How apply works

```
dotm apply
      ↓
  1. Walk files/ directory
  2. Render .tmpl templates with Go template engine
  3. Write to dest (skip unchanged files)
  4. Create symlinks from [symlinks] map
  5. Apply permission rules from perms file
  6. Run lifecycle scripts
  7. Record manifest for orphan tracking
```

## Why not chezmoi
Chezmoi have large codebase to support dozen of tools and that bloats one.\
Chezmoi literaly don't have somewhat saneful permission system.\
Chezmoi is not intended to use with dest = "/" which I want to be used.

## Docs
- **[usage](docs/usage.md)**
    - [Example repo layout](docs/usage.md#example-repo-layout)
        - [dotfiles](docs/usage.md#dotfiles)
        - [rootfiles](docs/usage.md#rootfiles)
    - [Configuration](docs/usage.md#configuration)
        - [Scripts](docs/usage.md#scripts)
        - [Built-in variables](docs/usage.md#built-in-variables)
        - [Custom functions](docs/usage.md#custom-functions)
        - [Example: secrets via sops](docs/usage.md#example-secrets-via-sops)
        - [Example: conditional blocks](docs/usage.md#example-conditional-blocks)
        - [Ignore patterns](docs/usage.md#ignore-patterns)
    - [Permission management](docs/usage.md#permission-management)
    - [Package and service management](docs/usage.md#package-and-service-management)
        - [Conditional packages and services](docs/usage.md#conditional-packages-and-services)
        - [Adding a manager](docs/usage.md#adding-a-manager)
- **[notes](docs/notes.md)**
    - [Temporary files](docs/notes.md#temporary-files)
    - [State](docs/notes.md#state)
- **[Migrate](MIGRATE.md)**
## Installation

### AUR

```bash
yay -S dotm
```

### [gitpkg](https://github.com/fkzys/gitpkg)
```bash
gitpkg install dotm
```

### Manual

```bash
git clone https://github.com/fkzys/dotm.git
cd dotm
sudo make install
```

### Build

```bash
make build
```
or directly:
```bash
go build -o dotm ./cmd/dotm/
```

## Usage

By default, `apply`, `status`, and `diff` operate on files only. Add scope words to include packages and services.

```bash
dotm init
dotm apply             # deploy everything but pkgs/services
dotm apply -n          # dry run
dotm apply --all       # deploy files + packages + services
dotm diff
dotm diff pkgs
dotm diff services
dotm status            # show sync state
dotm status pkgs
dotm status services
dotm status -v         # include clean files
dotm status -q         # exit 1 if problems exist, no output
dotm help
```

### Example dry run

```
$ dotm apply -n
mkdir /home/user/.config/hypr
write /home/user/.config/hypr/hyprland.conf (2847 bytes)
write /home/user/.zshrc (1204 bytes)
symlink /home/user/.local/bin/editor -> /home/user/.nix-profile/bin/nvim
run scripts/reload.sh
```

### Example status

```
$ dotm status --all --verbose
  modified   .config/waybar/style.css
  missing    .config/foot/foot.ini
  orphan     .config/sway/config

Packages:
  OK       hyprland (pacman)
  MISSING  brightnessctl (pacman)
  OBSOLETE old-tool (pacman) — still installed

Services:
  ENABLED  firewalld (systemd)
  DISABLED waybar (systemd-user)

```

Four states:
- **clean** — dest matches rendered source
- **modified** — dest differs from source
- **missing** — in source but not yet in dest
- **orphan** — was deployed previously, no longer in source, still in dest

dotm never auto-deletes orphans.

### How pkg management works

1. `dotm apply pkgs` reads the config and loads the previous manifest
2. For each package: runs `check` → if not installed, runs `install`
3. For each package in manifest but not in config: if still installed, runs `remove`
4. For each service: runs `check` → if not enabled, runs `enable`
5. For each service in manifest but not in config: if still enabled, runs `disable`
6. Saves new manifest to state

Use `dotm diff pkgs` to preview what would be installed/removed, and `dotm status pkgs` to see current package status.

## Tests

```bash
make test               # run all tests
make test-root          # run permission tests
```

Permission tests (`internal/perms/`) need root for `chmod`/`chown` verification.

## Dependencies

Build: Go 1.24+.

Go module dependencies:
- `github.com/BurntSushi/toml` — TOML parsing
- `gopkg.in/yaml.v3` — `fromYaml` template function

External optional tools:
- `sops` — if your templates call `output "sops" ...`
- `diff` — used by `dotm diff`
- `bash` — script execution

## License

AGPL-3.0-or-later
