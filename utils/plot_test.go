package utils

import (
	"fmt"
	"testing"
)

func TestPlotLine(t *testing.T) {
	tests := []struct {
		base       []float64
		prediction []float64
	}{
		{
			base:       []float64{1, 2, 3, 4, 5, 6},
			prediction: []float64{7, 8, 9, 10, 11},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if err := PlotLine(tt.base, tt.prediction, "test"); err != nil {
				t.Error(err)
				return
			}
		})
	}
}
