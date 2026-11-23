# dotm

## Example repo layout

Files under `files/` are copied to `dest` preserving directory structure.\
Files ending in `.tmpl` are rendered as Go templates, with the suffix stripped in the output.

### dotfiles

```
~/dotfiles/
├── dotm.toml           # required config
├── files/              # your actual dotfiles
│   ├── .config/
│   │   ├── hypr/
│   │   │   └── hyprland.conf
│   │   └── waybar/
│   │       ├── config.jsonc
│   │       └── style.css.tmpl
│   ├── .zshrc.tmpl
├── perms               # optional
├── ignore.tmpl         # optional
└── scripts/            # optional lifecycle scripts
    └── reload.sh
```

### rootfiles

You can use this function in your .zshrc to pass your age file to the sops root process.\
Of course you need to change `keyfile` according to your age key location.
```bash
sops() {
    local keyfile="$HOME/keys/age/dotm-pq.txt"
    local user runuser_bin cat_bin key_cmd

    user="$(id -un)" || return
    runuser_bin="$(command -v runuser)" || return
    cat_bin="$(command -v /usr/bin/cat)" || return

    key_cmd="$(printf '%q ' "$runuser_bin" -u "$user" -- "$cat_bin" -- "$keyfile")" || return
    key_cmd="${key_cmd% }"

    sudo -v || return
    sudo -n "SOPS_AGE_KEY_CMD=$key_cmd" "$@"
}
```

```
~/rootfiles/
├── dotm.toml           # required config
├── files/              # your actual rootfiles
│   ├── etc/
│   │   ├── mkinitcpio.d/
│   │   │   └── linux.preset
│   │   ├── systemd/
│   │   │   └── journald.conf
│   │   └── pacman.conf
│   ├── root/
│   │   └── .zshrc
├── perms               # optional
├── ignore.tmpl         # optional
└── scripts/            # optional lifecycle scripts
    └── reload.sh
```

## Configuration

`dotm.toml`:

```toml
dest = "~"

# Values available in templates as {{ .use_nvidia }}
[prompts]
use_nvidia = { type = "bool", question = "Enable NVIDIA config?" }
git_email  = { type = "string", question = "Git email address" }

# Relative to dest link -> target
[symlinks]
".local/bin/editor" = "{{ .homeDir }}/.nix-profile/bin/nvim"
```

For system-wide configuration:

```toml
dest = "/"
```

Files in `files/etc/pacman.conf` deploy to `/etc/pacman.conf`.

### Scripts

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

### Example: conditional blocks

`files/.zshrc.tmpl`:

```bash
export EDITOR="{{ .editor }}"

{{ if .use_nvidia -}}
export __GL_SHADER_DISK_CACHE_PATH="{{ .homeDir }}/.cache/nv"
export __GL_SHADER_DISK_CACHE_SIZE=1073741824
{{ end -}}
```

### Ignore patterns

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

Glob patterns support `*`, `?`, `**`. Rules are evaluated top-to-bottom; last match wins.\
Directory rules (trailing `/`) only match directories; file rules only match files.

If no `perms` file exists, all deployed files receive default `0o644` permissions and directories receive `0o755` after write.\
Files are initially written with `0o600` to prevent a window where sensitive files are world-readable, then lifted to the defaults once the apply completes.\
Add a `perms` file to override these defaults for specific paths.

This is the primary reason dotm exists as a separate tool — managing `/etc` permissions correctly matters, and encoding `0640 root:polkitd` in a filename is not a serious approach.

## Package and service management

Packages and services are managed via declarative managers.\
A manager defines command templates for `check`, `install`, `remove`, `enable`, and `disable`. Groups reference a manager by name and list packages or services.

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
check   = "pkg={{.Name}}; pkg=${pkg%@*}; test -d ~/.npm-global/lib/node_modules/$pkg"
install = "npm install -g --prefix ~/.npm-global {{.Name}}"
remove  = "npm uninstall -g --prefix ~/.npm-global {{.Name}}"

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

### Adding a manager

Just add a section to `[managers]`:

```toml
[managers.flatpak]
check   = "flatpak list --app --columns=application | grep -qxF {{.Name}}"
install = "flatpak install -y --noninteractive {{.Name}}"
remove  = "flatpak uninstall -y --noninteractive {{.Name}}"

[flatpak]
packages = ["org.mozilla.firefox"]
```
