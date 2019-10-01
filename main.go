package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// default env value for the data directory path.
const envDataPath = "NEXTWORD_DATA_PATH"

// flags
var dataPath = flag.String("data", os.Getenv(envDataPath), "path to data directory")
var maxCandidatesNum = flag.Int("candidates-num", 100, "max candidates num (positive int)")
var helpFlag = flag.Bool("h", false, "show this message")

func main() {
	os.Exit(run())
}

func run() int {
	if !parseArgs() {
		help := `The nextword prints the most likely English words that follow the stdin sentence.

The space character at the end of line plays an important role. If the line ends
with a space, the command show the next suggested words. However, if the line
ends with an alphabetic character, the suggested words start with the last word
of the line.

This command needs an external dataset. The dataset path should be set in an
environment value "$NEXTWORD_DATA_PATH". `

		flag.Usage()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, help)
		return 1
	}

	if err := serve(); err != nil {
		fmt.Fprintf(os.Stderr, "serve error: %v", err)
		return 1
	}
	return 0
}

func serve() error {
	sg := NewSuggester(*dataPath, *maxCandidatesNum)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		candidates, err := sg.Suggest(sc.Text())
		if err != nil {
			return fmt.Errorf("suggest error: %w", err)
		}
		fmt.Println(strings.Join(candidates, " "))
	}

	if sc.Err() != nil {
		return fmt.Errorf("read error: %w", sc.Err())
	}

	return nil
}

func parseArgs() bool {
	flag.Parse()
	if *helpFlag {
		return false
	}
	if *maxCandidatesNum < 1 {
		return false
	}
	if *dataPath == "" {
		return false
	}
	return true
}
