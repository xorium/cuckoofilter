package cuckoo

import (
	"reflect"
	"testing"
)

func TestBucket_Reset(t *testing.T) {
	var bkt bucket
	for i := fingerprint(0); i < bucketSize; i++ {
		bkt[i] = i
	}
	bkt.reset()

	var want bucket
	if !reflect.DeepEqual(bkt, want) {
		t.Errorf("bucket.reset() got %v, want %v", bkt, want)
	}
}
