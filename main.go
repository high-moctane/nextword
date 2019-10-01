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
		flag.Usage()
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
