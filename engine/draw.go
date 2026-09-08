// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// Draw detection. These are the three drawing rules a search needs to know
// about; stalemate falls out of move generation instead (see IsStalemate).

// IsRepetition reports whether the current position has occurred earlier in the
// move history. A single repetition counts, which is what a search wants: if a
// line can be repeated once it can usually be repeated twice.
func (p *Position) IsRepetition() bool {
	// Only positions since the last irreversible move can repeat, and the same
	// side must be to move, so step back two plies at a time.
	oldest := max(len(p.history)-int(p.halfMoves), 0)
	for i := len(p.history) - 2; i >= oldest; i -= 2 {
		if p.history[i].key == p.key {
			return true
		}
	}
	return false
}

// IsFiftyMoveDraw reports whether the fifty-move rule applies, meaning a hundred
// half moves have passed without a capture or a pawn move.
func (p *Position) IsFiftyMoveDraw() bool { return p.halfMoves >= 100 }

// HasInsufficientMaterial reports whether neither side can possibly deliver
// mate: bare kings, or a single minor piece against a bare king.
func (p *Position) HasInsufficientMaterial() bool {
	if p.byType[Pawn]|p.byType[Rook]|p.byType[Queen] != 0 {
		return false
	}
	minors := p.byType[Knight] | p.byType[Bishop]
	return !minors.MoreThanOne()
}

// IsDraw reports whether the position is drawn by repetition, by the fifty-move
// rule or by insufficient material.
func (p *Position) IsDraw() bool {
	return p.IsFiftyMoveDraw() || p.HasInsufficientMaterial() || p.IsRepetition()
}
