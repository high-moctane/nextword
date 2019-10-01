package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var EnvDataPath = os.Getenv("NEXTWORD_DATA_PATH")

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
			[]string{"is", "and", "process", "instruments", "instrument"},
			nil,
		},
		{
			10,
			"EDM",
			[]string{"EDM", "EDMA", "EDMAN", "EDMD", "EDMK", "EDMOND", "EDMONDS",
				"EDMONDSON", "EDMONSON", "EDMONSTON"},
			nil,
		},
		{
			20,
			"you could not buy ",
			[]string{
				"a", "the", "it", "them", "anything",
				"any", "or", "him", "his", "that",
				"one", "in", "her", "land", "food",
				"their", "into", "from", "this", "me",
			},
			nil,
		},
		{
			20,
			"you could not buy ",
			[]string{
				"a", "the", "it", "them", "anything",
				"any", "or", "him", "his", "that",
				"one", "in", "her", "land", "food",
				"their", "into", "from", "this", "me",
			},
			nil,
		},
		{
			20,
			"may be until day ",
			[]string{
				"after", "of", "and", "in", "to",
				"the", "or", "I", "when", "for",
				"he", "before", "was", "is", "that",
				"at", "on", "by", "with", "we",
			},
			nil,
		},
		{
			15,
			"just for a few m",
			[]string{
				"minutes", "moments", "months", "more", "miles",
				"minor", "men", "million", "milliseconds", "members",
				"m", "m!", "m'", "m'1", "m'2",
			},
			nil,
		},
		{
			20,
			"aaaaaaaaa bbbbbbbbbb ccccccccccc dddddddddd eeeeeeeeaaa",
			nil,
			nil,
		},
	}

	for idx, test := range tests {
		sg := NewSuggester(EnvDataPath, test.candidatesLen)

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
	}
}

func BenchmarkSuggester_Suggest(b *testing.B) {
	sg := NewSuggester(EnvDataPath, 100)

	for i := 0; i < b.N; i++ {
		sg.Suggest("The quick brown fox ju")
	}
}

func TestSuggester_ParseQuery(t *testing.T) {
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
	tests := []struct {
		words      []string
		candidates []string
	}{
		{
			[]string{"objectivation"},
			[]string{"of", "and"},
		},
		{
			[]string{"committee", "feels"},
			[]string{"that"},
		},
		{
			[]string{"these", "steps", "are"},
			[]string{"taken", "not", "completed"},
		},
		{
			[]string{"Have", "you", "seen", "or"},
			[]string{"heard"},
		},
		{
			[]string{"brousa"},
			nil,
		},
		{
			[]string{"I felt paint"},
			nil,
		},
		{
			[]string{"0000000000"},
			nil,
		},
		{
			[]string{"🤔"},
			nil,
		},
	}

	for idx, test := range tests {
		sg := NewSuggester(EnvDataPath, 100)
		cand, err := sg.suggestNgram(test.words)
		if err != nil {
			t.Errorf("[%d] unexpected error: %v", idx, err)
			continue
		}
		if !reflect.DeepEqual(test.candidates, cand) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.candidates, cand)
		}
	}
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
		{
			5,
			"Kumu",
			[]string{"Kumu", "Kumud", "Kumuda", "Kumudini", "Kumuhonua"},
			nil,
		},
		{
			5,
			"NT/2000",
			[]string{"NT/2000", "NT/2000/XP"},
			nil,
		},
		{
			5,
			"Sehoraaaaaaaa",
			nil,
			nil,
		},
	}

	for idx, test := range tests {
		sg := NewSuggester(EnvDataPath, test.candidatesLen)

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
	}
}

func TestSuggester_UniqCandidates(t *testing.T) {
	tests := []struct {
		in, out []string
	}{
		{
			[]string{},
			nil,
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
			nil,
		},
		{
			[]string{},
			"prefix",
			nil,
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
			nil,
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
	tests := []struct {
		filePath string
		query    string
		offset   int64
	}{
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			"A",
			0,
		},
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			"zł",
			11346418,
		},
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			"Recu",
			4901265,
		},
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			"Latemaa",
			3303588,
		},
		{
			filepath.Join(EnvDataPath, "2gram-e.txt"),
			"ELSE",
			24641,
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

			offset, err := sg.binSearch(f, info.Size(), test.query)
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
	tests := []struct {
		filePath string
		offset   int64
		head     int64
	}{
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			0,
			0,
		},
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			3,
			2,
		},
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			30749,
			30734,
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

			head, err := sg.findHeadOfLine(f, test.offset)
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

func TestSuggester_ReadLine(t *testing.T) {
	tests := []struct {
		filePath string
		offset   int64
		line     string
	}{
		{
			filepath.Join(EnvDataPath, "1gram.txt"),
			0,
			"A",
		},
		{
			filepath.Join(EnvDataPath, "2gram-e.txt"),
			13617,
			"EDUCA\tTION",
		},
	}

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

			sg := new(Suggester)
			line, err := sg.readLine(f, test.offset, info.Size())
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}

			if test.line != line {
				t.Errorf("[%d] expected %#v, but got %#v", idx, test.line, line)
			}
		}()
	}
}
