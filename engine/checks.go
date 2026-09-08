package engine

// Attack, check and pin queries. Move generation and the search both need to ask
// "who attacks this square?", and they need it to be fast, so every query works
// straight off the bitboards without playing a move.

// AttackersTo returns every piece of either color that attacks s, resolving
// sliders against the given occupancy. Passing an occupancy that omits a piece is
// how the generator answers questions such as "would the king still be attacked
// there if it moved off this square?".
func (p *Position) AttackersTo(s Square, occupied Bitboard) Bitboard {
	return (PawnAttacks(White, s) & p.PieceBB(Black, Pawn)) |
		(PawnAttacks(Black, s) & p.PieceBB(White, Pawn)) |
		(KnightAttacks(s) & p.byType[Knight]) |
		(RookAttacks(s, occupied) & (p.byType[Rook] | p.byType[Queen])) |
		(BishopAttacks(s, occupied) & (p.byType[Bishop] | p.byType[Queen])) |
		(KingAttacks(s) & p.byType[King])
}

// Checkers returns the pieces giving check to the side to move.
func (p *Position) Checkers() Bitboard {
	return p.AttackersTo(p.KingSquare(p.side), p.Occupied()) & p.byColor[p.side.Flip()]
}

// InCheck reports whether the side to move is in check.
func (p *Position) InCheck() bool { return p.Checkers() != 0 }

// pinnedPieces returns the pieces of color c that shield c's king from an enemy
// slider. Such a piece may only move along the line between the slider and the
// king, which is what the generator uses it for.
func (p *Position) pinnedPieces(c Color) Bitboard {
	ksq := p.KingSquare(c)
	occupied := p.Occupied()

	// Sliders that would attack the king on an empty board are the only ones that
	// can be pinning anything.
	snipers := (RookAttacks(ksq, EmptyBB) & (p.byType[Rook] | p.byType[Queen])) |
		(BishopAttacks(ksq, EmptyBB) & (p.byType[Bishop] | p.byType[Queen]))
	snipers &= p.byColor[c.Flip()]

	var pinned Bitboard
	for snipers != 0 {
		between := BetweenBB(ksq, snipers.PopLSB()) & occupied
		if between != 0 && !between.MoreThanOne() {
			pinned |= between & p.byColor[c]
		}
	}
	return pinned
}

// CheckSquares returns, for each piece type, the squares from which a piece of
// the side to move would attack the enemy king. Moving a piece of that type onto
// one of its squares therefore gives check.
//
// Discovered checks are not included, so this is a fast approximation, suited to
// search decisions such as "never prune a checking move". Use GivesCheck when the
// answer has to be exact.
func (p *Position) CheckSquares() [PieceTypeCount]Bitboard {
	them := p.side.Flip()
	ksq := p.KingSquare(them)
	occupied := p.Occupied()

	rook := RookAttacks(ksq, occupied)
	bishop := BishopAttacks(ksq, occupied)

	return [PieceTypeCount]Bitboard{
		Pawn:   PawnAttacks(them, ksq),
		Knight: KnightAttacks(ksq),
		Bishop: bishop,
		Rook:   rook,
		Queen:  rook | bishop,
		King:   EmptyBB, // a king cannot give check
	}
}

// GivesCheck reports whether m, which must be legal in p, leaves the opponent in
// check. It plays and unplays the move, so prefer CheckSquares in hot loops.
func (p *Position) GivesCheck(m Move) bool {
	p.MakeMove(m)
	inCheck := p.InCheck()
	p.UnmakeMove()
	return inCheck
}
