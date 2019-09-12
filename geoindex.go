package geoindex

import "fmt"

// Index is a wrapper around Interface that provides extra features like a
// Nearby (kNN) function.
// This can be created like such:
//   var tree = rbang.New()
//   var index = index.Index{tree}
// Now you can use `index` just like tree but with the extra features.
type Index struct {
	Interface
}

// Interface is a tree-like structure that contains geospatial data
type Interface interface {
	// Insert an item into the structure
	Insert(min, max [2]float64, data interface{})
	// Delete an item from the structure
	Delete(min, max [2]float64, data interface{})
	// Search the structure for items that intersects the rect param
	Search(
		min, max [2]float64,
		iter func(min, max [2]float64, data interface{}) bool,
	)
	// Scan iterates through all data in tree in no specified order.
	Scan(iter func(min, max [2]float64, data interface{}) bool)
	// Len returns the number of items in tree
	Len() int
	// Bounds returns the minimum bounding box
	Bounds() (min, max [2]float64)
	// Children returns all children for parent node. If parent node is nil
	// then the root nodes should be returned. The min, max, data, and items
	// slices all must have the same lengths. And, each element from all slices
	// must be associated. Returns true for `items` when the the item at the
	// leaf level. The reuse buffers are empty length slices that can
	// optionally be used to avoid extra allocations.
	Children(
		parent interface{},
		reuseMin [][2]float64,
		reuseMax [][2]float64,
		reuseData []interface{},
		reuseItem []bool,
	) (min, max [][2]float64, datas []interface{}, items []bool)
}

// Insert an item into the index
func (index *Index) Insert(min, max [2]float64, data interface{}) {
	index.Interface.Insert(min, max, data)
}

// Search the index for items that intersects the rect param
func (index *Index) Search(
	min, max [2]float64,
	iter func(min, max [2]float64, data interface{}) bool,
) {
	index.Interface.Search(min, max, iter)
}

// Delete an item from the index
func (index *Index) Delete(min, max [2]float64, data interface{}) {
	index.Interface.Delete(min, max, data)
}

// Nearby performs a kNN-type operation on the index. It's expected that the
// caller provides the `dist` function, which is used to calculate the
// distance from a node or item to another object. The other object is unknown
// this operation, but is expected to be known by the caller. The iter will
// return all items from the smallest dist to the largest dist.
func (index *Index) Nearby(
	algo func(min, max [2]float64, data interface{}, item bool) (dist float64),
	iter func(min, max [2]float64, data interface{}, dist float64) bool,
) {
	var q queue
	var parent interface{}
	var mins [][2]float64
	var maxs [][2]float64
	var datas []interface{}
	var items []bool
	for {
		// gather all children for parent
		mins, maxs, datas, items = index.Interface.Children(parent,
			mins[:0], maxs[:0], datas[:0], items[:0])
		for i := 0; i < len(datas); i++ {
			q.push(qnode{
				dist:   algo(mins[i], maxs[i], datas[i], items[i]),
				min:    mins[i],
				max:    maxs[i],
				data:   datas[i],
				item:   items[i],
				filled: true,
			})
		}
		for {
			node := q.pop()
			if !node.filled {
				// nothing left in queue
				return
			}
			if node.item {
				if !iter(node.min, node.max, node.data, node.dist) {
					return
				}
			} else {
				// gather more children
				parent = node.data
				break
			}
		}
	}
}

// Priority Queue ordered by dist (smallest to largest)

type qnode struct {
	dist   float64
	min    [2]float64
	max    [2]float64
	data   interface{}
	item   bool
	filled bool
}

type queue struct {
	nodes []qnode
	len   int
	size  int
}

func (q *queue) push(node qnode) {
	if q.nodes == nil {
		q.nodes = make([]qnode, 2)
	} else {
		q.nodes = append(q.nodes, qnode{})
	}
	i := q.len + 1
	j := i / 2
	for i > 1 && q.nodes[j].dist > node.dist {
		q.nodes[i] = q.nodes[j]
		i = j
		j = j / 2
	}
	q.nodes[i] = node
	q.len++
}

