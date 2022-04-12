package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

var newStore func() (Store, error)

func testStore(t *testing.T, storeFactory func() (Store, error)) {
	newStore = storeFactory

	t.Run("NewStore", testNewStore)
	t.Run("NewStoreIsEmpty", testNewStoreIsEmpty)
	t.Run("NewStoreClearsStore", testNewStoreClearsStore)
	t.Run("Put", testPut)
	t.Run("PutTwice", testPutTwice)
	t.Run("Get", testGet)
	t.Run("GetNonExisting", testGetNonExisting)
	t.Run("Delete", testDelete)
	t.Run("DeleteNonExisting", testDeleteNonExisting)
	t.Run("Iterator", testIterator)
	t.Run("IteratorSingleKey", testIteratorSingleKey)
	t.Run("IteratorNoValues", testIteratorNoValues)
	t.Run("IteratorRace", testIteratorRace)
	t.Run("Transaction", testTransaction)
	t.Run("TransactionInternalChangesVisible", testTransactionInternalChangesVisible)
	t.Run("TransactionInternalChangesNotVisibleOutside", testTransactionInternalChangesNotVisibleOutside)
	t.Run("TransactionNoop", testTransactionNoop)
	t.Run("TransactionDeleteThenPut", testTransactionDeleteThenPut)
	t.Run("TransactionDelete", testTransactionDelete)
	t.Run("TransactionIterator", testTransactionIterator)
	t.Run("TransactionIterateNotSeeDeleted", testTransactionIterateNotSeeDeleted)
	t.Run("TransactionGetAfterCommit", testTransactionGetAfterCommit)
	t.Run("TransactionPutAfterCommit", testTransactionPutAfterCommit)
	t.Run("TransactionDeleteAfterCommit", testTransactionDeleteAfterCommit)
	t.Run("RollbackPut", testRollbackPut)
	t.Run("RollbackDelete", testRollbackDelete)
	t.Run("Concurrency", testConcurrency)
	t.Run("StoreByValue", testStoreByValue)
	t.Run("IndependentTest", testIndependentTest)
	t.Run("IndependentTestReader", testIndependentTestReader)
}

func testNewStore(t *testing.T) {
	require := require.New(t)

	_, err := newStore()

	require.NoError(err)
}

func testNewStoreIsEmpty(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	iter, err := store.Iterator("")
	require.NoError(err)
	require.False(iter.HasNext())
}

func testNewStoreClearsStore(t *testing.T) {
	require := require.New(t)
	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("key", []byte("value")))

	store, err = newStore()
	require.NoError(err)

	iter, err := store.Iterator("")
	require.NoError(err)
	require.False(iter.HasNext())
}

func testPut(t *testing.T) {
	require := require.New(t)
	store, err := newStore()
	require.NoError(err)

	err = store.Put("key", []byte("value"))

	require.NoError(err)
}

func testPutTwice(t *testing.T) {
	require := require.New(t)
	store, err := newStore()
	require.NoError(err)
	err = store.Put("key", []byte("value"))
	require.NoError(err)

	err = store.Put("key", []byte("newValue"))
	require.NoError(err)

	fetchedValue, err := store.Get("key")
	require.NoError(err)
	require.Equal([]byte("newValue"), fetchedValue)
}

func testGet(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	err = store.Put("key", []byte("value"))
	require.NoError(err)

	fetchedValue, err := store.Get("key")

	require.NoError(err)
	require.Equal([]byte("value"), fetchedValue)
}

func testGetNonExisting(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	_, err = store.Get("key")

	var unsetError *ValueUnsetError
	require.ErrorAs(err, &unsetError)
}

func testDelete(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	err = store.Put("key", []byte("value"))
	require.NoError(err)

	err = store.Delete("key")
	require.NoError(err)

	_, err = store.Get("key")
	var unsetError *ValueUnsetError
	require.ErrorAs(err, &unsetError)
}

// Deleting a non-existing key is fine, and should not result in error.
func testDeleteNonExisting(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	err = store.Delete("key")

	require.NoError(err)
}

func testIterator(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("iterate:1", []byte("one")))
	require.NoError(store.Put("iterate:2", []byte("two")))
	require.NoError(store.Put("iterate:3", []byte("three")))

	iter, err := store.Iterator("iterate")
	require.NoError(err)
	idx := 0
	for iter.HasNext() {
		idx++
		val, err := iter.GetNext()
		assert.NoError(err)
		assert.Contains(val, "iterate:")
	}
	assert.Equal(3, idx)
}

func testIteratorSingleKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("key", []byte("value")))

	iter, err := store.Iterator("key")
	require.NoError(err)
	key, err := iter.GetNext()
	require.NoError(err)
	assert.Equal("key", key)
	require.False(iter.HasNext())
}

func testIteratorNoValues(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	iter, err := store.Iterator("iterate")
	require.NoError(err)
	require.False(iter.HasNext())

	_, err = iter.GetNext()
	var noElementsLeftError *NoElementsLeftError
	require.ErrorAs(err, &noElementsLeftError)
}

// Test the race condition in the stdStore.
// This must be specified in the test through [-race].
func testIteratorRace(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Put("key", []byte("value")))
	go func() {
		_, err = store.Iterator("key")
		require.NoError(err)
	}()
	require.NoError(tx.Commit())
}

