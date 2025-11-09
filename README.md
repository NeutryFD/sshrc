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

# Monitor only (no helpers)
./sshrc --host example.com --monitor-only
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
  -m, --monitor-only      Skip helpers, only monitor
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
