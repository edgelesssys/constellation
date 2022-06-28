package main

import (
	"context"
	"encoding/base64"
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

	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/verify/verifyproto"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

var (
	coordIP         = flag.String("constell-ip", "", "Public IP of the Constellation")
	coordinatorPort = flag.String("constell-port", strconv.Itoa(constants.VerifyServiceNodePortGRPC), "NodePort of the Constellation's verification service")
	export          = flag.String("o", "", "Write PCRs, formatted as Go code, to file")
	format          = flag.String("format", "json", "Output format: json, yaml (default json)")
	quiet           = flag.Bool("q", false, "Set to disable output")
	timeout         = flag.Duration("timeout", 2*time.Minute, "Wait this duration for the verification service to become available")
)

func main() {
	flag.Parse()

	addr := net.JoinHostPort(*coordIP, *coordinatorPort)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	attDocRaw, err := getAttestation(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}

	pcrs, err := validatePCRAttDoc(attDocRaw)
	if err != nil {
		log.Fatal(err)
	}

	if !*quiet {
		if err := printPCRs(os.Stdout, pcrs, *format); err != nil {
			log.Fatal(err)
		}
	}
	if *export != "" {
		if err := exportToFile(*export, pcrs, &afero.Afero{Fs: afero.NewOsFs()}); err != nil {
			log.Fatal(err)
		}
	}
}

type Measurements map[uint32][]byte

// MarshalYAML forces that measurements are written as base64. Default would
// be to print list of bytes.
func (m Measurements) MarshalYAML() (interface{}, error) {
	base64Map := make(map[uint32]string)

	for key, value := range m {
		base64Map[key] = base64.StdEncoding.EncodeToString(value[:])
	}

	return base64Map, nil
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

	nonce, err := util.GenerateRandomBytes(32)
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
// format can be one of 'json' or 'yaml'. If it doesnt match defaults to 'json'.
func printPCRs(w io.Writer, pcrs map[uint32][]byte, format string) error {
	switch format {
	case "json":
		return printPCRsJSON(w, pcrs)
	case "yaml":
		return printPCRsYAML(w, pcrs)
	default:
		return printPCRsJSON(w, pcrs)
	}
}

func printPCRsYAML(w io.Writer, pcrs Measurements) error {
	pcrYAML, err := yaml.Marshal(pcrs)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s", string(pcrYAML))
	return nil
}

func printPCRsJSON(w io.Writer, pcrs map[uint32][]byte) error {
	pcrJSON, err := json.MarshalIndent(pcrs, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s", string(pcrJSON))
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
