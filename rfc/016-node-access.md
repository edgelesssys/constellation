---
status: approved, not implemented
---

# RFC 016: Node Access

## Background

A production Constellation cluster is currently configured not to allow any kind of remote administrative access.
This choice is deliberate: any mechanism for remote accesss can potentially be exploited, or may leak sensitive data.

However, some operations on a Kubernetes cluster require some form of access to the nodes.
A good class of examples are etcd cluster maintenance tasks, like backup and recovery, or emergency operations like removing a permanently failed member.
Some kubeadm operations, like certificate rotation, also require some form of cluster access.

While some tasks can be accomplished by DaemonSets, CronJobs and the like, relying on Kubernetes objects is insufficient.
Executing commands in a Kubernetes pod may fail because Kubernetes is not healthy, etcd is bricked or the network is down.
Administrative access to the nodes through a side channel would greatly help remediate, or at least debug, those situations.

## Requirements

Constellation admins can log into Constellation nodes for maintenance, subject to the following restrictions:

* Access must be encrypted end-to-end to protect from CSP snooping.
* Access must be possible even if the Kubernetes API server is down.

Nice-to-have:

* The method of access should not require long-term storage of a second secret.
* The method of access should be time-limited.

## Proposed Design

Core to the proposal is [certificate-based authentication for OpenSSH](https://en.wikibooks.org/wiki/OpenSSH/Cookbook/Certificate-based_Authentication).
We can derive a valid SSH key from the Constellation master secret.
An OpenSSH server on the node accepts certificates issued by this CA key.
Admins can derive the CA key from the master secret on demand, and issue certificates for arbitrary public keys.
An example program is in the [Appendix](#appendix).

### Key Details

We use an HKDF to derive an ed25519 private key from the master secret.
This private key acts as an SSH certificate authority, whose signed certs allow access to cluster nodes.
Since the master secret is available to both the cluster owner and the nodes, no communication with the cluster is needed to mint valid certificates.
The choice of curve allows to directly use the derived secret bytes as key.
This makes the implementation deterministic, and thus the CA key recoverable.

### Server-side Details

An OpenSSH server is added to the node image software stack.
It's configured with a `TrustedUserCAKeys` file and a `RevokedKeys` file, both being empty on startup.
All other means of authentication are disabled.

After initialization, the bootstrapper fills the `TrustedUserCAKeys` file with the derived CA's public key.
Joining nodes send their public host key as part of the `IssueJoinTokenRequest` and receive the CA certificate and an indefinitely valid certificate as response.

The `RevokedKeys` KRL is an option for the cluster administrator to revoke keys, but it's not managed by Constellation.

### Client-side Details

A new `ssh` subcommand is added to the CLI.
The exact name is TBD, but it should fit in with other key-related activity, like generating volume keys.
It takes the master secret file and an SSH pub key file as arguments, and writes a certificate to stdout.
Optional arguments may include principals or vailidity period.
The implementation could roughly follow the PoC in the [Appendix](#appendix).

As an extension, the subcommand could allow generating a key pair and a matching certificate in a temp dir, and `exec` the ssh program directly.
This would encourage use of very short-lived certificates.

## Security Considerations

Exposing an additional service to the outside world increases the attack surface.
We propose the following mitigations:

1. The SSH port is only exposed to the VPC.
   This restricts the attackers to malicious co-tenants and the CSP.
   In an emergency, admins need to add a load balancer to be able to reach the nodes.
2. A hardened OpenSSH config only allows the options strictly necessary for the scheme proposed here.
   Authorized keys and passwords must be disabled.
   Cipher suites should be restricted. etc.

## Alternatives Considered

### Enable Serial Console

Serial consoles for cloud VMs are tunnelled through the CSP in the clear.
To make this solution secure, an encrypted channel would need to be established on top of the serial connection.
The author is not aware of any software providing such a channel.

### SSH with Authorized Keys

We could ask users to add a public key to their `constellation-conf.yaml` and add that to `/root/.ssh/authorized_keys` after joining.
This would require the cluster owner to permanently manage a second secret, and there would be no built-in way to revoke access.

### Debug Pod

Some node administration tasks can be performed with a [debug pod].
If privileged access is required, it's usually necessary to schedule a custom pod.
This only works if the Kubernetes API server is still processing requests, pods can be scheduled on the target node and the network allows connecting to it.

[debug pod]: https://kubernetes.io/docs/tasks/debug/debug-cluster/kubectl-node-debug/

### Host an Admin API Server

There are alternatives to SSH that allow fine-grained authorization of node operations.
An example would be [SansShell], which verifies node access requests with a policy.
Setting up such a tool requires a detailed understanding of the use cases, of which some might be hard to foresee.
This may be better suited as an extension of the low-level emergency access mechanisms.

[SansShell]: https://github.com/Snowflake-Labs/sansshell

## Appendix

A proof-of-concept implementation of the certificate generation.
Constellation nodes would stop after deriving the CA public key.

```golang
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/ssh"
)

type secret struct {
	Key  []byte `json:"key,omitempty"`
	Salt []byte `json:"salt,omitempty"`
}

var permissions = ssh.Permissions{
	Extensions: map[string]string{
		"permit-port-forwarding": "yes",
		"permit-pty":             "yes",
	},
}

func main() {
	masterSecret := flag.String("secret", "", "")
	flag.Parse()

	secretJSON, err := os.ReadFile(*masterSecret)
	must(err)
	var secret secret
	must(json.Unmarshal(secretJSON, &secret))

	hkdf := hkdf.New(sha256.New, secret.Key, secret.Salt, []byte("ssh-ca"))

	_, priv, err := ed25519.GenerateKey(hkdf)
	must(err)

	ca, err := ssh.NewSignerFromSigner(priv)
	must(err)

	log.Printf("CA KEY: %s", string(ssh.MarshalAuthorizedKey(ca.PublicKey())))

	buf, err := os.ReadFile(flag.Arg(0))
	must(err)
	pub, _, _, _, err := ssh.ParseAuthorizedKey(buf)
	must(err)
	certificate := ssh.Certificate{
		Key:             pub,
		CertType:        ssh.UserCert,
		ValidAfter:      uint64(time.Now().Unix()),
		ValidBefore:     uint64(time.Now().Add(24 * time.Hour).Unix()),
		ValidPrincipals: []string{"root"},
		Permissions:     permissions,
	}
	must(certificate.SignCert(rand.Reader, ca))

	fmt.Printf("%s\n", string(ssh.MarshalAuthorizedKey(&certificate)))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
```
