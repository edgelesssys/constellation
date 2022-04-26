package proto

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// dummyAVPNActivateAsCoordinatorClient is a dummy and panics if Recv() is called.
type dummyAVPNActivateAsCoordinatorClient struct {
	grpc.ClientStream
}

func (c dummyAVPNActivateAsCoordinatorClient) Recv() (*pubproto.ActivateAsCoordinatorResponse, error) {
	panic("i'm a dummy, Recv() not implemented")
}

// dummyAVPNActivateAsCoordinatorClient is a dummy and panics if Recv() is called.
type dummyAVPNActivateAdditionalNodesClient struct {
	grpc.ClientStream
}

func (c dummyAVPNActivateAdditionalNodesClient) Recv() (*pubproto.ActivateAdditionalNodesResponse, error) {
	panic("i'm a dummy, Recv() not implemented")
}

// stubAVPNActivationAsCoordinatorClient recives responses from an predefined
// response stream iterator or a stub error.
type stubAVPNActivationAsCoordinatorClient struct {
	grpc.ClientStream

	stream  *stubActivateAsCoordinatorResponseIter
	recvErr error
}

func (c stubAVPNActivationAsCoordinatorClient) Recv() (*pubproto.ActivateAsCoordinatorResponse, error) {
	if c.recvErr != nil {
		return nil, c.recvErr
	}
	return c.stream.Next()
}

// stubActivateAsCoordinatorResponseIter is an iterator over a slice of
// ActivateAsCoordinatorResponses. It returns the messages in the order
// they occur in the slice and returns an io.EOF error when no response
// is left.
type stubActivateAsCoordinatorResponseIter struct {
	msgs []*pubproto.ActivateAsCoordinatorResponse
}

// Next returns the next message from the message slice or an io.EOF error
// if the message slice is empty.
func (q *stubActivateAsCoordinatorResponseIter) Next() (*pubproto.ActivateAsCoordinatorResponse, error) {
	if len(q.msgs) == 0 {
		return nil, io.EOF
	}
	msg := q.msgs[0]
	q.msgs = q.msgs[1:]
	return msg, nil
}

func TestNextLog(t *testing.T) {
	testClientVpnIp := "192.0.2.1"
	testCoordinatorVpnKey := []byte("32bytesWireGuardKeyForTheTesting")
	testCoordinatorVpnKey64 := []byte("MzJieXRlc1dpcmVHdWFyZEtleUZvclRoZVRlc3Rpbmc=")
	testKubeconfig := []byte("apiVersion:v1 kind:Config...")
	testConfigResp := &pubproto.ActivateAsCoordinatorResponse{
		Content: &pubproto.ActivateAsCoordinatorResponse_AdminConfig{
			AdminConfig: &pubproto.AdminConfig{
				AdminVpnIp:           testClientVpnIp,
				CoordinatorVpnPubKey: testCoordinatorVpnKey,
				Kubeconfig:           testKubeconfig,
			},
		},
	}
	testLogMessage := "some log message"
	testLogResp := &pubproto.ActivateAsCoordinatorResponse{
		Content: &pubproto.ActivateAsCoordinatorResponse_Log{
			Log: &pubproto.Log{
				Message: testLogMessage,
			},
		},
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		msgs       []*pubproto.ActivateAsCoordinatorResponse
		wantLogLen int
		wantState  bool
		recvErr    error
		wantErr    bool
	}{
		"some logs": {
			msgs:       []*pubproto.ActivateAsCoordinatorResponse{testLogResp, testLogResp, testLogResp},
			wantLogLen: 3,
		},
		"only admin config": {
			msgs:      []*pubproto.ActivateAsCoordinatorResponse{testConfigResp},
			wantState: true,
		},
		"logs and configs": {
			msgs:       []*pubproto.ActivateAsCoordinatorResponse{testLogResp, testConfigResp, testLogResp, testConfigResp},
			wantLogLen: 2,
			wantState:  true,
		},
		"no response": {
			msgs:       []*pubproto.ActivateAsCoordinatorResponse{},
			wantLogLen: 0,
		},
		"recv fail": {
			recvErr: someErr,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			respClient := stubAVPNActivationAsCoordinatorClient{
				stream: &stubActivateAsCoordinatorResponseIter{
					msgs: tc.msgs,
				},
				recvErr: tc.recvErr,
			}
			client := NewActivationRespClient(respClient)

			var logs []string
			var err error
			for err == nil {
				var log string
				log, err = client.NextLog()
				if err == nil {
					logs = append(logs, log)
				}
			}

			assert.Error(err)
			if tc.wantErr {
				assert.NotErrorIs(err, io.EOF)
				return
			}

			assert.ErrorIs(err, io.EOF)
			assert.Len(logs, tc.wantLogLen)

			if tc.wantState {
				ip, err := client.GetClientVpnIp()
				assert.NoError(err)
				assert.Equal(testClientVpnIp, ip)
				config, err := client.GetKubeconfig()
				assert.NoError(err)
				assert.Equal(string(testKubeconfig), config)
				key, err := client.GetCoordinatorVpnKey()
				assert.NoError(err)
				assert.Equal(string(testCoordinatorVpnKey64), key)
			}
		})
	}
}

