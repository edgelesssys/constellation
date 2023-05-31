/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Verify verifies an MAA JWT and prints the SNP ID key digest on success.
package main

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// IsolationTEE describes an Azure SNP TEE.
type IsolationTEE struct {
	IDKeyDigest   string `json:"x-ms-sevsnpvm-idkeydigest"`
	TEESvn        int    `json:"x-ms-sevsnpvm-tee-svn"`
	SNPFwSvn      int    `json:"x-ms-sevsnpvm-snpfw-svn"`
	MicrocodeSvn  int    `json:"x-ms-sevsnpvm-microcode-svn"`
	BootloaderSvn int    `json:"x-ms-sevsnpvm-bootloader-svn"`
	GuestSvn      int    `json:"x-ms-sevsnpvm-guestsvn"`
}

// PrintSVNs prints the relevant Security Version Numbers (SVNs).
func (i *IsolationTEE) PrintSVNs() {
	fmt.Println("\tTEE SVN:", i.TEESvn)
	fmt.Println("\tSNP FW SVN:", i.SNPFwSvn)
	fmt.Println("\tMicrocode SVN:", i.MicrocodeSvn)
	fmt.Println("\tBootloader SVN:", i.BootloaderSvn)
	fmt.Println("\tGuest SVN:", i.GuestSvn)
}

func main() {
	configAPIExportPath := flag.String("export-path", "azure-sev-snp-version.json", "Path to the exported config API file.")
	flag.Parse()
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "<maa-jwt>")
		return
	}
	report, err := getTEEReport(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully validated ID key digest:", report.IDKeyDigest)
	fmt.Println("Currently reported SVNs:")
	report.PrintSVNs()

	if *configAPIExportPath != "" {
		configAPIVersion := convertToConfigAPIFile(report)
		if err := exportToJSONFile(configAPIVersion, *configAPIExportPath); err != nil {
			panic(err)
		}
		fmt.Println("Successfully exported config API file to:", configAPIExportPath)
	}
}

func convertToConfigAPIFile(i IsolationTEE) configapi.AzureSEVSNPVersion {
	return configapi.AzureSEVSNPVersion{
		Bootloader: uint8(i.BootloaderSvn),
		TEE:        uint8(i.TEESvn),
		SNP:        uint8(i.SNPFwSvn),
		Microcode:  uint8(i.MicrocodeSvn),
	}
}

func exportToJSONFile(configAPIVersion configapi.AzureSEVSNPVersion, configAPIExportPath string) error {
	data, err := json.Marshal(configAPIVersion)
	if err != nil {
		return err
	}
	return os.WriteFile(configAPIExportPath, data, 0o644)
}

func getTEEReport(rawToken string) (IsolationTEE, error) {
	// Parse token.
	token, err := jwt.ParseSigned(rawToken)
	if err != nil {
		return IsolationTEE{}, err
	}

	// Get JSON Web Key set.
	keySetBytes, err := httpGet(context.Background(), "https://sharedeus.eus.attest.azure.net/certs")
	if err != nil {
		return IsolationTEE{}, err
	}
	keySet, err := parseKeySet(keySetBytes)
	if err != nil {
		return IsolationTEE{}, err
	}

	// Get claims. Private claims contain ID Key digest.

	var publicClaims jwt.Claims

	var privateClaims struct {
		IsolationTEE IsolationTEE `json:"x-ms-isolation-tee"`
	}

	if err := token.Claims(&keySet, &publicClaims, &privateClaims); err != nil {
		return IsolationTEE{}, err
	}
	if err := publicClaims.Validate(jwt.Expected{Time: time.Now()}); err != nil {
		return IsolationTEE{}, err
	}

	return privateClaims.IsolationTEE, nil
}

func parseKeySet(keySetBytes []byte) (jose.JSONWebKeySet, error) {
	var rawKeySet struct {
		Keys []struct {
			X5c []string
			Kid string
		}
	}
	if err := json.Unmarshal(keySetBytes, &rawKeySet); err != nil {
		return jose.JSONWebKeySet{}, err
	}

	var keySet jose.JSONWebKeySet
	for _, key := range rawKeySet.Keys {
		rawCert, _ := base64.StdEncoding.DecodeString(key.X5c[0])
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return jose.JSONWebKeySet{}, err
		}
		keySet.Keys = append(keySet.Keys, jose.JSONWebKey{KeyID: key.Kid, Key: cert.PublicKey})
	}

	return keySet, nil
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
