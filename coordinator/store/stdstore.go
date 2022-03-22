package store

import (
	"fmt"
	"strings"
	"sync"
)

// StdStore is the standard implementation of the Store interface.
type StdStore struct {
	data       map[string]string
	mut, txmut sync.Mutex
}

// NewStdStore creates and initializes a new StdStore object.
func NewStdStore() *StdStore {
	s := &StdStore{
		data: make(map[string]string),
	}

	return s
}

// Get retrieves a value from StdStore by Type and Name.
func (s *StdStore) Get(request string) ([]byte, error) {
	s.mut.Lock()
	value, ok := s.data[request]
	s.mut.Unlock()

	if ok {
		return []byte(value), nil
	}
	return nil, &StoreValueUnsetError{requestedValue: request}
}

// Put saves a value in StdStore by Type and Name.
func (s *StdStore) Put(request string, requestData []byte) error {
	tx, err := s.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.Put(request, requestData); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *StdStore) Delete(key string) error {
	tx, err := s.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.Delete(key); err != nil {
		return err
	}
	return tx.Commit()
}

// Iterator returns an iterator for keys saved in StdStore with a given prefix.
// For an empty prefix this is an iterator for all keys in StdStore.
func (s *StdStore) Iterator(prefix string) (Iterator, error) {
	keys := make([]string, 0)
	s.mut.Lock()
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	s.mut.Unlock()

	return &StdIterator{0, keys}, nil
}

// BeginTransaction starts a new transaction.
func (s *StdStore) BeginTransaction() (Transaction, error) {
	tx := stdTransaction{store: s, data: map[string]string{}}
	s.txmut.Lock()

	s.mut.Lock()
	for k, v := range s.data {
		tx.data[k] = v
	}
	s.mut.Unlock()

	return &tx, nil
}

func (s *StdStore) commit(data map[string]string) error {
	s.mut.Lock()
	s.data = data
	s.mut.Unlock()

	s.txmut.Unlock()

	return nil
}

func (s *StdStore) Transfer(newstore Store) error {
	s.mut.Lock()
	// copy key:value pairs from the old storage into etcd
	for key, value := range s.data {
		if err := newstore.Put(key, []byte(value)); err != nil {
			return err
		}
	}
	s.mut.Unlock()
	return nil
}

type stdTransaction struct {
	store *StdStore
	data  map[string]string
}

// Get retrieves a value.
func (t *stdTransaction) Get(request string) ([]byte, error) {
	if value, ok := t.data[request]; ok {
		return []byte(value), nil
	}
	return nil, &StoreValueUnsetError{requestedValue: request}
}

// Put saves a value.
func (t *stdTransaction) Put(request string, requestData []byte) error {
	t.data[request] = string(requestData)
	return nil
}

func (t *stdTransaction) Delete(key string) error {
	delete(t.data, key)
	return nil
}

// Iterator returns an iterator for all keys in the transaction with a given prefix.
func (t *stdTransaction) Iterator(prefix string) (Iterator, error) {
	keys := make([]string, 0)
	for k := range t.data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return &StdIterator{0, keys}, nil
}

// Commit ends a transaction and persists the changes.
func (t *stdTransaction) Commit() error {
	if err := t.store.commit(t.data); err != nil {
		return err
	}
	t.store = nil
	return nil
}

// Rollback aborts a transaction.
func (t *stdTransaction) Rollback() {
	if t.store != nil {
		t.store.txmut.Unlock()
	}
}

// StdIterator is the standard Iterator implementation.
type StdIterator struct {
	idx  int
	keys []string
}

// GetNext returns the next element of the iterator.
func (i *StdIterator) GetNext() (string, error) {
	if i.idx >= len(i.keys) {
		return "", fmt.Errorf("index out of range [%d] with length %d", i.idx, len(i.keys))
	}
	val := i.keys[i.idx]
	i.idx++
	return val, nil
}

// HasNext returns true if there are elements left to get with GetNext().
func (i *StdIterator) HasNext() bool {
	return i.idx < len(i.keys)
}
