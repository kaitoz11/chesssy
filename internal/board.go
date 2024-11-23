package internal

import (
	"errors"
	"fmt"
	"strings"
)

// type state struct {
//     isKingMoved bool
//     isKSideRookMoved bool
//     isQSideRookMoved bool
// }

type Player int

const (
    WHITE Player = 0
    BLACK Player = 1
)

type CastlingRights byte
const (
    WHITE_K CastlingRights = 0b1000
    WHITE_Q CastlingRights = 0b0100
    BLACK_k CastlingRights = 0b0010
    BLACK_q CastlingRights = 0b0001
)

type Board struct{
    holder [64]Piece

    gameState CastlingRights
    // blackState *state
    // whiteState *state

    nextMove Player
}


func (b *Board) UpdateCastle(side CastlingRights, p Player) error {
    var resetBytes CastlingRights
    if p == WHITE {
        resetBytes = ^(WHITE_K | WHITE_Q)
    } else {
        resetBytes = ^(BLACK_k | BLACK_q)
    }

    isInvalidPlayerSide := side & resetBytes != 0 
    if isInvalidPlayerSide {
        return errors.New("Not valid Player")
    }
    
    isNotCastlingAvailable := b.gameState & side == 0
    if isNotCastlingAvailable {
        return errors.New("Castling is not available")
    }

    b.gameState = b.gameState & resetBytes
    fmt.Printf("Update state to: %4b - %v\n", b.gameState, b.gameState)
    return nil
}

func (b *Board) Print(){
    for index, square := range b.holder {
        fmt.Printf("%c\t",square.ToChar())
        if (index + 1) % 8 == 0 {
            fmt.Print("\n\n")
        }
    }
    fmt.Println()
}

func NewBoard() *Board {
    boardHolder := [64]Piece{
        BLACK_ROOK, BLACK_KNIGHT, BLACK_BISHOP, BLACK_QUEEN, BLACK_KING, BLACK_BISHOP, BLACK_KNIGHT, BLACK_ROOK,
        BLACK_PAWN, BLACK_PAWN, BLACK_PAWN, BLACK_PAWN, BLACK_PAWN, BLACK_PAWN, BLACK_PAWN, BLACK_PAWN,
        NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE,
        NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE,
        NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE,
        NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE, NO_PIECE,
        WHITE_PAWN, WHITE_PAWN, WHITE_PAWN, WHITE_PAWN, WHITE_PAWN, WHITE_PAWN, WHITE_PAWN, WHITE_PAWN,
        WHITE_ROOK, WHITE_KNIGHT, WHITE_BISHOP, WHITE_QUEEN, WHITE_KING, WHITE_BISHOP, WHITE_KNIGHT, WHITE_ROOK,
    }
    return &Board{
        holder: boardHolder,

        // KQkq (Uppercase -> White)
        gameState: 0b1111,

        nextMove: WHITE,
    }
}


func NewBoardFromFEN(fen string) (*Board, error) {

    return nil, nil
}

func IsFen(fen string) bool {
    splitted := strings.Split(fen, " ")
    if len(splitted) != 6 {
        return false
    }
    positions := strings.Split(splitted[0], "/")
    
    if len(positions) != 8 {
        return false
    }

    return true
}
