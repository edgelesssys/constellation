/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
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
}

type spinner struct {
	out   io.Writer
	delay time.Duration
	wg    *sync.WaitGroup
	stop  int32
}

func newSpinner(writer io.Writer) *spinner {
	return &spinner{
		out:   writer,
		wg:    &sync.WaitGroup{},
		delay: 100 * time.Millisecond,
		stop:  0,
	}
}

// Start starts the spinner using the given text.
func (s *spinner) Start(text string, showDots bool) {
	atomic.StoreInt32(&s.stop, 0)
	s.wg.Add(1)
	fmt.Fprint(s.out, hideCursor)

	go func() {
		defer s.wg.Done()

		for i := 0; ; i = (i + 1) % len(spinnerStates) {
			if atomic.LoadInt32(&s.stop) != 0 {
				break
			}
			dotsState := ""
			if showDots {
				dotsState = dotsStates[i%len(dotsStates)]
			}
			state := fmt.Sprintf("\r%s %s%s", spinnerStates[i], text, dotsState)
			fmt.Fprint(s.out, state)
			time.Sleep(s.delay)
		}

		dotsState := ""
		if showDots {
			dotsState = dotsStates[len(dotsStates)-1]
		}
		finalState := fmt.Sprintf("\r%s%s  ", text, dotsState)
		fmt.Fprintln(s.out, finalState)
	}()
}

// Stop stops the spinner.
func (s *spinner) Stop() {
	atomic.StoreInt32(&s.stop, 1)
	s.wg.Wait()
	fmt.Fprint(s.out, showCursor)
}

// Write stops the spinner and writes the given bytes to the underlying writer.
func (s *spinner) Write(p []byte) (n int, err error) {
	s.Stop()
	return s.out.Write(p)
}

type nopSpinner struct{}

func (s nopSpinner) Start(string, bool) {}
func (s nopSpinner) Stop()              {}
