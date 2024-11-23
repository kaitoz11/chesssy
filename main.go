package main

import (
	"chesssy/internal"
	"fmt"
)

func main() {

	b := internal.NewBoard()
	b.Print()

	err := b.UpdateCastle(internal.BLACK_k, internal.BLACK)
	if err != nil {
		fmt.Println(err)
	}

	err = b.UpdateCastle(internal.BLACK_k, internal.BLACK)
	if err != nil {
		fmt.Println(err)
	}

	err = b.UpdateCastle(internal.BLACK_k, internal.WHITE)
	if err != nil {
		fmt.Println(err)
	}
    
    err = b.UpdateCastle(internal.WHITE_K, internal.WHITE)
	if err != nil {
		fmt.Println(err)
	}

	err = b.UpdateCastle(internal.WHITE_Q, internal.WHITE)
	if err != nil {
		fmt.Println(err)
	}

    err = b.UpdateCastle(internal.WHITE_Q, internal.BLACK)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println("WTF")
	// a := internal.ToPiece(3)
	// fmt.Println(a.Point())
	//
	//    var hehe rune 
	//    hehe = 'a'
	//    fmt.Println(hehe)
	//
	
    // fen := "r3k2r/pp1n2pp/2p2q2/b2p1n2/BP1Pp3/P1N2P2/2PB2PP/R2Q1RK1 w kq b3 0 13"
	// fmt.Prentln(internal.IsFen(fen))
}
