package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Suggester struct {
	dataPath     string
	candidateLen int
}

func NewSuggester(dataPath string, candidateLen int) *Suggester {
	return &Suggester{dataPath: dataPath, candidateLen: candidateLen}
}

func (*Suggester) fileName(n int, prefix string) string {
	if n == 1 {
		return "1gram.txt"
	}
	return fmt.Sprintf("%dgram-%s.txt", n, prefix)
}

func (sg *Suggester) filePath(n int, prefix string) string {
	return filepath.Join(sg.dataPath, sg.fileName(n, prefix))
}

func (sg *Suggester) Suggest(query string) (candidates []string, err error) {
	if query == "" {
		return
	}

	words, prefix := sg.parseQuery(query)

	// search n-gram in decscending order
	for i := 0; i < len(words); i++ {
		var cand []string
		cand, err = sg.suggestNgram(words[i:])
		if err != nil {
			return
		}
		candidates = append(candidates, cand...)
	}

	// search 1gram
	if prefix != "" {
		var cand []string
		cand, err = sg.suggest1gram(prefix)
		if err != nil {
			return
		}
		candidates = append(candidates, cand...)
	}

	// filter candidates
	candidates = sg.filterCandidates(candidates, prefix)
	candidates = sg.uniqCandidates(candidates)
	if len(candidates) > sg.candidateLen {
		candidates = candidates[:sg.candidateLen]
	}
	return
}

func (*Suggester) parseQuery(input string) (words []string, prefix string) {
	elems := strings.Split(input, " ")

	// If the end of the input is not " ", the last word in the input will be the prefix.
	if elems[len(elems)-1] != "" {
		prefix = elems[len(elems)-1]
		elems = elems[:len(elems)-1]
	}

	// collect up to last 4 words
	words = []string{}
	for i := len(elems) - 1; i >= 0; i-- {
		if elems[i] == "" {
			continue
		}
		words = append([]string{elems[i]}, words...)
		if len(words) >= 4 {
			break
		}
	}

	return
}

func (sg *Suggester) suggestNgram(words []string) (candidates []string, err error) {
	// open data
	n := len(words) + 1
	initial := strings.ToLower(string([]rune(words[0])[0]))
	f, err := os.Open(sg.filePath(n, initial))
	if err != nil {
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return
	}

	// search for a head offset which the query starts
	query := strings.Join(words, " ") + "\t"
	offset, err := sg.binSearch(f, info.Size(), query)
	if err != nil {
		return
	}

	entry, err := sg.readLine(f, offset, info.Size())
	if err != nil {
		return
	}
	if !strings.HasPrefix(entry, query) {
		// no matching
		return
	}

	candidates = strings.Split(strings.Split(entry, "\t")[1], " ")
	return
}

func (sg *Suggester) suggest1gram(prefix string) (candidates []string, err error) {
	// open 1gram file
	f, err := os.Open(sg.filePath(1, ""))
	if err != nil {
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return
	}

	// search for a head offset which the prefix starts
	offset, err := sg.binSearch(f, info.Size(), prefix)
	if err != nil {
		return
	}

	// read candidates
	sr := io.NewSectionReader(f, offset, info.Size()-offset)
	sc := bufio.NewScanner(sr)
	for i := 0; i < sg.candidateLen; i++ {
		if sc.Scan() {
			if !strings.HasPrefix(sc.Text(), prefix) {
				break
			}
			candidates = append(candidates, sc.Text())
		}
		if sc.Err() != nil {
			err = sc.Err()
			return
		}
	}

	return
}

func (*Suggester) uniqCandidates(candidates []string) []string {
	var res []string
	set := map[string]bool{} // set ot candidates

	for _, word := range candidates {
		if set[word] {
			continue
		}
		res = append(res, word)
		set[word] = true
	}

	return res
}

func (*Suggester) filterCandidates(candidates []string, prefix string) []string {
	var res []string
	for _, word := range candidates {
		if strings.HasPrefix(word, prefix) {
			res = append(res, word)
		}
	}
	return res
}

func (sg *Suggester) binSearch(r io.ReaderAt, size int64, query string) (offset int64, err error) {
	var left int64
	right := size

	for left <= right {
		mid := left + (right-left)/2

		offset, err = sg.findHeadOfLine(r, mid)
		if err != nil {
			return
		}

		var line string
		line, err = sg.readLine(r, offset, size)
		if err != nil {
			return
		}

		if query < line {
			right = mid - 1
		} else if query > line {
			left = mid + 1
		} else {
			return
		}
	}

	offset, err = sg.findHeadOfLine(r, left)
	if err != nil {
		return
	}

	return
}

func (sg *Suggester) findHeadOfLine(r io.ReaderAt, offset int64) (head int64, err error) {
	// The initial value of head is a previous value from the offset.
	for head = offset - 1; ; head-- {
		if head <= 0 {
			head = 0
			return
		}

		buf := make([]byte, 1)
		if _, err = r.ReadAt(buf, head); err != nil {
			return
		}

		if buf[0] == '\n' {
			head++
			return
		}
	}
}

func (*Suggester) readLine(r io.ReaderAt, offset, size int64) (line string, err error) {
	sr := io.NewSectionReader(r, offset, size-offset)
	sc := bufio.NewScanner(sr)
	if sc.Scan() {
		line = sc.Text()
	}
	if sc.Err() != nil {
		err = sc.Err()
		return
	}
	return
}
