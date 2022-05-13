package ssh

// SSHKey describes a public ssh key.
type SSHKey struct {
	Username string `yaml:"user"`
	KeyValue string `yaml:"pubkey"`
}
