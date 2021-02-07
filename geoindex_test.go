package geoindex_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/tidwall/geoindex"
	"github.com/tidwall/rtree"
)

func init() {
	seed := time.Now().UnixNano()
	println("seed:", seed)
	rand.Seed(seed)
}

func TestGeoIndex(t *testing.T) {
	t.Run("BenchVarious", func(t *testing.T) {
		geoindex.Tests.TestBenchVarious(t, &rtree.RTree{}, 100000)
	})
	t.Run("RandomRects", func(t *testing.T) {
		geoindex.Tests.TestRandomRects(t, &rtree.RTree{}, 10000)
	})
	t.Run("RandomPoints", func(t *testing.T) {
		geoindex.Tests.TestRandomPoints(t, &rtree.RTree{}, 10000)
	})
	t.Run("ZeroPoints", func(t *testing.T) {
		geoindex.Tests.TestZeroPoints(t, &rtree.RTree{})
	})
	t.Run("CitiesSVG", func(t *testing.T) {
		geoindex.Tests.TestCitiesSVG(t, &rtree.RTree{})
	})
}

func BenchmarkRandomInsert(b *testing.B) {
	geoindex.Tests.BenchmarkRandomInsert(b, &rtree.RTree{})
}
