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
	t.Run("basic", testBasic)
	t.Run("iterator", testIterator)
	t.Run("rollback", testRollback)
	t.Run("testConcurrency", testConcurrency)
	t.Run("testTransaction", testTransaction)
	t.Run("testTransactionGet", testTransactionGet)
	t.Run("testTransactionDelete_1", testTransactionDelete1)
	t.Run("testTransactionDelete_2", testTransactionDelete2)
	t.Run("testIteratorKey", testIteratorKey)
	t.Run("testTransactionIterator", testTransactionIterator)
	t.Run("testTransactionDeleteIterator", testTransactionDeleteIterator)
	t.Run("testEmptyIterator", testEmptyIterator)
	t.Run("testIteratorRace", testIteratorRace)
	t.Run("testTransactionIteratorPutPrefix", testTransactionIteratorPutPrefix)
	t.Run("testStoreByValue", testStoreByValue)
}

func testBasic(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	testData1 := []byte("test data")
	testData2 := []byte("more test data")

	// request unset value
	_, err = store.Get("test:input")
	assert.Error(err)

	// test Put method
	tx, err := store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx.Put("test:input", testData1))
	assert.NoError(tx.Put("another:input", testData2))
	assert.NoError(tx.Commit())

	// make sure values have been set
	val, err := store.Get("test:input")
	require.NoError(err)
	assert.Equal(testData1, val)
	val, err = store.Get("another:input")
	require.NoError(err)
	assert.Equal(testData2, val)

	_, err = store.Get("invalid:key")
	assert.Error(err)
	var unsetErr *StoreValueUnsetError
	assert.ErrorAs(err, &unsetErr)
	assert.NoError(store.Delete("test:input"))
	assert.NoError(store.Delete("another:input"))

	_, err = store.Get("test:input")
	assert.Error(err)
}

func testIterator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	require.NoError(store.Put("test:1", []byte{0x00, 0x11}))
	require.NoError(store.Put("test:2", []byte{0x00, 0x11}))
	require.NoError(store.Put("test:3", []byte{0x00, 0x11}))
	require.NoError(store.Put("value:1", []byte{0x00, 0x11}))
	require.NoError(store.Put("something:1", []byte{0x00}))

	iter, err := store.Iterator("test")
	require.NoError(err)
	idx := 0
	for iter.HasNext() {
		idx++
		val, err := iter.GetNext()
		assert.NoError(err)
		assert.Contains(val, "test:")
	}
	assert.EqualValues(3, idx)

	iter, err = store.Iterator("value")
	require.NoError(err)
	idx = 0
	for iter.HasNext() {
		idx++
		val, err := iter.GetNext()
		assert.NoError(err)
		assert.Contains(val, "value:")
	}
	assert.EqualValues(1, idx)

	iter, err = store.Iterator("")
	require.NoError(err)
	idx = 0
	for iter.HasNext() {
		idx++
		_, err = iter.GetNext()
		assert.NoError(err)
	}
	assert.EqualValues(5, idx)

	iter, err = store.Iterator("empty")
	require.NoError(err)
	assert.False(iter.HasNext())

	_, err = iter.GetNext()
	assert.Error(err)

	require.NoError(store.Delete("test:2"))
	require.NoError(store.Delete("test:1"))
	require.NoError(store.Delete("test:3"))
	require.NoError(store.Delete("value:1"))
	require.NoError(store.Delete("something:1"))
}

func testRollback(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)

	testData1 := []byte("test data")
	testData2 := []byte("more test data")
	testData3 := []byte("and even more data")

	// save data to store and seal
	tx, err := store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx.Put("test:input", testData1))
	assert.NoError(tx.Commit())

	// save more data to store
	tx, err = store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx.Put("another:input", testData2))

	// rollback and verify only testData1 exists
	tx.Rollback()
	val, err := store.Get("test:input")
	require.NoError(err)
	assert.Equal(testData1, val)
	_, err = store.Get("another:input")
	assert.Error(err)

	// save something new
	tx, err = store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx.Put("last:input", testData3))
	assert.NoError(tx.Commit())

	// verify values
	val, err = store.Get("test:input")
	require.NoError(err)
	assert.Equal(testData1, val)
	val, err = store.Get("last:input")
	require.NoError(err)
	assert.Equal(testData3, val)
	_, err = store.Get("another:input")
	assert.Error(err)

	assert.NoError(store.Delete("test:input"))
	assert.NoError(store.Delete("last:input"))
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
	assert.NoError(tx1.Put("concurrent", []byte{0x00, 0x00}))

	go func() {
		time.Sleep(200 * time.Millisecond)
		require.NoError(tx1.Commit())
	}()
	tx2, err := store.BeginTransaction()
	require.NoError(err)
	result, err := tx2.Get("concurrent")
	require.NoError(err)
	assert.Equal(result, []byte{0x00, 0x00})
	assert.NoError(tx2.Put("concurrent", []byte{0x11, 0x11}))
	assert.NoError(tx2.Commit())
	result, err = store.Get("concurrent")
	require.NoError(err)
	assert.Equal(result, []byte{0x11, 0x11})

	assert.NoError(store.Delete("concurrent"))
}

