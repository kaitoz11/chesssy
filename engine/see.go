// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// Static exchange evaluation answers "if both sides keep capturing on this
// square, taking the cheapest attacker first, does the side that moves first
// come out at least `threshold` ahead?" It is used to skip losing captures in
// quiescence search and to order moves, so it must be cheap: no moves are
// played, only an occupancy mask is updated as pieces come off.

// seeValues are the exchange values used here, in centipawns. They are deliberately
// coarse: the algorithm only needs a consistent ordering, not an accurate one.
var seeValues = [PieceTypeCount]int{
	Pawn:   100,
	Knight: 320,
	Bishop: 330,
	Rook:   500,
	Queen:  950,
	King:   0,
}

// SeeValue returns the exchange value of a piece type in centipawns, so that
// callers ordering captures rank them the same way SeeGE does.
func SeeValue(pt PieceType) int { return seeValues[pt] }

// SeeGE reports whether the static exchange evaluation of m is greater than or
// equal to threshold. m must be legal in p.
func (p *Position) SeeGE(m Move, threshold int) bool {
	from, to := m.From(), m.To()

	// Castling never wins or loses material, and the king's path is already
	// known to be safe.
	if m.Kind() == CastlingMove {
		return threshold <= 0
	}

	gain := 0
	if captured := p.board[to]; captured != NoPiece {
		gain = seeValues[captured.Type()]
	}
	risked := seeValues[p.board[from].Type()]

	switch m.Kind() {
	case EnPassantMove:
		gain = seeValues[Pawn]
	case PromotionMove:
		gain += seeValues[m.Promotion()] - seeValues[Pawn]
		risked = seeValues[m.Promotion()]
	}

	// Best case: we keep the captured material and lose nothing.
	swap := gain - threshold
	if swap < 0 {
		return false
	}
	// Worst case: the moving piece is lost immediately and we still come out
	// ahead, so no further search is needed.
	swap = risked - swap
	if swap <= 0 {
		return true
	}

	occupied := (p.Occupied() ^ SquareBB(from)) | SquareBB(to)
	if m.Kind() == EnPassantMove {
		occupied ^= SquareBB(to.Shift(-PawnPush(p.side)))
	}

	attackers := p.AttackersTo(to, occupied) & occupied
	stm := p.side.Flip()
	result := 1

	for {
		stmAttackers := attackers & p.byColor[stm]
		if stmAttackers == 0 {
			break
		}
		result ^= 1

		// Recapture with the least valuable attacker.
		pt := Pawn
		var bb Bitboard
		for ; pt <= King; pt++ {
			if bb = stmAttackers & p.byType[pt]; bb != 0 {
				break
			}
		}
		if pt == King {
			// A king may only recapture when the square is otherwise undefended.
			if attackers & ^p.byColor[stm] != 0 {
				result ^= 1
			}
			return result != 0
		}

		if swap = seeValues[pt] - swap; swap < result {
			break
		}

		occupied ^= SquareBB(bb.LSB())
		// Removing a piece can uncover a slider behind it.
		if pt == Pawn || pt == Bishop || pt == Queen {
			attackers |= BishopAttacks(to, occupied) & (p.byType[Bishop] | p.byType[Queen])
		}
		if pt == Rook || pt == Queen {
			attackers |= RookAttacks(to, occupied) & (p.byType[Rook] | p.byType[Queen])
		}
		attackers &= occupied
		stm = stm.Flip()
	}
	return result != 0
}

// See returns the exchange value of m in centipawns, found by binary search over
// SeeGE. It is only used for reporting and tests; search paths use SeeGE.
func (p *Position) See(m Move) int {
	low, high := -2000, 2000
	for low < high {
		mid := (low + high + 1) / 2
		if p.SeeGE(m, mid) {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return low
}
