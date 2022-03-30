package store

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TODO: Generate certificates for the coordinator.
const (
	peerCertFilepath = "/etc/kubernetes/pki/etcd/peer.crt"
	keyFilepath      = "/etc/kubernetes/pki/etcd/peer.key"
	caCertFilepath   = "/etc/kubernetes/pki/etcd/server.crt"
	etcdPrefix       = "constellationRegion"
	dialTimeout      = 10 * time.Second
)

type EtcdStore struct {
	client *clientv3.Client
}

func NewEtcdStore(endpoint string, forceTls bool, logger *zap.Logger) (*EtcdStore, error) {
	var tlsConfig *tls.Config
	if forceTls {
		caCert, err := os.ReadFile(caCertFilepath)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		cert, err := tls.LoadX509KeyPair(peerCertFilepath, keyFilepath)
		if err != nil {
			return nil, err
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			MinVersion:   tls.VersionTLS12,
		}
	}

	// Blocks until connection is up
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{endpoint},
		TLS:         tlsConfig,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
		Logger:      logger,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdStore{client: cli}, nil
}

// Get retrieves a value from EtcdStore by Type and Name.
func (s *EtcdStore) Get(request string) ([]byte, error) {
	values, err := s.client.Get(context.TODO(), etcdPrefix+request)
	if err != nil {
		return nil, err
	}
	if values.Count == 0 {
		return nil, &ValueUnsetError{requestedValue: request}
	}
	if values.Count == 1 {
		return values.Kvs[0].Value, nil
	}
	return nil, fmt.Errorf("got multiple entries for key [%s] in etcd", request)
}

// Put saves a value in EtcdStore by Type and Name.
func (s *EtcdStore) Put(request string, requestData []byte) error {
	_, err := s.client.Put(context.TODO(), etcdPrefix+request, string(requestData))
	return err
}

// Iterator returns an Iterator for a given prefix.
func (s *EtcdStore) Iterator(prefix string) (Iterator, error) {
	resp, err := s.client.Get(context.TODO(), etcdPrefix+prefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		key := strings.TrimPrefix(string(kv.Key), etcdPrefix)
		keys = append(keys, key)
	}
	output := &EtcdIterator{keys: keys}
	return output, nil
}

// TODO: Implement this function, currently this function is never called.
func (s *EtcdStore) Transfer(store Store) error {
	panic("etcd store Transfer() function not implemented, should never be called")
}

// Delete deletes the store entry with the given key.
func (s *EtcdStore) Delete(key string) error {
	_, err := s.client.Delete(context.TODO(), etcdPrefix+key)
	return err
}

func (s *EtcdStore) BeginTransaction() (Transaction, error) {
	sess, err := concurrency.NewSession(s.client)
	if err != nil {
		return nil, err
	}
	mut := concurrency.NewLocker(sess, etcdPrefix)
	mut.Lock()

	return &EtcdTransaction{
		store:              s,
		dataInsert:         map[string][]byte{},
		dataDelete:         map[string]struct{}{},
		ongoingTransaction: true,
		mut:                mut,
		session:            sess,
	}, nil
}

// Close closes the etcd client.
func (s *EtcdStore) Close() error {
	return s.client.Close()
}

type EtcdTransaction struct {
	store              *EtcdStore
	dataInsert         map[string][]byte
	dataDelete         map[string]struct{}
	ongoingTransaction bool
	session            *concurrency.Session
	mut                sync.Locker
}

func (t *EtcdTransaction) Get(request string) ([]byte, error) {
	if !t.ongoingTransaction {
		return nil, fmt.Errorf("EtcdTransaction Pointer is nil, but Get function is called")
	}
	if value, ok := t.dataInsert[request]; ok {
		return value, nil
	}
	if _, ok := t.dataDelete[request]; ok {
		return nil, &ValueUnsetError{requestedValue: request}
	}
	return t.store.Get(request)
}

// Put saves a value.
func (t *EtcdTransaction) Put(request string, requestData []byte) error {
	if !t.ongoingTransaction {
		return fmt.Errorf("EtcdTransaction Pointer is nil, but Put function is called")
	}
	t.dataInsert[request] = requestData
	return nil
}

// Delete deletes the key if it exists. Only errors if there is no ongoing Transaction.
func (t *EtcdTransaction) Delete(key string) error {
	if !t.ongoingTransaction {
		return fmt.Errorf("EtcdTransaction Pointer is nil, but Delete function is called")
	}
	delete(t.dataInsert, key)
	t.dataDelete[key] = struct{}{}
	return nil
}

// Iterator returns an iterator for all keys in the transaction with a given prefix.
func (t *EtcdTransaction) Iterator(prefix string) (Iterator, error) {
	resp, err := t.store.client.Get(context.TODO(), etcdPrefix+prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, v := range resp.Kvs {
		key := strings.TrimPrefix(string(v.Key), etcdPrefix)
		if _, ok := t.dataDelete[key]; !ok {
			keys = append(keys, key)
		}
	}
	for k := range t.dataInsert {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	output := &EtcdIterator{idx: 0, keys: keys}
	return output, err
}

// Commit ends a transaction and persists the changes.
func (t *EtcdTransaction) Commit() error {
	if !t.ongoingTransaction {
		return fmt.Errorf("no ongoing transaction")
	}

	ops := make([]clientv3.Op, 0, len(t.dataInsert)+len(t.dataDelete))
	// add all transactions into one object; for future
	// implementations, we can also atomically delete elements
	// however, it's not compatible with stdstore atm
	for k, v := range t.dataInsert {
		ops = append(ops, clientv3.OpPut(etcdPrefix+k, string(v)))
	}
	for k := range t.dataDelete {
		// Each key is only allowed to occur once per transaction
		if _, ok := t.dataInsert[k]; !ok {
			ops = append(ops, clientv3.OpDelete(etcdPrefix+k))
		}
	}
	// transaction, so either everything gets applied or nothing
	_, err := t.store.client.Txn(context.TODO()).Then(ops...).Commit()
	if err != nil {
		return err
	}
	t.session.Close()
	t.mut.Unlock()
	t.ongoingTransaction = false
	return nil
}

// Rollback aborts a transaction.
func (t *EtcdTransaction) Rollback() {
	if t.ongoingTransaction {
		t.session.Close()
		t.mut.Unlock()
	}
	t.ongoingTransaction = false
}

type EtcdIterator struct {
	idx  int
	keys []string
}

// GetNext gets the next element.
func (i *EtcdIterator) GetNext() (string, error) {
	if i.idx >= len(i.keys) {
		return "", fmt.Errorf("index out of range [%d] with length %d", i.idx, len(i.keys))
	}
	key := i.keys[i.idx]
	i.idx++
	return key, nil
}

// HasNext returns true if there are elements left to get with GetNext().
func (i *EtcdIterator) HasNext() bool {
	return i.idx < len(i.keys)
}

// EtcdStoreFactory is a factory to create EtcdStores.
type EtcdStoreFactory struct {
	Endpoint string
	ForceTLS bool
	Logger   *zap.Logger
}

// New creates a new EtcdStore.
func (f *EtcdStoreFactory) New() (Store, error) {
	return NewEtcdStore(f.Endpoint, f.ForceTLS, f.Logger)
}
