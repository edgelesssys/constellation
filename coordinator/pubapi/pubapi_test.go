package pubapi

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	grpcpeer "google.golang.org/grpc/peer"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestGetRecoveryPeerFromContext(t *testing.T) {
	assert := assert.New(t)
	testIP := "192.0.2.1"
	testPort := 1234
	wantPeer := net.JoinHostPort(testIP, "9000")

	addr := &net.TCPAddr{IP: net.ParseIP(testIP), Port: testPort}
	ctx := grpcpeer.NewContext(context.Background(), &grpcpeer.Peer{Addr: addr})

	peer, err := GetRecoveryPeerFromContext(ctx)
	assert.NoError(err)
	assert.Equal(wantPeer, peer)

	_, err = GetRecoveryPeerFromContext(context.Background())
	assert.Error(err)
}
