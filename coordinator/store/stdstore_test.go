package store

import "testing"

func TestStdStore(t *testing.T) {
	testStore(t, func() (Store, error) {
		return NewStdStore(), nil
	})
}
