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

func TestQueue(t *testing.T) {
	var q queue

	q.push(qnode{
		dist: 2,
	})
	q.push(qnode{
		dist: 1,
	})
	q.push(qnode{
		dist: 5,
	})
	q.push(qnode{
		dist: 3,
	})
	q.push(qnode{
		dist: 4,
	})

	lastDist := float64(-1)
	for i := 0; i < 3; i++ {
		node, ok := q.pop()
		if !ok {
			t.Fatal("queue was empty")
		}
		if node.dist < lastDist {
			t.Fatal("queue was out of order")
		}
	}

	if len(q) != 2 {
		t.Fatal("queue was wrong size")
	}

	capBeforeInserts := cap(q)
	q.push(qnode{
		dist: 1,
	})
	q.push(qnode{
		dist: 10,
	})
	q.push(qnode{
		dist: 11,
	})

	if cap(q) != capBeforeInserts {
		t.Fatal("queue did not reuse space")
	}

	lastDist = -1
	for i := 0; i < 5; i++ {
		node, ok := q.pop()
		if !ok {
			t.Fatal("queue was empty")
		}
		if node.dist < lastDist {
			t.Fatal("queue was out of order")
		}
	}

	_, ok := q.pop()
	if ok {
		t.Fatal("queue was not empty")
	}
}

func BenchmarkQueue(b *testing.B) {
	var q queue

	for i := 0; i < b.N; i++ {
		r := rand.Float64()
		if r < 0.5 {
			q.push(qnode{dist: r})
		} else {
			q.pop()
		}
	}
}
