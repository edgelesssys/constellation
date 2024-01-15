//go:build cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/server"
	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"
	"libvirt.org/go/libvirt"
)

func main() {
	bindPort := flag.String("port", "8080", "Port to bind to")
	targetNetwork := flag.String("network", "constellation-network", "Name of the network in QEMU to use")
	libvirtURI := flag.String("libvirt-uri", "qemu:///system", "URI of the libvirt connection")
	initSecretHash := flag.String("initsecrethash", "", "brcypt hash of the init secret")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	conn, err := libvirt.NewConnect(*libvirtURI)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to connect to libvirt")
		os.Exit(1)
	}
	defer conn.Close()

	serv := server.New(log, *targetNetwork, *initSecretHash, &virtwrapper.Connect{Conn: conn})
	if err := serv.ListenAndServe(*bindPort); err != nil {
		log.With(slog.Any("error", err)).Error("Failed to serve")
		os.Exit(1)
	}
}
