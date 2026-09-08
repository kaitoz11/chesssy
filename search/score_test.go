package search

import (
	"testing"
)

func TestMateDistanceConversion(t *testing.T) {
	tests := []struct {
		score int
		want  int
	}{
		{mateScore - 1, 1},
		{mateScore - 2, 1},
		{mateScore - 3, 2},
		{-mateScore + 1, -1},
		{-mateScore + 3, -2},
		{0, 0},
		{250, 0},
		{-250, 0},
	}
	for _, tc := range tests {
		if got := mateDistance(tc.score); got != tc.want {
			t.Errorf("mateDistance(%d) = %d, want %d", tc.score, got, tc.want)
		}
	}
}
