package qemu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDs(t *testing.T) {
	assert := assert.New(t)
	testState := testInstances()
	wantIDs := []string{"id-9", "id-10", "id-11", "id-12"}
	assert.ElementsMatch(wantIDs, testState.IDs())
}

func TestPublicIPs(t *testing.T) {
	assert := assert.New(t)
	testState := testInstances()
	wantIPs := []string{"192.0.2.1", "192.0.2.3", "192.0.2.5", "192.0.2.7"}
	assert.ElementsMatch(wantIPs, testState.PublicIPs())
}

func TestPrivateIPs(t *testing.T) {
	assert := assert.New(t)
	testState := testInstances()
	wantIPs := []string{"192.0.2.2", "192.0.2.4", "192.0.2.6", "192.0.2.8"}
	assert.ElementsMatch(wantIPs, testState.PrivateIPs())
}

func TestGetOne(t *testing.T) {
	assert := assert.New(t)
	testState := testInstances()
	id, instance, err := testState.GetOne()
	assert.NoError(err)
	assert.Contains(testState, id)
	assert.Equal(testState[id], instance)
}

func TestGetOthers(t *testing.T) {
	assert := assert.New(t)
	testCases := testInstances().IDs()

	for _, id := range testCases {
		others := testInstances().GetOthers(id)
		assert.NotContains(others, id)
		wantInstances := testInstances()
		delete(wantInstances, id)
		assert.ElementsMatch(others.IDs(), wantInstances.IDs())
	}
}

func testInstances() Instances {
	return Instances{
		"id-9": {
			PublicIP:  "192.0.2.1",
			PrivateIP: "192.0.2.2",
		},
		"id-10": {
			PublicIP:  "192.0.2.3",
			PrivateIP: "192.0.2.4",
		},
		"id-11": {
			PublicIP:  "192.0.2.5",
			PrivateIP: "192.0.2.6",
		},
		"id-12": {
			PublicIP:  "192.0.2.7",
			PrivateIP: "192.0.2.8",
		},
	}
}
