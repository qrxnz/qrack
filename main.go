package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cheggaaa/pb/v3"
)

func main() {
	var dictPath, binPath, pattern string

	flag.StringVar(&dictPath, "dictionary", "", "dictionary path")
	flag.StringVar(&binPath, "binary", "", "binary path")
	flag.StringVar(&pattern, "pattern", "", "flag pattern")

	flag.Parse()

	dictFile, err := os.Open(dictPath)
	if err != nil {
		log.Fatalf("failed to open dictionary file: %v", err)
	}
	defer dictFile.Close()

	// count lines in the file for progress bar
	numLines, err := countLines(dictFile)
	if err != nil {
		log.Fatalf("failed to count lines in dictionary file: %v", err)
	}

	// reset file pointer to start of file
	if _, err := dictFile.Seek(0, 0); err != nil {
		log.Fatalf("failed to seek to start of dictionary file: %v", err)
	}

	scanner := bufio.NewScanner(dictFile)

	// set up progress bar
	bar := pb.StartNew(numLines)
	bar.SetMaxWidth(80)

	for scanner.Scan() {
		input := scanner.Text()

		cmd := exec.Command(binPath)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalf("failed to create stdin pipe: %v", err)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("failed to create stdout pipe: %v", err)
		}

		if err := cmd.Start(); err != nil {
			log.Fatalf("failed to start command: %v", err)
		}

		_, err = io.WriteString(stdin, input)
		if err != nil {
			log.Fatalf("failed to write to stdin: %v", err)
		}

		if err := stdin.Close(); err != nil {
			log.Fatalf("failed to close stdin pipe: %v", err)
		}

		output, err := io.ReadAll(stdout)
		if err != nil {
			log.Fatalf("failed to read from stdout: %v", err)
		}

		if err := cmd.Wait(); err != nil {
			log.Fatalf("command failed: %v", err)
		}

		if strings.Contains(string(output), pattern) {
			fmt.Printf("\nPassword: %s\n%s\n", input, output)
			os.Exit(0)
		}

		bar.Increment()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to read dictionary file: %v", err)
	}

	bar.Finish()
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
