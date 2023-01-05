//go:build ignore

/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/shortname"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"golang.org/x/tools/go/ast/astutil"
)

const byteArrayCompositeLitWidth = 235

func mustGetMeasurements(ctx context.Context, verifier rekorVerifier, cosignPublicKey []byte, client *http.Client, provider cloudprovider.Provider, image string) measurements.M {
	measurementsURL, err := measurementURL(provider, image, "measurements.json")
	if err != nil {
		panic(err)
	}
	signatureURL, err := measurementURL(provider, image, "measurements.json.sig")
	if err != nil {
		panic(err)
	}

	log.Println("Fetching measurements from", measurementsURL, "and signature from", signatureURL)
	var fetchedMeasurements measurements.M
	hash, err := fetchedMeasurements.FetchAndVerify(
		ctx, client,
		measurementsURL,
		signatureURL,
		cosignPublicKey,
		measurements.WithMetadata{
			CSP:   provider,
			Image: image,
		},
	)
	if err != nil {
		panic(err)
	}
	if err := verifyWithRekor(ctx, verifier, hash); err != nil {
		panic(err)
	}
	return fetchedMeasurements
}

func measurementURL(provider cloudprovider.Provider, image, file string) (*url.URL, error) {
	ref, stream, version, err := shortname.ToParts(image)
	if err != nil {
		return nil, fmt.Errorf("parsing image name: %w", err)
	}
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("parsing image version repository URL: %w", err)
	}
	url.Path = path.Join(constants.CDNAPIPrefix, "ref", ref, "stream", stream, version, "image", "csp", strings.ToLower(provider.String()), file)
	return url, nil
}

func verifyWithRekor(ctx context.Context, verifier rekorVerifier, hash string) error {
	uuids, err := verifier.SearchByHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("searching Rekor for hash: %w", err)
	}

	if len(uuids) == 0 {
		return fmt.Errorf("no matching entries in Rekor")
	}

	// We expect the first entry in Rekor to be our original entry.
	// SHA256 should ensure there is no entry with the same hash.
	// Any subsequent hashes are treated as potential attacks and are ignored.
	// Attacks on Rekor will be monitored from other backend services.
	artifactUUID := uuids[0]

	return verifier.VerifyEntry(
		ctx, artifactUUID,
		base64.StdEncoding.EncodeToString([]byte(constants.CosignPublicKey)),
	)
}

func byteArrayCompositeLit(hex [32]byte, pos token.Pos) *ast.CompositeLit {
	var elts []ast.Expr
	curPos := pos + 16 // 16 = len("[32]byte{") + padding
	for i, b := range hex {
		elts = append(elts, &ast.BasicLit{
			ValuePos: curPos,
			Kind:     token.INT,
			Value:    fmt.Sprintf("0x%02x", b),
		})
		if (i+1)%8 == 0 {
			curPos += 11 // line break
		} else {
			curPos += 6 // 6 = len("0x00, ")
		}
	}
	return &ast.CompositeLit{
		Type: &ast.ArrayType{
			Lbrack: pos,
			Len:    ast.NewIdent("32"),
			Elt:    ast.NewIdent("byte"),
		},
		Lbrace: pos + 8, // 8 = len("[32]byte")
		Elts:   elts,
		Rbrace: pos + 223, // 223 = len("[32]byte{0x00, 0x01, ...}") - 1
	}
}

