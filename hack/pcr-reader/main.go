/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

func main() {
	coordIP := flag.String("constell-ip", "", "Public IP of the Constellation")
	port := flag.String("constell-port", strconv.Itoa(constants.VerifyServiceNodePortGRPC), "NodePort of the Constellation's verification service")
	format := flag.String("format", "json", "Output format: json, yaml (default json)")
	quiet := flag.Bool("q", false, "Set to disable output")
	timeout := flag.Duration("timeout", 2*time.Minute, "Wait this duration for the verification service to become available")
	flag.Parse()

	if *coordIP == "" || *port == "" {
		flag.Usage()
		os.Exit(1)
	}

	addr := net.JoinHostPort(*coordIP, *port)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	attDocRaw, err := getAttestation(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}

	measurements, err := validatePCRAttDoc(attDocRaw)
	if err != nil {
		log.Fatal(err)
	}

	if !*quiet {
		if err := printPCRs(os.Stdout, measurements, *format); err != nil {
			log.Fatal(err)
		}
	}
}

// getAttestation connects to the Constellation verification service and returns its attestation document.
func getAttestation(ctx context.Context, addr string) ([]byte, error) {
	conn, err := grpc.DialContext(
		ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to verification service: %w", err)
	}
	defer conn.Close()

	nonce, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return nil, err
	}

	client := verifyproto.NewAPIClient(conn)
	res, err := client.GetAttestation(ctx, &verifyproto.GetAttestationRequest{Nonce: nonce, UserData: nonce})
	if err != nil {
		return nil, err
	}
	return res.Attestation, nil
}

// validatePCRAttDoc parses and validates PCRs of an attestation document.
func validatePCRAttDoc(attDocRaw []byte) (measurements.M, error) {
	attDoc := vtpm.AttestationDocument{}
	if err := json.Unmarshal(attDocRaw, &attDoc); err != nil {
		return nil, err
	}
	if attDoc.Attestation == nil {
		return nil, errors.New("empty attestation")
	}
	qIdx, err := vtpm.GetSHA256QuoteIndex(attDoc.Attestation.Quotes)
	if err != nil {
		return nil, err
	}

	m := measurements.M{}
	for idx, pcr := range attDoc.Attestation.Quotes[qIdx].Pcrs.Pcrs {
		if len(pcr) != 32 {
			return nil, fmt.Errorf("incomplete PCR at index: %d", idx)
		}

		m[idx] = measurements.Measurement{
			Expected: *(*[32]byte)(pcr),
			WarnOnly: true,
		}
	}
	return m, nil
}

// printPCRs formats and prints PCRs to the given writer.
// format can be one of 'json' or 'yaml'. If it doesn't match defaults to 'json'.
func printPCRs(w io.Writer, pcrs measurements.M, format string) error {
	switch format {
	case "json":
		return printPCRsJSON(w, pcrs)
	case "yaml":
		return printPCRsYAML(w, pcrs)
	default:
		return printPCRsJSON(w, pcrs)
	}
}

func printPCRsYAML(w io.Writer, pcrs measurements.M) error {
	pcrYAML, err := yaml.Marshal(pcrs)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s", string(pcrYAML))
	return nil
}

func printPCRsJSON(w io.Writer, pcrs measurements.M) error {
	pcrJSON, err := json.MarshalIndent(pcrs, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s", string(pcrJSON))
	return nil
}
