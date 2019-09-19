package algo

// SimpleBox performs box distance algorithm on rectangles.
func SimpleBox(targetMin, targetMax [2]float64) (
	algo func(
		min, max [2]float64, data interface{}, item bool,
		add func(min, max [2]float64, data interface{}, item bool, dist float64),
	),
) {
	return func(
		min, max [2]float64, data interface{}, item bool,
		add func(min, max [2]float64, data interface{}, item bool, dist float64),
	) {
		add(min, max, data, item, boxDist(targetMin, targetMax, min, max))
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
