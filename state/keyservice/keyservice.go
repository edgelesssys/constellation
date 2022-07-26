package keyservice

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/edgelesssys/constellation/joinservice/joinproto"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
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
	timeout           time.Duration
	keyproto.UnimplementedAPIServer
}

// New initializes a KeyAPI with the given parameters.
func New(log *logger.Logger, issuer QuoteIssuer, metadata metadata.InstanceLister, timeout time.Duration) *KeyAPI {
	return &KeyAPI{
		log:         log,
		metadata:    metadata,
		issuer:      issuer,
		keyReceived: make(chan struct{}, 1),
		timeout:     timeout,
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
	// set up for the select statement to immediately request a key, skipping the initial delay caused by using a ticker
	firstReq := make(chan struct{}, 1)
	firstReq <- struct{}{}

	ticker := time.NewTicker(a.timeout)
	defer ticker.Stop()
	for {
		select {
		// return if a key was received
		// a key can be send by
		// - a control-plane node, after the request rpc was received
		// - by a Constellation admin, at any time this loop is running on a node during boot
		case <-a.keyReceived:
			return
		case <-ticker.C:
			a.requestKey(uuid, creds, opts...)
		case <-firstReq:
			a.requestKey(uuid, creds, opts...)
		}
	}
}

func (a *KeyAPI) requestKey(uuid string, credentials credentials.TransportCredentials, opts ...grpc.DialOption) {
	// list available control-plane nodes
	endpoints, _ := metadata.JoinServiceEndpoints(context.Background(), a.metadata)

	a.log.With(zap.Strings("endpoints", endpoints)).Infof("Sending a key request to available control-plane nodes")
	// notify all available control-plane nodes to send a key to the node
	// any errors encountered here will be ignored, and the calls retried after a timeout
	for _, endpoint := range endpoints {
		ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
		defer cancel()

		// request rejoin ticket from JoinService
		conn, err := grpc.DialContext(ctx, endpoint, append(opts, grpc.WithTransportCredentials(credentials))...)
		if err != nil {
			continue
		}
		defer conn.Close()
		client := joinproto.NewAPIClient(conn)
		response, err := client.IssueRejoinTicket(ctx, &joinproto.IssueRejoinTicketRequest{DiskUuid: uuid})
		if err != nil {
			a.log.With(zap.Error(err), zap.String("endpoint", endpoint)).Warnf("Failed to request key")
			continue
		}

		// push key to own gRPC server
		pushKeyConn, err := grpc.DialContext(ctx, a.listenAddr, append(opts, grpc.WithTransportCredentials(credentials))...)
		if err != nil {
			continue
		}
		defer pushKeyConn.Close()
		pushKeyClient := keyproto.NewAPIClient(pushKeyConn)
		if _, err := pushKeyClient.PushStateDiskKey(
			ctx,
			&keyproto.PushStateDiskKeyRequest{StateDiskKey: response.StateDiskKey, MeasurementSecret: response.MeasurementSecret},
		); err != nil {
			a.log.With(zap.Error(err), zap.String("endpoint", a.listenAddr)).Errorf("Failed to push key")
			continue
		}
	}
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
