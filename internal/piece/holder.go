package piece

type PieceHolder int8

const (
	WHITE_PAWN   PieceHolder = 1
	WHITE_KNIGHT PieceHolder = 3
	WHITE_BISHOP PieceHolder = 4
	WHITE_ROOK   PieceHolder = 5
	WHITE_QUEEN  PieceHolder = 7
	WHITE_KING   PieceHolder = 10

	BLACK_PAWN   PieceHolder = -1
	BLACK_KNIGHT PieceHolder = -3
	BLACK_BISHOP PieceHolder = -4
	BLACK_ROOK   PieceHolder = -5
	BLACK_QUEEN  PieceHolder = -7
	BLACK_KING   PieceHolder = -10

	NO_PIECE PieceHolder = 0
)

func FromCharToPiece(c rune) PieceHolder {
	switch c {
	case 'p':
		return BLACK_PAWN
	case 'P':
		return WHITE_PAWN

	case 'n':
		return BLACK_KNIGHT
	case 'N':
		return WHITE_KNIGHT

	case 'b':
		return BLACK_BISHOP
	case 'B':
		return WHITE_BISHOP

	case 'r':
		return BLACK_ROOK
	case 'R':
		return WHITE_ROOK

	case 'q':
		return BLACK_QUEEN
	case 'Q':
		return WHITE_QUEEN

	case 'k':
		return BLACK_KING
	case 'K':
		return WHITE_KING

	default:
		return NO_PIECE
	}
}

func ToPiece(i int8) PieceHolder {
	switch PieceHolder(i) {
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

func (p PieceHolder) ToChar() rune {

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

func (p PieceHolder) Point() int {
    
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
