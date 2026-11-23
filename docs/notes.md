# dotm

## Security

### Temporary files

Scripts and diff operations use temporary files.\
By default, `dotm` creates these in `$XDG_RUNTIME_DIR/dotm/` (typically `/run/user/<uid>/`) or `$HOME/.local/state/dotm/tmp/`, both with mode `0700`.\
This prevents symlink race attacks that are possible when using the world-writable `/tmp`.

If neither secure directory is available, `dotm` falls back to the system
temp directory.

## State

dotm stores state in `~/.local/state/dotm/<hash>.toml`:

- **Prompt answers** — cached so you're not asked every apply
- **Script hashes** — for `on_change` trigger
- **Manifest** — list of deployed files, directories, symlinks for orphan detection
- **Pkg manifest** — list of deployed packages and services for orphan tracking

Each source repo gets its own state file (keyed by SHA-256 of absolute path). Re-run `dotm init` to re-answer prompts.
