package engine

import (
	"testing"
)

func TestBitboardShifts(t *testing.T) {
	tests := []struct {
		name      string
		from      Square
		direction Direction
		want      Bitboard
	}{
		{"north", D4, North, SquareBB(D5)},
		{"south", D4, South, SquareBB(D3)},
		{"east", D4, East, SquareBB(E4)},
		{"west", D4, West, SquareBB(C4)},
		{"north east", D4, NorthEast, SquareBB(E5)},
		{"south west", D4, SouthWest, SquareBB(C3)},
		{"east wraps off board", H4, East, EmptyBB},
		{"west wraps off board", A4, West, EmptyBB},
		{"north east wraps off board", H4, NorthEast, EmptyBB},
		{"north off board", D8, North, EmptyBB},
		{"south off board", D1, South, EmptyBB},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := SquareBB(tc.from).Shift(tc.direction); got != tc.want {
				t.Errorf("got:\n%v\nwant:\n%v", got, tc.want)
			}
		})
	}
}
