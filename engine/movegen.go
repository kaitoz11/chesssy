// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import "fmt"

// Move generation produces strictly legal moves. Rather than generating
// pseudo-legal moves and filtering them by playing each one, the generator works
// from three pieces of information:
//
//   - the checkers: when the king is attacked twice only a king move can help, and
//     when it is attacked once every other move must capture the checker or block
//     the line to it;
//   - the pinned pieces: a piece shielding its king from a slider may only move
//     along that line;
//   - king safety: king destinations are tested against an occupancy with the king
//     removed, so a slider is not blocked by the very piece that is moving.
//
// En passant is the one case that still needs a simulated occupancy, because
// removing the captured pawn can expose the king along a rank.

// genMode selects how much of the move list a caller wants.
type genMode int

const (
	// genAll generates every legal move.
	genAll genMode = iota
	// genTactical generates the moves a quiescence search looks at: captures, en
	// passant and queen promotions. In check it falls back to genAll, because an
	// evasion may be quiet.
	genTactical
)

// GenerateMoves fills list with every legal move in the position.
func (p *Position) GenerateMoves(list *MoveList) { p.generate(list, genAll) }

// GenerateCaptures fills list with the moves a quiescence search cares about:
// captures, en passant and queen promotions. When the side to move is in check it
// generates every legal move instead, since evasions may be quiet.
func (p *Position) GenerateCaptures(list *MoveList) { p.generate(list, genTactical) }

func (p *Position) generate(list *MoveList, mode genMode) {
	list.Reset()

	us, them := p.side, p.side.Flip()
	ourPieces, theirPieces := p.byColor[us], p.byColor[them]
	occupied := ourPieces | theirPieces
	ksq := p.KingSquare(us)
	checkers := p.Checkers()

	// Quiet moves are wanted unless the caller asked for tactical moves only, and
	// even then they are needed to escape check.
	wantQuiets := mode == genAll || checkers != 0

	p.generateKingMoves(list, ksq, ourPieces, theirPieces, occupied, wantQuiets)

	// Under double check no other piece can resolve the check.
	if checkers.MoreThanOne() {
		return
	}

	// Squares a non-king move may land on and still leave the king safe.
	resolveCheck := FullBB
	if checkers != 0 {
		checker := checkers.LSB()
		resolveCheck = BetweenBB(ksq, checker) | checkers
	}
	quietMask := ^occupied & resolveCheck
	captureMask := theirPieces & resolveCheck
	pinned := p.pinnedPieces(us)

	p.generatePawnMoves(list, quietMask, captureMask, pinned, ksq, wantQuiets)

	pieceMask := captureMask
	if wantQuiets {
		pieceMask |= quietMask
	}
	for pt := Knight; pt <= Queen; pt++ {
		for bb := p.PieceBB(us, pt); bb != 0; {
			from := bb.PopLSB()
			moves := PieceAttacks(pt, from, occupied) & pieceMask
			if pinned.Has(from) {
				moves &= LineBB(ksq, from)
			}
			for moves != 0 {
				list.Add(NewMove(from, moves.PopLSB()))
			}
		}
	}

	if wantQuiets && checkers == 0 {
		p.generateCastling(list, occupied, theirPieces)
	}
}

// generateKingMoves adds the king moves that do not walk into an attack. The
// occupancy used for the test omits the king itself, so a slider checking along a
// line does not appear to be blocked by the piece that is moving away.
func (p *Position) generateKingMoves(list *MoveList, ksq Square, ourPieces, theirPieces, occupied Bitboard, wantQuiets bool) {
	targets := KingAttacks(ksq) & ^ourPieces
	if !wantQuiets {
		targets &= theirPieces
	}
	occupiedNoKing := occupied ^ SquareBB(ksq)
	for targets != 0 {
		to := targets.PopLSB()
		if p.AttackersTo(to, occupiedNoKing)&theirPieces == 0 {
			list.Add(NewMove(ksq, to))
		}
	}
}

func (p *Position) generatePawnMoves(list *MoveList, quietMask, captureMask, pinned Bitboard, ksq Square, wantQuiets bool) {
	us, them := p.side, p.side.Flip()
	pawns := p.PieceBB(us, Pawn)
	if pawns == 0 {
		return
	}

	up := PawnPush(us)
	empty := ^p.Occupied()
	promoRank := RelativeRankBB(us, Rank8BB)
	doublePushRank := RelativeRankBB(us, Rank3BB)

	pushes := pawns.Shift(up) & empty
	doublePushes := (pushes & doublePushRank).Shift(up) & quietMask
	pushes &= quietMask
	if !wantQuiets {
		// A quiescence search only wants forcing pushes, which means promotions.
		pushes &= promoRank
		doublePushes = EmptyBB
	}

	// add records one pawn move, skipping it when a pinned pawn would step off the
	// line to its king, and expanding promotions.
	add := func(from, to Square) {
		if pinned.Has(from) && !Aligned(ksq, from, to) {
			return
		}
		switch {
		case !promoRank.Has(to):
			list.Add(NewMove(from, to))
		case wantQuiets:
			for promo := Queen; promo >= Knight; promo-- {
				list.Add(NewPromotion(from, to, promo))
			}
		default:
			list.Add(NewPromotion(from, to, Queen))
		}
	}

	for bb := pushes; bb != 0; {
		to := bb.PopLSB()
		add(to.Shift(-up), to)
	}
	for bb := doublePushes; bb != 0; {
		to := bb.PopLSB()
		add(to.Shift(-up).Shift(-up), to)
	}
	for _, dir := range [2]Direction{up + East, up + West} {
		for bb := pawns.Shift(dir) & captureMask; bb != 0; {
			to := bb.PopLSB()
			add(to.Shift(-dir), to)
		}
	}

	if p.enPassant != NoSquare {
		p.generateEnPassant(list, pawns, ksq, them, up)
	}
}

