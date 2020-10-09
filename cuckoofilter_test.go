package cuckoo

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// optFloatNear considers float64 as equal if the relative delta is small.
var optFloatNear = cmp.Comparer(func(x, y float64) bool {
	delta := math.Abs(x - y)
	mean := math.Abs(x+y) / 2.0
	return delta/mean < 0.00001
})

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

	if got, want := cf.Count(), lineCount; got != want {
		t.Errorf("After inserting: Count() = %d, want %d", got, want)
	}
	if got, want := cf.LoadFactor(), float64(0.097657); !cmp.Equal(got, want, optFloatNear) {
		t.Errorf("After inserting: LoadFactor() = %f, want %f.", got, want)
	}

	for _, v := range values {
		cf.Delete(v)
	}

	if got, want := cf.Count(), uint(0); got != want {
		t.Errorf("After deleting: Count() = %d, want %d", got, want)
	}
	if got, want := cf.LoadFactor(), float64(0); got != want {
		t.Errorf("After deleting: LoadFactor() = %f, want %f", got, want)
	}
}

func TestLookup(t *testing.T) {
	cf := NewFilter(4)
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

func TestFilter_LookupLarge(t *testing.T) {
	const size = 10000
	insertFail := 0
	cf := NewFilter(size)
	for i := 0; i < size; i++ {
		if !cf.Insert([]byte{byte(i)}) {
			insertFail++
		}
	}
	fn := 0
	for i := 0; i < size; i++ {
		if !cf.Lookup([]byte{byte(i)}) {
			fn++
		}
	}

	if fn != 0 {
		t.Errorf("cf.Lookup() with %d items. False negatives = %d, want 0. Insert failed %d times", size, fn, insertFail)
	}
}

func TestFilter_Insert(t *testing.T) {
	const cap = 10000
	filter := NewFilter(cap)

	var hash [32]byte

	for i := 0; i < 100; i++ {
		io.ReadFull(rand.Reader, hash[:])
		filter.Insert(hash[:])
	}

	if got, want := filter.Count(), uint(100); got != want {
		t.Errorf("inserting 100 items, Count() = %d, want %d", got, want)
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
	cf := NewFilter(10)
	cf.Insert([]byte{1})
	cf.Insert([]byte{2})
	cf.Insert([]byte{3})
	cf.Insert([]byte{4})
	cf.Insert([]byte{5})
	cf.Insert([]byte{6})
	cf.Insert([]byte{7})
	cf.Insert([]byte{8})
	cf.Insert([]byte{9})
	encoded := cf.Encode()
	got, err := Decode(encoded)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(cf, got) {
		t.Errorf("Decode = %v, want %v, encoded = %v", got, cf, encoded)
	}
}
