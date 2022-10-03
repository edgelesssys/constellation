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

	"github.com/stretchr/testify/require"
)

const (
	baseWait = 1
	baseText = "Loading"
)

func TestSpinnerInitialState(t *testing.T) {
	// Command
	cmd := NewInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	s := NewSpinner(cmd, baseText, true)
	s.Start()
	time.Sleep(baseWait * time.Second)
	s.Stop()
	require.True(t, out.Len() > 0)
	require.True(t, errOut.Len() == 0)

	outStr := out.String()
	require.True(t, strings.HasPrefix(outStr, generateAllStatesAsString(baseText, true)))
}

func TestSpinnerFinalState(t *testing.T) {
	// Command
	cmd := NewInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	s := NewSpinner(cmd, baseText, true)
	s.Start()
	time.Sleep(baseWait * time.Second)
	s.Stop()
	require.True(t, out.Len() > 0)
	require.True(t, errOut.Len() == 0)

	outStr := out.String()
	require.True(t, strings.HasSuffix(outStr, baseText+"...\n"))
}

func TestSpinnerDisabledShowDotsFlag(t *testing.T) {
	// Command
	cmd := NewInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	s := NewSpinner(cmd, baseText, false)
	s.Start()
	time.Sleep(baseWait * time.Second)
	s.Stop()
	require.True(t, out.Len() > 0)
	require.True(t, errOut.Len() == 0)

	outStr := out.String()
	require.True(t, strings.HasPrefix(outStr, generateAllStatesAsString(baseText, false)))
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
