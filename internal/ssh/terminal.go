package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// OpenInteractiveSession opens an interactive terminal session with the remote host
// If useCustomRC is true, it will use the temporary RC file with helpers
func (c *Client) OpenInteractiveSession(shellType string, useCustomRC bool) error {
	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Get terminal file descriptor
	fd := int(os.Stdin.Fd())

	// Get current terminal state
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %v", err)
	}
	defer terminal.Restore(fd, oldState)

	// Get terminal size
	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		termWidth = 80
		termHeight = 24
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %v", err)
	}

	// Set up stdin, stdout, stderr
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if useCustomRC {
		var shellCmd string
		switch shellType {
		case "zsh":
			shellCmd = "ZDOTDIR=/tmp zsh"
		default:
			shellCmd = "bash --rcfile /tmp/sshrc/.sshrc_bashrc"
		}
		if err := session.Start(shellCmd); err != nil {
			return fmt.Errorf("failed to start shell: %v", err)
		}

	} else {
		if err := session.Shell(); err != nil {
			return fmt.Errorf("failed to start shell: %v", err)
		}
	}

	// Wait for session to finish
	if err := session.Wait(); err != nil {
		if _, ok := err.(*ssh.ExitError); ok {
			// Normal exit
			return nil
		}
		return fmt.Errorf("session error: %v", err)
	}

	return nil
}
