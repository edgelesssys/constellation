package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/edgelesssys/constellation/cli/status"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/oid"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	coordIP         = flag.String("coord-ip", "", "IP of the VM the Coordinator is running on")
	coordinatorPort = flag.String("coord-port", "9000", "Port of the Coordinator's pub API")
	export          = flag.String("o", "", "Write PCRs, formatted as Go code, to file")
	quiet           = flag.Bool("q", false, "Set to disable output")
)

func main() {
	flag.Parse()

	fmt.Printf("connecting to Coordinator at %s:%s\n", *coordIP, *coordinatorPort)
	addr := net.JoinHostPort(*coordIP, *coordinatorPort)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// wait for coordinator to come online
	waiter := status.NewWaiter()
	waiter.InitializeValidators(nil)
	if err := waiter.WaitFor(ctx, addr, state.AcceptingInit, state.ActivatingNodes, state.IsNode, state.NodeWaitingForClusterJoin); err != nil {
		log.Fatal(err)
	}

	attDocRaw := []byte{}
	tlsConfig, err := atls.CreateUnverifiedClientTLSConfig()
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig.VerifyPeerCertificate = getVerifyPeerCertificateFunc(&attDocRaw)
	if err := connectToCoordinator(ctx, addr, tlsConfig); err != nil {
		log.Fatal(err)
	}

	pcrs, err := validatePCRAttDoc(attDocRaw)
	if err != nil {
		log.Fatal(err)
	}

	if !*quiet {
		if err := printPCRs(os.Stdout, pcrs); err != nil {
			log.Fatal(err)
		}
	}
	if *export != "" {
		if err := exportToFile(*export, pcrs, &afero.Afero{Fs: afero.NewOsFs()}); err != nil {
			log.Fatal(err)
		}
	}
}

// connectToCoordinator connects to the Constellation Coordinator and returns its attestation document.
func connectToCoordinator(ctx context.Context, addr string, tlsConfig *tls.Config) error {
	conn, err := grpc.DialContext(
		ctx, addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.GetState(ctx, &pubproto.GetStateRequest{})
	return err
}

// getVerifyPeerCertificateFunc returns a VerifyPeerCertificate function, which writes the attestation document extension to the given byte slice pointer.
func getVerifyPeerCertificateFunc(attDoc *[]byte) func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return errors.New("rawCerts is empty")
		}
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return err
		}

		for _, ex := range cert.Extensions {
			if ex.Id.Equal(oid.Azure{}.OID()) || ex.Id.Equal(oid.GCP{}.OID()) || ex.Id.Equal(oid.GCPNonCVM{}.OID()) {
				if err := json.Unmarshal(ex.Value, attDoc); err != nil {
					*attDoc = ex.Value
				}
			}
		}

		if len(*attDoc) == 0 {
			return errors.New("did not receive attestation document in certificate extension")
		}
		return nil
	}
}

// validatePCRAttDoc parses and validates PCRs of an attestation document.
func validatePCRAttDoc(attDocRaw []byte) (map[uint32][]byte, error) {
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
	for idx, pcr := range attDoc.Attestation.Quotes[qIdx].Pcrs.Pcrs {
		if len(pcr) != 32 {
			return nil, fmt.Errorf("incomplete PCR at index: %d", idx)
		}
	}
	return attDoc.Attestation.Quotes[qIdx].Pcrs.Pcrs, nil
}

// printPCRs formates and prints PCRs to the given writer.
func printPCRs(w io.Writer, pcrs map[uint32][]byte) error {
	pcrJSON, err := json.MarshalIndent(pcrs, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "PCRs:\n%s\n", string(pcrJSON))
	return nil
}

// exportToFile writes pcrs to a file, formatted to be valid Go code.
// Validity of the PCR map is not checked, and should be handled by the caller.
func exportToFile(path string, pcrs map[uint32][]byte, fs *afero.Afero) error {
	goCode := `package pcrs

var pcrs = map[uint32][]byte{%s
}
`
	pcrsFormatted := ""
	for i := 0; i < len(pcrs); i++ {
		pcrHex := fmt.Sprintf("%#02X", pcrs[uint32(i)][0])
		for j := 1; j < len(pcrs[uint32(i)]); j++ {
			pcrHex = fmt.Sprintf("%s, %#02X", pcrHex, pcrs[uint32(i)][j])
		}

		pcrsFormatted = pcrsFormatted + fmt.Sprintf("\n\t%d: {%s},", i, pcrHex)
	}

	return fs.WriteFile(path, []byte(fmt.Sprintf(goCode, pcrsFormatted)), 0o644)
}
