/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package issues

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestMap(t *testing.T) {
	assert := assert.New(t)

	m := New()
	assert.Equal(0, len(m))
	assert.False(m.FileHasIssues("file1"))
	m.Set("file1", map[string][]error{
		"rule1": {errors.New("r1_issue1"), errors.New("r1_issue2")},
		"rule2": {errors.New("r2_issue1")},
	})
	assert.Equal(3, m.IssuesPerFile("file1"))
	assert.True(m.FileHasIssues("file1"))

	// let report write to a buffer
	b := new(bytes.Buffer)
	m.Report(b)
	rep := b.String()
	assert.Equal(rep, `File file1 (3 issues total):
  Rule rule1 (2 issues total):
    r1_issue1
    r1_issue2
  Rule rule2 (1 issues total):
    r2_issue1
`)
}