func measurementsEntryKeyValueExpr(pcr uint32, measuremnt measurements.Measurement, pos token.Pos) (*ast.KeyValueExpr, token.Pos) {
	key := fmt.Sprintf("%d", pcr)
	keyLen := len(key)
	colon := pos + token.Pos(keyLen)
	valuePos := token.Pos(colon + 2)
	expectedKeyPos := token.Pos(valuePos + 5)
	expectedColon := token.Pos(expectedKeyPos + 8)
	expectedValuePos := token.Pos(expectedColon + 2)
	warnOnlyKeyPos := token.Pos(expectedColon + 1 + byteArrayCompositeLitWidth)
	warnOnlyColon := token.Pos(warnOnlyKeyPos + 9)
	warnOnlyValuePos := token.Pos(warnOnlyColon + 2)
	var rbrace token.Pos
	if measuremnt.WarnOnly {
		rbrace = token.Pos(warnOnlyValuePos + 9) // 9 = len("true") + padding
	} else {
		rbrace = token.Pos(warnOnlyValuePos + 10) // 10 = len("false") + padding
	}

	return &ast.KeyValueExpr{
		Key: &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.INT,
			Value:    key,
		},
		Colon: colon,
		Value: &ast.CompositeLit{
			// Type:   ast.NewIdent("Measurement"),
			Lbrace: valuePos,
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key:   &ast.Ident{NamePos: expectedKeyPos, Name: "Expected"},
					Colon: expectedColon,
					Value: byteArrayCompositeLit(measuremnt.Expected, expectedValuePos),
				},
				&ast.KeyValueExpr{
					Key:   &ast.Ident{NamePos: warnOnlyKeyPos, Name: "WarnOnly"},
					Colon: warnOnlyColon,
					Value: &ast.Ident{NamePos: warnOnlyValuePos, Name: strconv.FormatBool(measuremnt.WarnOnly)},
				},
			},
			Rbrace: rbrace,
		},
	}, rbrace + 1
}

func measurementsCompositeLiteral(measurements measurements.M, pos token.Pos) *ast.CompositeLit {
	lbrace := pos + 1
	var elts []ast.Expr
	pcrs := make([]uint32, 0, len(measurements))
	for pcr := range measurements {
		pcrs = append(pcrs, pcr)
	}
	sort.Slice(pcrs, func(i, j int) bool { return pcrs[i] < pcrs[j] })
	entryPos := lbrace + 5
	for _, pcr := range pcrs {
		kvExpr, newEntryPos := measurementsEntryKeyValueExpr(pcr, measurements[pcr], entryPos)
		elts = append(elts, kvExpr)
		entryPos = newEntryPos + 4 // 4 = padding and newline
	}
	return &ast.CompositeLit{
		Lbrace: lbrace,
		Type: &ast.Ident{
			NamePos: pos,
			Name:    "M",
		},
		Elts:   elts,
		Rbrace: entryPos - 1,
	}
}

func main() {
	defaultConf := config.Default()
	log.Printf("Generating measurements for %s\n", defaultConf.Image)

	const filePath = "./measurements_enterprise.go"

	ctx := context.Background()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	rekor, err := sigstore.NewRekor()
	if err != nil {
		log.Fatal(err)
	}
	newFile := astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		n := cursor.Node()

		if clause, ok := n.(*ast.CaseClause); ok && len(clause.List) > 0 && len(clause.Body) > 0 {
			sel, ok := clause.List[0].(*ast.SelectorExpr)
			if !ok {
				return true
			}
			returnStmt, ok := clause.Body[0].(*ast.ReturnStmt)
			if !ok || len(returnStmt.Results) == 0 {
				return true
			}

			provider := cloudprovider.FromString(sel.Sel.Name)
			if provider == cloudprovider.Unknown {
				return true
			}
			log.Println("Found", provider)
			measuremnts := mustGetMeasurements(ctx, rekor, []byte(constants.CosignPublicKey), http.DefaultClient, provider, defaultConf.Image)
			returnStmt.Results[0] = measurementsCompositeLiteral(measuremnts, returnStmt.Return+7)
		}
		return true
	}, nil,
	)

	var buf bytes.Buffer
	printConfig := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}

	if err = printConfig.Fprint(&buf, fset, newFile); err != nil {
		log.Fatalf("error formatting file %s: %s", filePath, err)
	}
	if err := os.WriteFile(filePath, buf.Bytes(), 0o644); err != nil {
		log.Fatalf("error writing file %s: %s", filePath, err)
	}
	log.Println("Successfully generated hashes.")
}

type rekorVerifier interface {
	SearchByHash(context.Context, string) ([]string, error)
	VerifyEntry(context.Context, string, string) error
}
