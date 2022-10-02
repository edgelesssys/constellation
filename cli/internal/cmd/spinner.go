package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"sync"
	"sync/atomic"
	"time"
)

var spinnerStates = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
var dotsStates = []string{".", "..", "..."}

type Spinner struct {
	out      *cobra.Command
	text     string
	showDots bool
	delay    time.Duration
	wg       *sync.WaitGroup
	stop     int32
}

func NewSpinner(c *cobra.Command, text string, showDots bool) *Spinner {
	return &Spinner{out: c,
		text:     text,
		showDots: showDots,
		wg:       &sync.WaitGroup{},
		delay:    100 * time.Millisecond,
		stop:     0,
	}
}

func (s *Spinner) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

	out:
		for {
			for i := 0; i < len(spinnerStates); i++ {
				if atomic.LoadInt32(&s.stop) == 0 {
					dotsState := ""
					if s.showDots {
						dotsState = dotsStates[i%len(dotsStates)]
					}
					state := fmt.Sprintf("\r%s %s%s", spinnerStates[i], s.text, dotsState)
					s.out.Print(state)
					time.Sleep(s.delay)
				} else {
					break out
				}
			}
		}

		dotsState := ""
		if s.showDots {
			dotsState = dotsStates[len(dotsStates)-1]
		}
		finalState := fmt.Sprintf("\r%s%s", s.text, dotsState)
		s.out.Println(finalState)
	}()
}

func (s *Spinner) Stop() {
	atomic.StoreInt32(&s.stop, 1)
	s.wg.Wait()
}
