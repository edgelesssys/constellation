package cmd

import wgquick "github.com/nmiculinic/wg-quick-go"

type vpnHandler interface {
	Create(coordinatorPubKey string, coordinatorPubIP string, clientPrivKey string, clientVPNIP string, mtu int) (*wgquick.Config, error)
	Apply(*wgquick.Config) error
	Marshal(*wgquick.Config) ([]byte, error)
}
