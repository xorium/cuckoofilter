package cuckoo

import (
	"encoding/binary"

	metro "github.com/dgryski/go-metro"
)

var (
	masks = [65]uint{}
)

func init() {
	for i := uint(0); i <= 64; i++ {
		masks[i] = (1 << i) - 1
	}
}

func getAltIndex(fp fingerprint, i uint, bucketPow uint) uint {
	mask := masks[bucketPow]
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(fp))
	hash := uint(metro.Hash64(b, 1337))
	return (i ^ hash) & mask
}

func getFingerprint(hash uint64) fingerprint {
	// Use least significant bits for fingerprint.
	fp := fingerprint(hash%(1<<fingerprintSizeBits-1) + 1)
	return fp
}

// getIndicesAndFingerprint returns the 2 bucket indices and fingerprint to be used
func getIndicesAndFingerprint(data []byte, bucketPow uint) (uint, uint, fingerprint) {
	hash := metro.Hash64(data, 1337)
	f := getFingerprint(hash)
	// Use most significant bits for deriving index.
	i1 := uint(hash>>32) & masks[bucketPow]
	i2 := getAltIndex(f, i1, bucketPow)
	return i1, i2, f
}

func getNextPow2(n uint64) uint {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return uint(n)
}
