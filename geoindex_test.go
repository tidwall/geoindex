package geoindex

import (
	"math/rand"
	"testing"
	"time"

	"github.com/tidwall/rbang"
)

func init() {
	seed := time.Now().UnixNano()
	println("seed:", seed)
	rand.Seed(seed)
}

func TestGeoIndex(t *testing.T) {
	t.Run("BenchVarious", func(t *testing.T) {
		Tests.TestBenchVarious(t, &rbang.RTree{}, 100000)
	})
	t.Run("RandomRects", func(t *testing.T) {
		Tests.TestRandomRects(t, &rbang.RTree{}, 10000)
	})
	t.Run("RandomPoints", func(t *testing.T) {
		Tests.TestRandomPoints(t, &rbang.RTree{}, 10000)
	})
	t.Run("ZeroPoints", func(t *testing.T) {
		Tests.TestZeroPoints(t, &rbang.RTree{})
	})
	t.Run("CitiesSVG", func(t *testing.T) {
		Tests.TestCitiesSVG(t, &rbang.RTree{})
	})
}

func BenchmarkRandomInsert(b *testing.B) {
	Tests.BenchmarkRandomInsert(b, &rbang.RTree{})
}
