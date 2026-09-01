package parser

import (
	"chesssy/internal/engine"
	"chesssy/internal/piece"
	"errors"
	"fmt"
	"strings"
)


func NewBoardFromFEN(fen string) (*engine.Board, error) {
	splitted := strings.Split(fen, " ")
	if len(splitted) != 6 {
		return nil, errors.New("invalid FEN format")
	}
	positions := strings.Split(splitted[0], "/")

	if len(positions) != 8 {
		return nil, errors.New("invalid FEN format")
	}

	var boardHolder [64]piece.PieceHolder

	bIndex := 0
	for _, row := range positions {
		// fmt.Printf("%d - row: %v\n", i, row)
		for _, slot := range row {
			// fmt.Printf("%d _ %c\n",bIndex, slot)
			s := int(slot)
			if s >= '1' && s <= '8' {
				bIndex = bIndex + (s - '1') + 1
				// fmt.Println("Updated")
			} else {
				boardHolder[bIndex] = piece.FromCharToPiece(slot)
				bIndex++
			}
		}
	}

	var nextPlayer engine.Player
	if splitted[1] == "w" {
		nextPlayer = engine.WHITE
	} else if splitted[1] == "b" {
		nextPlayer = engine.BLACK
	} else {
		return nil, errors.New("invalid FEN format")
	}

	var castlingState engine.CastlingRights

	castlingState = 0b0000

	if splitted[2] != "-" {
		for k, v := range map[rune]engine.CastlingRights{
			'K': engine.WHITE_K,
			'Q': engine.WHITE_Q,
			'k': engine.BLACK_k,
			'q': engine.BLACK_q,
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

	return engine.NewBoardWith(
		boardHolder,
		nextPlayer,
		castlingState,
		enPassantSquare,
		0,
	), nil
}

func BoardToPosition(bPos int) (string, error) {
	if bPos < 0 || bPos >= 64 {
		return "", errors.New("invalid board position")
	}
	rank := 8 - ((bPos) / 8)
	file := 'a' + ((bPos) % 8)
	return fmt.Sprintf("%c%v", file, rank), nil
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
	rank := '8' - rpos + 1

	return file - 1 + (rank-1)*8, nil
}
