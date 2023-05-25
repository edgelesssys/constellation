/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package configapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// UseDummyConfigAPIServer starts a config api server which returns dummy data and updates the BaseURL so that subsequent API requests are directed to the dummy server. IMPORTANT: The returned cancel function must be called to stop the server (i.e. when the test is finished).
func UseDummyConfigAPIServer(port uint) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	addr := fmt.Sprintf(":%d", port)
	go startDummyConfigAPIServer(ctx, addr)
	baseURL = fmt.Sprintf("http://localhost%s", addr)
	return cancel
}

// startDummyConfigAPIServer starts a config api server which returns dummy data. It stops when the context is done.
func startDummyConfigAPIServer(ctx context.Context, addr string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/constellation/v1/attestation/azure-sev-snp/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]string{"2021-01-01-01-01.json"}); err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("/constellation/v1/attestation/azure-sev-snp/2021-01-01-01-01.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]int{"bootloader": 2, "tee": 0, "snp": 6, "microcode": 93}); err != nil {
			panic(err)
		}
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
