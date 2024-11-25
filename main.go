package main

import (
	"chesssy/internal"
	"fmt"
)

func main() {

	// b := internal.NewBoard()
	// b.Print()
	//
	// err := b.UpdateCastle(internal.BLACK_k, internal.BLACK)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// err = b.UpdateCastle(internal.BLACK_k, internal.BLACK)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// err = b.UpdateCastle(internal.BLACK_k, internal.WHITE)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	//    err = b.UpdateCastle(internal.WHITE_K, internal.WHITE)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// err = b.UpdateCastle(internal.WHITE_Q, internal.WHITE)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	//    err = b.UpdateCastle(internal.WHITE_Q, internal.BLACK)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("WTF")
	// a := internal.ToPiece(3)
	// fmt.Println(a.Point())
	//
	//    var hehe rune 
	//    hehe = 'a'
	//    fmt.Println(hehe)
	//
	
    fen := "rnbqkb1r/ppp1pppp/3p1n2/8/3PP3/2N5/PPP2PPP/R1BQKBNR b KQkq d3 0 3"
    board, err := internal.NewBoardFromFEN(fen)
    if err != nil {
		fmt.Println(err)
	}
    board.Print()
}

