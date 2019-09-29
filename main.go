package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const dataPathEnv = "NEXTWORD_DATA_PATH"

func main() {
	os.Exit(run())
}

func run() int {
	if err := serve(); err != nil {
		fmt.Fprintf(os.Stderr, "serve error: %v", err)
		return 1
	}
	return 0
}

func serve() error {
	sg := NewSuggester(os.Getenv(dataPathEnv), 100)

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
