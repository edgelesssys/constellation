/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package fetcher_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := ":8081"
	go startDummyConfigAPIServer(ctx, addr)
	configapi.BaseURL = fmt.Sprintf("http://localhost%s", addr)

	fetcher := fetcher.NewConfigAPIFetcher()
	res, err := fetcher.FetchLatestAzureSEVSNPVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, configapi.AttestationVersion{
		Value:    2,
		IsLatest: false,
	}, res.Bootloader)
}

func startDummyConfigAPIServer(ctx context.Context, addr string) {
	http.HandleFunc("/constellation/v1/attestation/azure-sev-snp/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]string{"2021-01-01-01-01.json"}); err != nil {
			panic(err)
		}
	})
	http.HandleFunc("/constellation/v1/attestation/azure-sev-snp/2021-01-01-01-01.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]int{"bootloader": 2, "tee": 0, "snp": 6, "microcode": 93}); err != nil {
			panic(err)
		}
	})

	server := &http.Server{
		Addr: addr,
	}
	// wait for context to be done
	go func() {
		<-ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	// Start the HTTP server
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
