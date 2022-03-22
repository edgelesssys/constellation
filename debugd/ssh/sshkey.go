package ssh

// SSHKey describes a public ssh key.
type SSHKey struct {
	Username string `json:"user"`
	KeyValue string `json:"pubkey"`
}
