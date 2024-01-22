//go:build integration && cgo && linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package integration

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"testing"

	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/diskencryption"
	"github.com/martinjungblut/go-cryptsetup"
)

func BenchmarkMapper(b *testing.B) {
	cryptsetup.SetDebugLevel(cryptsetup.CRYPT_LOG_ERROR)
	cryptsetup.SetLogCallback(func(_ int, message string) { fmt.Println(message) })

	testPath := *diskPath
	if testPath == "" {
		// no disk specified, use 1GB loopback disk
		testPath = devicePath
		if err := setup(1); err != nil {
			b.Fatal("Failed to setup test environment:", err)
		}

		defer func() {
			if err := teardown(); err != nil {
				b.Fatal("failed to delete test disk:", err)
			}
		}()
	}

	passphrase := "benchmark"
	mapper, free, err := diskencryption.New(testPath, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	if err != nil {
		b.Fatal("Failed to create mapper:", err)
	}
	defer free()

	if err := mapper.FormatDisk(passphrase); err != nil {
		b.Fatal("Failed to format disk:", err)
	}

	testCases := map[string]struct {
		wipeBlockSize int
	}{
		"16KiB": {
			wipeBlockSize: int(math.Pow(2, 14)),
		},
		"32KiB": {
			wipeBlockSize: int(math.Pow(2, 15)),
		},
		"64KiB": {
			wipeBlockSize: int(math.Pow(2, 16)),
		},
		"128KiB": {
			wipeBlockSize: int(math.Pow(2, 17)),
		},
		"256KiB": {
			wipeBlockSize: int(math.Pow(2, 18)),
		},
		"512KiB": {
			wipeBlockSize: int(math.Pow(2, 19)),
		},
		"1MiB": {
			wipeBlockSize: int(math.Pow(2, 20)),
		},
		"2MiB": {
			wipeBlockSize: int(math.Pow(2, 21)),
		},
		"4MiB": {
			wipeBlockSize: int(math.Pow(2, 22)),
		},
		"8MiB": {
			wipeBlockSize: int(math.Pow(2, 23)),
		},
		"16MiB": {
			wipeBlockSize: int(math.Pow(2, 24)),
		},
		"32MiB": {
			wipeBlockSize: int(math.Pow(2, 25)),
		},
		"64MiB": {
			wipeBlockSize: int(math.Pow(2, 26)),
		},
		"128MiB": {
			wipeBlockSize: int(math.Pow(2, 27)),
		},
		"256MiB": {
			wipeBlockSize: int(math.Pow(2, 28)),
		},
		"512MiB": {
			wipeBlockSize: int(math.Pow(2, 29)),
		},
	}

	for name, tc := range testCases {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := mapper.Wipe(tc.wipeBlockSize); err != nil {
					b.Fatal("Failed to wipe disk:", err)
				}
			}
		})
	}
}
