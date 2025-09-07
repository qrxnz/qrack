package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle      = lipgloss.NewStyle().Margin(1, 2)
	foundStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	notFoundStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	quitTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type model struct {
	totalPasswords     int
	processedPasswords int64
	foundPassword      string
	err                error
	progress           progress.Model
	quitting           bool
	width              int
	height             int
}

type progressMsg struct{ processed int64 }
type foundMsg struct{ password string }
type errorMsg struct{ err error }

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10
		return m, nil

	case progressMsg:
		atomic.StoreInt64(&m.processedPasswords, msg.processed)
		percent := float64(msg.processed) / float64(m.totalPasswords)
		if percent > 1.0 {
			percent = 1.0
		}

		// Create the command to set the percentage
		cmd := m.progress.SetPercent(percent)

		// If there's a command, execute it and update the progress model
		if cmd != nil {
			// Get the message from the command
			msgForProgressBar := cmd()
			// Update the progress bar model
			updatedProgressModel, newCmd := m.progress.Update(msgForProgressBar)
			// Assign the new model back
			m.progress = updatedProgressModel.(progress.Model)
			// Return the new command
			return m, newCmd
		}

		return m, nil

	case foundMsg:
		m.foundPassword = msg.password
		m.quitting = true
		return m, tea.Quit

	case errorMsg:
		m.err = msg.err
		m.quitting = true
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.quitting {
		var s string
		if m.err != nil {
			s = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		} else if m.foundPassword != "" {
			s = foundStyle.Render(fmt.Sprintf("Password found: %s", m.foundPassword))
		} else {
			s = notFoundStyle.Render("Password not found.")
		}
		return docStyle.Render(s)
	}

	return docStyle.Render(
		"Qrack is running...\n\n" +
			m.progress.View() + "\n\n" +
			fmt.Sprintf("Processed: %d/%d", atomic.LoadInt64(&m.processedPasswords), m.totalPasswords) + "\n\n" +
			quitTextStyle.Render("(Press 'q' to quit)"),
	)
}


func main() {
	var dictPath, binPath, pattern string
	var concurrency int

	flag.StringVar(&dictPath, "dictionary", "", "dictionary path")
	flag.StringVar(&binPath, "binary", "", "binary path")
	flag.StringVar(&pattern, "pattern", "Password correct!", "flag pattern")
	flag.IntVar(&concurrency, "concurrency", 4, "concurrency level")
	flag.Parse()

	if dictPath == "" || binPath == "" {
		fmt.Println("Usage: qrack --dictionary <dict_path> --binary <binary_path> [--pattern <flag_pattern>] [--concurrency <level>]")
		os.Exit(1)
	}

	dictFile, err := os.Open(dictPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open dictionary file: %v\n", err)
		os.Exit(1)
	}
	defer dictFile.Close()

	totalPasswords, err := countLines(dictFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to count lines in dictionary: %v\n", err)
		os.Exit(1)
	}
	_, err = dictFile.Seek(0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to seek to the beginning of dictionary file: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model{
		totalPasswords: totalPasswords,
		progress:       progress.New(progress.WithDefaultGradient()),
	})

	go runCracker(p, dictPath, binPath, pattern, concurrency)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func runCracker(p *tea.Program, dictPath, binPath, pattern string, concurrency int) {
	dictFile, err := os.Open(dictPath)
	if err != nil {
		p.Send(errorMsg{err})
		return
	}
	defer dictFile.Close()

	scanner := bufio.NewScanner(dictFile)
	passwordChan := make(chan string, concurrency*10)
	foundChan := make(chan string, 1)
	var wg sync.WaitGroup
	var processedPasswords int64
	var foundOnce, errorOnce sync.Once
	const progressBatchSize = 100
	bytePattern := []byte(pattern)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for password := range passwordChan {
				select {
				case <-foundChan:
					return
				default:
				}

				cmd := exec.Command(binPath, password)
				output, err := cmd.CombinedOutput()
				if err != nil {
					errorOnce.Do(func() {
						p.Send(errorMsg{fmt.Errorf("cmd execution for password '%s' failed: %w, output: %s", password, err, string(output))})
						close(foundChan) // Signal other workers to stop
					})
					return
				}

				if bytes.Contains(output, bytePattern) {
					foundOnce.Do(func() {
						p.Send(foundMsg{password: password})
						close(foundChan)
					})
					return
				}

				processed := atomic.AddInt64(&processedPasswords, 1)
				if processed%progressBatchSize == 0 {
					p.Send(progressMsg{processed: processed})
				}
			}
		}()
	}

	go func() {
		for scanner.Scan() {
			passwordChan <- scanner.Text()
		}
		close(passwordChan)
	}()

	wg.Wait()

	// Send final status update after all workers are done
	foundOnce.Do(func() {
		// Ensure the progress bar reaches 100%
		p.Send(progressMsg{processed: atomic.LoadInt64(&processedPasswords)})
		p.Send(foundMsg{password: ""}) // Not found
	})
}

func waitWorkers(wg *sync.WaitGroup) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

func countLines(file *os.File) (int, error) {
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}