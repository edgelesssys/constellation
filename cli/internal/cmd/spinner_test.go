/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	baseWait = 300
	baseText = "Loading"
)

func TestSpinnerInitialState(t *testing.T) {
	assert := assert.New(t)

	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.delay = time.Millisecond * 10
	s.spinFunc = spinTTY
	s.Start(baseText, true)
	time.Sleep(baseWait * time.Millisecond)
	s.Stop()
	assert.Greater(out.Len(), 0)

	outStr := out.String()
	prefix := hideCursor + generateAllStatesAsString(t, baseText, true)
	assert.True(strings.HasPrefix(outStr, prefix), fmt.Sprintf("\nOutStr: %#v\nPrefix: %#v\n", outStr, prefix))
}

func TestSpinnerFinalState(t *testing.T) {
	assert := assert.New(t)
	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.delay = time.Millisecond * 10
	s.spinFunc = spinTTY
	s.Start(baseText, true)
	time.Sleep(baseWait * time.Millisecond)
	s.Stop()
	assert.Greater(out.Len(), 0)

	outStr := out.String()
	assert.True(strings.HasSuffix(outStr, baseText+"...  \n"+showCursor))
}

func TestSpinnerDisabledShowDotsFlag(t *testing.T) {
	assert := assert.New(t)
	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.delay = time.Millisecond * 10
	s.spinFunc = spinTTY
	s.Start(baseText, false)
	time.Sleep(baseWait * time.Millisecond)
	s.Stop()
	assert.True(out.Len() > 0)

	outStr := out.String()
	assert.True(strings.HasPrefix(outStr, hideCursor+generateAllStatesAsString(t, baseText, false)))
	assert.True(strings.HasSuffix(outStr, baseText+"  \n"+showCursor))
}

func TestSpinnerInterruptWriter(t *testing.T) {
	assert := assert.New(t)

	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.spinFunc = spinTTY
	s.Start(baseText, false)
	time.Sleep(200 * time.Millisecond)
	_, err := s.Write([]byte("test"))
	assert.NoError(err)
	assert.True(strings.HasSuffix(out.String(), "test"))
}

func TestSpinNoTTY(t *testing.T) {
	assert := assert.New(t)

	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.spinFunc = spinNoTTY
	s.Start(baseText, true)
	time.Sleep(baseWait * time.Millisecond)
	s.Stop()
	assert.Greater(out.Len(), 0)
	assert.Equal(baseText+"...\n", out.String())
}

func generateAllStatesAsString(t *testing.T, text string, showDots bool) string {
	t.Helper()

	var builder strings.Builder

	for i := 0; i < len(spinnerStates); i++ {
		dotsState := ""
		if showDots {
			dotsState = dotsStates[i%len(dotsStates)]
		}
		builder.WriteString(fmt.Sprintf("\r%s %s%s", spinnerStates[i], text, dotsState))
	}

	return builder.String()
}
