package ssh

import (
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client wraps the SSH client connection
type Client struct {
	*ssh.Client
}

// Config holds the SSH connection configuration
type Config struct {
	Host    string
	Port    string
	User    string
	KeyPath string
}

// NewClient creates a new SSH client connection
func NewClient(cfg Config) (*Client, error) {
	// Read private key
	key, err := ioutil.ReadFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	// Create SSH config
	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect to remote host
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return &Client{Client: client}, nil
}
