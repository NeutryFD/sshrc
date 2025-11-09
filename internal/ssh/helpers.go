package ssh

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const sshrcTmpDir = "/tmp/sshrc"

const monitoringScript = `#!/bin/bash

# SSH Monitoring Script
# This script monitors the specific SSH session by TTY
# Usage: ./ssh-mon.bash [TTY]

LOG_FILE="/var/log/ssh-monitor.log"
current_USER=$USER

# Get TTY from parameter or SSH_TTY environment variable
if [ -n "$1" ]; then
    MONITOR_TTY="$1"
elif [ -n "$SSH_TTY" ]; then
    MONITOR_TTY="$SSH_TTY"
else
    TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
    echo "[$TIMESTAMP] ERROR: No TTY specified and SSH_TTY not set" >> "$LOG_FILE"
    exit 1
fi

# Log start of monitoring and record start time
START_TIME=$(date +%s)
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
echo "[$TIMESTAMP] Started monitoring SSH session for user $current_USER on TTY $MONITOR_TTY" >> "$LOG_FILE"

# Get the basename of the TTY (e.g., /dev/pts/5 -> pts/5)
TTY_BASENAME=$(echo "$MONITOR_TTY" | sed 's#^/dev/##')

# Function to calculate duration
calculate_duration() {
    local start=$1
    local end=$2
    local duration=$((end - start))
    
    local hours=$((duration / 3600))
    local minutes=$(((duration % 3600) / 60))
    local seconds=$((duration % 60))
    
    if [ $hours -gt 0 ]; then
        echo "${hours}h ${minutes}m ${seconds}s"
    elif [ $minutes -gt 0 ]; then
        echo "${minutes}m ${seconds}s"
    else
        echo "${seconds}s"
    fi
}

# Monitor the SSH session
while true; do
    # Check if the TTY device still exists and has processes attached
    if [ ! -e "$MONITOR_TTY" ]; then
        END_TIME=$(date +%s)
        DURATION=$(calculate_duration $START_TIME $END_TIME)
        TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
        echo "[$TIMESTAMP] SSH session ended - TTY device $MONITOR_TTY no longer exists (Duration: $DURATION)" >> "$LOG_FILE"
        # Restore original RC files and cleanup
        mv ~/.bashrc.sshrc_backup ~/.bashrc 2>/dev/null || true
        mv ~/.zshrc.sshrc_backup ~/.zshrc 2>/dev/null || true
        rm -rf /tmp/sshrc
        exit 0
    fi
    
    # Check if there are any bash/shell processes on this TTY
    SHELL_ON_TTY=$(ps aux | grep "$TTY_BASENAME" | grep -E "(bash|sh|zsh)" | grep -v grep | grep "$current_USER")
    
    if [ -z "$SHELL_ON_TTY" ]; then
        END_TIME=$(date +%s)
        DURATION=$(calculate_duration $START_TIME $END_TIME)
        TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
        echo "[$TIMESTAMP] SSH session ended - No shell process found on TTY $MONITOR_TTY (Duration: $DURATION)" >> "$LOG_FILE"
        # Restore original RC files and cleanup
        mv ~/.bashrc.sshrc_backup ~/.bashrc 2>/dev/null || true
        mv ~/.zshrc.sshrc_backup ~/.zshrc 2>/dev/null || true
        rm -rf /tmp/sshrc
        exit 0
    fi
    
    sleep 2
done

# This should never be reached, but just in case
END_TIME=$(date +%s)
DURATION=$(calculate_duration $START_TIME $END_TIME)
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
echo "[$TIMESTAMP] Monitoring script ended for user $current_USER on TTY $MONITOR_TTY (Duration: $DURATION)" >> "$LOG_FILE"
# Restore original RC files and cleanup
mv ~/.bashrc.sshrc_backup ~/.bashrc 2>/dev/null || true
mv ~/.zshrc.sshrc_backup ~/.zshrc 2>/dev/null || true
rm -rf /tmp/sshrc
exit 0
`

// CopyMonitoringScript copies the embedded monitoring script to the remote host
func (c *Client) CopyMonitoringScript() error {
	// Use embedded script content
	scriptContent := []byte(monitoringScript)

	// Create session for file transfer
	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Ensure /tmp/sshrc exists, then create remote file and write content
	remoteDir := "/tmp/sshrc"
	remotePath := "/tmp/sshrc/ssh-mon.bash"
	cmd := fmt.Sprintf("mkdir -p %s && cat > %s && chmod +x %s", remoteDir, remotePath, remotePath)

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	go func() {
		defer stdin.Close()
		stdin.Write(scriptContent)
	}()

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to write monitoring script: %v", err)
	}

	return nil
}

