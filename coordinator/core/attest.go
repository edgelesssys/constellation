package core

import (
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
