/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

const (
	kekLen               = 32
	importedPublicKeyLen = 2048
)

// fakeAWSClient mocks the AWS KMS client.
type fakeAWSClient struct {
	kekPool       map[string][]byte
	keyIDCount    int
	keyPairStruct *rsa.PrivateKey
}

// GenerateDataKey creates a new DEK.
func (m *fakeAWSClient) GenerateDataKey(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
	if params.KeyId == nil {
		return nil, fmt.Errorf("Missing paramerter KeyId")
	}

	kekID := *params.KeyId
	if _, ok := m.kekPool[kekID]; !ok {
		return nil, kmsInterface.ErrKEKUnknown
	}

	var keyLen int32
	if params.NumberOfBytes != nil {
		keyLen = *params.NumberOfBytes
	} else {
		switch params.KeySpec {
		case types.DataKeySpecAes128:
			keyLen = 128
		case types.DataKeySpecAes256:
			keyLen = 256
		}
	}

	dek := make([]byte, keyLen)

	// should not be random, but dependent on the context and the kekId
	contextKeyID := params.EncryptionContext[DEKContext]
	i := 0
	for i < kekLen/2 {
		dek[i] = contextKeyID[i%len(contextKeyID)]
		i++
	}
	for i < kekLen {
		dek[i] = []byte(*params.KeyId)[i%len(*params.KeyId)]
		i++
	}
	output := &kms.GenerateDataKeyOutput{
		CiphertextBlob: dek,
		KeyId:          &kekID,
		Plaintext:      dek,
	}
	return output, nil
}

// Listed in order called by CreateKEK

// DescribeKey takes an alias and searches in the kekPool whether this alias exists.
// If the alias exists, return the keyId the alias refers to (identifiable via the same key value). Otherwise, indicate that the key was not found.
func (m *fakeAWSClient) DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	if aliasKeyValue, ok := m.kekPool[*params.KeyId]; ok {
		awsGeneratedKeyID := m.getIDPartnerByKeyValue(*params.KeyId, aliasKeyValue)
		if awsGeneratedKeyID == "" {
			return nil, errors.New("could not determine id partner")
		}
		describeKeyOutput := &kms.DescribeKeyOutput{
			KeyMetadata: &types.KeyMetadata{
				KeyId: &awsGeneratedKeyID,
			},
		}
		return describeKeyOutput, nil
	}
	return nil, &types.NotFoundException{Message: aws.String("not found exception error")}
}

// CreateKey stores a key in fakeKMSStore.
func (m *fakeAWSClient) CreateKey(ctx context.Context, params *kms.CreateKeyInput, optFns ...func(*kms.Options)) (*kms.CreateKeyOutput, error) {
	m.keyIDCount++
	kekID := strconv.Itoa(m.keyIDCount)

	if _, ok := m.kekPool[kekID]; ok {
		return nil, errors.New("Key with this id already exists")
	}
	// Use m.keyIdCount to make the key values unique for the search in describeKey
	kekValue := make([]byte, kekLen)
	for i := 0; i < kekLen; i++ {
		kekValue[i] = byte(m.keyIDCount)
	}
	m.kekPool[kekID] = kekValue

	keyOutput := &kms.CreateKeyOutput{
		KeyMetadata: &types.KeyMetadata{
			KeyId: &kekID,
		},
	}
	return keyOutput, nil
}

// CreateAlias changes keyId value to alias.
func (m *fakeAWSClient) CreateAlias(ctx context.Context, params *kms.CreateAliasInput, optFns ...func(*kms.Options)) (*kms.CreateAliasOutput, error) {
	kekValue, ok := m.kekPool[*params.TargetKeyId]
	if !ok {
		return nil, kmsInterface.ErrKEKUnknown
	}
	// Copy the key bytes of the KEK to a new entry using the alias as keyId.
	// As with the KMS, each key can be addressed using either its alias or the incremented keyId generated during creation
	m.kekPool[*params.AliasName] = kekValue
	return &kms.CreateAliasOutput{}, nil
}

// GetParametersForImport returns an empty public key and empty import token.
func (m *fakeAWSClient) GetParametersForImport(ctx context.Context, params *kms.GetParametersForImportInput, optFns ...func(*kms.Options)) (*kms.GetParametersForImportOutput, error) {
	var err error
	m.keyPairStruct, err = rsa.GenerateKey(rand.Reader, importedPublicKeyLen)
	if err != nil {
		return nil, errors.New("Error generating Key Pair")
	}
	var pki any = &m.keyPairStruct.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(pki)
	if err != nil {
		return nil, err
	}
	importToken := make([]byte, 1)
	getParametersForImportOutput := &kms.GetParametersForImportOutput{
		PublicKey:   publicKeyBytes,
		ImportToken: importToken,
	}
	return getParametersForImportOutput, nil
}

