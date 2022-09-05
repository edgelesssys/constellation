/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"

	"github.com/edgelesssys/constellation/hack/qemu-metadata-api/server"
	"github.com/edgelesssys/constellation/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"libvirt.org/go/libvirt"
)

func main() {
	bindPort := flag.String("port", "8080", "Port to bind to")
	flag.Parse()

	log := logger.New(logger.JSONLog, zapcore.InfoLevel)

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to connect to libvirt")
	}
	defer conn.Close()

	serv := server.New(log, &virtwrapper.Connect{Conn: conn}, file.NewHandler(afero.NewOsFs()))
	if err := serv.ListenAndServe(*bindPort); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to serve")
	}
}
