package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	label string
	stop  chan struct{}
	done  sync.WaitGroup
}

func NewSpinner(label string) *Spinner {
	return &Spinner{
		label: label,
		stop:  make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	s.done.Add(1)
	go func() {
		defer s.done.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				return
			default:
				frame := spinnerFrames[i%len(spinnerFrames)]
				fmt.Printf("\r  %s %s", style.Render(frame), s.label)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop(success bool) {
	close(s.stop)
	s.done.Wait()
	fmt.Print("\r\033[2K")
	if success {
		fmt.Printf("  %s %s\n", SuccessStyle.Render("✓"), s.label)
	} else {
		fmt.Printf("  %s %s\n", ErrorStyle.Render("✗"), s.label)
	}
}
