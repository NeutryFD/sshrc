# SSHRC - SSH Remote Command Tool

Like `sshuttle` but for your shell environment! Bring your local aliases, functions, and scripts to any remote host temporarily.

## Features

- 🔐 SSH key-based authentication
- 📦 Auto-copy your helper scripts to remote hosts
- 📊 Built-in session monitoring with duration tracking
- 🧹 Auto-cleanup on disconnect (everything removed)
- 💻 Works with bash and zsh

## Project Structure
```
sshrc/
├── cmd/              # CLI commands
├── internal/
│   ├── logger/       # Logging utilities
│   └── ssh/          # SSH client, helpers, terminal
├── helpers/          # Default helper scripts (auto-copied to remote)
│   ├── aliases.sh
│   └── functions.sh
├── README.md         # Project documentation
├── main.go           # Entry point
```
## Usage Examples

```sh
# Basic usage (with helpers)
go run ./main.go -H <host>

# Use a custom RC file (from marker onward)
go run ./main.go -H <host> --local-rc /path/to/.zshrc

# Monitor-only mode (no helpers)
go run ./main.go -H <host> --monitor-only
```

### Flags
- `-H, --host <host>`: Remote host to connect to (required)
- `-p, --port <port>`: SSH port (default: 22)
- `-u, --user <user>`: SSH user (default: root)
- `-k, --key <path>`: Path to SSH private key (default: ~/.ssh/id_rsa)
- `-d, --helpers <dir>`: Path to helpers directory (default: ./helpers)
- `-m, --monitor-only`: Only monitor session without copying helpers
- `--local-rc <path>`: Path to local RC file to copy from marker onward (e.g. ~/.zshrc)
  
## Detailed Features

- Sync local shell helpers to a remote SSH session
- Monitor and auto-clean up temp files on disconnect
- All temp files managed under `/tmp/sshrc` (configurable via constant)
- `--local-rc` flag: injects any local RC file (bashrc, zshrc, etc) from a marker onward into the remote temp RC file
- Robust marker detection (case-insensitive, partial match)
- Works with bash or zsh (auto-detects remote shell)
- The `--local-rc` flag can point to any shell RC file. Only lines from the marker (e.g. `# HELPERS`) onward are injected.
- All temp files and helpers are placed under `/tmp/sshrc` for easy cleanup.

## Helpers
Place your helper scripts in the `helpers/` directory. They will be copied to the remote and sourced automatically.

```bash
# helpers/aliases.sh
alias ll='ls -alF'
alias gs='git status'

# helpers/functions.sh
mkcd() {
    mkdir -p "$@" && cd "$@"
}
```

These become available automatically on the remote host!

## Requirements
- Go 1.19+
- SSH access to remote host
- Remote host should have bash or zsh installed

## Development
- All temp file paths use the `sshrcTmpDir` constant for maintainability.
- Marker detection is robust: any line containing `# HELPERS` (case-insensitive) is matched.

---
MIT License