func TestPrintLogStream(t *testing.T) {
	assert := assert.New(t)

	//
	// 10 logs a 10 byte
	//
	var msgs []*pubproto.ActivateAsCoordinatorResponse
	for i := 0; i < 10; i++ {
		msgs = append(msgs, &pubproto.ActivateAsCoordinatorResponse{
			Content: &pubproto.ActivateAsCoordinatorResponse_Log{
				Log: &pubproto.Log{
					Message: "10BytesLog",
				},
			},
		})
	}
	respClient := stubAVPNActivationAsCoordinatorClient{
		stream: &stubActivateAsCoordinatorResponseIter{
			msgs: msgs,
		},
	}
	client := NewActivationRespClient(respClient)
	out := &bytes.Buffer{}
	assert.NoError(client.WriteLogStream(out))
	assert.Equal(out.Len(), 10*11) // 10 messages * (len(message) + 1 newline)

	//
	// Check error handling.
	//
	someErr := errors.New("failed")
	respClient = stubAVPNActivationAsCoordinatorClient{
		recvErr: someErr,
	}
	client = NewActivationRespClient(respClient)
	assert.Error(client.WriteLogStream(&bytes.Buffer{}))
}

func TestGetKubeconfig(t *testing.T) {
	assert := assert.New(t)

	client := NewActivationRespClient(dummyAVPNActivateAsCoordinatorClient{})
	_, err := client.GetKubeconfig()
	assert.Error(err)

	client.kubeconfig = "apiVersion:v1 kind:Config..."
	config, err := client.GetKubeconfig()
	assert.NoError(err)
	assert.Equal("apiVersion:v1 kind:Config...", config)
}

func TestGetCoordinatorVpnKey(t *testing.T) {
	assert := assert.New(t)

	client := NewActivationRespClient(dummyAVPNActivateAsCoordinatorClient{})
	_, err := client.GetCoordinatorVpnKey()
	assert.Error(err)

	client.coordinatorVpnKey = "32bytesWireGuardKeyForTheTesting"
	key, err := client.GetCoordinatorVpnKey()
	assert.NoError(err)
	assert.Equal("32bytesWireGuardKeyForTheTesting", key)
}

func TestGetClientVpnIp(t *testing.T) {
	assert := assert.New(t)

	client := NewActivationRespClient(dummyAVPNActivateAsCoordinatorClient{})
	_, err := client.GetClientVpnIp()
	assert.Error(err)

	client.clientVpnIp = "192.0.2.1"
	ip, err := client.GetClientVpnIp()
	assert.NoError(err)
	assert.Equal("192.0.2.1", ip)
}
