package gcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/kms/config"
	kmsInterface "github.com/edgelesssys/constellation/kms/kms"
	"github.com/edgelesssys/constellation/kms/kms/util"
	"github.com/edgelesssys/constellation/kms/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var (
	testKey    = []byte{0x52, 0xFD, 0xFC, 0x07, 0x21, 0x82, 0x65, 0x4F, 0x16, 0x3F, 0x5F, 0x0F, 0x9A, 0x62, 0x1D, 0x72, 0x95, 0x66, 0xC7, 0x4D, 0x10, 0x03, 0x7C, 0x4D, 0x7B, 0xBB, 0x04, 0x07, 0xD1, 0xE2, 0xC6, 0x49}
	testKeyRSA = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAu+OepfHCTiTi27nkTGke
dn+AIkiM1AIWWDwqfqG85aNulcj60mGQGXIYV8LoEVkyKOhYBIUmJUaVczB4ltqq
ZhR7l46RQw2vnv+XiUmfK555d4ZDInyjTusO69hE6tkuYKdXLlG1HzcrhJ254LE2
wXtE1Yf9DygOsWet+S32gmpfH2whUY1mRTdwW4zoY4c3qtmmWImhVVNr6qR8Z95X
Y49EteCoNIomQNEZH7EnMlBsh34L7doOsckh1aTvQcrJorQSrBkWKbdV6kvuBKZp
fLK0DZiOh9BwZCZANtOqgH3V+AuNk338iON8eKCFRjoiQ40YGM6xKH3E6PHVnuKt
uIO0MPvE0qdV8Lvs+nCCrvwP5sJKZuciM40ioEO1pV1y3491xIxYhx3OfN4gg2h8
cgdKob/R8qwxqTrfceO36FBFb1vXCUApsm5oy6WxmUtIUgoYhK+6JYpVWDyOJYwP
iMJhdJA65n2ZliN8NxEhsaFoMgw76BOiD0wkt/CKPmNbOm5MGS3/fiZCt6A6u3cn
Ubhn4tvjy/q5XzVqZtBeoseW2TyyrsAN53LBkSqag5tG/264CQDigQ6Y/OADOE2x
n08MyrFHIL/wFMscOvJo7c2Eo4EW1yXkEkAy5tF5PZgnfRObakj4gdqPeq18FNzc
Y+t5OxL3kL15VzY1Ob0d5cMCAwEAAQ==
-----END PUBLIC KEY-----`
)

// Google KMS testing implementation taken from: https://github.com/googleapis/google-cloud-go/blob/kms/v1.1.0/kms/apiv1/mock_test.go
//
// To keep the tests simple this only implements the methods required by our Google KMS client.
// More methods can be added as needed.
type mockKeyManagementServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	kmspb.KeyManagementServiceServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

// CreateCryptoKey creates a new KEK.
func (s *mockKeyManagementServer) CreateCryptoKey(ctx context.Context, req *kmspb.CreateCryptoKeyRequest) (*kmspb.CryptoKey, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.popResponse().(*kmspb.CryptoKey), nil
}

// Decrypt performs decryption.
func (s *mockKeyManagementServer) Decrypt(ctx context.Context, req *kmspb.DecryptRequest) (*kmspb.DecryptResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	res := s.popResponse().(*kmspb.DecryptResponse)
	res.Plaintext = make([]byte, len(req.Ciphertext))
	for i, v := range req.Ciphertext {
		res.Plaintext[len(res.Plaintext)-1-i] = v
	}
	return res, nil
}

// Encrypt performs encryption.
func (s *mockKeyManagementServer) Encrypt(ctx context.Context, req *kmspb.EncryptRequest) (*kmspb.EncryptResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	res := s.popResponse().(*kmspb.EncryptResponse)
	// Reverse the input string to generate a ciphertext
	res.Ciphertext = make([]byte, len(req.Plaintext))
	for i, v := range req.Plaintext {
		res.Ciphertext[len(res.Ciphertext)-1-i] = v
	}
	return res, nil
}

// CreateImportJob creates a new import job.
func (s *mockKeyManagementServer) CreateImportJob(ctx context.Context, req *kmspb.CreateImportJobRequest) (*kmspb.ImportJob, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.popResponse().(*kmspb.ImportJob), nil
}

// ImportCryptoKeyVersion imports a KEK using an import job.
func (s *mockKeyManagementServer) ImportCryptoKeyVersion(ctx context.Context, req *kmspb.ImportCryptoKeyVersionRequest) (*kmspb.CryptoKeyVersion, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.popResponse().(*kmspb.CryptoKeyVersion), nil
}

// UpdateCryptoKeyPrimaryVersion sets the primary version of a KEK.
func (s *mockKeyManagementServer) UpdateCryptoKeyPrimaryVersion(ctx context.Context, req *kmspb.UpdateCryptoKeyPrimaryVersionRequest) (*kmspb.CryptoKey, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.popResponse().(*kmspb.CryptoKey), nil
}

// GetImportJob returns information about a running import job.
func (s *mockKeyManagementServer) GetImportJob(ctx context.Context, req *kmspb.GetImportJobRequest) (*kmspb.ImportJob, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.popResponse().(*kmspb.ImportJob), nil
}

func (s *mockKeyManagementServer) popResponse() proto.Message {
	resp := s.resps[0]
	if len(s.resps) > 1 {
		s.resps = s.resps[1:]
	}
	return resp
}

func TestGoogleKMS(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	serv := grpc.NewServer()
	defer serv.GracefulStop()
	var mockKeyManagement mockKeyManagementServer
	kmspb.RegisterKeyManagementServiceServer(serv, &mockKeyManagement)
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	go serv.Serve(lis)

	project := "test-project"
	location := "global"
	keyRing := "test-key-ring"
	kekName := "test-kek"
	dekName := "test-dek"
	plainDEK := []byte("plain DEK")

	// load responses
	mockKeyManagement.resps = []proto.Message{
		&kmspb.CryptoKey{
			Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyRing, kekName),
		},
		&kmspb.EncryptResponse{
			Name: dekName,
		},
		&kmspb.DecryptResponse{
			Plaintext: plainDEK,
		},
		&kmspb.CryptoKey{
			Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyRing, kekName),
		},
		&kmspb.ImportJob{
			Name: "import-job",
		},
		&kmspb.ImportJob{
			Name:  "import-job",
			State: kmspb.ImportJob_ACTIVE,
			PublicKey: &kmspb.ImportJob_WrappingPublicKey{
				Pem: testKeyRSA,
			},
		},
		&kmspb.CryptoKeyVersion{
			Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s/cryptoKeyVersions/1", project, location, keyRing, kekName),
		},
		&kmspb.CryptoKey{
			Name: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyRing, kekName),
		},
	}

	store := storage.NewMemMapStorage()
	client := New(project, location, keyRing, store, kmspb.ProtectionLevel_SOFTWARE)

	// redirect client calls to mock kms
	// since the connection is closed after each call, we need to reset this option every time
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	ctx := context.Background()

	// Create KEK
	assert.NoError(client.CreateKEK(ctx, kekName, nil))

	// Encrypt and save new DEK
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	err = client.putDEK(ctx, kekName, dekName, plainDEK)
	assert.NoError(err)
	savedDEK, err := store.Get(ctx, dekName)
	require.NoError(err)
	assert.NotEqual(plainDEK, savedDEK)

	// Decrypt DEK
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	res, err := client.GetDEK(ctx, kekName, dekName, config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Equal(plainDEK, res)

	// Import a key
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	assert.NoError(client.CreateKEK(ctx, kekName, testKey))
}

func TestGetNewDEK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	serv := grpc.NewServer()
	defer serv.GracefulStop()
	var mockKeyManagement mockKeyManagementServer
	kmspb.RegisterKeyManagementServiceServer(serv, &mockKeyManagement)
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	go serv.Serve(lis)

	project := "test-project"
	location := "global"
	keyRing := "test-key-ring"
	kekName := "test-kek"
	dekName := "test-dek"
	largeDEKName := "test-dek-large"

	store := storage.NewMemMapStorage()
	client := New(project, location, keyRing, store, kmspb.ProtectionLevel_SOFTWARE)

	mockKeyManagement.resps = []proto.Message{
		&kmspb.EncryptResponse{
			Name: dekName,
		},
		&kmspb.DecryptResponse{},
		&kmspb.EncryptResponse{
			Name: largeDEKName,
		},
	}
	ctx := context.Background()

	// Requesting an unset DEK should generate a new one, which we can then fetch in a second request
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	res1, err := client.GetDEK(ctx, kekName, dekName, config.SymmetricKeyLength)
	assert.NoError(err)
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	res2, err := client.GetDEK(ctx, kekName, dekName, config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Equal(res1, res2)

	// Requesting larger key sizes should be possible
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	res3, err := client.GetDEK(ctx, kekName, largeDEKName, 96)
	assert.NoError(err)
	assert.Len(res3, 96)
}

func TestUnknownKEK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	serv := grpc.NewServer()
	defer serv.GracefulStop()
	var mockKeyManagement mockKeyManagementServer
	kmspb.RegisterKeyManagementServiceServer(serv, &mockKeyManagement)
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	go serv.Serve(lis)

	mockKeyManagement.err = errors.New("rpc error: code = NotFound")

	store := storage.NewMemMapStorage()
	client := New("test-project", "global", "test-key-ring", store, kmspb.ProtectionLevel_SOFTWARE)
	ctx := context.Background()

	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	err = client.putDEK(ctx, "invalid-kek", "test-dek", []byte("dek"))
	assert.Error(err)
	assert.ErrorIs(err, kmsInterface.ErrKEKUnknown)

	require.NoError(store.Put(ctx, "test-dek", []byte("Test Key")))
	client.clientOpts = []option.ClientOption{getConnection(lis.Addr().String(), require)}
	_, err = client.GetDEK(ctx, "invalid-kek", "test-dek", config.SymmetricKeyLength)
	assert.Error(err)
	assert.ErrorIs(err, kmsInterface.ErrKEKUnknown)
}

func getConnection(lisAddr string, r *require.Assertions) option.ClientOption {
	conn, err := grpc.Dial(lisAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	r.NoError(err)
	return option.WithGRPCConn(conn)
}

func TestWrapKeyRSA(t *testing.T) {
	assert := assert.New(t)

	rsaPub, err := util.ParsePEMtoPublicKeyRSA([]byte(testKeyRSA))
	assert.NoError(err)

	res, err := wrapCryptoKey(testKey, rsaPub)
	assert.NoError(err)
	assert.Equal(552, len(res))
}