// generateEnPassant adds en passant captures. Legality needs the occupancy that
// results from the capture, because the moving pawn and the captured pawn both
// leave the board and can uncover a slider along the rank.
func (p *Position) generateEnPassant(list *MoveList, pawns Bitboard, ksq Square, them Color, up Direction) {
	captured := SquareBB(p.enPassant.Shift(-up))
	for bb := PawnAttacks(them, p.enPassant) & pawns; bb != 0; {
		from := bb.PopLSB()
		after := (p.Occupied() ^ SquareBB(from) ^ captured) | SquareBB(p.enPassant)
		if p.AttackersTo(ksq, after)&p.byColor[them] & ^captured == 0 {
			list.Add(NewSpecialMove(from, p.enPassant, EnPassantMove))
		}
	}
}

// generateCastling adds the castling moves whose path is clear and unattacked.
// The caller has already established that the king is not in check.
func (p *Position) generateCastling(list *MoveList, occupied, theirPieces Bitboard) {
	us := p.side
	rook := NewPiece(us, Rook)

	for i := 2 * int(us); i < 2*int(us)+2; i++ {
		if p.castling&CastlingRights(1<<uint(i)) == 0 {
			continue
		}
		info := &castlings[i]
		// The rook is guaranteed by the rights in a legal position, but a FEN can
		// claim rights that the placement does not support.
		if occupied&info.emptyPath != 0 || p.board[info.rookFrom] != rook {
			continue
		}
		if p.anyAttacked(info.safePath, theirPieces, occupied) {
			continue
		}
		list.Add(NewSpecialMove(info.kingFrom, info.kingTo, CastlingMove))
	}
}

// anyAttacked reports whether any square in squares is attacked by a piece in
// byPieces.
func (p *Position) anyAttacked(squares, byPieces, occupied Bitboard) bool {
	for squares != 0 {
		if p.AttackersTo(squares.PopLSB(), occupied)&byPieces != 0 {
			return true
		}
	}
	return false
}

// HasLegalMoves reports whether the side to move has at least one legal move.
func (p *Position) HasLegalMoves() bool {
	var list MoveList
	p.GenerateMoves(&list)
	return list.Len() > 0
}

// IsCheckmate reports whether the side to move is checkmated.
func (p *Position) IsCheckmate() bool { return p.InCheck() && !p.HasLegalMoves() }

// IsStalemate reports whether the side to move has no legal move but is not in
// check.
func (p *Position) IsStalemate() bool { return !p.InCheck() && !p.HasLegalMoves() }

// IsCapture reports whether m takes a piece, including en passant.
func (p *Position) IsCapture(m Move) bool {
	return p.board[m.To()] != NoPiece || m.Kind() == EnPassantMove
}

// IsQuiet reports whether m neither captures nor promotes.
func (p *Position) IsQuiet(m Move) bool { return !p.IsCapture(m) && !m.IsPromotion() }

// ParseMove converts UCI notation such as "e2e4" or "e7e8q" into a legal move of
// the current position. Matching against the generated moves is what fills in the
// move kind, and it rejects illegal input as a side effect.
func (p *Position) ParseMove(s string) (Move, error) {
	if len(s) != 4 && len(s) != 5 {
		return NoMove, fmt.Errorf("%w: %q is not 4 or 5 characters", ErrIllegalMove, s)
	}
	from, err := SquareFromString(s[0:2])
	if err != nil {
		return NoMove, err
	}
	to, err := SquareFromString(s[2:4])
	if err != nil {
		return NoMove, err
	}

	promo := NoPiece
	if len(s) == 5 {
		// UCI writes the promotion piece in lower case; accept either case.
		if promo = PieceFromChar(s[4] | 0x20); promo == NoPiece {
			return NoMove, fmt.Errorf("%w: unknown promotion piece in %q", ErrIllegalMove, s)
		}
	}

	var list MoveList
	p.GenerateMoves(&list)
	for _, m := range list.Moves() {
		if m.From() != from || m.To() != to {
			continue
		}
		if m.IsPromotion() != (promo != NoPiece) {
			continue
		}
		if promo != NoPiece && m.Promotion() != promo.Type() {
			continue
		}
		return m, nil
	}
	return NoMove, fmt.Errorf("%w: %q", ErrIllegalMove, s)
}
