/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
)

var (
	spinnerStates = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	dotsStates    = []string{".", "..", "..."}
)

type spinner struct {
	out      *cobra.Command
	text     string
	showDots bool
	delay    time.Duration
	wg       *sync.WaitGroup
	stop     int32
}

func newSpinner(c *cobra.Command, text string, showDots bool) *spinner {
	return &spinner{
		out:      c,
		text:     text,
		showDots: showDots,
		wg:       &sync.WaitGroup{},
		delay:    100 * time.Millisecond,
		stop:     0,
	}
}

func (s *spinner) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for i := 0; ; i = (i + 1) % len(spinnerStates) {
			if atomic.LoadInt32(&s.stop) != 0 {
				break
			}
			dotsState := ""
			if s.showDots {
				dotsState = dotsStates[i%len(dotsStates)]
			}
			state := fmt.Sprintf("\r%s %s%s", spinnerStates[i], s.text, dotsState)
			s.out.Print(state)
			time.Sleep(s.delay)
		}

		dotsState := ""
		if s.showDots {
			dotsState = dotsStates[len(dotsStates)-1]
		}
		finalState := fmt.Sprintf("\r%s%s  ", s.text, dotsState)
		s.out.Println(finalState)
	}()
}

func (s *spinner) Stop() {
	atomic.StoreInt32(&s.stop, 1)
	s.wg.Wait()
}