func (q *queue) pop() qnode {
	if q.len == 0 {
		return qnode{}
	}
	n := q.nodes[1]
	q.nodes[1] = q.nodes[q.len]
	q.len--
	var j, k int
	i := 1
	for i != q.len+1 {
		k = q.len + 1
		j = 2 * i
		if j <= q.len && q.nodes[j].dist < q.nodes[k].dist {
			k = j
		}
		if j+1 <= q.len && q.nodes[j+1].dist < q.nodes[k].dist {
			k = j + 1
		}
		q.nodes[i] = q.nodes[k]
		i = k
	}
	return n
}

// SimpleBoxAlgo ...
func SimpleBoxAlgo(targetMin, targetMax [2]float64) (
	dist func(min, max [2]float64, data interface{}, item bool) (dist float64),
) {
	return func(min, max [2]float64, data interface{}, item bool) float64 {
		return boxDist(targetMin, targetMax, min, max)
	}
}

func boxDist(amin, amax, bmin, bmax [2]float64) float64 {
	var dist float64
	var min, max float64
	if amin[0] > bmin[0] {
		min = amin[0]
	} else {
		min = bmin[0]
	}
	if amax[0] < bmax[0] {
		max = amax[0]
	} else {
		max = bmax[0]
	}
	squared := min - max
	if squared > 0 {
		dist += squared * squared
	}
	if amin[1] > bmin[1] {
		min = amin[1]
	} else {
		min = bmin[1]
	}
	if amax[1] < bmax[1] {
		max = amax[1]
	} else {
		max = bmax[1]
	}
	squared = min - max
	if squared > 0 {
		dist += squared * squared
	}
	return dist
}

func (index *Index) svg(data interface{}, item bool, min, max [2]float64, height int) []byte {
	var out []byte
	point := true
	for i := 0; i < 2; i++ {
		if min[i] != max[i] {
			point = false
			break
		}
	}
	if point { // is point
		out = append(out, fmt.Sprintf(
			"<rect x=\"%.0f\" y=\"%0.f\" width=\"%0.f\" height=\"%0.f\" "+
				"stroke=\"%s\" fill=\"purple\" "+
				"fill-opacity=\"0\" stroke-opacity=\"1\" "+
				"rx=\"15\" ry=\"15\"/>\n",
			(min[0])*svgScale,
			(min[1])*svgScale,
			(max[0]-min[0]+1/svgScale)*svgScale,
			(max[1]-min[1]+1/svgScale)*svgScale,
			strokes[height%len(strokes)])...)
	} else { // is rect
		out = append(out, fmt.Sprintf(
			"<rect x=\"%.0f\" y=\"%0.f\" width=\"%0.f\" height=\"%0.f\" "+
				"stroke=\"%s\" fill=\"purple\" "+
				"fill-opacity=\"0\" stroke-opacity=\"1\"/>\n",
			(min[0])*svgScale,
			(min[1])*svgScale,
			(max[0]-min[0]+1/svgScale)*svgScale,
			(max[1]-min[1]+1/svgScale)*svgScale,
			strokes[height%len(strokes)])...)
	}
	if !item {
		mins, maxs, datas, items :=
			index.Children(data, nil, nil, nil, nil)
		for i := 0; i < len(datas); i++ {
			out = append(out, index.svg(datas[i], items[i], mins[i], maxs[i], height+1)...)
		}
	}
	return out
}

const (
	// Continue to first child rectangle and/or next sibling.
	Continue = iota
	// Ignore child rectangles but continue to next sibling.
	Ignore
	// Stop iterating
	Stop
)

const svgScale = 4.0

var strokes = [...]string{"black", "#cccc00", "green", "red", "purple"}

// SVG prints 2D rtree in wgs84 coordinate space
func (index *Index) SVG() string {
	var out string
	out += fmt.Sprintf("<svg viewBox=\"%.0f %.0f %.0f %.0f\" "+
		"xmlns =\"http://www.w3.org/2000/svg\">\n",
		-190.0*svgScale, -100.0*svgScale,
		380.0*svgScale, 190.0*svgScale)

	out += fmt.Sprintf("<g transform=\"scale(1,-1)\">\n")

	var outb []byte
	mins, maxs, datas, items :=
		index.Children(nil, nil, nil, nil, nil)
	for i := 0; i < len(datas); i++ {
		outb = append(outb, index.svg(datas[i], items[i], mins[i], maxs[i], 1)...)
	}

	out += string(outb)
	out += fmt.Sprintf("</g>\n")
	out += fmt.Sprintf("</svg>\n")
	return out
}
