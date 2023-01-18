/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"path/filepath"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/setup"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/keyservice/internal/server"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	port := flag.String("port", strconv.Itoa(constants.KeyservicePort), "Port gRPC server listens on")
	masterSecretPath := flag.String("master-secret", filepath.Join(constants.ServiceBasePath, constants.ConstellationMasterSecretKey), "Path to the Constellation master secret")
	saltPath := flag.String("salt", filepath.Join(constants.ServiceBasePath, constants.ConstellationSaltKey), "Path to the Constellation salt")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))

	log.With(zap.String("version", constants.VersionInfo)).
		Infof("Constellation Key Management Service")

	// read master secret and salt
	file := file.NewHandler(afero.NewOsFs())
	masterKey, err := file.Read(*masterSecretPath)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to read master secret")
	}
	if len(masterKey) < crypto.MasterSecretLengthMin {
		log.With(zap.Error(errors.New("invalid key length"))).Fatalf("Provided master secret is smaller than the required minimum of %d bytes", crypto.MasterSecretLengthMin)
	}
	salt, err := file.Read(*saltPath)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to read salt")
	}
	if len(salt) < crypto.RNGLengthDefault {
		log.With(zap.Error(errors.New("invalid salt length"))).Fatalf("Expected salt to be %d bytes, but got %d", crypto.RNGLengthDefault, len(salt))
	}
	keyURI := setup.ClusterKMSURI + "?salt=" + base64.URLEncoding.EncodeToString(salt)

	// set up Key Management Service
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	conKMS, err := setup.KMS(ctx, setup.NoStoreURI, keyURI)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to setup KMS")
	}
	if err := conKMS.CreateKEK(ctx, "Constellation", masterKey); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create KMS KEK from MasterKey")
	}

	if err := server.New(log.Named("keyservice"), conKMS).Run(*port); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to run keyservice server")
	}
}
