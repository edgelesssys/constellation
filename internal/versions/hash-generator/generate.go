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

	"golang.org/x/tools/go/ast/astutil"
)

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
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, url+".sha256", http.NoBody)
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

	return fmt.Sprintf("\"sha256:%x\"", fileHash)
}

func main() {
	fmt.Println("Generating hashes...")

	const filePath = "./versions.go"

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var componentListsCtr, componentCtr int

	newFile := astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		n := cursor.Node()

		//
		// Find CompositeLit of type 'components.Components'
		//
		comp, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		selExpr, ok := comp.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if selExpr.Sel.Name != "Components" {
			return true
		}
		xIdent, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return true
		}
		if xIdent.Name != "components" {
			return true
		}
		componentListsCtr++

		//
		// Iterate over the components
		//
		for _, componentElt := range comp.Elts {
			component := componentElt.(*ast.CompositeLit)
			componentCtr++

			var url *ast.KeyValueExpr
			var hash *ast.KeyValueExpr

			for _, e := range component.Elts {
				kv, ok := e.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				ident, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}
				switch ident.Name {
				case "Url":
					url = kv
				case "Hash":
					hash = kv
				}
			}

			fmt.Println("Generating hash for", url.Value.(*ast.BasicLit).Value)
			hash.Value.(*ast.BasicLit).Value = mustGetHash(url.Value.(*ast.BasicLit).Value)
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
	if componentCtr == 0 {
		log.Fatalf("no components lists found")
	}

	fmt.Printf("Successfully generated hashes for %d components in %d component lists.\n", componentCtr, componentListsCtr)
}
