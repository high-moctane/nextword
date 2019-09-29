package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const EnvDataPath = "NEXTWORD_DATA_PATH"

func TestSuggester_Suggest(t *testing.T) {
	tests := []struct {
		candidatesLen int
		query         string
		candidates    []string
		err           error
	}{
		{
			10,
			"",
			nil,
			nil,
		},
		{
			10,
			"EDM ",
			[]string{"is", "andprocess", "instruments", "instrument"},
			nil,
		},
		{
			10,
			"EDM",
			[]string{"EDM", "EDMA", "EDMAN", "EDMD", "EDMK", "EDMOND", "EDMONDS",
				"EDMONDSON", "EDMONSON", "EDMONSTON"},
			nil,
		},
	}

	for idx, test := range tests {
		sg, err := NewSuggester(os.Getenv(EnvDataPath), test.candidatesLen)
		if err != nil {
			t.Fatal(err)
		}

		cand, err := sg.Suggest(test.query)
		if err != nil {
			t.Errorf("[%d] expected %v, but got %v", idx, test.err, err)
		}
		if err != nil {
			continue
		}
		if !reflect.DeepEqual(test.candidates, cand) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.candidates, cand)
		}

		sg.Close()
	}
}

func TestSuggester_ParseInput(t *testing.T) {
	tests := []struct {
		query  string
		words  []string
		prefix string
	}{
		{
			"abc",
			[]string{},
			"abc",
		},
		{
			"abc ",
			[]string{"abc"},
			"",
		},
		{
			"abc def ",
			[]string{"abc", "def"},
			"",
		},
		{
			"abc def g",
			[]string{"abc", "def"},
			"g",
		},
		{
			"abc def ghi jkl ",
			[]string{"abc", "def", "ghi", "jkl"},
			"",
		},
		{
			"abc def ghi jkl mno ",
			[]string{"def", "ghi", "jkl", "mno"},
			"",
		},
		{
			"abc def ghi jkl mno pqr",
			[]string{"def", "ghi", "jkl", "mno"},
			"pqr",
		},
	}

	for idx, test := range tests {
		s := new(Suggester)
		words, prefix := s.parseQuery(test.query)
		if !reflect.DeepEqual(test.words, words) {
			t.Errorf("[%d] expect %v, but got %v", idx, test.words, words)
		}
		if test.prefix != prefix {
			t.Errorf("[%d] expected %s, but got %v", idx, test.prefix, prefix)
		}
	}
}

func TestSuggester_SuggestNgram(t *testing.T) {
	t.FailNow()
}

func TestSuggester_Suggest1gram(t *testing.T) {
	tests := []struct {
		candidatesLen int
		prefix        string
		candidates    []string
		err           error
	}{
		{
			10,
			"pound",
			[]string{"pound", "pound.1", "pounda", "poundage", "poundages",
				"poundal", "poundals", "poundcake", "pounde", "pounded"},
			nil,
		},
	}

	for idx, test := range tests {
		sg, err := NewSuggester(os.Getenv(EnvDataPath), test.candidatesLen)
		if err != nil {
			t.Fatalf("[%d] cannot create Suggester: %v", idx, err)
		}

		cand, err := sg.suggest1gram(test.prefix)
		if !reflect.DeepEqual(test.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.err, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(test.candidates, cand) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.candidates, cand)
		}

		sg.Close()
	}
}

func TestSuggester_UniqCandidates(t *testing.T) {
	tests := []struct {
		in, out []string
	}{
		{
			[]string{},
			[]string{},
		},
		{
			[]string{"abc"},
			[]string{"abc"},
		},
		{
			[]string{"abc", "abc"},
			[]string{"abc"},
		},
		{
			[]string{"abc", "def"},
			[]string{"abc", "def"},
		},
		{
			[]string{"abc", "def", "abc"},
			[]string{"abc", "def"},
		},
		{
			[]string{"abc", "def", "abc", "def", "abc"},
			[]string{"abc", "def"},
		},
	}

	for idx, test := range tests {
		s := new(Suggester)
		out := s.uniqCandidates(test.in)
		if !reflect.DeepEqual(test.out, out) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out, out)
		}
	}
}

