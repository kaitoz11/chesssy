package engine

import "testing"

func TestPositionAccessors(t *testing.T) {
	p, err := NewPositionFromFEN("rnbqkb1r/ppp1pppp/3p1n2/8/3PP3/2N5/PPP2PPP/R1BQKBNR b KQkq d3 3 7")
	if err != nil {
		t.Fatal(err)
	}

	if got := p.SideToMove(); got != Black {
		t.Errorf("SideToMove() = %v, want black", got)
	}
	if got := p.CastlingRights(); got != AllCastling {
		t.Errorf("CastlingRights() = %v, want %v", got, AllCastling)
	}
	if got := p.EnPassantSquare(); got != D3 {
		t.Errorf("EnPassantSquare() = %v, want d3", got)
	}
	if got := p.HalfMoveClock(); got != 3 {
		t.Errorf("HalfMoveClock() = %d, want 3", got)
	}
	if got := p.FullMoveNumber(); got != 7 {
		t.Errorf("FullMoveNumber() = %d, want 7", got)
	}
	if got := p.Ply(); got != 0 {
		t.Errorf("Ply() = %d, want 0 for a freshly parsed position", got)
	}
	if got := p.PieceAt(E4); got != WhitePawn {
		t.Errorf("PieceAt(e4) = %v, want a white pawn", got)
	}
	if got := p.KingSquare(White); got != E1 {
		t.Errorf("KingSquare(white) = %v, want e1", got)
	}
	if got := p.Count(Black, Pawn); got != 8 {
		t.Errorf("Count(black, pawn) = %d, want 8", got)
	}
	if got := p.Occupied().Count(); got != 32 {
		t.Errorf("Occupied().Count() = %d, want 32", got)
	}
	if got := p.PieceBB(White, Knight); got != SquareBB(C3)|SquareBB(G1) {
		t.Errorf("PieceBB(white, knight) =\n%v", got)
	}
	if got := p.TypeBB(Queen); got != SquareBB(D1)|SquareBB(D8) {
		t.Errorf("TypeBB(queen) =\n%v", got)
	}

	// Ply counts the moves played since the position was set.
	m, err := p.ParseMove("b8c6")
	if err != nil {
		t.Fatal(err)
	}
	p.MakeMove(m)
	if got := p.Ply(); got != 1 {
		t.Errorf("Ply() after one move = %d, want 1", got)
	}
	if got := p.FullMoveNumber(); got != 8 {
		t.Errorf("FullMoveNumber() after black's move = %d, want 8", got)
	}
}

func TestGivesCheck(t *testing.T) {
	// A rook on d1 with the black king on e8: Rd1e1 does not check, Rd8 does.
	p, err := NewPositionFromFEN("4k3/8/8/8/8/8/8/3RK3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	before := p.FEN()

	tests := []struct {
		move string
		want bool
	}{
		{"d1d8", true},
		{"d1d7", false},
		{"e1e2", false},
	}
	for _, tc := range tests {
		m, err := p.ParseMove(tc.move)
		if err != nil {
			t.Fatalf("%s: %v", tc.move, err)
		}
		if got := p.GivesCheck(m); got != tc.want {
			t.Errorf("GivesCheck(%s) = %v, want %v", tc.move, got, tc.want)
		}
	}
	if got := p.FEN(); got != before {
		t.Errorf("GivesCheck modified the position:\ngot  %s\nwant %s", got, before)
	}
}
