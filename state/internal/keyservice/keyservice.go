package keyservice

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/edgelesssys/constellation/joinservice/joinproto"
	"github.com/edgelesssys/constellation/state/keyproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"k8s.io/utils/clock"
)

// KeyAPI is the interface called by control-plane or an admin during restart of a node.
type KeyAPI struct {
	listenAddr        string
	log               *logger.Logger
	mux               sync.Mutex
	metadata          metadata.InstanceLister
	issuer            QuoteIssuer
	key               []byte
	measurementSecret []byte
	keyReceived       chan struct{}

	clock    clock.WithTicker
	timeout  time.Duration
	interval time.Duration

	keyproto.UnimplementedAPIServer
}

// New initializes a KeyAPI with the given parameters.
func New(log *logger.Logger, issuer QuoteIssuer, metadata metadata.InstanceLister, timeout time.Duration, interval time.Duration) *KeyAPI {
	return &KeyAPI{
		log:         log,
		metadata:    metadata,
		issuer:      issuer,
		keyReceived: make(chan struct{}, 1),
		clock:       clock.RealClock{},
		timeout:     timeout,
		interval:    interval,
	}
}

// PushStateDiskKey is the rpc to push state disk decryption keys to a restarting node.
func (a *KeyAPI) PushStateDiskKey(ctx context.Context, in *keyproto.PushStateDiskKeyRequest) (*keyproto.PushStateDiskKeyResponse, error) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if len(a.key) != 0 {
		return nil, status.Error(codes.FailedPrecondition, "node already received a passphrase")
	}
	if len(in.StateDiskKey) != crypto.StateDiskKeyLength {
		return nil, status.Errorf(codes.InvalidArgument, "received invalid passphrase: expected length: %d, but got: %d", crypto.StateDiskKeyLength, len(in.StateDiskKey))
	}
	if len(in.MeasurementSecret) != crypto.RNGLengthDefault {
		return nil, status.Errorf(codes.InvalidArgument, "received invalid measurement secret: expected length: %d, but got: %d", crypto.RNGLengthDefault, len(in.MeasurementSecret))
	}

	a.key = in.StateDiskKey
	a.measurementSecret = in.MeasurementSecret
	a.keyReceived <- struct{}{}
	return &keyproto.PushStateDiskKeyResponse{}, nil
}

// WaitForDecryptionKey notifies control-plane nodes to send a decryption key and waits until a key is received.
func (a *KeyAPI) WaitForDecryptionKey(uuid, listenAddr string) (diskKey, measurementSecret []byte, err error) {
	if uuid == "" {
		return nil, nil, errors.New("received no disk UUID")
	}
	a.listenAddr = listenAddr
	creds := atlscredentials.New(a.issuer, nil)
	server := grpc.NewServer(grpc.Creds(creds))
	keyproto.RegisterAPIServer(server, a)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, nil, err
	}
	defer listener.Close()

	a.log.Infof("Waiting for decryption key. Listening on: %s", listener.Addr().String())
	go server.Serve(listener)
	defer server.GracefulStop()

	a.requestKeyLoop(uuid)
	return a.key, a.measurementSecret, nil
}

// ResetKey resets a previously set key.
func (a *KeyAPI) ResetKey() {
	a.key = nil
}

// requestKeyLoop continuously requests decryption keys from all available control-plane nodes, until the KeyAPI receives a key.
func (a *KeyAPI) requestKeyLoop(uuid string, opts ...grpc.DialOption) {
	// we do not perform attestation, since the restarting node does not need to care about notifying the correct node
	// if an incorrect key is pushed by a malicious actor, decrypting the disk will fail, and the node will not start
	creds := atlscredentials.New(a.issuer, nil)

	ticker := a.clock.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		endpoints, err := a.getJoinServiceEndpoints()
		if err != nil {
			a.log.With(zap.Error(err)).Errorf("Failed to get JoinService endpoints")
		} else {
			a.log.Infof("Received list with JoinService endpoints: %v", endpoints)
			for _, endpoint := range endpoints {
				a.requestKey(endpoint, uuid, creds, opts...)
			}
		}

		select {
		case <-a.keyReceived:
			// return if a key was received
			// a key can be send by
			// - a control-plane node, after the request rpc was received
			// - by a Constellation admin, at any time this loop is running on a node during boot
			return
		case <-ticker.C():
		}
	}
}

func (a *KeyAPI) getJoinServiceEndpoints() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()
	return metadata.JoinServiceEndpoints(ctx, a.metadata)
}

func (a *KeyAPI) requestKey(endpoint, uuid string, credentials credentials.TransportCredentials, opts ...grpc.DialOption) {
	opts = append(opts, grpc.WithTransportCredentials(credentials))

	a.log.With(zap.String("endpoint", endpoint)).Infof("Requesting rejoin ticket")
	rejoinTicket, err := a.requestRejoinTicket(endpoint, uuid, opts...)
	if err != nil {
		a.log.With(zap.Error(err), zap.String("endpoint", endpoint)).Errorf("Failed to request rejoin ticket")
		return
	}

	a.log.With(zap.String("endpoint", endpoint)).Infof("Pushing key to own server")
	if err := a.pushKeyToOwnServer(rejoinTicket.StateDiskKey, rejoinTicket.MeasurementSecret, opts...); err != nil {
		a.log.With(zap.Error(err), zap.String("endpoint", a.listenAddr)).Errorf("Failed to push key to own server")
		return
	}
}

func (a *KeyAPI) requestRejoinTicket(endpoint, uuid string, opts ...grpc.DialOption) (*joinproto.IssueRejoinTicketResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("dialing gRPC: %w", err)
	}
	defer conn.Close()
	client := joinproto.NewAPIClient(conn)
	req := &joinproto.IssueRejoinTicketRequest{DiskUuid: uuid}
	return client.IssueRejoinTicket(ctx, req)
}

func (a *KeyAPI) pushKeyToOwnServer(stateDiskKey, measurementSecret []byte, opts ...grpc.DialOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, a.listenAddr, opts...)
	if err != nil {
		return fmt.Errorf("dialing gRPC: %w", err)
	}
	defer conn.Close()
	client := keyproto.NewAPIClient(conn)
	req := &keyproto.PushStateDiskKeyRequest{StateDiskKey: stateDiskKey, MeasurementSecret: measurementSecret}
	_, err = client.PushStateDiskKey(ctx, req)
	return err
}

// QuoteValidator validates quotes.
type QuoteValidator interface {
	oid.Getter
	// Validate validates a quote and returns the user data on success.
	Validate(attDoc []byte, nonce []byte) ([]byte, error)
}

// QuoteIssuer issues quotes.
type QuoteIssuer interface {
	oid.Getter
	// Issue issues a quote for remote attestation for a given message
	Issue(userData []byte, nonce []byte) (quote []byte, err error)
}
