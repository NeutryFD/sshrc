# SSHRC - SSH Remote Command Tool

Like `sshuttle` but for your shell environment! Bring your local aliases, functions, and scripts to any remote host temporarily.

## Features

- 🔐 SSH key-based authentication
- 📦 Auto-copy your helper scripts to remote hosts
- 📊 Built-in session monitoring with duration tracking
- 🧹 Auto-cleanup on disconnect (everything removed)
- 💻 Works with bash and zsh

## Quick Start

```bash
# Build
go build -o sshrc

# Connect with default helpers
./sshrc --host example.com

# Use custom helpers directory
./sshrc --host example.com --helpers ~/my-tools

```

## Usage

```bash
./sshrc [flags]

Flags:
  -H, --host string       Remote host (required)
  -d, --helpers string    Helpers directory (default: ./helpers)
  -u, --user string       SSH user (default: root)
  -p, --port string       SSH port (default: 22)
  -k, --key string        SSH private key (default: ~/.ssh/id_rsa)
  -m, --monitor-only      Monitor session without copying helpers
  --local-rc string       Local RC file to copy from marker onward (e.g. ~/.zshrc)
```

## How It Works

1. Connects to remote host via SSH
2. Copies helper scripts to `/tmp/sshrc_helpers/`
3. Creates temporary shell RC that loads your helpers
4. Opens interactive session with monitoring
5. Cleans up everything when you disconnect

## Helper Scripts

Create scripts in `./helpers/` (or your custom directory):

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

## Project Structure

```
sshrc/
├── cmd/              # CLI commands
├── internal/
│   ├── logger/       # Logging utilities
│   └── ssh/          # SSH client, helpers, terminal
├── helpers/          # Default helper scripts
│   ├── aliases.sh
│   └── functions.sh
└── main.go
```
## License

See LICENSE file for details.

## Detailed Features

- Sync local shell helpers to a remote SSH session
- Monitor and auto-clean up temp files on disconnect
- All temp files managed under `/tmp/sshrc` (configurable via constant)
- `--local-rc` flag: injects any local RC file (bashrc, zshrc, etc) from a marker onward into the remote temp RC file
- Robust marker detection (case-insensitive, partial match)
- Works with bash or zsh (auto-detects remote shell)

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

- The `--local-rc` flag can point to any shell RC file. Only lines from the marker (e.g. `# HELPERS`) onward are injected.
- All temp files and helpers are placed under `/tmp/sshrc` for easy cleanup.

## Helpers
Place your helper scripts in the `helpers/` directory. They will be copied to the remote and sourced automatically.

## Requirements
- Go 1.19+
- SSH access to remote host
- Remote host should have bash or zsh installed

## Development
- All temp file paths use the `sshrcTmpDir` constant for maintainability.
- Marker detection is robust: any line containing `# HELPERS` (case-insensitive) is matched.

---
MIT License
