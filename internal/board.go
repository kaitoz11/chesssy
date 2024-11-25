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
    enPassantSquare int
    moveNumber int
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
    fmt.Println(b.enPassantSquare)
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

        enPassantSquare: -1,

        nextMove: WHITE,
    }
}


func NewBoardFromFEN(fen string) (*Board, error) {
    splitted := strings.Split(fen, " ")
    if len(splitted) != 6 {
        return nil, errors.New("invalid FEN format")
    }
    positions := strings.Split(splitted[0], "/")
    
    if len(positions) != 8 {
        return nil, errors.New("invalid FEN format")
    }

    var boardHolder [64]Piece

    bIndex := 0
    for i, row := range positions {
        fmt.Printf("%d - row: %v\n", i, row)
        for _, slot := range row {
            fmt.Printf("%d _ %c\n",bIndex, slot)
            s := int(slot)
            if s >= '1' && s <= '8' {
                bIndex = bIndex + (s - '1') + 1
                fmt.Println("Updated")
            } else {
                boardHolder[bIndex] = FromCharToPiece(slot)
                bIndex ++
            }
        }
    }

    var nextPlayer Player
    if splitted[1] == "w" {
        nextPlayer = WHITE
    } else if splitted[1] == "b" {
        nextPlayer = BLACK
    } else {
        return nil, errors.New("invalid FEN format")
    }


    var castlingState CastlingRights

    castlingState = 0b0000

    if splitted[2] != "-" {
        for k, v := range map[rune]CastlingRights {
            'K': WHITE_K,
            'Q': WHITE_Q,
            'k': BLACK_k,
            'q': BLACK_q,
        } {
            if strings.ContainsRune(splitted[2], k) {
                castlingState = castlingState | v
            }
        }
    }


    enPassantSquare := -1
    if splitted[3] != "-" {
        pos, err := PositionToBoard(splitted[3])
        if err != nil {
            return nil, err
        }
        enPassantSquare = pos
    }

    return &Board{
        holder: boardHolder,
        nextMove: nextPlayer,
        gameState: castlingState,
        enPassantSquare: enPassantSquare,
    }, nil
}

func BoardToPosition(bPos int) (string, error) {
    if bPos < 0 || bPos >= 64 {
        return "", errors.New("invalid board position") 
    }
    rank := 8 - (( bPos ) / 8)
    file := 'a' + ((bPos) % 8)
    return fmt.Sprintf("%c%v",file, rank), nil
}

func PositionToBoard(pos string) (int, error) {
    if len(pos) != 2 {
        return -1, errors.New("invalid position")
    }

    fpos := int(pos[0])
    rpos := int(pos[1])

    if fpos < 'a' || fpos > 'h' {
        return -1, errors.New("invalid position")
    }

    if rpos < '1' || rpos > '8' {
        return -1, errors.New("invalid position")
    }

    file := fpos - 'a' + 1
    rank := '8' - rpos  + 1

    return file - 1 + (rank - 1) * 8, nil
}
