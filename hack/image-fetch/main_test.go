package main

import "testing"

func TestNop(t *testing.T) {
	t.Skip("This is a nop-test to catch build-time errors in this package.")
}
