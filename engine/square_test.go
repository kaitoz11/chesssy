package engine

import (
	"testing"
)

func TestSquareHelpers(t *testing.T) {
	if got := NewSquare(FileE, Rank4); got != E4 {
		t.Errorf("NewSquare(FileE, Rank4) = %v, want e4", got)
	}
	if got := E4.String(); got != "e4" {
		t.Errorf("E4.String() = %q, want %q", got, "e4")
	}
	if got := A1.Flip(); got != A8 {
		t.Errorf("A1.Flip() = %v, want a8", got)
	}
	if got, err := SquareFromString("h8"); err != nil || got != H8 {
		t.Errorf("SquareFromString(\"h8\") = %v, %v, want h8, nil", got, err)
	}
	if _, err := SquareFromString("j9"); err == nil {
		t.Error("SquareFromString(\"j9\") succeeded, want error")
	}
	for s := A1; s < SquareCount; s++ {
		if got, err := SquareFromString(s.String()); err != nil || got != s {
			t.Fatalf("round trip %v failed: %v, %v", s, got, err)
		}
	}
}
