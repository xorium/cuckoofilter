package cuckoo

import (
	"encoding/binary"
	"fmt"
	"math/rand"
)

const maxCuckooCount = 500

// Filter is a probabilistic counter
type Filter struct {
	buckets    []bucket
	count      uint
	numBuckets uint
}

// NewFilter returns a new cuckoofilter sutable for the given number of elements.
// When inserting more elements, insertion speed will drop significantly.
// A capacity of 1000000 is a normal default, which allocates
// about ~1MB on 64-bit machines.
func NewFilter(numElements uint) *Filter {
	numBuckets := getNextPow2(uint64(numElements / bucketSize))
	if float64(numElements)/float64(numBuckets*bucketSize) > 0.96 {
		numBuckets <<= 1
	}
	if numBuckets == 0 {
		numBuckets = 1
	}
	buckets := make([]bucket, numBuckets)
	return &Filter{
		buckets:    buckets,
		count:      0,
		numBuckets: numBuckets,
	}
}

// Lookup returns true if data is in the counter
func (cf *Filter) Lookup(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.numBuckets)
	i2 := getAltIndex(fp, i1, cf.numBuckets)
	b1, b2 := cf.buckets[i1], cf.buckets[i2]
	return b1.getFingerprintIndex(fp) > -1 || b2.getFingerprintIndex(fp) > -1
}

// Reset removes all items from the filter, setting count to 0.
func (cf *Filter) Reset() {
	for i := range cf.buckets {
		cf.buckets[i].reset()
	}
	cf.count = 0
}

func randi(i1, i2 uint) uint {
	if rand.Int31()%2 == 0 {
		return i1
	}
	return i2
}

// Insert inserts data into the counter and returns true upon success
func (cf *Filter) Insert(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.numBuckets)
	if cf.insert(fp, i1) {
		return true
	}
	i2 := getAltIndex(fp, i1, cf.numBuckets)
	if cf.insert(fp, i2) {
		return true
	}
	return cf.reinsert(fp, randi(i1, i2))
}

func (cf *Filter) insert(fp fingerprint, i uint) bool {
	if cf.buckets[i].insert(fp) {
		cf.count++
		return true
	}
	return false
}

func (cf *Filter) reinsert(fp fingerprint, i uint) bool {
	for k := 0; k < maxCuckooCount; k++ {
		j := rand.Intn(bucketSize)
		// Swap fp with bucket entry.
		cf.buckets[i][j], fp = fp, cf.buckets[i][j]

		// look in the alternate location for that random element
		i = getAltIndex(fp, i, cf.numBuckets)
		if cf.insert(fp, i) {
			return true
		}
	}
	return false
}

// Delete data from counter if exists and return if deleted or not
func (cf *Filter) Delete(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.numBuckets)
	i2 := getAltIndex(fp, i1, cf.numBuckets)
	return cf.delete(fp, i1) || cf.delete(fp, i2)
}

func (cf *Filter) delete(fp fingerprint, i uint) bool {
	if cf.buckets[i].delete(fp) {
		cf.count--
		return true
	}
	return false
}

// Count returns the number of items in the counter
func (cf *Filter) Count() uint {
	return cf.count
}

// Encode returns a byte slice representing a Cuckoofilter
func (cf *Filter) Encode() []byte {
	bytes := make([]byte, 0, len(cf.buckets)*bucketSize*fingerprintSizeBits/8)
	for _, b := range cf.buckets {
		for _, f := range b {
			next := make([]byte, 2)
			binary.LittleEndian.PutUint16(next, uint16(f))
			bytes = append(bytes, next...)
		}
	}
	return bytes
}

// Decode returns a Cuckoofilter from a byte slice
func Decode(bytes []byte) (*Filter, error) {
	var count uint
	if len(bytes)%bucketSize != 0 {
		return nil, fmt.Errorf("expected bytes to be multiple of %d, got %d", bucketSize, len(bytes))
	}
	buckets := make([]bucket, len(bytes)/4*8/fingerprintSizeBits)
	for i, b := range buckets {
		for j := range b {
			var next []byte
			next, bytes = bytes[0:2], bytes[2:]

			if fp := fingerprint(binary.LittleEndian.Uint16(next)); fp != 0 {
				buckets[i][j] = fp
				count++
			}
		}
	}
	return &Filter{
		buckets:    buckets,
		count:      count,
		numBuckets: uint(len(buckets)),
	}, nil
}
