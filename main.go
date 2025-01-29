package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {
	var dictPath, binPath, pattern string
	var concurrency int

	flag.StringVar(&dictPath, "dictionary", "", "dictionary path")
	flag.StringVar(&binPath, "binary", "", "binary path")
	flag.StringVar(&pattern, "pattern", "Password correct!", "flag pattern")
	flag.IntVar(&concurrency, "concurrency", 4, "concurrency level")

	flag.Parse()

	if dictPath == "" || binPath == "" || pattern == "" {
		fmt.Println("Usage: qrack --dictionary <dict_path> --binary <binary_path> --pattern <flag_pattern> [--concurrency <level>]")
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
		fmt.Fprintf(os.Stderr, "Warning: Failed to count lines in dictionary, progress bar might be inaccurate: %v\n", err)
		totalPasswords = -1
	}
	_, err = dictFile.Seek(0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to seek to the beginning of dictionary file: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(dictFile)
	// Increased buffer size for passwordChan
	passwordChan := make(chan string, concurrency*10) // Increased buffer size
	foundChan := make(chan string)
	var wg sync.WaitGroup
	var attempts int64
	var processedPasswords int64

	// Initialize progress bar
	if totalPasswords > 0 {
		updateProgressBar(0, totalPasswords)
	} else {
		fmt.Printf("\rProgress: [Counting lines...] 0%%\r")
	}

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(binPath, pattern, passwordChan, foundChan, &wg, &attempts, totalPasswords, &processedPasswords)
	}

	go func() {
		for scanner.Scan() {
			passwordChan <- scanner.Text()
		}
		close(passwordChan)
	}()

	foundPassword := ""

	select {
	case foundPassword = <-foundChan:
		fmt.Printf("\nPassword found: %s\n", foundPassword)
		// No need to wait for workers to finish, password is found.
	case <-waitWorkers(&wg):
		if totalPasswords > 0 {
			updateProgressBar(totalPasswords, totalPasswords)
		} else {
			fmt.Printf("\rProgress: [??????????] 100%%\r\n")
		}

		if scanner.Err() != nil {
			fmt.Fprintf(os.Stderr, "failed to read dictionary file: %v\n", scanner.Err())
			os.Exit(1)
		}
		fmt.Println("\nPassword not found.")
	}

	if foundPassword != "" {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func waitWorkers(wg *sync.WaitGroup) <-chan bool {
	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

func worker(binPath, pattern string, passwordChan <-chan string, foundChan chan<- string, wg *sync.WaitGroup, attempts *int64, totalPasswords int, processedPasswords *int64) {
	defer wg.Done()
	for password := range passwordChan {
		atomic.AddInt64(attempts, 1)

		cmd := exec.Command(binPath, password)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			// log.Printf("failed to create stdout pipe: %v", err)
			continue
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			// log.Printf("failed to create stderr pipe: %v", err)
			continue
		}

		if err := cmd.Start(); err != nil {
			// log.Printf("failed to start command: %v", err)
			continue
		}

		// Buffer outputChan, errorChan, waitChan
		outputChan := make(chan []byte, 1) // Buffered channels
		errorChan := make(chan []byte, 1)  // Buffered channels
		waitChan := make(chan error, 1)    // Buffered channels
		var ioWaitGroup sync.WaitGroup

		ioWaitGroup.Add(1)
		go func() {
			defer ioWaitGroup.Done()
			outputBytes, readErr := io.ReadAll(stdout)
			if readErr != nil {
				// log.Printf("failed to read stdout: %v", readErr)
			}
			outputChan <- outputBytes
			close(outputChan) // Close outputChan after reading stdout
		}()

		ioWaitGroup.Add(1)
		go func() {
			defer ioWaitGroup.Done()
			errorBytes, readErr := io.ReadAll(stderr)
			if readErr != nil {
				// log.Printf("failed to read stderr: %v", readErr)
			}
			errorChan <- errorBytes
			close(errorChan) // Close errorChan after reading stderr
		}()

		go func() {
			defer close(waitChan) // Ensure waitChan is closed when cmd.Wait finishes
			waitErr := cmd.Wait()
			waitChan <- waitErr
		}()

		ioWaitGroup.Wait() // Wait for io.ReadAll goroutines to finish
		output := <-outputChan
		errorOutput := <-errorChan
		<-waitChan // Wait for cmd.Wait to finish (and channel to close)

		if strings.Contains(string(output), pattern) {
			fmt.Printf("\nPassword: %s\nStdout: %s\nStderr: %s\n", password, output, errorOutput)
			foundChan <- password
			return
		}
		atomic.AddInt64(processedPasswords, 1)
		if totalPasswords > 0 {
			updateProgressBar(int(atomic.LoadInt64(processedPasswords)), totalPasswords)
		} else if totalPasswords == -1 {
			updateIndeterminateProgressBar(int(atomic.LoadInt64(processedPasswords)))
		}
	}
}

func countLines(file *os.File) (int, error) {
	reader := bufio.NewReader(file)
	count := 0
	for {
		_, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

func updateProgressBar(current int, total int) {
	if total <= 0 {
		return
	}
	progress := float64(current) / float64(total)
	barLength := 30
	filledLength := int(progress * float64(barLength))
	bar := strings.Repeat("=", filledLength) + strings.Repeat(" ", barLength-filledLength)
	percentage := int(progress * 100)
	fmt.Printf("\rProgress: [%s] %d%% (%d/%d)", bar, percentage, current, total)
}

var indeterminateCounter int

func updateIndeterminateProgressBar(current int) {
	animationFrames := []string{"\\", "|", "/", "-"}
	frameIndex := indeterminateCounter % len(animationFrames)
	bar := animationFrames[frameIndex] + strings.Repeat(".", 9)
	percentage := "?"
	fmt.Printf("\rProgress: [%s] %s%% (Attempt %d)", bar, percentage, current)
	indeterminateCounter++
}
