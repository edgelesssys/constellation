package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/watcher"
	"github.com/edgelesssys/constellation/kms/internal/server"
	"github.com/edgelesssys/constellation/kms/setup"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	port := flag.String("port", strconv.Itoa(constants.KMSPort), "Port gRPC server listens on")
	portATLS := flag.String("atls-port", strconv.Itoa(constants.KMSNodePort), "Port aTLS server listens on")
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	masterSecretPath := flag.String("master-secret", filepath.Join(constants.ServiceBasePath, constants.MasterSecretFilename), "Path to the Constellation master secret")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))

	log.With(zap.String("version", constants.VersionInfo), zap.String("cloudProvider", *provider)).
		Infof("Constellation Key Management Service")

	validator, err := watcher.NewValidator(log.Named("validator"), *provider, file.NewHandler(afero.NewOsFs()))
	if err != nil {
		flag.Usage()
		log.With(zap.Error(err)).Fatalf("Failed to create validator")
	}
	creds := atlscredentials.New(nil, []atls.Validator{validator})

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

	// set up listeners
	atlsListener, err := net.Listen("tcp", net.JoinHostPort("", *portATLS))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to listen on port %s", *portATLS)
	}
	plainListener, err := net.Listen("tcp", net.JoinHostPort("", *port))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to listen on port %s", *port)
	}

	// start the measurements file watcher
	watcher, err := watcher.New(log.Named("fileWatcher"), validator)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create watcher for measurements updates")
	}
	defer watcher.Close()

	go func() {
		log.Infof("starting file watcher for measurements file %s", filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename))
		if err := watcher.Watch(filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename)); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to watch measurements file")
		}
	}()

	// start the server
	if err := server.New(log.Named("server"), conKMS).Run(atlsListener, plainListener, creds); err != nil {
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
	if len(secretBytes) < constants.MasterSecretLengthMin {
		return nil, fmt.Errorf("provided master secret is smaller than the required minimum of %d bytes", constants.MasterSecretLengthMin)
	}

	return secretBytes, nil
}
