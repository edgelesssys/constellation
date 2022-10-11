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
)

const (
	baseWait = 1
	baseText = "Loading"
)

func TestSpinnerInitialState(t *testing.T) {
	assert := assert.New(t)

	cmd := NewInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	s, _ := newSpinner(cmd, nil)
	s.Start(baseText, true)
	time.Sleep(baseWait * time.Second)
	s.Stop()
	assert.True(out.Len() > 0)
	assert.True(errOut.Len() == 0)

	outStr := out.String()
	assert.True(strings.HasPrefix(outStr, generateAllStatesAsString(baseText, true)))
}

func TestSpinnerFinalState(t *testing.T) {
	assert := assert.New(t)

	cmd := NewInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	s, _ := newSpinner(cmd, nil)
	s.Start(baseText, true)
	time.Sleep(baseWait * time.Second)
	s.Stop()
	assert.True(out.Len() > 0)
	assert.True(errOut.Len() == 0)

	outStr := out.String()
	assert.True(strings.HasSuffix(outStr, baseText+"...  \n"))
}

func TestSpinnerDisabledShowDotsFlag(t *testing.T) {
	assert := assert.New(t)

	cmd := NewInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	s, _ := newSpinner(cmd, nil)
	s.Start(baseText, false)
	time.Sleep(baseWait * time.Second)
	s.Stop()
	assert.True(out.Len() > 0)
	assert.True(errOut.Len() == 0)

	outStr := out.String()
	assert.True(strings.HasPrefix(outStr, generateAllStatesAsString(baseText, false)))
	assert.True(strings.HasSuffix(outStr, baseText+"  \n"))
}

func TestSpinnerInterruptWriter(t *testing.T) {
	assert := assert.New(t)

	cmd := NewInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	s, interruptWriter := newSpinner(cmd, &out)
	s.Start(baseText, false)
	time.Sleep(200 * time.Millisecond)
	_, err := interruptWriter.Write([]byte("test"))
	assert.NoError(err)
	assert.Equal(int32(1), atomic.LoadInt32(&s.stop))
	assert.True(strings.HasSuffix(out.String(), "test"))
}

func generateAllStatesAsString(text string, showDots bool) string {
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
