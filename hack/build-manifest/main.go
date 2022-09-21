/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/edgelesssys/constellation/v2/hack/build-manifest/azure"
	"github.com/edgelesssys/constellation/v2/hack/build-manifest/gcp"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap/zapcore"
)

const (
	AzureSubscriptionIDEnv = "AZURE_SUBSCRIPTION_ID"
)

func main() {
	ctx := context.Background()
	log := logger.New(logger.PlainLog, zapcore.InfoLevel)
	manifests := OldManifests()

	fetchAzureImages(ctx, manifests, log)
	fetchGCPImages(ctx, manifests, log)

	if err := json.NewEncoder(os.Stdout).Encode(&manifests); err != nil {
		log.Fatalf("%v", err)
	}
}

func fetchAzureImages(ctx context.Context, manifests Manifest, log *logger.Logger) {
	options := azure.DefaultOptions()
	if err := options.SetSubscription(os.Getenv(AzureSubscriptionIDEnv)); err != nil {
		log.Fatalf("please provide a valid subscription UUID via '%s' envar", AzureSubscriptionIDEnv)
	}

	client := azure.NewClient(log, options)
	images, err := client.FetchImages(ctx)
	if err != nil {
		log.Fatalf("unable to fetch Azure image: %v", err)
	}

	for version, image := range images {
		manifests.SetAzureImage(version, image)
	}
}

func fetchGCPImages(ctx context.Context, manifests Manifest, log *logger.Logger) {
	options := gcp.DefaultOptions()
	client := gcp.NewClient(ctx, log, options)
	images, err := client.FetchImages(ctx)
	if err != nil {
		log.Fatalf("unable to fetch GCP images: %v", err)
	}

	for version, image := range images {
		manifests.SetGCPImage(version, image)
	}
}
