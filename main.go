package main

import (
	"chesssy/internal"
	"fmt"
)

func main() {

	// b := internal.NewBoard()
	//
	// b.Print()
	// fmt.Println("WTF")
	// a := internal.ToPiece(3)
	// fmt.Println(a.Point())
	//
	//    var hehe rune 
	//    hehe = 'a'
	//    fmt.Println(hehe)
	//
	   fen := "r3k2r/pp1n2pp/2p2q2/b2p1n2/BP1Pp3/P1N2P2/2PB2PP/R2Q1RK1 w kq b3 0 13"
	   fmt.Println(internal.IsFen(fen))

    var castling byte 
    // KQkq
    castling = 0b1111
    K := byte(0b1000)
    // Q := byte(0b0100)
    // k := byte(0b0010)
    // q := byte(0b0001)

    canCastle := castling & K != 0

    fmt.Printf("can castle: %v\n",canCastle)

    fmt.Println(castling)
}
