package store

import (
	"fmt"
)

// Store is the interface for persistence.
type Store interface {
	// Get returns a value from store by key.
	Get(string) ([]byte, error)
	// Put saves a value to store by key.
	Put(string, []byte) error
	// Delete deletes the key.
	Delete(string) error
	// Iterator returns an Iterator for a given prefix.
	Iterator(string) (Iterator, error)
	// BeginTransaction starts a new transaction.
	BeginTransaction() (Transaction, error)
	// Transfer copies the whole store Database.
	Transfer(Store) error
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

// ValueUnsetError is an error raised by unset values in the store.
type ValueUnsetError struct {
	requestedValue string
}

// Error implements the Error interface.
func (s *ValueUnsetError) Error() string {
	return fmt.Sprintf("store: requested value not set: %s", s.requestedValue)
}

// NoElementsLeftError occurs when trying to get an element from an interator that
// doesn't have elements left.
type NoElementsLeftError struct {
	idx int
}

// Error implements the Error interface.
func (n *NoElementsLeftError) Error() string {
	return fmt.Sprintf("index out of range [%d]", n.idx)
}

// TransactionAlreadyCommittedError occurs when further operations:
// Get, Put, Delete or Iterate are called on a committed transaction.
type TransactionAlreadyCommittedError struct {
	op string
}

func (t *TransactionAlreadyCommittedError) Error() string {
	return fmt.Sprintf("transaction is already committed, but %s is called", t.op)
}
