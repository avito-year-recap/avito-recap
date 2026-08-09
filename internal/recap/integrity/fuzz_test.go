package integrity_test

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"testing"
)

func FuzzValidMetricsPipeline(f *testing.F) {
	f.Add([]byte("active seller"))
	f.Add([]byte("researcher"))
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		hash := fnv.New64a()
		_, _ = hash.Write(data)
		seedBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(seedBytes, hash.Sum64())
		seed := int64(binary.LittleEndian.Uint64(seedBytes))
		random := rand.New(rand.NewSource(seed))
		assertPipelineInvariants(t, randomValidMetrics(random), randomValidActionableState(random))
	})
}
