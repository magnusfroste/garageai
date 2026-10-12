package portal

import (
	"context"
	"errors"
	"time"
)

// TunnelGrant is what the portal gives a garage that uses the Bridge tunnel (docs/bridge-tunnel.md).
type TunnelGrant struct {
	TunnelIP     string `json:"tunnel_ip"`
	HubPublicKey string `json:"hub_public_key"`
	HubEndpoint  string `json:"hub_endpoint"`
	HubIP        string `json:"hub_ip"`
}

// TunnelAssign registers this garage's tunnel public key with the portal and gets its tunnel
// address and the hub's key and endpoint. The garage's private key never leaves the machine.
func (c *Client) TunnelAssign(ctx context.Context, publicKey string) (*TunnelGrant, error) {
	var g TunnelGrant
	code, err := c.post(ctx, c.sibling("tunnel-assign"), map[string]string{"tunnel_public_key": publicKey}, &g, 20*time.Second)
	if code == 401 || code == 403 {
		return nil, ErrRejected
	}
	if err != nil {
		return nil, err
	}
	if g.TunnelIP == "" || g.HubPublicKey == "" || g.HubEndpoint == "" {
		return nil, errors.New("tunnel-assign: incomplete answer (no tunnel_ip, hub_public_key or hub_endpoint)")
	}
	return &g, nil
}

// TunnelRelease tells the portal this garage is back on NetBird.
func (c *Client) TunnelRelease(ctx context.Context) error {
	_, err := c.post(ctx, c.sibling("tunnel-assign"), map[string]bool{"release": true}, nil, 20*time.Second)
	return err
}
