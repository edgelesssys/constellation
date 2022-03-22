package cmd

type stubVPNConfigurer struct {
	configured   bool
	configureErr error
}

func (c *stubVPNConfigurer) Configure(clientVpnIp, coordinatorPubKey, coordinatorPubIP, clientPrivKey string) error {
	c.configured = true
	return c.configureErr
}

type dummyVPNConfigurer struct{}

func (c *dummyVPNConfigurer) Configure(clientVpnIp, coordinatorPubKey, coordinatorPubIP, clientPrivKey string) error {
	panic("dummy doesn't implement this function")
}