// ImportKeyMaterial stores the encrypted KEK in the KMS.
func (m *fakeAWSClient) ImportKeyMaterial(ctx context.Context, params *kms.ImportKeyMaterialInput, optFns ...func(*kms.Options)) (*kms.ImportKeyMaterialOutput, error) {
	if _, ok := m.kekPool[*params.KeyId]; !ok {
		return nil, kmsInterface.ErrKEKUnknown
	}

	// decrypt the encryptedKeyMaterial generated using RSAES_OAEP_SHA_256 to get the imported key bytes
	kekValue, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, m.keyPairStruct, params.EncryptedKeyMaterial, nil)
	if err != nil {
		return nil, errors.New("Error decrypting the wrapped key")
	}

	// Set imported key bytes for keyId and alias
	currentKeyValue := m.kekPool[*params.KeyId]
	aliasID := m.getIDPartnerByKeyValue(*params.KeyId, currentKeyValue)
	if aliasID == "" {
		return nil, errors.New("could not determine id partner")
	}
	m.kekPool[aliasID] = kekValue
	m.kekPool[*params.KeyId] = kekValue

	return &kms.ImportKeyMaterialOutput{}, nil
}

// PutKeyPolicy sets the policy for an existing KEK in the kms.
func (m *fakeAWSClient) PutKeyPolicy(ctx context.Context, params *kms.PutKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.PutKeyPolicyOutput, error) {
	return &kms.PutKeyPolicyOutput{}, nil
}

// Decrypt performs decryption.
// Since the keys are saved in plain text during testing, decrypt returns the saved value.
func (m *fakeAWSClient) Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
	decryptOutput := &kms.DecryptOutput{
		Plaintext: params.CiphertextBlob,
	}
	return decryptOutput, nil
}

// getIDPartnerByKeyValue returns a the alternative name of a key, if it exists.
// For every key there should exist an alias after successful execution of CreateKEK.
// Both key and its alias are stored with the same key value.
// For consistency, they should be updated simultaneously, since they must store the same key value.
func (m *fakeAWSClient) getIDPartnerByKeyValue(keyID1 string, keyValue1 []byte) string {
	// Search for the keyId with the same keyValue the input key has
	for keyID2, keyValue2 := range m.kekPool {
		if bytes.Equal(keyValue2, keyValue1) && keyID2 != keyID1 {
			return keyID2
		}
	}
	return ""
}

// DeleteAlias is a stub.
func (m *fakeAWSClient) DeleteAlias(ctx context.Context, params *kms.DeleteAliasInput, optFns ...func(*kms.Options)) (*kms.DeleteAliasOutput, error) {
	return &kms.DeleteAliasOutput{}, nil
}

// ScheduleKeyDeletion is a stub.
func (m *fakeAWSClient) ScheduleKeyDeletion(ctx context.Context, params *kms.ScheduleKeyDeletionInput, optFns ...func(*kms.Options)) (*kms.ScheduleKeyDeletionOutput, error) {
	return &kms.ScheduleKeyDeletionOutput{}, nil
}

// GenerateDataKeyWithoutPlaintext is a stub.
func (m *fakeAWSClient) GenerateDataKeyWithoutPlaintext(ctx context.Context, params *kms.GenerateDataKeyWithoutPlaintextInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyWithoutPlaintextOutput, error) {
	return nil, errors.New("Not implemented")
}

// Encrypt is a stub.
func (m *fakeAWSClient) Encrypt(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error) {
	return nil, errors.New("Not implemented")
}

type stubKeyPolicyProducer struct {
	createKeyPolicyErr error
}

// CreateKeyPolicy creates a key policy.
func (m *stubKeyPolicyProducer) CreateKeyPolicy(keyID string) (string, error) {
	return "", m.createKeyPolicyErr
}

