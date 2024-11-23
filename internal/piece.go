package internal

type Piece int32

const (
	WHITE_PAWN   Piece = 1
	WHITE_KNIGHT Piece = 3
	WHITE_BISHOP Piece = 4
	WHITE_ROOK   Piece = 5
	WHITE_QUEEN  Piece = 7
	WHITE_KING   Piece = 10

	BLACK_PAWN   Piece = -1
	BLACK_KNIGHT Piece = -3
	BLACK_BISHOP Piece = -4
	BLACK_ROOK   Piece = -5
	BLACK_QUEEN  Piece = -7
	BLACK_KING   Piece = -10

	NO_PIECE Piece = 0
)

func ToPiece(i int32) Piece {
	switch Piece(i) {
	case WHITE_PAWN:
		return WHITE_PAWN
	case WHITE_KNIGHT:
		return WHITE_KNIGHT
	case WHITE_BISHOP:
		return WHITE_BISHOP
	case WHITE_ROOK:
		return WHITE_ROOK
	case WHITE_QUEEN:
		return WHITE_QUEEN
	case WHITE_KING:
		return WHITE_KING
	case BLACK_PAWN:
		return BLACK_PAWN
	case BLACK_KNIGHT:
		return BLACK_KNIGHT
	case BLACK_BISHOP:
		return BLACK_BISHOP
	case BLACK_ROOK:
		return BLACK_ROOK
	case BLACK_QUEEN:
		return BLACK_QUEEN
	case BLACK_KING:
		return BLACK_KING
	default:
		return NO_PIECE
	}
}

func (p Piece) ToChar() rune {

    switch p {
        case BLACK_PAWN:
            return 'p'
        case WHITE_PAWN:
            return 'P'

        case BLACK_KNIGHT:
            return 'n'
        case WHITE_KNIGHT:
            return 'N'

        case BLACK_BISHOP:
            return 'b'
        case WHITE_BISHOP:
            return 'B'

        case BLACK_ROOK:
            return 'r'
        case WHITE_ROOK:
            return 'R'

        case BLACK_QUEEN:
            return 'q'
        case WHITE_QUEEN:
            return 'Q'

        case BLACK_KING:
            return 'k'
        case WHITE_KING:
            return 'K'

	   default:
	       return '.'
	}
}

func (p Piece) Point() int {
    
	switch p {
        case BLACK_PAWN:
	    case WHITE_PAWN:
	       return 1

       case BLACK_KNIGHT:
	   case WHITE_KNIGHT:
	       return 3

	   case BLACK_BISHOP:
	   case WHITE_BISHOP:
	       return 3

	   case BLACK_ROOK:
	   case WHITE_ROOK:
	       return 5

	   case BLACK_QUEEN:
	   case WHITE_QUEEN:
	       return 9

	   case BLACK_KING:
	   case WHITE_KING:
	       return 0

	   default:
	       return -1
	}
    return -1
}