// CopyHelpers copies all helper scripts from a local directory to the remote host
func (c *Client) CopyHelpers(helpersDir string) ([]string, error) {
	if helpersDir == "" {
		return nil, nil
	}

	// Check if helpers directory exists
	if _, err := os.Stat(helpersDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("helpers directory does not exist: %s", helpersDir)
	}

	// Read all files in the helpers directory
	files, err := ioutil.ReadDir(helpersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read helpers directory: %v", err)
	}

	var copiedFiles []string

	// Copy each file to remote host
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		localPath := filepath.Join(helpersDir, file.Name())
		remotePath := fmt.Sprintf("/tmp/sshrc/helpers/%s", file.Name())

		// Read file content
		content, err := ioutil.ReadFile(localPath)
		if err != nil {
			return copiedFiles, fmt.Errorf("failed to read file %s: %v", file.Name(), err)
		}

		// Create remote directory and copy file
		session, err := c.NewSession()
		if err != nil {
			return copiedFiles, fmt.Errorf("failed to create session: %v", err)
		}

		cmd := fmt.Sprintf("mkdir -p /tmp/sshrc/helpers && cat > %s && chmod +x %s", remotePath, remotePath)
		stdin, err := session.StdinPipe()
		if err != nil {
			session.Close()
			return copiedFiles, fmt.Errorf("failed to create stdin pipe: %v", err)
		}

		go func() {
			defer stdin.Close()
			stdin.Write(content)
		}()

		if err := session.Run(cmd); err != nil {
			session.Close()
			return copiedFiles, fmt.Errorf("failed to copy file %s: %v", file.Name(), err)
		}
		session.Close()

		copiedFiles = append(copiedFiles, remotePath)
	}

	return copiedFiles, nil
}

// SetupShellRC creates a temporary shell RC file that sources helpers and user's original RC
func (c *Client) SetupShellRC(shellType string) error {
	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var originalRC string
	var rcFile string

	switch shellType {
	case "bash":
		originalRC = "~/.bashrc"
		rcFile = "/tmp/sshrc/.sshrc_bashrc"
	case "zsh":
		originalRC = "~/.zshrc"
		rcFile = "/tmp/sshrc/.sshrc_zshrc"
	default:
		originalRC = "~/.bashrc"
		rcFile = "/tmp/sshrc/.sshrc_bashrc"
	}

	rcContent := fmt.Sprintf(`#!/bin/bash
# SSHRC Temporary Shell Configuration
# This file is auto-generated and will be cleaned up on exit

# Source original RC if it exists
if [ -f %s ]; then
    source %s
fi

# Add helper scripts to PATH
export PATH="/tmp/sshrc/helpers:$PATH"

# Source all helper scripts
for helper in /tmp/sshrc/helpers/*; do
    if [ -f "$helper" ] && [ -x "$helper" ]; then
        source "$helper" 2>/dev/null || true
    fi
done

# Launch monitoring script in background only once per connection
if [ -n "$SSH_TTY" ]; then
    LOCKFILE="/tmp/sshrc/.monitor_$(basename $SSH_TTY)"
    if [ ! -f "$LOCKFILE" ]; then
        touch "$LOCKFILE"
        nohup /tmp/sshrc/ssh-mon.bash "$SSH_TTY" </dev/null >/dev/null 2>&1 &
    fi
fi

`, originalRC, originalRC)

	cmd := fmt.Sprintf("cat > %s << 'SSHRC_EOF'\n%s\nSSHRC_EOF\nchmod +x %s", rcFile, rcContent, rcFile)

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to create RC file: %v", err)
	}

	return nil
}

// DetectRemoteShell detects which shell is being used on the remote host
func (c *Client) DetectRemoteShell() (string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput("echo $SHELL")
	if err != nil {
		return "bash", nil // Default to bash
	}

	shellPath := strings.TrimSpace(string(output))
	if strings.Contains(shellPath, "zsh") {
		return "zsh", nil
	} else if strings.Contains(shellPath, "bash") {
		return "bash", nil
	}

	return "bash", nil // Default to bash
}
