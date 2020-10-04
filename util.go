package cuckoo

import (
	"encoding/binary"

	metro "github.com/dgryski/go-metro"
)

func getAltIndex(fp fingerprint, i uint, maxIndex uint) uint {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(fp))
	hash := uint(metro.Hash64(b, 1337))
	return (i ^ hash) % maxIndex
}

func getFingerprint(hash uint64) fingerprint {
	// Use least significant bits for fingerprint.
	fp := fingerprint(hash%(1<<fingerprintSizeBits-1) + 1)
	return fp
}

// getIndexAndFingerprint returns the primary bucket index and fingerprint to be used
func getIndexAndFingerprint(data []byte, maxIndex uint) (uint, fingerprint) {
	hash := metro.Hash64(data, 1337)
	f := getFingerprint(hash)
	// Use most significant bits for deriving index.
	i1 := uint(hash>>32) % maxIndex
	return i1, f
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