func TestSuggester_FilterCandidates(t *testing.T) {
	tests := []struct {
		cand   []string
		prefix string
		out    []string
	}{
		{
			[]string{},
			"",
			[]string{},
		},
		{
			[]string{},
			"prefix",
			[]string{},
		},
		{
			[]string{"abc"},
			"",
			[]string{"abc"},
		},
		{
			[]string{"abc"},
			"ab",
			[]string{"abc"},
		},
		{
			[]string{"abc"},
			"ae",
			[]string{},
		},
		{
			[]string{"abc", "bcd", "abd", "abe", "ade", "absent"},
			"ab",
			[]string{"abc", "abd", "abe", "absent"},
		},
	}

	for idx, test := range tests {
		s := new(Suggester)
		out := s.filterCandidates(test.cand, test.prefix)
		if !reflect.DeepEqual(test.out, out) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out, out)
		}
	}
}

func TestSuggester_BinSearch(t *testing.T) {
	dataPath := os.Getenv(EnvDataPath)

	tests := []struct {
		filePath string
		query    []byte
		delim    []byte
		offset   int64
	}{
		{
			filepath.Join(dataPath, "1gram.txt"),
			[]byte("A"),
			[]byte{'\n'},
			0,
		},
		{
			filepath.Join(dataPath, "1gram.txt"),
			[]byte("Signifi"),
			[]byte{'\n'},
			5423021,
		},
		{
			filepath.Join(dataPath, "1gram.txt"),
			[]byte("zzzzzz"),
			[]byte{'\n'},
			11346353,
		},
		{
			filepath.Join(dataPath, "1gram.txt"),
			[]byte("LIPKINAA"),
			[]byte{'\n'},
			34242490,
		},
		{
			filepath.Join(dataPath, "1gram.txt"),
			[]byte("dna0"),
			[]byte{'\n'},
			7736920,
		},
		{
			filepath.Join(dataPath, "5gram-q"),
			[]byte{0x46, 0xAD, 0x24},
			[]byte{0xFF, 0xFF, 0xFF},
			0,
		},
		{
			filepath.Join(dataPath, "5gram-q"),
			[]byte{0x99, 0x68, 0x56, 0xA5},
			[]byte{0xFF, 0xFF, 0xFF},
			0x6ABF,
		},
		{
			filepath.Join(dataPath, "5gram-q"),
			[]byte{0x99, 0x47, 0xA5},
			[]byte{0xFF, 0xFF, 0xFF},
			0x16E84,
		},
	}

	sg := new(Suggester)
	for idx, test := range tests {
		func() {
			f, err := os.Open(test.filePath)
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}
			defer f.Close()
			info, err := f.Stat()
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}

			offset, err := sg.binSearch(f, info.Size(), test.query, test.delim)
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}

			if test.offset != offset {
				t.Errorf("[%d] expected %d, but got %d", idx, test.offset, offset)
			}
		}()
	}
}

func TestSuggester_FindHeadOfLine(t *testing.T) {
	dataPath := os.Getenv(EnvDataPath)

	tests := []struct {
		filePath string
		offset   int64
		delim    []byte
		head     int64
	}{
		{
			filepath.Join(dataPath, "1gram.txt"),
			15651,
			[]byte{'\n'},
			15643,
		},
		{
			filepath.Join(dataPath, "1gram.txt"),
			6503483,
			[]byte{'\n'},
			6503483,
		},
		{
			filepath.Join(dataPath, "1gram.txt"),
			407317,
			[]byte{'\n'},
			407314,
		},
		{
			filepath.Join(dataPath, "5gram-q"),
			0xC2FA,
			[]byte{'\n'},
			0xC2EE,
		},
		{
			filepath.Join(dataPath, "5gram-q"),
			0xC339,
			[]byte{'\n'},
			0xC348,
		},
	}

	sg := new(Suggester)
	for idx, test := range tests {
		func() {
			f, err := os.Open(test.filePath)
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}
			defer f.Close()

			head, err := sg.findHeadOfLine(f, test.offset, test.delim)
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}

			if test.head != head {
				t.Errorf("[%d] expected %d, but got %d", idx, test.head, head)
			}
		}()
	}
}
