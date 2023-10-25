//go:build cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"log/slog"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/server"
	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
	"libvirt.org/go/libvirt"
)

func main() {
	bindPort := flag.String("port", "8080", "Port to bind to")
	targetNetwork := flag.String("network", "constellation-network", "Name of the network in QEMU to use")
	libvirtURI := flag.String("libvirt-uri", "qemu:///system", "URI of the libvirt connection")
	initSecretHash := flag.String("initsecrethash", "", "brcypt hash of the init secret")
	flag.Parse()

	log := logger.New(logger.JSONLog, slog.LevelInfo)

	conn, err := libvirt.NewConnect(*libvirtURI)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to connect to libvirt")
	}
	defer conn.Close()

	serv := server.New(log, *targetNetwork, *initSecretHash, &virtwrapper.Connect{Conn: conn})
	if err := serv.ListenAndServe(*bindPort); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to serve")
	}
}
