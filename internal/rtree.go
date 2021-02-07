// Copyright 2021 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package internal

import "github.com/tidwall/geoindex/child"

// This file is an internal rtree implementation for the specific purpose of
// testing the geoindex. It's a sane implementation that is copied verbatim
// from github.com/tidwall/rtree v1.2.5.
//
// Please do not use this outside of testing

const (
	maxEntries = 32
	minEntries = maxEntries * 20 / 100
)

type rect struct {
	min, max [2]float64
	data     interface{}
}

type node struct {
	count int
	rects [maxEntries + 1]rect
}

// RTree ...
type RTree struct {
	height   int
	root     rect
	count    int
	reinsert []rect
}

func (r *rect) expand(b *rect) {
	if b.min[0] < r.min[0] {
		r.min[0] = b.min[0]
	}
	if b.max[0] > r.max[0] {
		r.max[0] = b.max[0]
	}
	if b.min[1] < r.min[1] {
		r.min[1] = b.min[1]
	}
	if b.max[1] > r.max[1] {
		r.max[1] = b.max[1]
	}
}

func (r *rect) area() float64 {
	return (r.max[0] - r.min[0]) * (r.max[1] - r.min[1])
}

func (r *rect) overlapArea(b *rect) float64 {
	area := 1.0
	var max, min float64
	if r.max[0] < b.max[0] {
		max = r.max[0]
	} else {
		max = b.max[0]
	}
	if r.min[0] > b.min[0] {
		min = r.min[0]
	} else {
		min = b.min[0]
	}
	if max > min {
		area *= max - min
	} else {
		return 0
	}
	if r.max[1] < b.max[1] {
		max = r.max[1]
	} else {
		max = b.max[1]
	}
	if r.min[1] > b.min[1] {
		min = r.min[1]
	} else {
		min = b.min[1]
	}
	if max > min {
		area *= max - min
	} else {
		return 0
	}
	return area
}

func (r *rect) enlargedArea(b *rect) float64 {
	area := 1.0
	if b.max[0] > r.max[0] {
		if b.min[0] < r.min[0] {
			area *= b.max[0] - b.min[0]
		} else {
			area *= b.max[0] - r.min[0]
		}
	} else {
		if b.min[0] < r.min[0] {
			area *= r.max[0] - b.min[0]
		} else {
			area *= r.max[0] - r.min[0]
		}
	}
	if b.max[1] > r.max[1] {
		if b.min[1] < r.min[1] {
			area *= b.max[1] - b.min[1]
		} else {
			area *= b.max[1] - r.min[1]
		}
	} else {
		if b.min[1] < r.min[1] {
			area *= r.max[1] - b.min[1]
		} else {
			area *= r.max[1] - r.min[1]
		}
	}
	return area
}

// Insert inserts an item into the RTree
func (tr *RTree) Insert(min, max [2]float64, value interface{}) {
	var item rect
	fit(min, max, value, &item)
	tr.insert(&item)
}

func (tr *RTree) insert(item *rect) {
	if tr.root.data == nil {
		fit(item.min, item.max, new(node), &tr.root)
	}
	grown := tr.root.insert(item, tr.height)
	if grown {
		tr.root.expand(item)
	}
	if tr.root.data.(*node).count == maxEntries+1 {
		newRoot := new(node)
		tr.root.splitLargestAxisEdgeSnap(&newRoot.rects[1])
		newRoot.rects[0] = tr.root
		newRoot.count = 2
		tr.root.data = newRoot
		tr.root.recalc()
		tr.height++
	}
	tr.count++
}

const inlineEnlargedArea = true

func (r *rect) chooseLeastEnlargement(b *rect) (index int) {
	n := r.data.(*node)
	j, jenlargement, jarea := -1, 0.0, 0.0
	for i := 0; i < n.count; i++ {
		var earea float64
		if inlineEnlargedArea {
			earea = 1.0
			if b.max[0] > n.rects[i].max[0] {
				if b.min[0] < n.rects[i].min[0] {
					earea *= b.max[0] - b.min[0]
				} else {
					earea *= b.max[0] - n.rects[i].min[0]
				}
			} else {
				if b.min[0] < n.rects[i].min[0] {
					earea *= n.rects[i].max[0] - b.min[0]
				} else {
					earea *= n.rects[i].max[0] - n.rects[i].min[0]
				}
			}
			if b.max[1] > n.rects[i].max[1] {
				if b.min[1] < n.rects[i].min[1] {
					earea *= b.max[1] - b.min[1]
				} else {
					earea *= b.max[1] - n.rects[i].min[1]
				}
			} else {
				if b.min[1] < n.rects[i].min[1] {
					earea *= n.rects[i].max[1] - b.min[1]
				} else {
					earea *= n.rects[i].max[1] - n.rects[i].min[1]
				}
			}
		} else {
			earea = n.rects[i].enlargedArea(b)
		}
		area := n.rects[i].area()
		enlargement := earea - area
		if j == -1 || enlargement < jenlargement ||
			(enlargement == jenlargement && area < jarea) {
			j, jenlargement, jarea = i, enlargement, area
		}
	}
	return j
}