func TestAWSKMSClient(t *testing.T) {
	assert := assert.New(t)

	awsClient := &fakeAWSClient{kekPool: make(map[string][]byte, 2)}

	client := &KMSClient{
		awsClient:      awsClient,
		policyProducer: &stubKeyPolicyProducer{},
		storage:        storage.NewMemMapStorage(),
	}

	awsClient.keyIDCount = -1

	testKEK1ID := "testKEK1"
	testKEK1 := []byte("test KEK")
	testKEK2ID := "testKEK2"
	testKEK2 := []byte("more test KEK")
	ctx := context.Background()

	// try to get a DEK before setting the KEK
	_, err := client.GetDEK(ctx, testKEK1ID, "volume01", config.SymmetricKeyLength)
	assert.Error(err)
	assert.ErrorIs(err, kmsInterface.ErrKEKUnknown)

	// test CreateKEK method
	assert.NoError(client.CreateKEK(ctx, testKEK1ID, testKEK1))
	assert.Equal(testKEK1, awsClient.kekPool[strconv.Itoa(awsClient.keyIDCount)])

	// make sure that CreateKEK is idempotent
	assert.NoError(client.CreateKEK(ctx, testKEK1ID, testKEK1))
	assert.Equal(testKEK1, awsClient.kekPool[strconv.Itoa(awsClient.keyIDCount)])

	// test setting a second KEK
	assert.NoError(client.CreateKEK(ctx, testKEK2ID, testKEK2))
	assert.Equal(testKEK2, awsClient.kekPool[strconv.Itoa(awsClient.keyIDCount)])

	// test GetDEK method
	dek1, err := client.GetDEK(ctx, testKEK1ID, "volume01", config.SymmetricKeyLength)
	assert.NoError(err)
	dek2, err := client.GetDEK(ctx, testKEK2ID, "volume02", config.SymmetricKeyLength)
	assert.NoError(err)

	// make sure that GetDEK is idempotent
	dek1Copy, err := client.GetDEK(ctx, testKEK1ID, "volume01", config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Equal(dek1, dek1Copy)
	dek2Copy, err := client.GetDEK(ctx, testKEK2ID, "volume02", config.SymmetricKeyLength)
	assert.NoError(err)
	assert.Equal(dek2, dek2Copy)
}

func TestCreateKEK(t *testing.T) {
	someErr := errors.New("error")
	importKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	importPubKey, _ := pem.Decode([]byte(`-----BEGIN PUBLIC KEY-----
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
-----END PUBLIC KEY-----`))

	testCases := map[string]struct {
		client          *stubAWSClient
		policyProducer  KeyPolicyProducer
		importKey       []byte
		cleanupRequired bool
		wantErr         bool
	}{
		"create new kek successful": {
			client:         &stubAWSClient{createKeyID: "key-id"},
			policyProducer: &stubKeyPolicyProducer{},
		},
		"CreateKeyPolicy fails on existing": {
			client:         &stubAWSClient{},
			policyProducer: &stubKeyPolicyProducer{createKeyPolicyErr: someErr},
			wantErr:        true,
		},
		"CreateKeyPolicy fails on new": {
			client:          &stubAWSClient{describeKeyErr: &types.NotFoundException{}},
			policyProducer:  &stubKeyPolicyProducer{createKeyPolicyErr: someErr},
			cleanupRequired: true,
			wantErr:         true,
		},
		"PutKeyPolicy fails on new": {
			client: &stubAWSClient{
				describeKeyErr:  &types.NotFoundException{},
				putKeyPolicyErr: someErr,
				createKeyID:     "key-id",
			},
			policyProducer:  &stubKeyPolicyProducer{},
			cleanupRequired: true,
			wantErr:         true,
		},
		"CreateAlias fails on new": {
			client: &stubAWSClient{
				describeKeyErr: &types.NotFoundException{},
				createAliasErr: someErr,
				createKeyID:    "key-id",
			},
			policyProducer:  &stubKeyPolicyProducer{},
			cleanupRequired: true,
			wantErr:         true,
		},
		"CreateKey fails on new": {
			client:         &stubAWSClient{describeKeyErr: &types.NotFoundException{}, createKeyErr: someErr},
			policyProducer: &stubKeyPolicyProducer{},
			wantErr:        true,
		},
		"DescribeKey fails": {
			client:         &stubAWSClient{describeKeyErr: someErr},
			policyProducer: &stubKeyPolicyProducer{},
			wantErr:        true,
		},
		"DescribeKey fails with not found error": {
			client:         &stubAWSClient{describeKeyErr: &types.NotFoundException{}},
			policyProducer: &stubKeyPolicyProducer{},
		},
		"import kek successful": {
			client:         &stubAWSClient{getParametersForImportPubKey: importPubKey.Bytes},
			policyProducer: &stubKeyPolicyProducer{},
			importKey:      importKey,
		},
		"GetParametersForImport fails on new": {
			client: &stubAWSClient{
				describeKeyErr:            &types.NotFoundException{},
				getParametersForImportErr: someErr,
				createKeyID:               "key-id",
			},
			policyProducer:  &stubKeyPolicyProducer{},
			importKey:       importKey,
			cleanupRequired: true,
			wantErr:         true,
		},
		"ImportKeyMaterial fails on new": {
			client: &stubAWSClient{
				describeKeyErr:       &types.NotFoundException{},
				importKeyMaterialErr: someErr,
				createKeyID:          "key-id",
			},
			policyProducer:  &stubKeyPolicyProducer{},
			importKey:       importKey,
			cleanupRequired: true,
			wantErr:         true,
		},
		"GetParametersForImport fails on existing": {
			client:         &stubAWSClient{getParametersForImportErr: someErr},
			policyProducer: &stubKeyPolicyProducer{},
			importKey:      importKey,
			wantErr:        true,
		},
		"ImportKeyMaterial fails on existing": {
			client:         &stubAWSClient{importKeyMaterialErr: someErr},
			policyProducer: &stubKeyPolicyProducer{},
			importKey:      importKey,
			wantErr:        true,
		},
		"errors during cleanup don't stop execution": {
			client: &stubAWSClient{
				describeKeyErr: &types.NotFoundException{},
				deleteAliasErr: someErr,
				createKeyID:    "key-id",
			},
			policyProducer:  &stubKeyPolicyProducer{createKeyPolicyErr: someErr},
			cleanupRequired: true,
			wantErr:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := KMSClient{
				awsClient:      tc.client,
				storage:        storage.NewMemMapStorage(),
				policyProducer: tc.policyProducer,
			}

			err := client.CreateKEK(context.Background(), "test-key", tc.importKey)
			if tc.wantErr {
				assert.Error(err)
				if tc.cleanupRequired {
					assert.True(tc.client.cleanUpCalled, "failed to clean up")
				} else {
					assert.False(tc.client.cleanUpCalled, "cleaned up when not necessary")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubAWSClient struct {
	cleanUpCalled                      bool
	createAliasErr                     error
	createKeyErr                       error
	createKeyID                        string
	decryptErr                         error
	deleteAliasErr                     error
	describeKeyErr                     error
	encryptErr                         error
	generateDataKeyErr                 error
	generateDataKeyWithoutPlaintextErr error
	getParametersForImportErr          error
	getParametersForImportPubKey       []byte
	importKeyMaterialErr               error
	putKeyPolicyErr                    error
	scheduleKeyDeletionErr             error
}

func (s *stubAWSClient) CreateAlias(ctx context.Context, params *kms.CreateAliasInput, optFns ...func(*kms.Options)) (*kms.CreateAliasOutput, error) {
	return &kms.CreateAliasOutput{}, s.createAliasErr
}

func (s *stubAWSClient) CreateKey(ctx context.Context, params *kms.CreateKeyInput, optFns ...func(*kms.Options)) (*kms.CreateKeyOutput, error) {
	return &kms.CreateKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: aws.String(s.createKeyID)}}, s.createKeyErr
}

func (s *stubAWSClient) Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
	return &kms.DecryptOutput{}, s.decryptErr
}

func (s *stubAWSClient) DeleteAlias(ctx context.Context, params *kms.DeleteAliasInput, optFns ...func(*kms.Options)) (*kms.DeleteAliasOutput, error) {
	return &kms.DeleteAliasOutput{}, s.deleteAliasErr
}

func (s *stubAWSClient) DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	return &kms.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: params.KeyId}}, s.describeKeyErr
}

