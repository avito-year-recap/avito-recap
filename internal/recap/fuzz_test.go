package recap

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"testing"
)

func FuzzEnrichMetricsRatesStayWithinUnitInterval(f *testing.F) {
	f.Add(uint64(0), uint64(0), uint64(0), uint64(0))
	f.Add(uint64(1), uint64(1), uint64(1), uint64(1))
	f.Add(uint64(100), uint64(40), uint64(10), uint64(3))
	f.Add(^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0))

	f.Fuzz(func(t *testing.T, views, uniqueRaw, chats, linkedChatsRaw uint64) {
		unique := boundUint64(uniqueRaw, views)
		linkedChats := boundUint64(linkedChatsRaw, chats)
		metrics := EnrichMetrics(Metrics{
			TotalViews:        views,
			UniqueListings:    unique,
			RepeatedViews:     views - unique,
			ChatsStarted:      chats,
			ChatsWithPurchase: linkedChats,
		})
		assertRateInUnitInterval(t, "repeat rate", metrics.RepeatRate)
		assertRateInUnitInterval(t, "purchase rate", metrics.PurchaseRate)
	})
}

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

func boundUint64(value, maximum uint64) uint64 {
	if maximum == ^uint64(0) {
		return value
	}
	return value % (maximum + 1)
}
