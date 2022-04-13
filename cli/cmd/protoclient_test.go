package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/cli/proto"
)

type stubProtoClient struct {
	conn        bool
	respClient  proto.ActivationResponseClient
	connectErr  error
	closeErr    error
	activateErr error

	activateUserPublicKey         []byte
	activateMasterSecret          []byte
	activateEndpoints             []string
	activateAutoscalingNodeGroups []string
	cloudServiceAccountURI        string
}

func (c *stubProtoClient) Connect(_, _ string, _, _ map[uint32][]byte) error {
	c.conn = true
	return c.connectErr
}

func (c *stubProtoClient) Close() error {
	c.conn = false
	return c.closeErr
}

func (c *stubProtoClient) Activate(ctx context.Context, userPublicKey, masterSecret []byte, endpoints []string, autoscalingNodeGroups []string, cloudServiceAccountURI string) (proto.ActivationResponseClient, error) {
	c.activateUserPublicKey = userPublicKey
	c.activateMasterSecret = masterSecret
	c.activateEndpoints = endpoints
	c.activateAutoscalingNodeGroups = autoscalingNodeGroups
	c.cloudServiceAccountURI = cloudServiceAccountURI

	return c.respClient, c.activateErr
}

type stubActivationRespClient struct {
	nextLogErr              *error
	getKubeconfigErr        error
	getCoordinatorVpnKeyErr error
	getClientVpnIpErr       error
	getOwnerIDErr           error
	getClusterIDErr         error
	writeLogStreamErr       error
}

func (s *stubActivationRespClient) NextLog() (string, error) {
	if s.nextLogErr == nil {
		return "", io.EOF
	}
	return "", *s.nextLogErr
}

func (s *stubActivationRespClient) WriteLogStream(io.Writer) error {
	return s.writeLogStreamErr
}

func (s *stubActivationRespClient) GetKubeconfig() (string, error) {
	return "", s.getKubeconfigErr
}

func (s *stubActivationRespClient) GetCoordinatorVpnKey() (string, error) {
	return "", s.getCoordinatorVpnKeyErr
}

func (s *stubActivationRespClient) GetClientVpnIp() (string, error) {
	return "", s.getClientVpnIpErr
}

func (s *stubActivationRespClient) GetOwnerID() (string, error) {
	return "", s.getOwnerIDErr
}

func (s *stubActivationRespClient) GetClusterID() (string, error) {
	return "", s.getClusterIDErr
}

type fakeProtoClient struct {
	conn       bool
	respClient proto.ActivationResponseClient
}

func (c *fakeProtoClient) Connect(_, _ string, _, _ map[uint32][]byte) error {
	c.conn = true
	return nil
}

func (c *fakeProtoClient) Close() error {
	c.conn = false
	return nil
}

func (c *fakeProtoClient) Activate(ctx context.Context, userPublicKey, masterSecret []byte, endpoints []string, autoscalingNodeGroups []string, cloudServiceAccountURI string) (proto.ActivationResponseClient, error) {
	if !c.conn {
		return nil, errors.New("client is not connected")
	}
	return c.respClient, nil
}

type fakeActivationRespClient struct {
	responses         []fakeActivationRespMessage
	kubeconfig        string
	coordinatorVpnKey string
	clientVpnIp       string
	ownerID           string
	clusterID         string
}

func (c *fakeActivationRespClient) NextLog() (string, error) {
	for len(c.responses) > 0 {
		resp := c.responses[0]
		c.responses = c.responses[1:]
		if len(resp.log) > 0 {
			return resp.log, nil
		}
		c.kubeconfig = resp.kubeconfig
		c.coordinatorVpnKey = resp.coordinatorVpnKey
		c.clientVpnIp = resp.clientVpnIp
		c.ownerID = resp.ownerID
		c.clusterID = resp.clusterID
	}
	return "", io.EOF
}

func (c *fakeActivationRespClient) WriteLogStream(w io.Writer) error {
	log, err := c.NextLog()
	for err == nil {
		fmt.Fprint(w, log)
		log, err = c.NextLog()
	}
	if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func (c *fakeActivationRespClient) GetKubeconfig() (string, error) {
	if c.kubeconfig == "" {
		return "", errors.New("kubeconfig is empty")
	}
	return c.kubeconfig, nil
}

func (c *fakeActivationRespClient) GetCoordinatorVpnKey() (string, error) {
	if c.coordinatorVpnKey == "" {
		return "", errors.New("coordinator public VPN key is empty")
	}
	return c.coordinatorVpnKey, nil
}

func (c *fakeActivationRespClient) GetClientVpnIp() (string, error) {
	if c.clientVpnIp == "" {
		return "", errors.New("client VPN IP is empty")
	}
	return c.clientVpnIp, nil
}

func (c *fakeActivationRespClient) GetOwnerID() (string, error) {
	if c.ownerID == "" {
		return "", errors.New("init secret is empty")
	}
	return c.ownerID, nil
}

func (c *fakeActivationRespClient) GetClusterID() (string, error) {
	if c.clusterID == "" {
		return "", errors.New("cluster identifier is empty")
	}
	return c.clusterID, nil
}

type fakeActivationRespMessage struct {
	log               string
	kubeconfig        string
	coordinatorVpnKey string
	clientVpnIp       string
	ownerID           string
	clusterID         string
}