func testTransaction(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	assert.NoError(store.Put("transactionDelete", []byte{0x00, 0x00}))

	tx1, err := store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx1.Put("transaction", []byte{0x11, 0x11}))
	assert.NoError(tx1.Delete("transactionDelete"))
	assert.NoError(tx1.Commit())

	result, err := store.Get("transaction")
	require.NoError(err)
	assert.Equal(result, []byte{0x11, 0x11})
	_, err = store.Get("transactionDelete")
	assert.Error(err)
}

func testTransactionGet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("transactionGet", []byte{0x00, 0x00}))

	tx1, err := store.BeginTransaction()
	require.NoError(err)
	result, err := tx1.Get("transactionGet")
	require.NoError(err)
	assert.Equal(result, []byte{0x00, 0x00})
	assert.NoError(tx1.Put("transactionGet", []byte{0x11, 0x11}))
	result, err = tx1.Get("transactionGet")
	require.NoError(err)
	assert.Equal(result, []byte{0x11, 0x11})
	assert.NoError(tx1.Commit())

	result, err = store.Get("transactionGet")
	require.Equal(result, []byte{0x11, 0x11})
	assert.NoError(err)
}

func testTransactionDelete1(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("transactionDelete", []byte{0x00, 0x00}))

	tx1, err := store.BeginTransaction()
	require.NoError(err)

	require.NoError(tx1.Put("transactionDelete", []byte{0x00, 0x11}))
	assert.NoError(tx1.Delete("transactionDelete"))
	require.NoError(tx1.Put("transactionDelete", []byte{0x7, 0x8}))
	assert.NoError(tx1.Commit())

	result, err := store.Get("transactionDelete")
	assert.NoError(err)
	assert.Equal([]byte{0x7, 0x8}, result)
}

func testTransactionDelete2(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("transactionDelete", []byte{0x00, 0x00}))

	tx1, err := store.BeginTransaction()
	require.NoError(err)
	assert.NoError(tx1.Delete("transactionDelete"))
	_, err = tx1.Get("transactionDelete")
	assert.Error(err)
	assert.NoError(tx1.Commit())
}

func testIteratorKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("testIteratorKey", []byte{0xF1, 0xFF}))
	iter, err := store.Iterator("testIteratorKey")
	require.NoError(err)
	key, err := iter.GetNext()
	require.NoError(err)
	assert.Equal("testIteratorKey", key)
}

func testTransactionIterator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("testTransactionIterator", []byte{0xF1, 0xFF}))
	txdata, err := store.BeginTransaction()
	require.NoError(err)
	iter, err := txdata.Iterator("testTransactionIterator")
	require.NoError(err)
	key, err := iter.GetNext()
	require.NoError(err)
	assert.Equal("testTransactionIterator", key)
	require.NoError(txdata.Commit())
}

func testTransactionDeleteIterator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	require.NoError(store.Put("testTransactionDeleteIterator", []byte{0xF1, 0xFF}))
	require.NoError(store.Put("testTransactionDeleteIterator15", []byte{0x5, 0xFF}))
	txdata, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(txdata.Delete("testTransactionDeleteIterator"))
	iter, err := txdata.Iterator("testTransactionDeleteIterator")
	require.NoError(err)
	key, err := iter.GetNext()
	require.NoError(err)
	assert.Equal("testTransactionDeleteIterator15", key)
	require.NoError(txdata.Commit())
}

func testEmptyIterator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	txdata, err := store.BeginTransaction()
	require.NoError(err)
	iter, err := txdata.Iterator("ThisKeyTestsAnEmptyIterator")
	require.NoError(err)
	assert.False(iter.HasNext())
	_, err = iter.GetNext()
	assert.Error(err)
	require.NoError(txdata.Commit())
}

// Test the race condition in the stdStore.
// This must be specified in the test through [-race].
func testIteratorRace(t *testing.T) {
	require := require.New(t)

	store, err := newStore()
	require.NoError(err)
	tx, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(tx.Put("testIteratorRace", []byte{1, 2, 3, 4, 5}))
	go func() {
		_, err = store.Iterator("testIteratorRace")
		require.NoError(err)
	}()
	require.NoError(tx.Commit())
}

func testTransactionIteratorPutPrefix(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	expectedKey := "testTransactionIteratorPutPrefix"
	store, err := newStore()
	require.NoError(err)
	txdata, err := store.BeginTransaction()
	require.NoError(err)
	require.NoError(txdata.Put("IteratorNotInIterator", []byte{0xF2, 0xFF}))
	require.NoError(txdata.Put(expectedKey, []byte{0xF1, 0xFF}))
	iter, err := txdata.Iterator(expectedKey)
	require.NoError(err)
	require.True(iter.HasNext())
	key, err := iter.GetNext()
	require.NoError(err)
	assert.Equal(expectedKey, key)
	assert.False(iter.HasNext())
	assert.NoError(txdata.Commit())
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
