package kubeadm

import (
	"testing"
	"time"

	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/stretchr/testify/assert"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/utils/clock"
	testclock "k8s.io/utils/clock/testing"
)

func TestKeyManager(t *testing.T) {
	testCases := map[string]struct {
		clock       clock.Clock
		client      clientset.Interface
		ttl         time.Time
		key         string
		shouldReuse bool
		wantErr     bool
	}{
		"no key exists": {
			clock:  testclock.NewFakeClock(time.Time{}),
			client: fake.NewSimpleClientset(),
		},
		"key exists and is valid": {
			clock:       testclock.NewFakeClock(time.Time{}),
			client:      fake.NewSimpleClientset(),
			ttl:         time.Time{}.Add(time.Hour),
			key:         "key",
			shouldReuse: true,
		},
		"key has expired": {
			clock:  testclock.NewFakeClock(time.Time{}.Add(time.Hour)),
			client: fake.NewSimpleClientset(),
			ttl:    time.Time{},
			key:    "key",
		},
		"key expires in the next 30 seconds": {
			clock:       testclock.NewFakeClock(time.Time{}),
			client:      fake.NewSimpleClientset(),
			ttl:         time.Time{}.Add(30 * time.Second),
			key:         "key",
			shouldReuse: true,
		},
		"uploading certs fails": {
			clock: testclock.NewFakeClock(time.Time{}),
			client: &failingClient{
				fake.NewSimpleClientset(),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			km := &keyManager{
				expirationDate: tc.ttl,
				key:            tc.key,
				clock:          tc.clock,
				log:            logger.NewTest(t),
				client:         fake.NewSimpleClientset(),
			}

			key, err := km.getCertificatetKey()
			if tc.wantErr {
				assert.Error(err)
				return
			}
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

type failingClient struct {
	*fake.Clientset
}

func (f *failingClient) CoreV1() corev1.CoreV1Interface {
	return &failingCoreV1{
		&fakecorev1.FakeCoreV1{Fake: &f.Clientset.Fake},
	}
}
