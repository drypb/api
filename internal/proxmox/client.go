// Package proxmox is an adapter to a proxmox API.
package proxmox

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"

	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// Client is a wrapper around [proxmoxAPI.Client].
type Client struct {
	client *proxmoxAPI.Client
}

// NewClient creates a new [Client].
func NewClient() *Client {
	url := proxmoxURL()
	id := tokenID()
	secret := tokenSecret()
	// Maybe change to use TLS when in production.
	insecureHTTPClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	c := &Client{
		client: proxmoxAPI.NewClient(
			url,
			proxmoxAPI.WithHTTPClient(insecureHTTPClient),
			proxmoxAPI.WithAPIToken(id, secret),
		),
	}
	if c.client == nil {
		log.Fatal("Failed to create proxmox client")
	}
	return c
}

// TokenID returns the proxmox token id from PROXMOX_TOKEN_ID.
func tokenID() string {
	id := os.Getenv("PROXMOX_TOKEN_ID")
	if id == "" {
		log.Fatal("PROXMOX_TOKEN_ID is empty")
	}
	return id
}

// TokenSecret returns the proxmox token secret from PROXMOX_TOKEN_SECRET.
func tokenSecret() string {
	s := os.Getenv("PROXMOX_TOKEN_SECRET")
	if s == "" {
		log.Fatal("PROXMOX_TOKEN_SECRET is empty")
	}
	return s
}

// ProxmoxURL returns the proxmox server URL from PROXMOX_URL.
func proxmoxURL() string {
	url := os.Getenv("PROXMOX_URL")
	if url == "" {
		log.Fatal("PROXMOX_URL is empty")
	}
	return url
}

// Node returns a [Node] given it's name.
func (c *Client) Node(ctx context.Context, name string) (*Node, error) {
	n, err := c.client.Node(ctx, name)
	if err != nil {
		return nil, err
	}
	return &Node{node: n}, err
}
