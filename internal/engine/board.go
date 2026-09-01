package engine

import (
	"chesssy/internal/piece"
	"errors"
	"fmt"
)

// type state struct {
//     isKingMoved bool
//     isKSideRookMoved bool
//     isQSideRookMoved bool
// }


type CastlingRights byte

const (
	WHITE_K CastlingRights = 0b1000
	WHITE_Q CastlingRights = 0b0100
	BLACK_k CastlingRights = 0b0010
	BLACK_q CastlingRights = 0b0001
)

type Board struct {
	holder [64]piece.PieceHolder

	pieces Bitmap

	gameState       CastlingRights
	enPassantSquare int
	moveNumber      int
	// blackState *state
	// whiteState *state

	nextMove Player
}

func NewBoardWith(holder [64]piece.PieceHolder, nextMove Player, gameState CastlingRights, enPassantSquare, moveNumber int) *Board {
	return &Board{
		holder:          holder,
		nextMove:        nextMove,
		gameState:       gameState,
		enPassantSquare: enPassantSquare,
		moveNumber:      moveNumber,
	}
}

func (b *Board) UpdateCastle(side CastlingRights, p Player) error {
	var resetBytes CastlingRights
	if p == WHITE {
		resetBytes = ^(WHITE_K | WHITE_Q)
	} else {
		resetBytes = ^(BLACK_k | BLACK_q)
	}

	isInvalidPlayerSide := side&resetBytes != 0
	if isInvalidPlayerSide {
		return errors.New("Not valid Player")
	}

	isNotCastlingAvailable := b.gameState&side == 0
	if isNotCastlingAvailable {
		return errors.New("Castling is not available")
	}

	b.gameState = b.gameState & resetBytes
	fmt.Printf("Update state to: %4b - %v\n", b.gameState, b.gameState)
	return nil
}

func (b *Board) Print() {
	for index, square := range b.holder {
		fmt.Printf("%c\t", square.ToChar())
		if (index+1)%8 == 0 {
			fmt.Print("\n\n")
		}
	}
	fmt.Println(b.enPassantSquare)
}

func NewBoard() *Board {
	boardHolder := [64]piece.PieceHolder{
		piece.BLACK_ROOK, piece.BLACK_KNIGHT, piece.BLACK_BISHOP, piece.BLACK_QUEEN, piece.BLACK_KING, piece.BLACK_BISHOP, piece.BLACK_KNIGHT, piece.BLACK_ROOK,
		piece.BLACK_PAWN, piece.BLACK_PAWN, piece.BLACK_PAWN, piece.BLACK_PAWN, piece.BLACK_PAWN, piece.BLACK_PAWN, piece.BLACK_PAWN, piece.BLACK_PAWN,
		piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE,
		piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE,
		piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE,
		piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE, piece.NO_PIECE,
		piece.WHITE_PAWN, piece.WHITE_PAWN, piece.WHITE_PAWN, piece.WHITE_PAWN, piece.WHITE_PAWN, piece.WHITE_PAWN, piece.WHITE_PAWN, piece.WHITE_PAWN,
		piece.WHITE_ROOK, piece.WHITE_KNIGHT, piece.WHITE_BISHOP, piece.WHITE_QUEEN, piece.WHITE_KING, piece.WHITE_BISHOP, piece.WHITE_KNIGHT, piece.WHITE_ROOK,
	}
	return &Board{
		holder: boardHolder,

		// KQkq (Uppercase -> White)
		gameState: 0b1111,

		enPassantSquare: -1,

		nextMove: WHITE,
	}
}

func (b *Board) Move(from, to Square) error {
	if from == to {
		return errors.New("invalid move")
	}

	if b.holder[from] == piece.NO_PIECE {
		return errors.New("invalid move")
	}

	b.holder[from] = piece.NO_PIECE
	b.holder[to] = b.holder[from]
	b.holder[from] = piece.NO_PIECE
	return nil
}

