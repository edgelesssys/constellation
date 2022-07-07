package kubeadm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/clock"
	testclock "k8s.io/utils/clock/testing"
)

func TestKeyManager(t *testing.T) {
	testCases := map[string]struct {
		clock       clock.Clock
		ttl         time.Time
		key         string
		shouldReuse bool
	}{
		"no key exists": {
			clock: testclock.NewFakeClock(time.Time{}),
		},
		"key exists and is valid": {
			clock:       testclock.NewFakeClock(time.Time{}),
			ttl:         time.Time{}.Add(time.Hour),
			key:         "key",
			shouldReuse: true,
		},
		"key has expired": {
			clock: testclock.NewFakeClock(time.Time{}.Add(time.Hour)),
			ttl:   time.Time{},
			key:   "key",
		},
		"key expires in the next 30 seconds": {
			clock:       testclock.NewFakeClock(time.Time{}),
			ttl:         time.Time{}.Add(30 * time.Second),
			key:         "key",
			shouldReuse: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			km := &keyManager{
				expirationDate: tc.ttl,
				key:            tc.key,
				clock:          tc.clock,
			}

			key, err := km.getCertificatetKey()
			assert.NoError(err)
			assert.True(km.expirationDate.After(tc.clock.Now().Add(2 * time.Minute)))

			if tc.shouldReuse {
				assert.Equal(tc.key, key)
			} else {
				assert.Equal(km.key, key)
				assert.NotEqual(tc.key, key)
			}
		})
	}
}
