/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// hash-generator updates the binary hashes and kubeadm patches in versions.go in place.
//
// This binary is usually invoked by the //bazel/ci:go_generate target, but you can run it
// manually, too. Clear a hash or a data URL in versions.go and execute
//
//	bazel run //internal/versions/hash-generator -- --update=false $PWD/internal/versions/versions.go
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/ref"
	"github.com/vincent-petithory/dataurl"
	"golang.org/x/tools/go/ast/astutil"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
)

const (
	defaultRegistry = "registry.k8s.io"
	etcdComponent   = "etcd"
	defaultFilePath = "./versions.go"
)

var supportedComponents = []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler", "etcd"}

func quote(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}

func unquote(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(s, `"`), `"`)
}

// pinKubernetesImage takes a component and a version and returns the corresponding container image pinned by hash.
//
// The version string is a Kubernetes version tag, which is used to derive the tags of the component images.
// The image hash is obtained directly from the default registry, registry.k8s.io.
func pinKubernetesImage(comp, ver string) (string, error) {
	if !slices.Contains(supportedComponents, comp) {
		return "", fmt.Errorf("k8s component %q not supported: valid components: %#v", comp, supportedComponents)
	}
	ref := ref.Ref{Scheme: "reg", Registry: defaultRegistry, Repository: comp, Tag: ver}
	if comp == etcdComponent {
		cfg := &kubeadm.ClusterConfiguration{
			KubernetesVersion: ver,
			ImageRepository:   defaultRegistry,
		}

		img := images.GetEtcdImage(cfg)
		_, tag, _ := strings.Cut(img, ":")
		ref.Tag = tag
	}
	log.Printf("Getting hash for image %#v", ref)

	rc := regclient.New()
	m, err := rc.ManifestGet(context.Background(), ref)
	if err != nil {
		return "", fmt.Errorf("could not obtain image manifest: %w", err)
	}

	return fmt.Sprintf("%s/%s:%s@%s", ref.Registry, ref.Repository, ref.Tag, m.GetDescriptor().Digest.String()), nil
}

func generateKubeadmPatch(comp, ver string) (string, error) {
	img, err := pinKubernetesImage(comp, ver)
	if err != nil {
		return "", err
	}
	content, err := json.Marshal([]map[string]string{{
		"op":    "replace",
		"path":  "/spec/containers/0/image",
		"value": img,
	}})
	if err != nil {
		return "", err
	}
	return dataurl.New(content, "application/json").String(), nil
}

// hashURLContent downloads a binary blob from the given URL and calculates its SHA256 hash.
//
// URLs passed to this function are expected to have upstream signatures with a .sha256 suffix. This upstream signature
// will be verified, too.
//
// nolint:noctx // This is a cli that does not benefit from passing contexts around.
func hashURLContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP response code: %d", resp.StatusCode)
	}

	// Generate SHA256 hash of the file
	sha := sha256.New()
	if _, err := io.Copy(sha, resp.Body); err != nil {
		return "", fmt.Errorf("could not calculate response body hash: %w", err)
	}
	fileHash := sha.Sum(nil)

	resp, err = http.Get(url + ".sha256")
	if err != nil {
		return "", fmt.Errorf("could not fetch upstream digest: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP response code for upstream digest: %d", resp.StatusCode)
	}

	// Compare hashes

	// Take the first 64 ascii characters = 32 bytes.
	// Some .sha256 files contain additional information afterwards.
	upstreamHash := make([]byte, 64)
	if _, err = resp.Body.Read(upstreamHash); err != nil {
		return "", fmt.Errorf("could not read upstream hash: %w", err)
	}
	if string(upstreamHash) != fmt.Sprintf("%x", fileHash) {
		return "", fmt.Errorf("computed hash %x does not match upstream hash %s", fileHash, string(upstreamHash))
	}

	return fmt.Sprintf("sha256:%x", fileHash), nil
}

type updater struct {
	k8sVersion string
}

// maybeSetVersion keeps track of the ambient ClusterVersion of components.
func (u *updater) maybeSetVersion(n ast.Node) {
	kv, ok := n.(*ast.KeyValueExpr)
	if !ok {
		return
	}
	key, ok := kv.Key.(*ast.Ident)
	if !ok || key.Name != "ClusterVersion" {
		return
	}
	val, ok := kv.Value.(*ast.BasicLit)
	if !ok || val.Kind != token.STRING {
		return
	}

	u.k8sVersion = val.Value[1 : len(val.Value)-1]
}

func (u *updater) updateComponents(cursor *astutil.Cursor) bool {
	n := cursor.Node()

	u.maybeSetVersion(n)
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

	log.Printf("Iterating over components for cluster version %q", u.k8sVersion)

	//
	// Iterate over the components
	//
	for _, componentElt := range comp.Elts {
		component := componentElt.(*ast.CompositeLit)

		var url, hash, installPath *ast.KeyValueExpr

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
			case "InstallPath":
				installPath = kv
			}
		}

		urlValue := unquote(url.Value.(*ast.BasicLit).Value)
		if urlValue == "" || strings.HasPrefix(urlValue, "data:") {
			// This can't be a downloadable component, so we assume this is supposed to be a kubeadm patch.
			if urlValue != "" && !*updateHash {
				continue
			}
			// all patch InstallPaths look like `patchFilePath("$COMPONENT")`
			comp := unquote(installPath.Value.(*ast.CallExpr).Args[0].(*ast.BasicLit).Value)
			log.Println("Generating kubeadm patch for", comp)
			dataURL, err := generateKubeadmPatch(comp, u.k8sVersion)
			if err != nil {
				log.Fatalf("Could not generate kubeadm patch for %q: %v", comp, err)
			}
			url.Value.(*ast.BasicLit).Value = quote(dataURL)
		} else {
			if hash.Value.(*ast.BasicLit).Value != `""` && !*updateHash {
				continue
			}
			log.Println("Generating hash for", urlValue)
			h, err := hashURLContent(urlValue)
			if err != nil {
				log.Fatalf("Could not hash URL %q: %v", urlValue, err)
			}
			hash.Value.(*ast.BasicLit).Value = quote(h)
		}
	}

	return true
}

var updateHash = flag.Bool("update", true, "update existing hashes and data URLs")

func main() {
	log.Println("Generating hashes...")

	flag.Parse()

	filePath := flag.Arg(0)
	if filePath == "" {
		filePath = defaultFilePath
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Could not parse file %q: %v", filePath, err)
	}

	updater := &updater{}
	newFile := astutil.Apply(file, updater.updateComponents, nil)

	var buf bytes.Buffer
	printConfig := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}

	if err = printConfig.Fprint(&buf, fset, newFile); err != nil {
		log.Fatalf("Could not format file %q: %v", filePath, err)
	}
	if err := os.WriteFile(filePath, buf.Bytes(), 0o644); err != nil {
		log.Fatalf("Could not write file %q: %v", filePath, err)
	}
}
