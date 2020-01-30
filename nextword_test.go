package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var NextwordTestDataPath string

func init() {
	if err := setNextwordTestDataPath(); err != nil {
		panic(err)
	}
}

func setNextwordTestDataPath() error {
	var ok bool
	NextwordTestDataPath, ok = os.LookupEnv("NEXTWORD_TEST_DATA_PATH")
	if !ok {
		return errors.New(`"NextwordTestDataPath" environment variable is not set`)
	}
	return nil
}

func TestNewNextword(t *testing.T) {
	tests := []struct {
		params *NextwordParams
		ok     bool
	}{
		{
			&NextwordParams{NextwordTestDataPath, 10, false},
			true,
		},
		{
			&NextwordParams{"", 10, false},
			false,
		},
		{
			&NextwordParams{"/invalid/invalid/invalid/invalid/invalid/invalid", 10, false},
			false,
		},
		{
			&NextwordParams{NextwordTestDataPath, 0, false},
			false,
		},
	}

	for idx, test := range tests {
		_, err := NewNextword(test.params)
		if (err == nil) != test.ok {
			t.Errorf("[%d] unexpected error: %v", idx, err)
		}
	}
}

func TestNextword_Suggest(t *testing.T) {
	tests := []struct {
		input      string
		params     *NextwordParams
		candidates []string
		err        error
	}{
		// 1gram
		{
			"misera",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			[]string{"misera", "miserable", "miserables", "miserably"},
			nil,
		},

		// n-gram
		{
			"by thermodynamic ",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			[]string{"considerations", "calculations", "equilibrium", "methods", "and"},
			nil,
		},

		// n-gram with prefix
		{
			"with the English a",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 5,
				Greedy:       false,
			},
			[]string{"and", "army", "at", "against", "as"},
			nil,
		},

		// CandidateNum
		{
			"young ",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			[]string{"man", "men", "people", "woman", "and", "children",
				"women", "lady", "girl", "girls"},
			nil,
		},

		// Greedy
		{
			"the consumptions ",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       true,
			},
			[]string{"of", "and", "are", "in"},
			nil,
		},

		// not found
		{
			"gur bjf ",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			nil,
			nil,
		},

		// not found
		{
			"gurbjf",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			nil,
			nil,
		},

		// not found
		{
			"zzzzzzzzzzzzzzzzzzzzz zzzzzzzzzzzzzzzz",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			nil,
			nil,
		},

		// not alphabet
		{
			"-",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			nil,
			nil,
		},

		// not alphabet
		{
			"- ",
			&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
				Greedy:       false,
			},
			nil,
			nil,
		},
	}

	for idx, test := range tests {
		nw, err := NewNextword(test.params)
		if err != nil {
			t.Fatal(err)
		}

		candidates, err := nw.Suggest(test.input)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("[%d] err: got %v, expected %v", idx, err, test.err)
		}
		if !reflect.DeepEqual(candidates, test.candidates) {
			t.Errorf("[%d] candidates: got %v, expected %v", idx, candidates, test.candidates)
		}
	}
}

func BenchmarkNextword_Suggest(b *testing.B) {
	queries := []string{
		"unsettled ",
		"zero at the ",
		"a ",
		"to the bottom and ",
		"on ",
		"even more pronounced for t",
		"in assessment of the a",
		"the ",
	}

	params := &NextwordParams{
		DataPath:     NextwordTestDataPath,
		CandidateNum: 100000,
		Greedy:       true,
	}
	nw, err := NewNextword(params)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		if _, err := nw.Suggest(query); err != nil {
			b.Fatal(err)
		}
	}
}

func TestNextword_ParseInput(t *testing.T) {
	tests := []struct {
		input  string
		ngram  []string
		prefix string
	}{
		{
			"The quick brown fox jumps over the lazy dog ",
			[]string{"over", "the", "lazy", "dog"},
			"",
		},
		{
			"The quick brown fox jumps over the lazy dog",
			[]string{"jumps", "over", "the", "lazy"},
			"dog",
		},
		{
			"The quick ",
			[]string{"The", "quick"},
			"",
		},
		{
			"",
			nil,
			"",
		},
	}

	for idx, test := range tests {
		nw, err := NewNextword(&NextwordParams{
			DataPath:     NextwordTestDataPath,
			CandidateNum: 10,
		})
		if err != nil {
			t.Fatal(err)
		}

		ngram, prefix := nw.parseInput(test.input)
		if !reflect.DeepEqual(ngram, test.ngram) {
			t.Errorf("[%d] ngram: got %v, expect %v", idx, ngram, test.ngram)
		}
		if prefix != test.prefix {
			t.Errorf("[%d] prefix: got %#v, expected %#v", idx, prefix, test.prefix)
		}
	}
}