func (r *rect) recalc() {
	n := r.data.(*node)
	r.min = n.rects[0].min
	r.max = n.rects[0].max
	for i := 1; i < n.count; i++ {
		r.expand(&n.rects[i])
	}
}

// contains return struct when b is fully contained inside of n
func (r *rect) contains(b *rect) bool {
	if b.min[0] < r.min[0] || b.max[0] > r.max[0] {
		return false
	}
	if b.min[1] < r.min[1] || b.max[1] > r.max[1] {
		return false
	}
	return true
}

func (r *rect) largestAxis() (axis int, size float64) {
	if r.max[1]-r.min[1] > r.max[0]-r.min[0] {
		return 1, r.max[1] - r.min[1]
	}
	return 0, r.max[0] - r.min[0]
}

func (r *rect) splitLargestAxisEdgeSnap(right *rect) {
	axis, _ := r.largestAxis()
	left := r
	leftNode := left.data.(*node)
	rightNode := new(node)
	right.data = rightNode

	var equals []rect
	for i := 0; i < leftNode.count; i++ {
		minDist := leftNode.rects[i].min[axis] - left.min[axis]
		maxDist := left.max[axis] - leftNode.rects[i].max[axis]
		if minDist < maxDist {
			// stay left
		} else {
			if minDist > maxDist {
				// move to right
				rightNode.rects[rightNode.count] = leftNode.rects[i]
				rightNode.count++
			} else {
				// move to equals, at the end of the left array
				equals = append(equals, leftNode.rects[i])
			}
			leftNode.rects[i] = leftNode.rects[leftNode.count-1]
			leftNode.rects[leftNode.count-1].data = nil
			leftNode.count--
			i--
		}
	}
	for _, b := range equals {
		if leftNode.count < rightNode.count {
			leftNode.rects[leftNode.count] = b
			leftNode.count++
		} else {
			rightNode.rects[rightNode.count] = b
			rightNode.count++
		}
	}
	left.recalc()
	right.recalc()
}

func (r *rect) insert(item *rect, height int) (grown bool) {
	n := r.data.(*node)
	if height == 0 {
		n.rects[n.count] = *item
		n.count++
		grown = !r.contains(item)
		return grown
	}

	// choose subtree
	index := -1
	narea := 0.0
	// first take a quick look for any nodes that contain the rect
	for i := 0; i < n.count; i++ {
		if n.rects[i].contains(item) {
			area := n.rects[i].area()
			if index == -1 || area < narea {
				narea = area
				index = i
			}
		}
	}
	// found nothing, now go the slow path
	if index == -1 {
		index = r.chooseLeastEnlargement(item)
	}
	// insert the item into the child node
	child := &n.rects[index]
	grown = child.insert(item, height-1)
	if grown {
		child.expand(item)
		grown = !r.contains(item)
	}
	if child.data.(*node).count == maxEntries+1 {
		child.splitLargestAxisEdgeSnap(&n.rects[n.count])
		n.count++
	}
	return grown
}

// fit an external item into a rect type
func fit(min, max [2]float64, value interface{}, target *rect) {
	target.min = min
	target.max = max
	target.data = value
}

// contains return struct when b is fully contained inside of n
func (r *rect) intersects(b *rect) bool {
	if b.min[0] > r.max[0] || b.max[0] < r.min[0] {
		return false
	}
	if b.min[1] > r.max[1] || b.max[1] < r.min[1] {
		return false
	}
	return true
}

func (r *rect) search(
	target rect, height int,
	iter func(min, max [2]float64, value interface{}) bool,
) bool {
	n := r.data.(*node)
	if height == 0 {
		for i := 0; i < n.count; i++ {
			if target.intersects(&n.rects[i]) {
				if !iter(n.rects[i].min, n.rects[i].max, n.rects[i].data) {
					return false
				}
			}
		}
	} else {
		for i := 0; i < n.count; i++ {
			if target.intersects(&n.rects[i]) {
				if !n.rects[i].search(target, height-1, iter) {
					return false
				}
			}
		}
	}
	return true
}

func (tr *RTree) search(
	target rect,
	iter func(min, max [2]float64, value interface{}) bool,
) {
	if tr.root.data == nil {
		return
	}
	if target.intersects(&tr.root) {
		tr.root.search(target, tr.height, iter)
	}
}

// Search ...
func (tr *RTree) Search(
	min, max [2]float64,
	iter func(min, max [2]float64, value interface{}) bool,
) {
	tr.search(rect{min: min, max: max}, iter)
}

func (r *rect) scan(
	height int,
	iter func(min, max [2]float64, value interface{}) bool,
) bool {
	n := r.data.(*node)
	if height == 0 {
		for i := 0; i < n.count; i++ {
			if !iter(n.rects[i].min, n.rects[i].max, n.rects[i].data) {
				return false
			}
		}
	} else {
		for i := 0; i < n.count; i++ {
			if !n.rects[i].scan(height-1, iter) {
				return false
			}
		}
	}
	return true
}

