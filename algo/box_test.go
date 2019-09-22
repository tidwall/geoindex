package algo

import (
	"testing"
)

func TestBoxDist(t *testing.T) {
	distA := BoxDistCalc(
		[2]float64{170, 33}, [2]float64{170, 33},
		[2]float64{-170, 33}, [2]float64{-170, 33},
		false,
	)

	distB := BoxDistCalc(
		[2]float64{170, 33}, [2]float64{170, 33},
		[2]float64{-170, 33}, [2]float64{-170, 33},
		true,
	)
	distC := BoxDistCalc(
		[2]float64{170 - 360, 33}, [2]float64{170 - 360, 33},
		[2]float64{-170, 33}, [2]float64{-170, 33},
		false,
	)
	if distA < distB || distC != distB {
		t.Fatalf("unexpected results")
	}
}
