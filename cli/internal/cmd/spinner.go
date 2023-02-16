/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	tty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

const (
	// hideCursor and showCursor are ANSI escape sequences to hide and show the cursor.
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
)

var (
	spinnerStates = []string{"⣷", "⣯", "⣟", "⡿", "⢿", "⣻", "⣽", "⣾"}
	dotsStates    = []string{".  ", ".. ", "..."}
)

type spinnerInterf interface {
	Start(text string, showDots bool)
	Stop()
	io.Writer
}

type spinner struct {
	out      io.Writer
	delay    time.Duration
	wg       *sync.WaitGroup
	stop     *atomic.Bool
	spinFunc func(out io.Writer, wg *sync.WaitGroup, stop *atomic.Bool, delay time.Duration, text string, showDots bool)
}

func newSpinnerOrStderr(cmd *cobra.Command) (spinnerInterf, error) {
	debug, err := cmd.Flags().GetBool("debug")
	noSpinner := os.Getenv(constants.EnvVarNoSpinner)
	if err != nil {
		return nil, err
	}
	if debug || noSpinner != "" {
		return &nopSpinner{cmd.ErrOrStderr()}, nil
	}
	return newSpinner(cmd.ErrOrStderr()), nil
}

func newSpinner(writer io.Writer) *spinner {
	s := &spinner{
		out:   writer,
		wg:    &sync.WaitGroup{},
		delay: time.Millisecond * 100,
		stop:  &atomic.Bool{},
	}

	s.spinFunc = spinTTY

	if !(writer == os.Stderr && tty.IsTerminal(os.Stderr.Fd())) {
		s.spinFunc = spinNoTTY
	}

	return s
}

// Start starts the spinner using the given text.
func (s *spinner) Start(text string, showDots bool) {
	s.wg.Add(1)

	go s.spinFunc(s.out, s.wg, s.stop, s.delay, text, showDots)
}

// Stop stops the spinner.
func (s *spinner) Stop() {
	s.stop.Store(true)
	s.wg.Wait()
}

// Write stops the spinner and writes the given bytes to the underlying writer.
func (s *spinner) Write(p []byte) (n int, err error) {
	s.Stop()
	return s.out.Write(p)
}

func spinTTY(out io.Writer, wg *sync.WaitGroup, stop *atomic.Bool, delay time.Duration, text string, showDots bool) {
	defer wg.Done()

	fmt.Fprint(out, hideCursor)

	for i := 0; ; i = (i + 1) % len(spinnerStates) {
		if stop.Load() {
			break
		}
		dotsState := ""
		if showDots {
			dotsState = dotsStates[i%len(dotsStates)]
		}
		state := fmt.Sprintf("\r%s %s%s", spinnerStates[i], text, dotsState)
		fmt.Fprint(out, state)
		time.Sleep(delay)
	}
	dotsState := ""
	if showDots {
		dotsState = dotsStates[len(dotsStates)-1]
	}
	finalState := fmt.Sprintf("\r%s%s  ", text, dotsState)
	fmt.Fprintln(out, finalState)
	fmt.Fprint(out, showCursor)
}

func spinNoTTY(out io.Writer, wg *sync.WaitGroup, _ *atomic.Bool, _ time.Duration, text string, _ bool) {
	defer wg.Done()
	fmt.Fprintln(out, text+"...")
}

type nopSpinner struct {
	io.Writer
}

func (s *nopSpinner) Start(string, bool) {}
func (s *nopSpinner) Stop()              {}
func (s *nopSpinner) Write(p []byte) (n int, err error) {
	return s.Writer.Write(p)
}
