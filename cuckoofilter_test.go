package cuckoo

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestInsertion(t *testing.T) {
	cf := NewFilter(1000000)
	fd, err := os.Open("/usr/share/dict/words")
	if err != nil {
		t.Fatalf("failed reading words: %v", err)
	}
	scanner := bufio.NewScanner(fd)

	var values [][]byte
	var lineCount uint
	for scanner.Scan() {
		s := []byte(scanner.Text())
		if cf.Insert(s) {
			lineCount++
		}
		values = append(values, s)
	}

	count := cf.Count()
	if count != lineCount {
		t.Errorf("Expected count = %d, instead count = %d", lineCount, count)
	}

	for _, v := range values {
		cf.Delete(v)
	}

	count = cf.Count()
	if count != 0 {
		t.Errorf("Expected count = 0, instead count == %d", count)
	}
}

func TestLookup(t *testing.T) {
	cf := NewFilter(10)
	cf.Insert([]byte("one"))
	cf.Insert([]byte("two"))
	cf.Insert([]byte("three"))

	testCases := []struct {
		word string
		want bool
	}{
		{"one", true},
		{"two", true},
		{"three", true},
		{"four", false},
		{"five", false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("cf.Lookup(%q)", tc.word), func(t *testing.T) {
			if got := cf.Lookup([]byte(tc.word)); got != tc.want {
				t.Errorf("cf.Lookup(%q) got %v, want %v", tc.word, got, tc.want)
			}
		})
	}
}

func TestFilter_Insert(t *testing.T) {
	const cap = 10000
	filter := NewFilter(cap)

	var hash [32]byte
	io.ReadFull(rand.Reader, hash[:])

	for i := 0; i < 100; i++ {
		filter.Insert(hash[:])
	}

	if got, want := filter.Count(), uint(100); got != want {
		t.Errorf("inserting same item 100 times, Count() = %d, want %d", got, want)
	}
}

func BenchmarkFilter_Reset(b *testing.B) {
	const cap = 10000
	filter := NewFilter(cap)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter.Reset()
	}
}

func BenchmarkFilter_Insert(b *testing.B) {
	const cap = 10000
	filter := NewFilter(cap)

	b.ResetTimer()

	var hash [32]byte
	for i := 0; i < b.N; i++ {
		io.ReadFull(rand.Reader, hash[:])
		filter.Insert(hash[:])
	}
}

func BenchmarkFilter_Lookup(b *testing.B) {
	const cap = 10000
	filter := NewFilter(cap)

	var hash [32]byte
	for i := 0; i < 10000; i++ {
		io.ReadFull(rand.Reader, hash[:])
		filter.Insert(hash[:])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		io.ReadFull(rand.Reader, hash[:])
		filter.Lookup(hash[:])
	}
}

func TestEncodeDecode(t *testing.T) {
	cf := NewFilter(8)
	cf.buckets = []bucket{
		[4]fingerprint{1, 2, 3, 4},
		[4]fingerprint{5, 6, 7, 8},
	}
	cf.count = 8
	encoded := cf.Encode()
	got, err := Decode(encoded)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(cf, got) {
		t.Errorf("Decode = %v, want %v, encoded = %v", got, cf, encoded)
	}
}
