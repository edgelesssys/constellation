/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package artifacts

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func TestContainerArtifacts(t *testing.T) {
	if !runsUnderBazel() {
		t.Skip("Skipping test as it is not running under Bazel")
	}
	compareArtifactsWithPrefix(t, "container_sums_")
}

func TestCLIArtifacts(t *testing.T) {
	if !runsUnderBazel() {
		t.Skip("Skipping test as it is not running under Bazel")
	}
	compareArtifactsWithPrefix(t, "cli_transitioned_to_")
}

type artifact struct {
	name, path string
}

func compareArtifactsWithPrefix(t *testing.T, prefix string) {
	artifacts := getAllArtifacts(prefix)
	if len(artifacts) < 2 {
		t.Fatal("No artifacts found")
	}
	hashes := make([]string, len(artifacts))
	for i, artifact := range artifacts {
		hash, err := hashForArtifact(artifact)
		if err != nil {
			t.Fatalf("hashing artifact %s: %v", artifact, err)
		}
		hashes[i] = hash
	}
	want := hashes[0]
	for i := 1; i < len(hashes)-1; i++ {
		if hashes[i] != want {
			t.Errorf("hash for %s: %s, want %s", artifacts[i].name, hashes[i], want)
		}
	}
}

func getAllArtifacts(prefix string) []artifact {
	envVars := os.Environ()
	var artifacts []artifact
	for _, envVar := range envVars {
		k, v := splitEnvVar(envVar)
		if strings.HasPrefix(k, prefix) {
			path, err := runfiles.Rlocation(v)
			if err != nil {
				panic("could not find path to artifact")
			}
			artifacts = append(artifacts, artifact{k, path})
		}
	}
	return artifacts
}

func splitEnvVar(envVar string) (string, string) {
	split := strings.SplitN(envVar, "=", 2)
	if len(split) == 0 {
		return "", ""
	}
	if len(split) == 1 {
		return split[0], ""
	}
	return split[0], split[1]
}

func hashForArtifact(artifact artifact) (string, error) {
	f, err := os.Open(artifact.path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func runsUnderBazel() bool {
	return runsUnder == "bazel"
}

// runsUnder is redefined only by the Bazel build at link time.
var runsUnder = "go"
