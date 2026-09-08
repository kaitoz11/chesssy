package engine

// Position is a chess position: piece placement plus the side to move, castling
// rights, en passant target and move clocks. It also keeps the history needed to
// undo moves, so a search can reuse one Position for the whole tree instead of
// copying the board at every node.
//
// A Position is mutable and is not safe for concurrent use. Use Clone to hand a
// copy to another goroutine.
type Position struct {
	// byType and byColor are the occupancy bitboards; board is the same
	// information keyed by square, which makes "what is on this square?" a
	// single load instead of a search through the bitboards.
	byType  [PieceTypeCount]Bitboard
	byColor [ColorCount]Bitboard
	board   [SquareCount]Piece

	side      Color
	castling  CastlingRights
	enPassant Square
	halfMoves uint16
	fullMoves uint16
	key       uint64

	// history grows with MakeMove and shrinks with UnmakeMove.
	history []state
}

// historyCapacity preallocates the move history for a deep search, so that
// MakeMove does not grow the slice mid-search.
const historyCapacity = 256

// NewPosition returns the standard starting position.
func NewPosition() *Position {
	p, err := NewPositionFromFEN(StartFEN)
	if err != nil {
		panic("engine: StartFEN is invalid: " + err.Error())
	}
	return p
}

// NewPositionFromFEN returns the position described by fen. The halfmove clock
// and fullmove number may be omitted.
func NewPositionFromFEN(fen string) (*Position, error) {
	p := &Position{history: make([]state, 0, historyCapacity)}
	if err := p.SetFEN(fen); err != nil {
		return nil, err
	}
	return p, nil
}

// Clone returns a deep copy of the position, including its move history.
func (p *Position) Clone() *Position {
	c := *p
	c.history = make([]state, len(p.history), max(cap(p.history), historyCapacity))
	copy(c.history, p.history)
	return &c
}

// SideToMove returns the color that moves next.
func (p *Position) SideToMove() Color { return p.side }

// CastlingRights returns the remaining castling privileges.
func (p *Position) CastlingRights() CastlingRights { return p.castling }

// EnPassantSquare returns the square a pawn may capture onto en passant, or
// NoSquare when there is none.
func (p *Position) EnPassantSquare() Square { return p.enPassant }

// HalfMoveClock returns the number of half moves since the last capture or pawn
// move, the counter the fifty-move rule uses.
func (p *Position) HalfMoveClock() int { return int(p.halfMoves) }

// FullMoveNumber returns the move number, starting at 1.
func (p *Position) FullMoveNumber() int { return int(p.fullMoves) }

// Key returns the Zobrist hash of the position, which identifies it for the
// transposition table and for repetition detection.
func (p *Position) Key() uint64 { return p.key }

// Ply returns the number of moves played since the position was last set with
// SetFEN.
func (p *Position) Ply() int { return len(p.history) }

// PieceAt returns the piece standing on s, or NoPiece.
func (p *Position) PieceAt(s Square) Piece { return p.board[s] }

// Occupied returns every occupied square.
func (p *Position) Occupied() Bitboard { return p.byColor[White] | p.byColor[Black] }

// ColorBB returns every square occupied by color c.
func (p *Position) ColorBB(c Color) Bitboard { return p.byColor[c] }

// TypeBB returns every square holding a piece of type pt, of either color.
func (p *Position) TypeBB(pt PieceType) Bitboard { return p.byType[pt] }

// PieceBB returns every square holding a piece of color c and type pt.
func (p *Position) PieceBB(c Color, pt PieceType) Bitboard { return p.byColor[c] & p.byType[pt] }

// KingSquare returns the square of c's king. Every position has exactly one king
// per side, which SetFEN enforces.
func (p *Position) KingSquare(c Color) Square { return (p.byColor[c] & p.byType[King]).LSB() }

// Count returns the number of pieces of color c and type pt.
func (p *Position) Count(c Color, pt PieceType) int { return p.PieceBB(c, pt).Count() }

// Board mutation. These three helpers are the only code that writes to the
// bitboards, the piece array and the Zobrist key, which is what keeps the three
// representations in agreement.

// put places piece on the empty square s.
func (p *Position) put(piece Piece, s Square) {
	bb := SquareBB(s)
	p.board[s] = piece
	p.byType[piece.Type()] |= bb
	p.byColor[piece.Color()] |= bb
	p.key ^= zobristPieces[piece][s]
}

// remove takes the piece standing on s off the board.
func (p *Position) remove(s Square) {
	piece := p.board[s]
	bb := ^SquareBB(s)
	p.board[s] = NoPiece
	p.byType[piece.Type()] &= bb
	p.byColor[piece.Color()] &= bb
	p.key ^= zobristPieces[piece][s]
}

// relocate moves the piece on from to the empty square to.
func (p *Position) relocate(from, to Square) {
	piece := p.board[from]
	bb := SquareBB(from) | SquareBB(to)
	p.board[from] = NoPiece
	p.board[to] = piece
	p.byType[piece.Type()] ^= bb
	p.byColor[piece.Color()] ^= bb
	p.key ^= zobristPieces[piece][from] ^ zobristPieces[piece][to]
}

// computeKey hashes the position from scratch. MakeMove maintains the same value
// incrementally; the tests compare the two to catch bookkeeping bugs.
func (p *Position) computeKey() uint64 {
	var key uint64
	for s := A1; s < SquareCount; s++ {
		if piece := p.board[s]; piece != NoPiece {
			key ^= zobristPieces[piece][s]
		}
	}
	key ^= zobristCastling[p.castling]
	if p.enPassant != NoSquare {
		key ^= zobristEnPassant[p.enPassant.File()]
	}
	if p.side == Black {
		key ^= zobristSide
	}
	return key
}