func (s *stubAWSClient) Encrypt(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error) {
	return &kms.EncryptOutput{}, s.encryptErr
}

func (s *stubAWSClient) GenerateDataKey(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
	return &kms.GenerateDataKeyOutput{}, s.generateDataKeyErr
}

func (s *stubAWSClient) GenerateDataKeyWithoutPlaintext(ctx context.Context, params *kms.GenerateDataKeyWithoutPlaintextInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyWithoutPlaintextOutput, error) {
	return &kms.GenerateDataKeyWithoutPlaintextOutput{}, s.generateDataKeyWithoutPlaintextErr
}

func (s *stubAWSClient) GetParametersForImport(ctx context.Context, params *kms.GetParametersForImportInput, optFns ...func(*kms.Options)) (*kms.GetParametersForImportOutput, error) {
	return &kms.GetParametersForImportOutput{
		PublicKey: s.getParametersForImportPubKey,
	}, s.getParametersForImportErr
}

func (s *stubAWSClient) ImportKeyMaterial(ctx context.Context, params *kms.ImportKeyMaterialInput, optFns ...func(*kms.Options)) (*kms.ImportKeyMaterialOutput, error) {
	return &kms.ImportKeyMaterialOutput{}, s.importKeyMaterialErr
}

func (s *stubAWSClient) PutKeyPolicy(ctx context.Context, params *kms.PutKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.PutKeyPolicyOutput, error) {
	return &kms.PutKeyPolicyOutput{}, s.putKeyPolicyErr
}

func (s *stubAWSClient) ScheduleKeyDeletion(ctx context.Context, params *kms.ScheduleKeyDeletionInput, optFns ...func(*kms.Options)) (*kms.ScheduleKeyDeletionOutput, error) {
	s.cleanUpCalled = true
	return &kms.ScheduleKeyDeletionOutput{}, s.scheduleKeyDeletionErr
}
