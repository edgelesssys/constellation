/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"io"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edgelesssys/constellation/v2/internal/constants"
)

func TestVersionCmd(t *testing.T) {
	assert := assert.New(t)

	cmd := NewVersionCmd()
	b := &bytes.Buffer{}
	cmd.SetOut(b)

	err := cmd.Execute()
	assert.NoError(err)

	s, err := io.ReadAll(b)
	assert.NoError(err)
	assert.Contains(string(s), constants.VersionInfo())
}

func TestParseBuildInfo(t *testing.T) {
	assert := assert.New(t)
	info := debug.BuildInfo{
		GoVersion: "go1.18.3",
		Settings: []debug.BuildSetting{
			{
				Key:   "-compiler",
				Value: "gc",
			},
			{
				Key:   "GOARCH",
				Value: "amd64",
			},
			{
				Key:   "GOOS",
				Value: "linux",
			},
			{
				Key:   "vcs.time",
				Value: "2022-06-20T11:57:25Z",
			},
			{
				Key:   "vcs.modified",
				Value: "true",
			},
			{
				Key:   "vcs.revision",
				Value: "abcdef",
			},
		},
	}

	commit, state, date, goVersion, compiler, platform := parseBuildInfo(&info)

	assert.Equal("abcdef", commit)
	assert.Equal("dirty", state)
	assert.Equal("2022-06-20T11:57:25Z", date)
	assert.Equal("go1.18.3", goVersion)
	assert.Equal("gc", compiler)
	assert.Equal("linux/amd64", platform)
}