func TestNextword_SearchNgram(t *testing.T) {
	tests := []struct {
		ngram      []string
		candidates []string
	}{
		// found
		{
			[]string{"role", "of", "agriculture", "in"},
			[]string{"the", "economic", "development"},
		},

		// not found
		{
			[]string{"role", "of", "aaaaaa", "bbbbb"},
			nil,
		},

		// not found
		{
			[]string{"Q", "A", "A", "A"},
			nil,
		},

		// not found
		{
			[]string{"qzzzzzzzzzzzzzzzzzzzzz", "zzzzz", "zzzzz", "zzzzzzz"},
			nil,
		},
	}

	for idx, test := range tests {
		nw, err := NewNextword(&NextwordParams{
			DataPath:     NextwordTestDataPath,
			CandidateNum: 10,
		})
		if err != nil {
			t.Fatal(err)
		}

		candidates, err := nw.searchNgram(test.ngram)
		if err != nil {
			t.Errorf("[%d] unexpected error: %v", idx, err)
			continue
		}

		if !reflect.DeepEqual(candidates, test.candidates) {
			t.Errorf("[%d] got %v, expected %v", idx, candidates, test.candidates)
		}
	}
}

func TestNextword_SearchOneGram(t *testing.T) {
	tests := []struct {
		prefix     string
		candidates []string
	}{
		// found
		{
			"Trace",
			[]string{"Trace", "Traceability", "Traced", "Tracer", "Tracers",
				"Tracery", "Traces", "Tracey"},
		},

		// not found
		{
			"AAAAAAAAAAAAAAAAAAAAAAa",
			nil,
		},

		// not found
		{
			"zzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			nil,
		},

		// invalid
		{
			"",
			nil,
		},
	}

	for idx, test := range tests {
		nw, err := NewNextword(&NextwordParams{
			DataPath:     NextwordTestDataPath,
			CandidateNum: 10,
		})
		if err != nil {
			t.Fatal(err)
		}

		candidates, err := nw.searchOneGram(test.prefix)
		if err != nil {
			t.Errorf("[%d] unexpected error: %v", idx, err)
			continue
		}

		if !reflect.DeepEqual(candidates, test.candidates) {
			t.Errorf("[%d] got %v, expected %v", idx, candidates, test.candidates)
		}
	}
}

func TestNextword_BinarySearch(t *testing.T) {
	tests := []struct {
		fname  string
		query  string
		offset int64
		err    error
	}{
		// found first
		{
			"1gram.txt",
			"A",
			0,
			nil,
		},

		// found last
		{
			"1gram.txt",
			"zz",
			2574630,
			nil,
		},

		// not found
		{
			"1gram.txt",
			"AAAAAAAAAAAAAAAAAAAAA",
			20,
			nil,
		},

		// not found
		{
			"1gram.txt",
			"zzzzzzzzzzzzzzzzzz",
			2574633,
			nil,
		},

		// found
		{
			"1gram.txt",
			"palpitations",
			2131290,
			nil,
		},

		// found prefix
		{
			"1gram.txt",
			"nightc",
			2074243,
			nil,
		},
	}

	for idx, test := range tests {
		nw, err := NewNextword(&NextwordParams{
			DataPath:     NextwordTestDataPath,
			CandidateNum: 10,
		})
		if err != nil {
			t.Fatal(err)
		}

		func() {
			path := filepath.Join(NextwordTestDataPath, test.fname)
			f, err := os.Open(path)
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}
			defer f.Close()

			fi, err := os.Stat(path)
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}

			offset, err := nw.binarySearch(f, 0, fi.Size(), test.query)
			if !errors.Is(err, test.err) {
				t.Errorf("[%d] err: got %v, expected %v", idx, err, test.err)
				return
			}
			if offset != test.offset {
				t.Errorf("[%d] offset: got %v, expected %v", idx, offset, test.offset)
			}
		}()
	}
}

func TestNextword_ReadLine(t *testing.T) {
	tests := []struct {
		bufSize int
		fname   string
		offset  int64
		line    string
		err     error
	}{
		// enough buf size
		{
			4096,
			"2gram-a.txt",
			83928,
			"ACID\tAND IN RAIN properties METABOLISM The COMPOSITION is CYCLE ON",
			nil,
		},

		// small buf size
		{
			10,
			"2gram-a.txt",
			83928,
			"ACID\tAND IN RAIN properties METABOLISM The COMPOSITION is CYCLE ON",
			nil,
		},

		// last
		{
			4096,
			"2gram-a.txt",
			3488928,
			"azygous\tvein",
			nil,
		},

		// EOF
		{
			4096,
			"2gram-a.txt",
			1000000000,
			"",
			io.EOF,
		},
	}

	for idx, test := range tests {
		func() {
			nw, err := NewNextword(&NextwordParams{
				DataPath:     NextwordTestDataPath,
				CandidateNum: 10,
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			nw.readLineBufSize = test.bufSize

			path := filepath.Join(NextwordTestDataPath, test.fname)
			f, err := os.Open(path)
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", idx, err)
				return
			}
			defer f.Close()

			line, err := nw.readLine(f, test.offset)
			if !errors.Is(err, test.err) {
				t.Errorf("[%d] err: got %v, expected %v", idx, err, test.err)
				return
			}

			if line != test.line {
				t.Errorf("[%d] line: got %#v, expected %#v", idx, line, test.line)
			}
		}()
	}
}
