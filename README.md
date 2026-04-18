# dotm

[![CI](https://github.com/rpPH4kQocMjkm2Ve/dotm/actions/workflows/ci.yml/badge.svg)](https://github.com/rpPH4kQocMjkm2Ve/dotm/actions/workflows/ci.yml)
![License](https://img.shields.io/github/license/rpPH4kQocMjkm2Ve/dotm)
[![Spec](https://img.shields.io/endpoint?url=https://gitlab.com/fkzys/specs/-/raw/main/version.json&maxAge=300)](https://gitlab.com/fkzys/specs)

Declarative dotfiles, packages, and services manager. Lightweight chezmoi alternative with normal file paths, delegated encryption, proper permission management, and first-class support for `dest = "/"`.

## How it works

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

By default, `apply`, `status`, and `diff` operate on files only. Add scope words to include packages and services:

```
dotm apply --all           # files + pkgs + services
dotm status pkgs           # package status only
dotm diff files services   # file diffs + service changes
```

The repo is the single source of truth. `apply` is a one-directional push. No bidirectional sync, no source state attributes, no magic prefixes.

## Why not chezmoi

| | dotm | chezmoi |
|---|---|---|
| **File naming** | `.config/hypr/hyprland.conf` | `dot_config/private_dot_hypr/hyprland.conf` |
| **Encryption** | Delegate to sops, age, etc. | Built-in age/gpg |
| **Permissions** | First-class `perms` file with glob patterns | Limited (encoded in filename prefixes) |
| **`dest = "/"`** | First-class support, works out of the box | Needs workarounds |
| **Packages** | Declarative, zero hardcoded managers | Via external tools |
| **Complexity** | ~4k LOC | ~50k LOC |

## Installation

### AUR

```bash
yay -S dotm
```

### [gitpkg](https://gitlab.com/fkzys/gitpkg)
```bash
gitpkg install dotm
```

### Manual

```bash
git clone https://gitlab.com/fkzys/dotm.git
cd dotm
sudo make install
```

### Build only

```bash
make build          # produces ./dotm binary
# or directly:
go build -o dotm ./cmd/dotm/
```

Requires Go 1.24+.

## Repository layout

```
~/dotfiles/
├── dotm.toml           # config (required)
├── files/              # your actual dotfiles (mirrors dest)
│   ├── .config/
│   │   ├── hypr/
│   │   │   └── hyprland.conf
│   │   └── waybar/
│   │       ├── config.jsonc
│   │       └── style.css.tmpl
│   ├── .zshrc.tmpl
│   └── etc/            # system files when dest = "/"
│       └── pacman.conf
├── perms               # permission rules (optional)
├── ignore.tmpl         # ignore patterns (optional)
└── scripts/            # lifecycle scripts (optional)
    └── reload.sh
```


Files under `files/` are copied to `dest` preserving directory structure. Files ending in `.tmpl` are rendered as Go templates, with the suffix stripped in the output.

## Configuration

`dotm.toml`:

```toml
dest = "~"

# Interactive prompts — values available in templates as {{ .use_nvidia }}
[prompts]
use_nvidia = { type = "bool", question = "Enable NVIDIA config?" }
git_email  = { type = "string", question = "Git email address" }

# Symlinks: link (relative to dest) → target
[symlinks]
".local/bin/editor" = "{{ .homeDir }}/.nix-profile/bin/nvim"

# Lifecycle scripts
[[scripts]]
path = "scripts/reload.sh"
template = true         # render as template before running
trigger = "on_change"   # "always" or "on_change"
```

For system-wide configuration:

```toml
dest = "/"
```

Files in `files/etc/pacman.conf` deploy to `/etc/pacman.conf`.

## Usage

```bash
dotm init              # resolve prompts, create state cache
dotm apply             # deploy files, symlinks, perms, scripts
dotm apply -n          # dry run — show what would happen
dotm apply --all       # deploy files + packages + services
dotm diff              # unified diff between source and dest
dotm diff pkgs         # show packages to install/remove
dotm diff services     # show services to enable/disable
dotm status            # show sync state of managed files
dotm status pkgs       # show package status
dotm status services   # show service status
dotm status -v         # include clean files
dotm status -q         # exit 1 if problems exist, no output
dotm help              # show help
```

### Example apply

```
$ dotm apply
mkdir /home/user/.config/hypr
write /home/user/.config/hypr/hyprland.conf (2847 bytes)
write /home/user/.zshrc (1204 bytes)
symlink /home/user/.local/bin/editor -> /home/user/.nix-profile/bin/nvim
run scripts/reload.sh
```

### Example status

```
$ dotm status
  modified   .config/waybar/style.css
  missing    .config/foot/foot.ini
  orphan     .config/sway/config
```

Four states:
- **clean** — dest matches rendered source
- **modified** — dest differs from source
- **missing** — in source but not yet in dest
- **orphan** — was deployed previously, no longer in source, still in dest

dotm never auto-deletes orphans. It reports them; you decide.

## Templates

Files ending in `.tmpl` are rendered with Go's `text/template`.

### Built-in variables

| Variable | Value |
|----------|-------|
| `{{ .homeDir }}` | User home directory |
| `{{ .hostname }}` | System hostname |
| `{{ .username }}` | Current user |
| `{{ .sourceDir }}` | Absolute path to dotfiles repo |

Prompt values are available by name: `{{ .use_nvidia }}`, `{{ .git_email }}`.

### Custom functions

| Function | Description |
|----------|-------------|
| `output "cmd" "arg1" "arg2"` | Run command, return stdout |
| `fromYaml` | Parse YAML string into map |
| `joinPath "a" "b"` | `filepath.Join` |
| `hasKey $map "key"` | Check if map contains key |
| `replace "old" "new" $s` | Replace all occurrences |
| `default "fallback" $value` | Return fallback if value is empty/nil |

### Example: secrets via sops

`files/.config/app/config.yaml.tmpl`:

```yaml
{{ $s := output "sops" "-d" (joinPath .sourceDir "secrets.enc.yaml") | fromYaml -}}
api_key: {{ index $s "api_key" }}
db_password: {{ index $s "db_password" }}
```

No built-in encryption. sops/age/gpg handle decryption; dotm handles templating and deployment.

### Example: conditional blocks

`files/.zshrc.tmpl`:

```bash
export EDITOR="{{ .editor }}"

{{ if .use_nvidia -}}
export __GL_SHADER_DISK_CACHE_PATH="{{ .homeDir }}/.cache/nv"
export __GL_SHADER_DISK_CACHE_SIZE=1073741824
{{ end -}}
```

## Ignore patterns

`ignore.tmpl` (rendered as template, then parsed as glob patterns):

```
# Always ignore
.git/**
*.swp
.DS_Store

# Conditional
{{ if not .use_nvidia -}}
.config/nvidia/**
{{ end -}}
```

Patterns are matched against paths relative to `files/`. Supports `*`, `?`, `**`.

## Permission management

The `perms` file sets mode, owner, and group on deployed files:

```bash
# pattern              mode   owner  group
# Trailing / = directories only, no / = files only
# - = don't change that attribute
# Last matching rule wins

etc/**/                0755   root   root
etc/**                 0644   root   root

etc/security/          0700   root   root
etc/security/**        0600   root   root

etc/polkit-1/rules.d/**  0640  root  polkitd

root/**/               0700   root   root
root/**                0600   root   root
```

Glob patterns support `*`, `?`, `**`. Rules are evaluated top-to-bottom; last match wins. Directory rules (trailing `/`) only match directories; file rules only match files.

If no `perms` file exists, all deployed files receive default `0o644` permissions and directories receive `0o755` after write. Files are initially written with `0o600` to prevent a window where sensitive files are world-readable, then lifted to the defaults once the apply completes. Add a `perms` file to override these defaults for specific paths.

This is the primary reason dotm exists as a separate tool — managing `/etc` permissions correctly matters, and encoding `0640 root:polkitd` in a filename is not a serious approach.

## Scripts

```toml
[[scripts]]
path = "scripts/reload-hypr.sh"
template = false
trigger = "always"        # run on every apply

[[scripts]]
path = "scripts/setup.sh.tmpl"
template = true           # render before running
trigger = "on_change"     # run only when content changes
```

Scripts are executed with `bash`. `on_change` tracks content hash in state — if the rendered script hasn't changed since last apply, it's skipped.

## Package and service management

Packages and services are managed via declarative **managers**. A manager defines command templates for `check`, `install`, `remove`, `enable`, and `disable`. Groups reference a manager by name and list packages or services.

```toml
[managers.pacman]
check   = "pacman -Q {{.Name}}"
install = "sudo pacman -S --needed {{.Name}}"
remove  = "sudo pacman -Rns {{.Name}}"

[managers.aur]
check   = "pacman -Q {{.Name}}"
install = "aur sync --no-view -n {{.Name}} && sudo pacman -S --needed {{.Name}}"
remove  = "sudo pacman -Rns {{.Name}}"

[managers.systemd]
check   = "systemctl is-enabled {{.Name}}"
enable  = "sudo systemctl enable {{.Name}}"
disable = "sudo systemctl disable {{.Name}}"

[managers.systemd-user]
check   = "systemctl --user is-enabled {{.Name}}"
enable  = "systemctl --user enable {{.Name}}"
disable = "systemctl --user disable {{.Name}}"

[managers.flatpak]
check   = "flatpak info '{{.Name}}' >/dev/null 2>&1"
install = "flatpak install -y --noninteractive {{.Name}}"
remove  = "flatpak uninstall -y --noninteractive {{.Name}}"

[managers.gitpkg]
check   = "gitpkg list 2>/dev/null | grep -q '^{{.Name}} '"
install = "sudo gitpkg install {{.Name}}"
remove  = "sudo gitpkg remove {{.Name}}"

[managers.npm]
check   = "pkg={{.Name}}; pkg=${pkg%@*}; test -d $(npm root -g)/$pkg"
install = "npm install -g {{.Name}}"
remove  = "npm uninstall -g {{.Name}}"

[pacman]
packages = [
    "hyprland",
    "neovim",
    "{{ if .laptop }}brightnessctl{{ end }}",
]

[aur]
packages = [
    "kopia-bin",
    "coolercontrol-bin",
]

[systemd]
services = ["firewalld", "systemd-oomd"]

[systemd-user]
services = ["hypridle", "waybar", "mpd"]

[flatpak]
packages = [
    "org.mozilla.firefox",
    "org.telegram.desktop",
    "{{ if .portproton }}ru.linux_gaming.PortProton{{ end }}",
]

[gitpkg]
packages = [
    "verify-lib",
    "bwrap-common",
    "hardened_malloc",
]

[npm]
packages = [
    "@qwen-code/qwen-code@latest",
]
```

Package and service names may contain Go template expressions. If a name renders to an empty string, the entry is skipped.

### Conditional packages and services

Package ande service names may include Go template expressions. If a rendered name is empty, it's skipped.

**Option 1: Inline conditionals per package:**
```toml
[pacman]
packages = [
    "hyprland",
    "{{ if .laptop }}brightnessctl{{ end }}",
    "{{ if .laptop }}tpm2-tss{{ end }}",
    "{{ if .laptop }}bluez{{ end }}",
]
```

**Option 2: Multi-line block with one condition:**
```toml
[pacman]
packages = [
    "hyprland",
    """
    {{ if .laptop }}
    brightnessctl
    tpm2-tss
    bluez
    {{ end }}
    """
]
```

### How it works

1. `dotm apply pkgs` reads the config and loads the previous manifest
2. For each package: runs `check` → if not installed, runs `install`
3. For each package in manifest but not in config: if still installed, runs `remove`
4. For each service: runs `check` → if not enabled, runs `enable`
5. For each service in manifest but not in config: if still enabled, runs `disable`
6. Saves new manifest to state

Use `dotm diff pkgs` to preview what would be installed/removed, and `dotm status pkgs` to see current package status.

### Adding a manager

No code changes needed — just add a section to `[managers]`:

```toml
[managers.flatpak]
check   = "flatpak list --app --columns=application | grep -qxF {{.Name}}"
install = "flatpak install -y --noninteractive {{.Name}}"
remove  = "flatpak uninstall -y --noninteractive {{.Name}}"

[flatpak]
packages = ["org.mozilla.firefox"]
```

### Status output

```
$ dotm status
  modified   .config/waybar/style.css
  missing    .config/foot/foot.ini

Packages:
  OK       hyprland (pacman)
  MISSING  brightnessctl (pacman)
  OBSOLETE old-tool (pacman) — still installed

Services:
  ENABLED  firewalld (systemd)
  DISABLED waybar (systemd-user)
```

## Security

### Temporary files

Scripts and diff operations use temporary files. By default, `dotm` creates
these in `$XDG_RUNTIME_DIR/dotm/` (typically `/run/user/<uid>/`) or
`$HOME/.local/state/dotm/tmp/`, both with mode `0700`. This prevents
symlink race attacks that are possible when using the world-writable `/tmp`.

If neither secure directory is available, `dotm` falls back to the system
temp directory.

## State

dotm stores state in `~/.local/state/dotm/<hash>.toml`:

- **Prompt answers** — cached so you're not asked every apply
- **Script hashes** — for `on_change` trigger
- **Manifest** — list of deployed files, directories, symlinks for orphan detection
- **Pkg manifest** — list of deployed packages and services for orphan tracking

Each source repo gets its own state file (keyed by SHA-256 of absolute path). Re-run `dotm init` to re-answer prompts.

## Tests

```bash
make test               # run all tests
make test-root          # run permission tests (requires root)
```

Permission tests (`internal/perms/`) need root for `chmod`/`chown` verification.

## Dependencies

Runtime: none (static Go binary).

Build: Go 1.24+.

Go module dependencies:
- `github.com/BurntSushi/toml` — TOML parsing
- `gopkg.in/yaml.v3` — `fromYaml` template function

External tools (optional, used by templates at runtime):
- `sops` — if your templates call `output "sops" ...`
- `diff` — used by `dotm diff` (present on any unix system)
- `bash` — script execution

## License

AGPL-3.0-or-later
