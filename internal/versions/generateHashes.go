//go:build ignore

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"golang.org/x/tools/go/ast/astutil"
)

var kuberntesMinorRegex = regexp.MustCompile(`^.*\.(?P<Minor>\d+)\..*(kubelet|kubeadm|kubectl)+$`)

func mustGetHash(url string) string {
	// remove quotes around url
	url = url[1 : len(url)-1]

	// Get the data
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		panic("bad status: " + resp.Status)
	}

	// Generate SHA256 hash of the file
	sha := sha256.New()
	if _, err := io.Copy(sha, resp.Body); err != nil {
		panic(err)
	}
	fileHash := sha.Sum(nil)

	// Get upstream hash
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, url+".sha256", nil)
	if err != nil {
		panic(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		panic("bad status: " + resp.Status)
	}

	// Compare hashes

	// Take the first 64 ascii characters = 32 bytes.
	// Some .sha256 files contain additional information afterwards.
	upstreamHash := make([]byte, 64)
	if _, err = resp.Body.Read(upstreamHash); err != nil {
		panic(err)
	}
	if string(upstreamHash) != fmt.Sprintf("%x", fileHash) {
		panic("hash mismatch")
	}

	// Verify cosign signature if available
	// Currently, we verify the signature of kubeadm, kubelet and kubectl with minor version >=1.26.
	minorVersion := kuberntesMinorRegex.FindStringSubmatch(url)
	if minorVersion == nil {
		return fmt.Sprintf("\"sha256:%x\"", fileHash)
	}
	minorVersionIndex := kuberntesMinorRegex.SubexpIndex("Minor")
	if minorVersionIndex != -1 {
		minorVersionNumber, err := strconv.Atoi(minorVersion[minorVersionIndex])
		if err != nil {
			panic(err)
		}
		if minorVersionNumber >= 26 {
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			if err := verifyCosignSignature(content, url); err != nil {
				panic(err)
			}
		}
	}

	return fmt.Sprintf("\"sha256:%x\"", fileHash)
}

func verifyCosignSignature(content []byte, url string) error {
	// Get the signature
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url+".sig", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		panic("bad status: " + resp.Status)
	}

	sig, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Get the certificate
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, url+".cert", nil)
	if err != nil {
		panic(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		panic("bad status: " + resp.Status)
	}

	cert, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return sigstore.VerifySignature(content, sig, cert)
}

func main() {
	fmt.Println("Generating hashes...")

	const filePath = "./versions.go"

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	newFile := astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		n := cursor.Node()

		if x, ok := n.(*ast.CompositeLit); ok {
			ident, ok := x.Type.(*ast.Ident)
			if !ok {
				return true
			}
			if ident.Name == "ComponentVersions" {
				for _, elt := range x.Elts {
					// component is one list element
					component := elt.(*ast.CompositeLit)

					var url *ast.KeyValueExpr
					var hash *ast.KeyValueExpr
					// Find the URL field
					for _, e := range component.Elts {
						kv, ok := e.(*ast.KeyValueExpr)
						if !ok {
							continue
						}
						ident, ok := kv.Key.(*ast.Ident)
						if !ok {
							continue
						}
						if ident.Name == "URL" {
							url = kv
							break
						}
					}
					// Find the Hash field
					for _, e := range component.Elts {
						kv, ok := e.(*ast.KeyValueExpr)
						if !ok {
							continue
						}
						ident, ok := kv.Key.(*ast.Ident)
						if !ok {
							continue
						}
						if ident.Name == "Hash" {
							hash = kv
							break
						}
					}

					// Generate the hash
					fmt.Println("Generating hash for", url.Value.(*ast.BasicLit).Value)
					hash.Value.(*ast.BasicLit).Value = mustGetHash(url.Value.(*ast.BasicLit).Value)
				}
			}
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
	fmt.Println("Successfully generated hashes.")
}
