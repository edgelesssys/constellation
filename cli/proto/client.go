package proto

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client wraps a AVPNClient and the connection to it.
// The client offers a method to activate the connected
// AVPNServer as Coordinator.
type Client struct {
	conn *grpc.ClientConn
	avpn pubproto.APIClient
}

// Connect connects the client to a given server, using the handed
// Validators for the attestation of the connection.
// The connection must be closed using Close(). If connect is
// called on a client that already has a connection, the old
// connection is closed.
func (c *Client) Connect(ip, port string, gcpPCRs, AzurePCRs map[uint32][]byte) error {
	addr := net.JoinHostPort(ip, port)

	validators := []atls.Validator{
		gcp.NewValidator(gcpPCRs),
		gcp.NewNonCVMValidator(map[uint32][]byte{}), // TODO: Remove once we no longer use non cvms
		azure.NewValidator(map[uint32][]byte{}),
	}

	tlsConfig, err := atls.CreateAttestationClientTLSConfig(validators)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return err
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.conn = conn
	c.avpn = pubproto.NewAPIClient(conn)
	return nil
}

// Close closes the grpc connection of the client.
// Close is idempotent and can be called on non connected clients
// without returning an error.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	c.conn = nil
	return nil
}

// Activate activates the Constellation coordinator via a grpc call.
// The handed IP addresses must be the private IP addresses of running AWS or GCP instances,
// and the userPublicKey is the VPN key of the users WireGuard interface.
func (c *Client) Activate(ctx context.Context, userPublicKey, masterSecret []byte, ips, autoscalingNodeGroups []string, cloudServiceAccountURI string) (ActivationResponseClient, error) {
	if c.avpn == nil {
		return nil, errors.New("client is not connected")
	}
	if len(userPublicKey) == 0 {
		return nil, errors.New("parameter userPublicKey is empty")
	}
	if len(ips) == 0 {
		return nil, errors.New("parameter ips is empty")
	}

	pubKey, err := wgtypes.ParseKey(string(userPublicKey))
	if err != nil {
		return nil, err
	}

	avpnRequest := &pubproto.ActivateAsCoordinatorRequest{
		AdminVpnPubKey:         pubKey[:],
		NodePublicIps:          ips,
		AutoscalingNodeGroups:  autoscalingNodeGroups,
		MasterSecret:           masterSecret,
		KmsUri:                 kms.ClusterKMSURI,
		StorageUri:             kms.NoStoreURI,
		KeyEncryptionKeyId:     "",
		UseExistingKek:         false,
		CloudServiceAccountUri: cloudServiceAccountURI,
	}

	client, err := c.avpn.ActivateAsCoordinator(ctx, avpnRequest)
	if err != nil {
		return nil, err
	}
	return NewActivationRespClient(client), nil
}

// ActivationResponseClient has methods to read messages from a stream of
// ActivateAsCoordinatorResponses.
type ActivationResponseClient interface {
	// NextLog reads responses from the response stream and returns the
	// first received log.
	// If AdminConfig responses are received before the first log response
	// occurs, the state of the client is updated with those configs. An
	// io.EOF error is returned at the end of the stream.
	NextLog() (string, error)

	// WriteLogStream reads responses from the response stream and
	// writes log responses to the handed writer.
	// Occurring AdminConfig responses update the state of the client.
	WriteLogStream(io.Writer) error

	// GetKubeconfig returns the kubeconfig that was received in the
	// latest AdminConfig response or an error if the field is empty.
	GetKubeconfig() (string, error)

	// GetCoordinatorVpnKey returns the Coordinator's VPN key that was
	// received in the latest AdminConfig response or an error if the field
	// is empty.
	GetCoordinatorVpnKey() (string, error)

	// GetClientVpnIp returns the client VPN IP that was received
	// in the latest AdminConfig response or an error if the field is empty.
	GetClientVpnIp() (string, error)

	// GetOwnerID returns the owner identifier, derived from the client's master secret
	// or an error if the field is empty.
	GetOwnerID() (string, error)

	// GetClusterID returns the cluster's unique identifier
	// or an error if the field is empty.
	GetClusterID() (string, error)
}
