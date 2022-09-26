/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/server"
	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"libvirt.org/go/libvirt"
)

func main() {
	bindPort := flag.String("port", "8080", "Port to bind to")
	targetNetwork := flag.String("network", "constellation-network", "Name of the network in QEMU to use")
	flag.Parse()

	log := logger.New(logger.JSONLog, zapcore.InfoLevel)

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to connect to libvirt")
	}
	defer conn.Close()

	serv := server.New(log, *targetNetwork, &virtwrapper.Connect{Conn: conn})
	if err := serv.ListenAndServe(*bindPort); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to serve")
	}
}
