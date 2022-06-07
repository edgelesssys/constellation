package proto

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ActivationRespClient has methods to read messages from a stream of
// ActivateAsCoordinatorResponses. It wraps an API_ActivateAsCoordinatorClient.
type ActivationRespClient struct {
	client            pubproto.API_ActivateAsCoordinatorClient
	kubeconfig        string
	coordinatorVpnKey string
	clientVpnIp       string
	ownerID           string
	clusterID         string
}

// NewActivationRespClient creates a new ActivationRespClient with the handed
// API_ActivateAsCoordinatorClient.
func NewActivationRespClient(client pubproto.API_ActivateAsCoordinatorClient) *ActivationRespClient {
	return &ActivationRespClient{
		client: client,
	}
}

// NextLog reads responses from the response stream and returns the
// first received log.
func (a *ActivationRespClient) NextLog() (string, error) {
	for {
		resp, err := a.client.Recv()
		if err != nil {
			return "", err
		}
		switch x := resp.Content.(type) {
		case *pubproto.ActivateAsCoordinatorResponse_Log:
			return x.Log.Message, nil
		case *pubproto.ActivateAsCoordinatorResponse_AdminConfig:
			config := x.AdminConfig
			a.kubeconfig = string(config.Kubeconfig)

			coordinatorVpnKey, err := wgtypes.NewKey(config.CoordinatorVpnPubKey)
			if err != nil {
				return "", err
			}
			a.coordinatorVpnKey = coordinatorVpnKey.String()
			a.clientVpnIp = config.AdminVpnIp
			a.ownerID = base64.StdEncoding.EncodeToString(config.OwnerId)
			a.clusterID = base64.StdEncoding.EncodeToString(config.ClusterId)
		}
	}
}

// WriteLogStream reads responses from the response stream and
// writes log responses to the handed writer.
func (a *ActivationRespClient) WriteLogStream(w io.Writer) error {
	log, err := a.NextLog()
	for err == nil {
		fmt.Fprintln(w, log)
		log, err = a.NextLog()
	}
	if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// GetKubeconfig returns the kubeconfig that was received in the
// latest AdminConfig response or an error if the field is empty.
func (a *ActivationRespClient) GetKubeconfig() (string, error) {
	if a.kubeconfig == "" {
		return "", errors.New("kubeconfig is empty")
	}
	return a.kubeconfig, nil
}

// GetCoordinatorVpnKey returns the Coordinator's VPN key that was
// received in the latest AdminConfig response or an error if the field
// is empty.
func (a *ActivationRespClient) GetCoordinatorVpnKey() (string, error) {
	if a.coordinatorVpnKey == "" {
		return "", errors.New("coordinator public VPN key is empty")
	}
	return a.coordinatorVpnKey, nil
}

// GetClientVpnIp returns the client VPN IP that was received
// in the latest AdminConfig response or an error if the field is empty.
func (a *ActivationRespClient) GetClientVpnIp() (string, error) {
	if a.clientVpnIp == "" {
		return "", errors.New("client VPN IP is empty")
	}
	return a.clientVpnIp, nil
}

// GetOwnerID returns the owner identifier, derived from the client's master secret
// or an error if the field is empty.
func (a *ActivationRespClient) GetOwnerID() (string, error) {
	if a.ownerID == "" {
		return "", errors.New("secret identifier is empty")
	}
	return a.ownerID, nil
}

// GetClusterID returns the cluster's unique identifier
// or an error if the field is empty.
func (a *ActivationRespClient) GetClusterID() (string, error) {
	if a.clusterID == "" {
		return "", errors.New("cluster identifier is empty")
	}
	return a.clusterID, nil
}
