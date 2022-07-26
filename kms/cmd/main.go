package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/kms/internal/server"
	"github.com/edgelesssys/constellation/kms/setup"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	port := flag.String("port", strconv.Itoa(constants.KMSPort), "Port gRPC server listens on")
	masterSecretPath := flag.String("master-secret", filepath.Join(constants.ServiceBasePath, constants.MasterSecretFilename), "Path to the Constellation master secret")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))

	log.With(zap.String("version", constants.VersionInfo)).
		Infof("Constellation Key Management Service")

	// set up Key Management Service
	masterKey, err := readMainSecret(*masterSecretPath)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to read master secret")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	conKMS, err := setup.SetUpKMS(ctx, setup.NoStoreURI, setup.ClusterKMSURI)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to setup KMS")
	}
	if err := conKMS.CreateKEK(ctx, "Constellation", masterKey); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create KMS KEK from MasterKey")
	}

	if err := server.New(log.Named("kms"), conKMS).Run(*port); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to run KMS server")
	}
}

// readMainSecret reads the base64 encoded main secret file from specified path and returns the secret as bytes.
func readMainSecret(fileName string) ([]byte, error) {
	if fileName == "" {
		return nil, errors.New("no filename to master secret provided")
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	secretBytes, err := fileHandler.Read(fileName)
	if err != nil {
		return nil, err
	}
	if len(secretBytes) < crypto.MasterSecretLengthMin {
		return nil, fmt.Errorf("provided master secret is smaller than the required minimum of %d bytes", crypto.MasterSecretLengthMin)
	}

	return secretBytes, nil
}
