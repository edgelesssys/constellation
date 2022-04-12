package keyservice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/coordinator/atls"
	azurecloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/azure"
	gcpcloud "github.com/edgelesssys/constellation/coordinator/cloudprovider/gcp"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// keyAPI is the interface called by the Coordinator or an admin during restart of a node.
type keyAPI struct {
	metadata    core.ProviderMetadata
	mux         sync.Mutex
	key         []byte
	keyReceived chan bool
	timeout     time.Duration
}

func (a *keyAPI) waitForDecryptionKey() {
	// go server.Start()
	// block until a key is pushed
	if <-a.keyReceived {
		return
	}
}

func (a *keyAPI) requestKeyFromCoordinator(uuid string, opts ...grpc.DialOption) error {
	// we do not perform attestation, since the restarting node does not need to care about notifying the correct Coordinator
	// if an incorrect key is pushed by a malicious actor, decrypting the disk will fail, and the node will not start
	tlsClientConfig, err := atls.CreateUnverifiedClientTLSConfig()
	if err != nil {
		return err
	}

	for {
		select {
		// return if a key was received by any means
		// a key can be send by
		// - a Coordinator, after the request rpc was received
		// - by a Constellation admin, at any time this loop is running on a node during boot
		case <-a.keyReceived:
			return nil
		default:
			// list available Coordinators
			endpoints, _ := core.CoordinatorEndpoints(context.Background(), a.metadata)
			// notify the all available Coordinators to send a key to the node
			// any errors encountered here will be ignored, and the calls retried after a timeout
			for _, endpoint := range endpoints {
				ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
				conn, err := grpc.DialContext(ctx, endpoint, append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsClientConfig)))...)
				if err == nil {
					client := pubproto.NewAPIClient(conn)
					_, _ = client.RequestStateDiskKey(ctx, &pubproto.RequestStateDiskKeyRequest{DiskUuid: uuid})
					conn.Close()
				}

				cancel()
			}
			time.Sleep(a.timeout)
		}
	}
}

// WaitForDecryptionKey notifies the Coordinator to send a decryption key and waits until a key is received.
func WaitForDecryptionKey(csp, uuid string) ([]byte, error) {
	if uuid == "" {
		return nil, errors.New("received no disk UUID")
	}

	keyWaiter := &keyAPI{
		keyReceived: make(chan bool, 1),
		timeout:     20 * time.Second, // try to request a key every 20 seconds
	}
	go keyWaiter.waitForDecryptionKey()

	switch strings.ToLower(csp) {
	case "azure":
		metadata, err := azurecloud.NewMetadata(context.Background())
		if err != nil {
			return nil, err
		}
		keyWaiter.metadata = metadata
	case "gcp":
		gcpClient, err := gcpcloud.NewClient(context.Background())
		if err != nil {
			return nil, err
		}
		keyWaiter.metadata = gcpcloud.New(gcpClient)
	default:
		fmt.Fprintf(os.Stderr, "warning: csp %q is not supported, unable to automatically request decryption keys\n", csp)
		keyWaiter.metadata = stubMetadata{}
	}

	if err := keyWaiter.requestKeyFromCoordinator(uuid); err != nil {
		return nil, err
	}

	return keyWaiter.key, nil
}

type stubMetadata struct {
	listResponse []core.Instance
}

func (s stubMetadata) List(ctx context.Context) ([]core.Instance, error) {
	return s.listResponse, nil
}

func (s stubMetadata) Self(ctx context.Context) (core.Instance, error) {
	return core.Instance{}, nil
}

func (s stubMetadata) GetInstance(ctx context.Context, providerID string) (core.Instance, error) {
	return core.Instance{}, nil
}

func (s stubMetadata) SignalRole(ctx context.Context, role role.Role) error {
	return nil
}

func (s stubMetadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	return nil
}

func (s stubMetadata) Supported() bool {
	return true
}