func testTransaction(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("toBeDeleted", []byte{0x00, 0x00}))

	tx1, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx1.Put("newKey", []byte{0x11, 0x11}))
	require.NoError(tx1.Delete("toBeDeleted"))
	require.NoError(tx1.Commit())

	result, err := store.Get("newKey")
	require.NoError(err)
	assert.Equal(result, []byte{0x11, 0x11})
	_, err = store.Get("toBeDeleted")
	require.Error(err)
}

func testTransactionInternalChangesVisible(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Put(("key"), []byte("value")))

	fetchedValue, err := tx.Get("key")
	require.NoError(err)
	require.Equal([]byte("value"), fetchedValue)
}

func testTransactionInternalChangesNotVisibleOutside(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Put(("key"), []byte("value")))

	_, err = store.Get("key")
	var valueUnsetError *ValueUnsetError
	require.ErrorAs(err, &valueUnsetError)
}

func testTransactionNoop(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NotNil(tx)

	err = tx.Commit()
	require.NoError(err)
}

func testTransactionDeleteThenPut(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("key", []byte{0x00, 0x00}))

	tx1, err := store.BeginTransaction()
	require.NoError(err)

	require.NoError(tx1.Put("key", []byte{0x00, 0x11}))
	assert.NoError(tx1.Delete("key"))
	require.NoError(tx1.Put("key", []byte{0x7, 0x8}))
	assert.NoError(tx1.Commit())

	result, err := store.Get("key")
	assert.NoError(err)
	assert.Equal([]byte{0x7, 0x8}, result)
}

func testTransactionDelete(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("key", []byte{0x00, 0x00}))

	tx1, err := store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx1.Delete("key"))
	_, err = tx1.Get("key")
	assert.Error(err)
	assert.NoError(tx1.Commit())
}

func testTransactionIterator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("key", []byte("value")))

	tx, err := store.BeginTransaction()
	require.NoError(err)
	iter, err := tx.Iterator("key")
	require.NoError(err)
	key, err := iter.GetNext()
	require.NoError(err)
	assert.Equal("key", key)
	require.NoError(tx.Commit())
	require.False(iter.HasNext())
}

func testTransactionIterateNotSeeDeleted(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("key:1", []byte("value")))
	require.NoError(store.Put("key:2", []byte("otherValue")))

	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Delete("key:1"))
	iter, err := tx.Iterator("key")
	require.NoError(err)
	key, err := iter.GetNext()
	require.NoError(err)
	assert.Equal("key:2", key)
	require.NoError(tx.Commit())
	require.False(iter.HasNext())
}

func testTransactionGetAfterCommit(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Put("key", []byte("value")))
	require.NoError(tx.Commit())

	_, err = tx.Get("key")
	var alreadyCommittedError *TransactionAlreadyCommittedError
	require.ErrorAs(err, &alreadyCommittedError)
}

func testTransactionPutAfterCommit(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Put("key", []byte("value")))
	require.NoError(tx.Commit())

	err = tx.Put("key", []byte("newValue"))
	var alreadyCommittedError *TransactionAlreadyCommittedError
	require.ErrorAs(err, &alreadyCommittedError)
}

func testTransactionDeleteAfterCommit(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Put("key", []byte("value")))
	require.NoError(tx.Commit())

	err = tx.Delete("key")
	var alreadyCommittedError *TransactionAlreadyCommittedError
	require.ErrorAs(err, &alreadyCommittedError)
}

func testRollbackPut(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	tx, err := store.BeginTransaction()
	require.NoError(err)
	err = tx.Put("key", []byte("value"))
	require.NoError(err)
	tx.Rollback()

	_, err = store.Get("key")
	var valueUnsetError *ValueUnsetError
	assert.ErrorAs(err, &valueUnsetError)
}

func testRollbackDelete(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("key", []byte("value")))

	tx, err := store.BeginTransaction()
	require.NoError(err)
	err = tx.Delete("key")
	require.NoError(err)
	tx.Rollback()

	fetchedValue, err := store.Get("key")
	require.NoError(err)
	assert.Equal([]byte("value"), fetchedValue)
}

// Test explicitly the storeLocking mechanism.
// This could fail for non-blocking transactions.
func testConcurrency(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	tx1, err := store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx1.Put("key", []byte("one")))

	go func() {
		time.Sleep(200 * time.Millisecond)
		require.NoError(tx1.Commit())
	}()
	tx2, err := store.BeginTransaction()
	require.NoError(err)
	result, err := tx2.Get("key")
	require.NoError(err)
	assert.Equal(result, []byte("one"))
	assert.NoError(tx2.Put("key", []byte("two")))
	assert.NoError(tx2.Commit())
	result, err = store.Get("key")
	require.NoError(err)
	assert.Equal(result, []byte("two"))
}

func testStoreByValue(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	testValue := []byte{0xF1, 0xFF}
	require.NoError(store.Put("StoreByValue", testValue))

	testValue[0] = 0x00
	storeValue, err := store.Get("StoreByValue")

	require.NoError(err)
	assert.NotEqual(storeValue[0], testValue[0])
}

func testIndependentTest(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	err = store.Put("uniqueTestKey253", []byte("value"))
	require.NoError(err)

	value, err := store.Get("uniqueTestKey253")
	require.NoError(err)
	assert.Equal([]byte("value"), value)
}

func testIndependentTestReader(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	// This test should not see & depend on the key `testIndependentTest` sets.
	_, err = store.Get("uniqueTestKey253")
	var unsetErr *ValueUnsetError
	assert.ErrorAs(err, &unsetErr)
}
