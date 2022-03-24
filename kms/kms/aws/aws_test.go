package aws

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/edgelesssys/constellation/kms/config"
	kmsInterface "github.com/edgelesssys/constellation/kms/kms"
	"github.com/edgelesssys/constellation/kms/storage"
	"github.com/stretchr/testify/assert"
)

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
	_, ok := m.kekPool[kekID]
	if !ok {
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
	m.keyIDCount += 1
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
	var pki interface{} = &m.keyPairStruct.PublicKey
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

type stubKeyPolicyProducer struct{}

// CreateKeyPolicy creates a key policy.
func (m *stubKeyPolicyProducer) CreateKeyPolicy(keyID string) (string, error) {
	return "", nil
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
