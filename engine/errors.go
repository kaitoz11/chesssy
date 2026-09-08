package engine

import "errors"

// Errors reported when parsing external input. They are wrapped with detail, so
// classify them with errors.Is rather than by comparing messages:
//
//	if errors.Is(err, engine.ErrIllegalMove) { ... }
var (
	// ErrInvalidFEN means a Forsyth-Edwards notation string is malformed or
	// describes a position that cannot occur, such as one without kings.
	ErrInvalidFEN = errors.New("invalid FEN")

	// ErrInvalidSquare means a square was not written as a file letter followed
	// by a rank digit, such as "e4".
	ErrInvalidSquare = errors.New("invalid square")

	// ErrIllegalMove means a move is well formed but not legal in the position
	// it was given for.
	ErrIllegalMove = errors.New("illegal move")
)
