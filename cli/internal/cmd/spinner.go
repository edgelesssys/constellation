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

	"github.com/spf13/cobra"
)

var (
	spinnerStates = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	dotsStates    = []string{".", "..", "..."}
)

type spinnerInterf interface {
	Start(text string, showDots bool)
	Stop()
}

type spinner struct {
	out   *cobra.Command
	delay time.Duration
	wg    *sync.WaitGroup
	stop  int32
}

func newSpinner(c *cobra.Command, writer io.Writer) (*spinner, *interruptSpinWriter) {
	spinner := &spinner{
		out:   c,
		wg:    &sync.WaitGroup{},
		delay: 100 * time.Millisecond,
		stop:  0,
	}

	if writer != nil {
		interruptWriter := &interruptSpinWriter{
			writer:  writer,
			spinner: spinner,
		}
		return spinner, interruptWriter
	}

	return spinner, nil
}

func (s *spinner) Start(text string, showDots bool) {
	s.wg.Add(1)
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
			s.out.Print(state)
			time.Sleep(s.delay)
		}

		dotsState := ""
		if showDots {
			dotsState = dotsStates[len(dotsStates)-1]
		}
		finalState := fmt.Sprintf("\r%s%s  ", text, dotsState)
		s.out.Println(finalState)
	}()
}

func (s *spinner) Stop() {
	atomic.StoreInt32(&s.stop, 1)
	s.wg.Wait()
}

type interruptSpinWriter struct {
	spinner *spinner
	writer  io.Writer
}

func (w *interruptSpinWriter) Write(p []byte) (n int, err error) {
	w.spinner.Stop()
	return w.writer.Write(p)
}

type nopSpinner struct{}

func (s nopSpinner) Start(string, bool) {}
func (s nopSpinner) Stop()              {}
