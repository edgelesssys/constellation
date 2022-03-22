package store

import (
	"fmt"
)

// Store is the interface for persistence.
type Store interface {
	// BeginTransaction starts a new transaction.
	BeginTransaction() (Transaction, error)
	// Get returns a value from store by key.
	Get(string) ([]byte, error)
	// Put saves a value to store by key.
	Put(string, []byte) error
	// Iterator returns an Iterator for a given prefix.
	Iterator(string) (Iterator, error)
	// Transfer copies the whole store Database.
	Transfer(Store) error
	// Delete deletes the key.
	Delete(string) error
}

// Transaction is a Store transaction.
type Transaction interface {
	// Get returns a value from store by key.
	Get(string) ([]byte, error)
	// Put saves a value to store by key.
	Put(string, []byte) error
	// Delete deletes the key.
	Delete(string) error
	// Iterator returns an Iterator for a given prefix.
	Iterator(string) (Iterator, error)
	// Commit ends a transaction and persists the changes.
	Commit() error
	// Rollback aborts a transaction. Noop if already committed.
	Rollback()
}

// Iterator is an iterator for the store.
type Iterator interface {
	// GetNext returns the next element of the iterator.
	GetNext() (string, error)
	// HasNext returns true if there are elements left to get with GetNext().
	HasNext() bool
}

// StoreValueUnsetError is an error raised by unset values in the store.
type StoreValueUnsetError struct {
	requestedValue string
}

// Error implements the Error interface.
func (s *StoreValueUnsetError) Error() string {
	return fmt.Sprintf("store: requested value not set: %s", s.requestedValue)
}
