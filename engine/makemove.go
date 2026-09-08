// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

// Move execution. MakeMove updates the occupancy bitboards, the piece array and
// the Zobrist key incrementally, and pushes everything it cannot recompute onto
// a history stack so that UnmakeMove is cheap. That is what lets a search reuse
// a single Position for the whole tree.

// state holds the information MakeMove destroys and UnmakeMove restores.
type state struct {
	move      Move
	captured  Piece
	castling  CastlingRights
	enPassant Square
	halfMoves uint16
	key       uint64
}

// MakeMove plays m, which must be legal in p.
func (p *Position) MakeMove(m Move) {
	us := p.side
	from, to := m.From(), m.To()
	moved := p.board[from]

	p.history = append(p.history, state{
		move:      m,
		captured:  p.board[to],
		castling:  p.castling,
		enPassant: p.enPassant,
		halfMoves: p.halfMoves,
		key:       p.key,
	})

	p.clearEnPassant()
	p.halfMoves++

	switch m.Kind() {
	case CastlingMove:
		info := &castlings[castlingIndexOf(us, to)]
		p.relocate(info.kingFrom, info.kingTo)
		p.relocate(info.rookFrom, info.rookTo)

	case EnPassantMove:
		captureSquare := to.Shift(-PawnPush(us))
		p.currentState().captured = p.board[captureSquare]
		p.remove(captureSquare)
		p.relocate(from, to)
		p.halfMoves = 0

	case PromotionMove:
		if p.board[to] != NoPiece {
			p.remove(to)
		}
		p.remove(from)
		p.put(NewPiece(us, m.Promotion()), to)
		p.halfMoves = 0

	default:
		if p.board[to] != NoPiece {
			p.remove(to)
			p.halfMoves = 0
		}
		p.relocate(from, to)
		if moved.Type() == Pawn {
			p.halfMoves = 0
			if isDoublePush(from, to) {
				p.setEnPassant(from.Shift(PawnPush(us)))
			}
		}
	}

	p.revokeCastlingRights(castlingLoss[from] | castlingLoss[to])
	p.passTurn()
}

// UnmakeMove undoes the most recent move.
func (p *Position) UnmakeMove() {
	st := p.popState()

	p.side = p.side.Flip()
	if p.side == Black {
		p.fullMoves--
	}

	from, to := st.move.From(), st.move.To()
	switch st.move.Kind() {
	case CastlingMove:
		info := &castlings[castlingIndexOf(p.side, to)]
		p.relocate(info.rookTo, info.rookFrom)
		p.relocate(info.kingTo, info.kingFrom)

	case EnPassantMove:
		p.relocate(to, from)
		p.put(st.captured, to.Shift(-PawnPush(p.side)))

	case PromotionMove:
		p.remove(to)
		p.put(NewPiece(p.side, Pawn), from)
		if st.captured != NoPiece {
			p.put(st.captured, to)
		}

	default:
		p.relocate(to, from)
		if st.captured != NoPiece {
			p.put(st.captured, to)
		}
	}

	p.restore(st)
}

// MakeNullMove passes the turn without moving a piece, which a search uses to
// ask "how bad is the position if I do nothing?". The side to move must not be
// in check.
func (p *Position) MakeNullMove() {
	p.history = append(p.history, state{
		move:      NoMove,
		captured:  NoPiece,
		castling:  p.castling,
		enPassant: p.enPassant,
		halfMoves: p.halfMoves,
		key:       p.key,
	})
	p.clearEnPassant()
	p.halfMoves++
	p.passTurn()
}

// UnmakeNullMove undoes MakeNullMove.
func (p *Position) UnmakeNullMove() {
	st := p.popState()
	p.side = p.side.Flip()
	if p.side == Black {
		p.fullMoves--
	}
	p.restore(st)
}

// isDoublePush reports whether a pawn move spans two ranks.
func isDoublePush(from, to Square) bool { return to-from == 16 || from-to == 16 }

func (p *Position) currentState() *state { return &p.history[len(p.history)-1] }

func (p *Position) popState() state {
	st := *p.currentState()
	p.history = p.history[:len(p.history)-1]
	return st
}

// restore puts back the fields that cannot be derived from the board.
func (p *Position) restore(st state) {
	p.castling = st.castling
	p.enPassant = st.enPassant
	p.halfMoves = st.halfMoves
	p.key = st.key
}

// passTurn hands the move to the other side, keeping the key and the move
// number in step.
func (p *Position) passTurn() {
	if p.side == Black {
		p.fullMoves++
	}
	p.side = p.side.Flip()
	p.key ^= zobristSide
}

func (p *Position) setEnPassant(s Square) {
	p.enPassant = s
	p.key ^= zobristEnPassant[s.File()]
}

func (p *Position) clearEnPassant() {
	if p.enPassant != NoSquare {
		p.key ^= zobristEnPassant[p.enPassant.File()]
		p.enPassant = NoSquare
	}
}

func (p *Position) revokeCastlingRights(lost CastlingRights) {
	if lost &= p.castling; lost == NoCastling {
		return
	}
	p.key ^= zobristCastling[p.castling]
	p.castling &= ^lost
	p.key ^= zobristCastling[p.castling]
}
