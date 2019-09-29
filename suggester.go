package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/high-moctane/nwenc"
)

type Suggester struct {
	dataPath     string
	om           nwenc.OffsetMapper
	fOneGram     *os.File
	candidateLen int
}

func NewSuggester(dataPath string, candidateLen int) (*Suggester, error) {
	s := &Suggester{dataPath: dataPath}

	var err error
	s.fOneGram, err = os.Open(s.filePath(1, ""))
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", s.filePath(1, ""), err)
	}
	info, err := s.fOneGram.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot get file info: %w", err)
	}

	s.om = nwenc.NewCachedSeekOffsetMapper(s.fOneGram, info.Size())

	if candidateLen < 1 {
		return nil, fmt.Errorf("candidateLen must be positive int, but %d", candidateLen)
	}
	s.candidateLen = candidateLen

	return s, nil
}

func (sg *Suggester) Close() error {
	return sg.fOneGram.Close()
}

func (*Suggester) fileName(n int, prefix string) string {
	if n == 1 {
		return "1gram.txt"
	}
	return fmt.Sprintf("%dgram-%s", n, prefix)
}

func (sg *Suggester) filePath(n int, prefix string) string {
	return filepath.Join(sg.dataPath, sg.fileName(n, prefix))
}

func (sg *Suggester) Suggest(query string) (candidates []string, err error) {
	candidates = []string{}

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

	candidates = sg.uniqCandidates(candidates)
	candidates = sg.filterCandidates(candidates, prefix)
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

	// search head offset which prefix starts
	offset, err := sg.binSearch(f, info.Size(), []byte(prefix), []byte{'\n'})
	if err != nil {
		return
	}

	// read candidates
	sr := io.NewSectionReader(f, offset, info.Size()-offset)
	sc := bufio.NewScanner(sr)
	for i := 0; i < sg.candidateLen; i++ {
		sc.Scan()
		if sc.Err() != nil {
			err = sc.Err()
			return
		}
		if !strings.HasPrefix(sc.Text(), prefix) {
			break
		}
		candidates = append(candidates, sc.Text())
	}

	return
}

func (*Suggester) uniqCandidates(candidates []string) []string {
	res := []string{}
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
	res := make([]string, 0, len(candidates))
	for _, word := range candidates {
		if strings.HasPrefix(word, prefix) {
			res = append(res, word)
		}
	}
	return res
}

func (sg *Suggester) binSearch(r io.ReaderAt, size int64, query []byte, delim []byte) (offset int64, err error) {
	var left int64
	right := size

	for left <= right {
		mid := left + (right-left)/2

		offset, err = sg.findHeadOfLine(r, mid, delim)
		if err != nil {
			return
		}

		var b []byte
		b, err = sg.readBytes(r, offset, delim)
		if err != nil {
			return
		}

		cmp := sg.cmpBytes(query, b)
		if cmp < 0 {
			right = mid - 1
		} else if cmp > 0 {
			left = mid + 1
		} else {
			return
		}
	}

	offset, err = sg.findHeadOfLine(r, left, delim)
	if err != nil {
		return
	}

	return
}

func (sg *Suggester) findHeadOfLine(r io.ReaderAt, offset int64, delim []byte) (head int64, err error) {
	// Because the data is encoded in fixed codeword length coding,
	// offsets are 0 mod len(delim).
	// The initial value of head is a previous value from the offset.
	delimLen := int64(len(delim))
	for head = offset - offset%delimLen - delimLen; head > 0; head -= delimLen {
		buf := make([]byte, delimLen)
		if _, err = r.ReadAt(buf, head); err != nil {
			return
		}

		if sg.cmpBytes(buf, delim) == 0 {
			head += delimLen
			return
		}
	}
	return
}

func (*Suggester) readBytes(r io.ReaderAt, offset int64, delim []byte) (b []byte, err error) {

}
