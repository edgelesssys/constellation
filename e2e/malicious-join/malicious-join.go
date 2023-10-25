/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// End-to-end test that issues various types of malicious join requests to a cluster.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
)

func main() {
	jsEndpoint := flag.String("js-endpoint", "", "Join service endpoint to use.")
	csp := flag.String("csp", "", "Cloud service provider to use.")
	attVariant := flag.String(
		"variant",
		"",
		fmt.Sprintf("Attestation variant to use. Set to \"default\" to use the default attestation variant for the CSP,"+
			"or one of: %s", variant.GetAvailableAttestationVariants()),
	)
	flag.Parse()
	fmt.Println(formatFlags(*attVariant, *csp, *jsEndpoint))

	testCases := map[string]struct {
		fn      func(attVariant, csp, jsEndpoint string) error
		wantErr bool
	}{
		"JoinFromUnattestedNode": {
			fn:      JoinFromUnattestedNode,
			wantErr: true,
		},
	}

	allPassed := true
	testOutput := &testOutput{
		TestCases: make(map[string]testCaseOutput),
	}
	for name, tc := range testCases {
		fmt.Printf("Running testcase %s\n", name)

		err := tc.fn(*attVariant, *csp, *jsEndpoint)

		switch {
		case err == nil && tc.wantErr:
			fmt.Printf("Test case %s failed: Expected error but got none\n", name)
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  false,
				Message: "Expected error but got none",
			}
			allPassed = false
		case !tc.wantErr && err != nil:
			fmt.Printf("Test case %s failed: Got unexpected error: %s\n", name, err)
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  false,
				Message: fmt.Sprintf("Got unexpected error: %s", err),
			}
			allPassed = false
		case tc.wantErr && err != nil:
			fmt.Printf("Test case %s succeeded\n", name)
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  true,
				Message: fmt.Sprintf("Got expected error: %s", err),
			}
		case !tc.wantErr && err == nil:
			fmt.Printf("Test case %s succeeded\n", name)
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  true,
				Message: "No error, as expected",
			}
		default:
			panic("invalid result")
		}
	}

	testOutput.AllPassed = allPassed
	out, err := json.Marshal(testOutput)
	if err != nil {
		panic(fmt.Sprintf("marshalling test output: %s", err))
	}
	fmt.Println(string(out))
}

type testOutput struct {
	AllPassed bool                      `json:"allPassed"`
	TestCases map[string]testCaseOutput `json:"testCases"`
}

type testCaseOutput struct {
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

func formatFlags(attVariant, csp, jsEndpoint string) string {
	var sb strings.Builder
	sb.WriteString("Using Flags:\n")
	sb.WriteString(fmt.Sprintf("\tjs-endpoint: %s\n", jsEndpoint))
	sb.WriteString(fmt.Sprintf("\tcsp: %s\n", csp))
	sb.WriteString(fmt.Sprintf("\tvariant: %s\n", attVariant))
	return sb.String()
}

// JoinFromUnattestedNode simulates a join request from a Node that uses a stub issuer
// and thus cannot be attested correctly.
func JoinFromUnattestedNode(attVariant, csp, jsEndpoint string) error {
	log := logger.New(logger.JSONLog, slog.LevelDebug)
	joiner, err := newMaliciousJoiner(attVariant, csp, jsEndpoint, log)
	if err != nil {
		return fmt.Errorf("creating malicious joiner: %w", err)
	}

	_, err = joiner.join(context.Background())
	if err != nil {
		return fmt.Errorf("joining cluster: %w", err)
	}
	return nil
}

// newMaliciousJoiner creates a new malicious joiner, i.e. a simulated node that issues
// an invalid join request.
func newMaliciousJoiner(attVariant, csp, endpoint string, log *logger.Logger) (*maliciousJoiner, error) {
	var attVariantOid variant.Variant
	var err error
	if strings.EqualFold(attVariant, "default") {
		attVariantOid = variant.GetDefaultAttestation(cloudprovider.FromString(csp))
	} else {
		attVariantOid, err = variant.FromString(attVariant)
		if err != nil {
			return nil, fmt.Errorf("parsing attestation variant: %w", err)
		}
	}

	issuer := newFakeIssuer(attVariantOid)

	return &maliciousJoiner{
		endpoint: endpoint,
		logger:   log,
		dialer:   dialer.New(issuer, nil, &net.Dialer{}),
	}, nil
}

// maliciousJoiner simulates a malicious node joining a cluster.
type maliciousJoiner struct {
	endpoint string
	logger   *logger.Logger
	dialer   *dialer.Dialer
}

// join issues a join request to the join service endpoint.
func (j *maliciousJoiner) join(ctx context.Context) (*joinproto.IssueJoinTicketResponse, error) {
	j.logger.Debugf("Dialing join service endpoint %s", j.endpoint)
	conn, err := j.dialer.Dial(ctx, j.endpoint)
	if err != nil {
		return nil, fmt.Errorf("dialing join service endpoint: %w", err)
	}
	defer conn.Close()
	j.logger.Debugf("Successfully dialed join service endpoint %s", j.endpoint)

	protoClient := joinproto.NewAPIClient(conn)

	j.logger.Debugf("Issuing join ticket")
	req := &joinproto.IssueJoinTicketRequest{
		DiskUuid:           "",
		CertificateRequest: []byte{},
		IsControlPlane:     false,
	}
	res, err := protoClient.IssueJoinTicket(ctx, req)
	j.logger.Debugf("Got join ticket response: %+v", res)
	if err != nil {
		return nil, fmt.Errorf("issuing join ticket: %w", err)
	}

	return res, nil
}

// newFakeIssuer creates a new fake issuer for a given attestation variant.
func newFakeIssuer(oid variant.Getter) *fakeIssuer {
	return &fakeIssuer{oid}
}

// fakeIssuer simulates an issuer that issues a fake / invalid attestation document.
type fakeIssuer struct {
	variant.Getter
}

// Issue issues a fake attestation document.
func (i *fakeIssuer) Issue(_ context.Context, userData, nonce []byte) ([]byte, error) {
	return json.Marshal(fakeAttestationDoc{UserData: userData, Nonce: nonce})
}

// fakeAttestationDoc is a fake attestation document.
type fakeAttestationDoc struct {
	UserData []byte
	Nonce    []byte
}
