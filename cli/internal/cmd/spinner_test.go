/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseWait = 100
	baseText = "Loading"
)

func TestSpinnerInitialState(t *testing.T) {
	assert := assert.New(t)

	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.delay = time.Millisecond * 10
	s.Start(baseText, true)
	time.Sleep(baseWait * time.Millisecond)
	s.Stop()
	assert.Greater(out.Len(), 0)

	outStr := out.String()
	assert.True(strings.HasPrefix(outStr, hideCursor+generateAllStatesAsString(t, baseText, true)))
}

func TestSpinnerFinalState(t *testing.T) {
	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.delay = time.Millisecond * 10
	s.Start(baseText, true)
	time.Sleep(baseWait * time.Millisecond)
	s.Stop()
	require.True(t, out.Len() > 0)

	outStr := out.String()
	require.True(t, strings.HasSuffix(outStr, baseText+"...  \n"+showCursor))
}

func TestSpinnerDisabledShowDotsFlag(t *testing.T) {
	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.delay = time.Millisecond * 10
	s.Start(baseText, false)
	time.Sleep(baseWait * time.Millisecond)
	s.Stop()
	require.True(t, out.Len() > 0)

	outStr := out.String()
	require.True(t, strings.HasPrefix(outStr, hideCursor+generateAllStatesAsString(t, baseText, false)))
	require.True(t, strings.HasSuffix(outStr, baseText+"  \n"+showCursor))
}

func TestSpinnerInterruptWriter(t *testing.T) {
	assert := assert.New(t)

	out := &bytes.Buffer{}

	s := newSpinner(out)
	s.Start(baseText, false)
	time.Sleep(200 * time.Millisecond)
	_, err := s.Write([]byte("test"))
	assert.NoError(err)
	assert.Equal(int32(1), atomic.LoadInt32(&s.stop))
	assert.True(strings.HasSuffix(out.String(), "test"))
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
