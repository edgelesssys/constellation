package core

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/internal/oid"
)

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

type mockAttDoc struct {
	UserData []byte
	Nonce    []byte
}

func newMockAttDoc(userData []byte, nonce []byte) *mockAttDoc {
	return &mockAttDoc{UserData: userData, Nonce: nonce}
}

type MockValidator struct {
	oid.Dummy
}

// NewMockValidator returns a new MockValidator object.
func NewMockValidator() *MockValidator {
	return &MockValidator{}
}

// Validate implements the Validator interface.
func (m *MockValidator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	var doc mockAttDoc

	if err := json.Unmarshal(attDoc, &doc); err != nil {
		return nil, err
	}

	if !bytes.Equal(doc.Nonce, nonce) {
		return nil, fmt.Errorf("attDoc not valid: nonce not found")
	}
	return doc.UserData, nil
}

// MockIssuer is a mockup quote issuer.
type MockIssuer struct {
	oid.Dummy
}

// NewMockIssuer returns a new MockIssuer object.
func NewMockIssuer() *MockIssuer {
	return &MockIssuer{}
}

// Issue implements the Issuer interface.
func (m *MockIssuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	return json.Marshal(newMockAttDoc(userData, nonce))
}
