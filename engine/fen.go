package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// StartFEN is the standard starting position in Forsyth-Edwards notation.
const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// SetFEN resets the position to the one described by fen, discarding any move
// history. The halfmove clock and fullmove number may be omitted.
//
// The position is left untouched when fen is invalid, so a rejected string
// cannot leave a half-built board behind.
func (p *Position) SetFEN(fen string) error {
	fields := strings.Fields(fen)
	if len(fields) < 4 {
		return fmt.Errorf("%w: needs at least 4 fields, got %d", ErrInvalidFEN, len(fields))
	}

	// Build into a copy so that an error leaves the receiver unchanged.
	next := *p
	next.byType = [PieceTypeCount]Bitboard{}
	next.byColor = [ColorCount]Bitboard{}
	next.key = 0
	for i := range next.board {
		next.board[i] = NoPiece
	}

	if err := next.parsePlacement(fields[0]); err != nil {
		return err
	}
	if err := next.parseSideToMove(fields[1]); err != nil {
		return err
	}
	if err := next.parseCastlingRights(fields[2]); err != nil {
		return err
	}
	if err := next.parseEnPassantField(fields[3]); err != nil {
		return err
	}
	if err := next.parseClocks(fields); err != nil {
		return err
	}

	for c := White; c < ColorCount; c++ {
		if kings := next.byColor[c] & next.byType[King]; kings.Count() != 1 {
			return fmt.Errorf("%w: %s has %d kings, want exactly 1", ErrInvalidFEN, c, kings.Count())
		}
	}

	next.key = next.computeKey()
	next.history = p.history[:0]
	*p = next
	return nil
}

// parsePlacement fills the board from the piece placement field, which lists the
// ranks from 8 down to 1.
func (p *Position) parsePlacement(placement string) error {
	ranks := strings.Split(placement, "/")
	if len(ranks) != RankCount {
		return fmt.Errorf("%w: needs 8 ranks, got %d", ErrInvalidFEN, len(ranks))
	}

	for i, row := range ranks {
		r := Rank(RankCount - 1 - i)
		f := FileA
		for j := 0; j < len(row); j++ {
			c := row[j]
			if c >= '1' && c <= '8' {
				f += File(c - '0')
				continue
			}
			piece := PieceFromChar(c)
			if piece == NoPiece {
				return fmt.Errorf("%w: unknown piece %q", ErrInvalidFEN, string(rune(c)))
			}
			if f > FileH {
				return fmt.Errorf("%w: rank %d is too long", ErrInvalidFEN, r+1)
			}
			p.put(piece, NewSquare(f, r))
			f++
		}
		if f != FileCount {
			return fmt.Errorf("%w: rank %d does not cover 8 files", ErrInvalidFEN, r+1)
		}
	}
	return nil
}

func (p *Position) parseSideToMove(field string) error {
	switch field {
	case "w":
		p.side = White
	case "b":
		p.side = Black
	default:
		return fmt.Errorf("%w: side to move %q is not w or b", ErrInvalidFEN, field)
	}
	return nil
}

func (p *Position) parseCastlingRights(field string) error {
	p.castling = NoCastling
	if field == "-" {
		return nil
	}
	for i := 0; i < len(field); i++ {
		switch field[i] {
		case 'K':
			p.castling |= WhiteKingSide
		case 'Q':
			p.castling |= WhiteQueenSide
		case 'k':
			p.castling |= BlackKingSide
		case 'q':
			p.castling |= BlackQueenSide
		default:
			return fmt.Errorf("%w: castling field %q", ErrInvalidFEN, field)
		}
	}
	return nil
}

func (p *Position) parseEnPassantField(field string) error {
	p.enPassant = NoSquare
	if field == "-" {
		return nil
	}
	sq, err := SquareFromString(field)
	if err != nil {
		return fmt.Errorf("%w: en passant target: %w", ErrInvalidFEN, err)
	}
	p.enPassant = sq
	return nil
}

// parseClocks reads the optional halfmove clock and fullmove number.
func (p *Position) parseClocks(fields []string) error {
	p.halfMoves, p.fullMoves = 0, 1
	if len(fields) > 4 {
		n, err := strconv.Atoi(fields[4])
		if err != nil || n < 0 {
			return fmt.Errorf("%w: halfmove clock %q", ErrInvalidFEN, fields[4])
		}
		p.halfMoves = uint16(n)
	}
	if len(fields) > 5 {
		n, err := strconv.Atoi(fields[5])
		if err != nil || n < 1 {
			return fmt.Errorf("%w: fullmove number %q", ErrInvalidFEN, fields[5])
		}
		p.fullMoves = uint16(n)
	}
	return nil
}

// FEN renders the position in Forsyth-Edwards notation.
func (p *Position) FEN() string {
	side := "w"
	if p.side == Black {
		side = "b"
	}
	return fmt.Sprintf("%s %s %s %s %d %d",
		p.placement(), side, p.castling, p.enPassant, p.halfMoves, p.fullMoves)
}

// placement renders the piece placement field, ranks 8 down to 1.
func (p *Position) placement() string {
	var sb strings.Builder
	for r := int(Rank8); r >= int(Rank1); r-- {
		empty := 0
		for f := FileA; f <= FileH; f++ {
			piece := p.board[NewSquare(f, Rank(r))]
			if piece == NoPiece {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteString(strconv.Itoa(empty))
				empty = 0
			}
			sb.WriteString(piece.String())
		}
		if empty > 0 {
			sb.WriteString(strconv.Itoa(empty))
		}
		if r > int(Rank1) {
			sb.WriteByte('/')
		}
	}
	return sb.String()
}

// String renders the board as a grid with rank 8 on top, followed by the FEN
// fields other than the placement.
func (p *Position) String() string {
	var sb strings.Builder
	for r := int(Rank8); r >= int(Rank1); r-- {
		fmt.Fprintf(&sb, "%d:", r+1)
		for f := FileA; f <= FileH; f++ {
			sb.WriteByte(' ')
			sb.WriteString(p.board[NewSquare(f, Rank(r))].String())
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("   a b c d e f g h\n\n")

	fields := strings.Fields(p.FEN())
	fmt.Fprintf(&sb, "   %s\n", strings.Join(fields[1:], " "))
	return sb.String()
}
