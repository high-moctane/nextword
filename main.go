package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const dataPathEnv = "NEXTWORD_DATA_PATH"

func main() {
	log.Fatal(run())
}

func run() error {
	return fmt.Errorf("serve error: %w", serve())
}

func serve() error {
	sg, err := NewSuggester(os.Getenv(dataPathEnv), 100)
	if err != nil {
		return fmt.Errorf("cannot create suggester: %w", err)
	}
	defer sg.Close()

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