// Scan iterates through all data in tree.
func (tr *RTree) Scan(iter func(min, max [2]float64, data interface{}) bool) {
	if tr.root.data == nil {
		return
	}
	tr.root.scan(tr.height, iter)
}

// Delete data from tree
func (tr *RTree) Delete(min, max [2]float64, data interface{}) {
	var item rect
	fit(min, max, data, &item)
	if tr.root.data == nil || !tr.root.contains(&item) {
		return
	}
	var removed, recalced bool
	removed, recalced = tr.root.delete(tr, &item, tr.height)
	if !removed {
		return
	}
	tr.count -= len(tr.reinsert) + 1
	if tr.count == 0 {
		tr.root = rect{}
		recalced = false
	} else {
		for tr.height > 0 && tr.root.data.(*node).count == 1 {
			tr.root = tr.root.data.(*node).rects[0]
			tr.height--
			tr.root.recalc()
		}
	}
	if recalced {
		tr.root.recalc()
	}
	if len(tr.reinsert) > 0 {
		for i := range tr.reinsert {
			tr.insert(&tr.reinsert[i])
			tr.reinsert[i].data = nil
		}
		tr.reinsert = tr.reinsert[:0]
	}
}

func (r *rect) delete(tr *RTree, item *rect, height int,
) (removed, recalced bool) {
	n := r.data.(*node)
	rects := n.rects[0:n.count]
	if height == 0 {
		for i := 0; i < len(rects); i++ {
			if rects[i].data == item.data {
				// found the target item to delete
				recalced = r.onEdge(&rects[i])
				rects[i] = rects[len(rects)-1]
				rects[len(rects)-1].data = nil
				n.count--
				if recalced {
					r.recalc()
				}
				return true, recalced
			}
		}
	} else {
		for i := 0; i < len(rects); i++ {
			if !rects[i].contains(item) {
				continue
			}
			removed, recalced = rects[i].delete(tr, item, height-1)
			if !removed {
				continue
			}
			if rects[i].data.(*node).count < minEntries {
				// underflow
				if !recalced {
					recalced = r.onEdge(&rects[i])
				}
				tr.reinsert = rects[i].flatten(tr.reinsert, height-1)
				rects[i] = rects[len(rects)-1]
				rects[len(rects)-1].data = nil
				n.count--
			}
			if recalced {
				r.recalc()
			}
			return removed, recalced
		}
	}
	return false, false
}

// flatten all leaf rects into a single list
func (r *rect) flatten(all []rect, height int) []rect {
	n := r.data.(*node)
	if height == 0 {
		all = append(all, n.rects[:n.count]...)
	} else {
		for i := 0; i < n.count; i++ {
			all = n.rects[i].flatten(all, height-1)
		}
	}
	return all
}

// onedge returns true when b is on the edge of r
func (r *rect) onEdge(b *rect) bool {
	if r.min[0] == b.min[0] || r.max[0] == b.max[0] {
		return true
	}
	if r.min[1] == b.min[1] || r.max[1] == b.max[1] {
		return true
	}
	return false
}

// Len returns the number of items in tree
func (tr *RTree) Len() int {
	return tr.count
}

// Bounds returns the minimum bounding rect
func (tr *RTree) Bounds() (min, max [2]float64) {
	if tr.root.data == nil {
		return
	}
	return tr.root.min, tr.root.max
}

// Children is a utility function that returns all children for parent node.
// If parent node is nil then the root nodes should be returned. The min, max,
// data, and items slices all must have the same lengths. And, each element
// from all slices must be associated. Returns true for `items` when the the
// item at the leaf level. The reuse buffers are empty length slices that can
// optionally be used to avoid extra allocations.
func (tr *RTree) Children(
	parent interface{},
	reuse []child.Child,
) []child.Child {
	children := reuse
	if parent == nil {
		if tr.Len() > 0 {
			// fill with the root
			children = append(children, child.Child{
				Min:  tr.root.min,
				Max:  tr.root.max,
				Data: tr.root.data,
				Item: false,
			})
		}
	} else {
		// fill with child items
		n := parent.(*node)
		item := true
		if n.count > 0 {
			if _, ok := n.rects[0].data.(*node); ok {
				item = false
			}
		}
		for i := 0; i < n.count; i++ {
			children = append(children, child.Child{
				Min:  n.rects[i].min,
				Max:  n.rects[i].max,
				Data: n.rects[i].data,
				Item: item,
			})
		}
	}
	return children
}

// Replace an item.
// This is effectively just a Delete followed by an Insert. Which means the
// new item will always be inserted, whether or not the old item was deleted.
func (tr *RTree) Replace(
	oldMin, oldMax [2]float64, oldData interface{},
	newMin, newMax [2]float64, newData interface{},
) {
	tr.Delete(oldMin, oldMax, oldData)
	tr.Insert(newMin, newMax, newData)
}
