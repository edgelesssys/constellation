/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package atlscredentials

import (
	"bytes"
	"context"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestATLSCredentials(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	oid := fakeOID{1, 3, 9900, 1}

	//
	// Create servers
	//

	serverCreds := New(fakeIssuer{fakeOID: oid}, nil)

	const serverCount = 15
	var listeners []*bufconn.Listener
	for i := 0; i < serverCount; i++ {
		api := &fakeAPI{}
		server := grpc.NewServer(grpc.Creds(serverCreds))
		initproto.RegisterAPIServer(server, api)

		listener := bufconn.Listen(1024)
		listeners = append(listeners, listener)

		defer server.GracefulStop()
		go server.Serve(listener)
	}

	//
	// Dial concurrently
	//

	clientCreds := New(nil, []atls.Validator{fakeValidator{fakeOID: oid}})

	errChan := make(chan error, serverCount)
	for _, listener := range listeners {
		lis := listener
		go func() {
			var err error
			defer func() { errChan <- err }()
			conn, err := grpc.DialContext(context.Background(), "", grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return lis.Dial()
			}), grpc.WithTransportCredentials(clientCreds))
			require.NoError(err)
			defer conn.Close()

			client := initproto.NewAPIClient(conn)
			_, err = client.Init(context.Background(), &initproto.InitRequest{})
		}()
	}

	for i := 0; i < serverCount; i++ {
		assert.NoError(<-errChan)
	}
}

type fakeIssuer struct {
	fakeOID
}

func (fakeIssuer) Issue(_ context.Context, userData []byte, nonce []byte) ([]byte, error) {
	return json.Marshal(fakeDoc{UserData: userData, Nonce: nonce})
}

type fakeValidator struct {
	fakeOID
	err error
}

func (v fakeValidator) Validate(_ context.Context, attDoc []byte, nonce []byte) ([]byte, error) {
	var doc fakeDoc
	if err := json.Unmarshal(attDoc, &doc); err != nil {
		return nil, err
	}
	if !bytes.Equal(doc.Nonce, nonce) {
		return nil, errors.New("invalid nonce")
	}
	return doc.UserData, v.err
}

type fakeOID asn1.ObjectIdentifier

func (o fakeOID) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier(o)
}

type fakeDoc struct {
	UserData []byte
	Nonce    []byte
}

type fakeAPI struct {
	initproto.UnimplementedAPIServer
}

func (f *fakeAPI) Init(_ *initproto.InitRequest, stream initproto.API_InitServer) error {
	_ = stream.Send(&initproto.InitResponse{})
	return nil
}
