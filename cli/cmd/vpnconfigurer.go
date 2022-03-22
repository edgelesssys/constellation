package cmd

type vpnConfigurer interface {
	Configure(clientVpnIp string, coordinatorPubKey string, coordinatorPubIP string, clientPrivKey string) error
}
