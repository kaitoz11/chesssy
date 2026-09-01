package engine

import "fmt"

type Bitmap uint64

func NewBitmap() Bitmap {
	return 0
}

func (b *Bitmap) Set(s Square) {
	*b  |= (1 << s)
}

func (b *Bitmap) Clear(s Square) {
	*b &= ^(1 << s)
}

func (b Bitmap) IsSet(s Square) bool {
	return (b & (1 << s)) != 0
}

type Direction uint8

const (
	UP Direction = iota
	UP_RIGHT
	RIGHT
	DOWN_RIGHT
	DOWN
	DOWN_LEFT
	LEFT
	UP_LEFT
)

func (b Bitmap) Shift(direction Direction) Bitmap {
	switch direction {
	case UP:
		b = b >> 8
	case UP_RIGHT:
		b = b >> 7
	case RIGHT:
		b = b << 1
	case DOWN_RIGHT:
		b = b << 9
	case DOWN:
		b = b << 8
	case DOWN_LEFT:
		b = b << 7
	case LEFT:
		b = b << 1
	case UP_LEFT:
		b = b >> 9
	}
		
	return b
}

func (b Bitmap) Print() {
	fmt.Println("=================")
	for i := range 64 {
		if i%8 == 0 {
			fmt.Printf("%d ",8-i/8)
		}
		if b.IsSet(Square(i)) {
			fmt.Print("x ")
		} else {
			fmt.Print(". ")
		}

		if (i+1)%8 == 0 {
			fmt.Println()
		}
	}
	fmt.Println("  a b c d e f g h")
	fmt.Println("-----------------")
}
