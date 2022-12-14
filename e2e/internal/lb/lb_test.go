//go:build e2elb

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package test

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/e2e/internal/kubectl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	namespaceName = "lb-test"
	serviceName   = "whoami"
	initialPort   = int32(8080)
	newPort       = int32(8044)
	numRequests   = 256
	numPods       = 3
	timeout       = time.Minute * 15
	interval      = time.Second * 5
)

func TestLoadBalancer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	k, err := kubectl.New()
	require.NoError(err)

	t.Log("Waiting for external IP to be registered")
	svc := testEventuallyExternalIPAvailable(t, k)
	loadBalancerIP := getIPOrHostname(t, svc)
	loadBalancerPort := svc.Spec.Ports[0].Port
	require.Equal(initialPort, loadBalancerPort)
	url := buildURL(t, loadBalancerIP, loadBalancerPort)

	t.Log("Checking service can be reached through LB")
	testEventuallyStatusOK(t, url)

	t.Log("Check that all pods receive traffic")
	var allHostnames []string
	for i := 0; i < numRequests; i++ {
		allHostnames = testEndpointAvailable(t, url, allHostnames)
	}
	assert.True(hasNUniqueStrings(allHostnames, numPods))
	allHostnames = allHostnames[:0]

	t.Log("Change port of service to 8044")
	svc.Spec.Ports[0].Port = newPort
	svc, err = k.CoreV1().Services(namespaceName).Update(context.Background(), svc, v1.UpdateOptions{})
	require.NoError(err)
	assert.Equal(newPort, svc.Spec.Ports[0].Port)

	t.Log("Wait for changed port to be available")
	newURL := buildURL(t, loadBalancerIP, newPort)
	testEventuallyStatusOK(t, newURL)

	t.Log("Check again that all pods receive traffic")
	for i := 0; i < numRequests; i++ {
		allHostnames = testEndpointAvailable(t, newURL, allHostnames)
	}
	assert.True(hasNUniqueStrings(allHostnames, numPods))
}

func getIPOrHostname(t *testing.T, svc *coreV1.Service) string {
	t.Helper()
	if ip := svc.Status.LoadBalancer.Ingress[0].IP; ip != "" {
		return ip
	}
	return svc.Status.LoadBalancer.Ingress[0].Hostname
}

func hasNUniqueStrings(elements []string, n int) bool {
	m := make(map[string]bool)
	for i := range elements {
		m[elements[i]] = true
	}

	numKeys := 0
	for range m {
		numKeys++
	}
	return numKeys == n
}

func buildURL(t *testing.T, ip string, port int32) string {
	t.Helper()
	return fmt.Sprintf("http://%s:%d", ip, port)
}

// testEventuallyStatusOK tests that the URL response with StatusOK within 5min.
func testEventuallyStatusOK(t *testing.T, url string) {
	assert := assert.New(t)
	require := require.New(t)

	assert.Eventually(func() bool {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
		require.NoError(err)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Log("Request failed: ", err.Error())
			return false
		}
		defer resp.Body.Close()

		statusOK := resp.StatusCode == http.StatusOK
		if !statusOK {
			t.Log("Status not OK: ", resp.StatusCode)
			return false
		}

		t.Log("Status OK")
		return true
	}, timeout, interval)
}

// testEventuallyExternalIPAvailable uses k to query if the whoami service is eventually available.
// Once the service is available the Service is returned.
func testEventuallyExternalIPAvailable(t *testing.T, k *kubernetes.Clientset) *coreV1.Service {
	var svc *coreV1.Service

	require.Eventually(t, func() bool {
		var err error
		svc, err = k.CoreV1().Services(namespaceName).Get(context.Background(), serviceName, v1.GetOptions{})
		if err != nil {
			t.Log("Getting service failed: ", err.Error())
			return false
		}
		t.Log("Successfully fetched service: ", svc.String())

		ingressAvailable := len(svc.Status.LoadBalancer.Ingress) > 0
		if !ingressAvailable {
			t.Log("Ingress not yet available")
			return false
		}

		t.Log("Ingress available")
		return true
	}, timeout, interval)

	return svc
}

// testEndpointAvailable GETs the provided URL. It expects a payload from
// traefik/whoami service and checks that the first body line is of form
// Hostname: <pod-name>
// If this works the <pod-name> value is appended to allHostnames slice and
// new allHostnames is returned.
func testEndpointAvailable(t *testing.T, url string, allHostnames []string) []string {
	assert := assert.New(t)
	require := require.New(t)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	require.NoError(err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(err)
	defer resp.Body.Close()
	assert.Equal(http.StatusOK, resp.StatusCode)
	// Force close of connections so that we see different backends
	http.DefaultClient.CloseIdleConnections()

	firstLine, err := bufio.NewReader(resp.Body).ReadString('\n')
	require.NoError(err)
	parts := strings.Split(firstLine, ": ")
	hostnameKey := parts[0]
	hostnameValue := parts[1]

	assert.Equal("Hostname", hostnameKey)
	require.NotEmpty(hostnameValue)

	return append(allHostnames, hostnameValue)
}
