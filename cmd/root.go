package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"sshrc/internal/logger"
	sshclient "sshrc/internal/ssh"

	"github.com/spf13/cobra"
)

var (
	host        string
	port        string
	user        string
	keyPath     string
	helpersDir  string
	monitorOnly bool
	localRC     string // now holds the path
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sshrc",
		Short: "SSH Remote Command - Connect with your helpers and shell configuration",
		Long:  `A CLI tool to connect to a remote host, copy helper scripts and shell configurations temporarily, with automatic cleanup on disconnect.`,
		Run:   runSSHRC,
	}

	// Define flags
	rootCmd.Flags().StringVarP(&host, "host", "H", "", "Remote host to connect to (required)")
	rootCmd.Flags().StringVarP(&port, "port", "p", "22", "SSH port (default: 22)")
	rootCmd.Flags().StringVarP(&user, "user", "u", "root", "SSH user (default: root)")
	rootCmd.Flags().StringVarP(&keyPath, "key", "k", "", "Path to SSH private key (default: ~/.ssh/id_rsa)")
	rootCmd.Flags().StringVarP(&helpersDir, "helpers", "d", "", "Path to helpers directory (default: ./helpers)")
	rootCmd.Flags().BoolVarP(&monitorOnly, "monitor-only", "m", false, "Only monitor session without copying helpers")
	rootCmd.PersistentFlags().StringVar(&localRC, "local-rc", "", "Path to local RC file to copy from # HELPERS (optional)")

	rootCmd.MarkFlagRequired("host")

	return rootCmd
}

func runSSHRC(cmd *cobra.Command, args []string) {
	logger.LogStep("Starting SSHRC execution")

	// Set default key path if not provided
	if keyPath == "" {
		keyPath = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}

	// Set default helpers directory
	if helpersDir == "" && !monitorOnly {
		helpersDir = "./helpers"
	}

	logger.LogStep(fmt.Sprintf("Connecting to %s@%s:%s", user, host, port))

	// Create SSH client configuration
	cfg := sshclient.Config{
		Host:    host,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
	}

	// Create SSH client
	client, err := sshclient.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create SSH client: %v", err)
	}
	defer client.Close()

	logger.LogStep("SSH connection established successfully")

	// Copy monitoring script (always enabled)
	logger.LogStep("Deploying embedded monitoring script")
	if err := client.CopyMonitoringScript(); err != nil {
		log.Fatalf("Failed to copy monitoring script: %v", err)
	}

	useCustomRC := false

	// Copy helper scripts if not in monitor-only mode
	if !monitorOnly && helpersDir != "" {
		// Get absolute path for better logging
		absHelpersDir, err := filepath.Abs(helpersDir)
		if err != nil {
			absHelpersDir = helpersDir
		}

		logger.LogStep(fmt.Sprintf("Loading helpers from: %s", absHelpersDir))
		copiedFiles, err := client.CopyHelpers(helpersDir)
		if err != nil {
			log.Fatalf("Failed to copy helpers: %v", err)
		}
		if len(copiedFiles) > 0 {
			logger.LogStep(fmt.Sprintf("Deployed %d helper file(s) to remote host", len(copiedFiles)))
			useCustomRC = true

			// Detect remote shell
			logger.LogStep("Detecting remote shell")
			shellType, err := client.DetectRemoteShell()
			if err != nil {
				log.Printf("Warning: Could not detect shell, defaulting to bash: %v", err)
				shellType = "bash"
			}
			logger.LogStep(fmt.Sprintf("Detected shell: %s", shellType))

			// Setup custom shell RC
			logger.LogStep("Setting up custom shell environment")
			if localRC != "" {
				if err := client.SetupShellRCWithLocal(shellType, localRC); err != nil {
					log.Fatalf("Failed to setup shell RC with local: %v", err)
				}
			} else {
				if err := client.SetupShellRC(shellType); err != nil {
					log.Fatalf("Failed to setup shell RC: %v", err)
				}
			}

			// Open interactive terminal with custom RC
			logger.LogStep("Opening interactive terminal session with helpers")
			if err := client.OpenInteractiveSession(shellType, useCustomRC); err != nil {
				log.Fatalf("Failed to open interactive session: %v", err)
			}
		} else {
			logger.LogStep("No helper files found, using monitor-only mode")
		}
	} else {
		// Monitor-only mode
		logger.LogStep("Opening interactive terminal session (monitor-only mode)")
		if err := client.OpenInteractiveSession("bash", false); err != nil {
			log.Fatalf("Failed to open interactive session: %v", err)
		}
	}

	logger.LogStep("Session closed (cleanup handled by monitoring script)")
}
