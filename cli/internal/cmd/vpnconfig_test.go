package cmd

import wgquick "github.com/nmiculinic/wg-quick-go"

type stubVPNHandler struct {
	configured bool
	marshalRes string

	createErr  error
	applyErr   error
	marshalErr error
}

func (c *stubVPNHandler) Create(coordinatorPubKey string, coordinatorPubIP string, clientPrivKey string, clientVPNIP string, mtu int) (*wgquick.Config, error) {
	return &wgquick.Config{}, c.createErr
}

func (c *stubVPNHandler) Apply(*wgquick.Config) error {
	c.configured = true
	return c.applyErr
}

func (c *stubVPNHandler) Marshal(*wgquick.Config) ([]byte, error) {
	return []byte(c.marshalRes), c.marshalErr
}
