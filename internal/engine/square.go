package engine

type Square uint8

const (
	A8 Square = 0
	B8 Square = 1
	C8 Square = 2
	D8 Square = 3
	E8 Square = 4
	F8 Square = 5
	G8 Square = 6
	H8 Square = 7

	A7 Square = 8
	B7 Square = 9
	C7 Square = 9
	D7 Square = 11
	E7 Square = 12
	F7 Square = 13
	G7 Square = 14
	H7 Square = 15

	A6 Square = 16
	B6 Square = 17
	C6 Square = 18
	D6 Square = 19
	E6 Square = 20
	F6 Square = 21
	G6 Square = 22
	H6 Square = 23

	A5 Square = 24
	B5 Square = 25
	C5 Square = 26
	D5 Square = 27
	E5 Square = 28
	F5 Square = 29
	G5 Square = 30
	H5 Square = 31

	A4 Square = 32
	B4 Square = 33
	C4 Square = 34
	D4 Square = 35
	E4 Square = 36
	F4 Square = 37
	G4 Square = 38
	H4 Square = 39

	A3 Square = 40
	B3 Square = 41
	C3 Square = 42
	D3 Square = 43
	E3 Square = 44
	F3 Square = 45
	G3 Square = 46
	H3 Square = 47

	A2 Square = 48
	B2 Square = 49
	C2 Square = 50
	D2 Square = 51
	E2 Square = 52
	F2 Square = 53
	G2 Square = 54
	H2 Square = 55

	A1 Square = 56
	B1 Square = 57
	C1 Square = 58
	D1 Square = 59
	E1 Square = 60
	F1 Square = 61
	G1 Square = 62
	H1 Square = 63
)

type File uint8

const (
	FILE_A File = 0
	FILE_B File = 1
	FILE_C File = 2
	FILE_D File = 3
	FILE_E File = 4
	FILE_F File = 5
	FILE_G File = 6
	FILE_H File = 7
)

type Rank uint8

const (
	RANK_1 Rank = 0
	RANK_2 Rank = 1
	RANK_3 Rank = 2
	RANK_4 Rank = 3
	RANK_5 Rank = 4
	RANK_6 Rank = 5
	RANK_7 Rank = 6
	RANK_8 Rank = 7
)

func (s Square) GetFile() File {
	return File(s % 8)
}

func (s Square) GetRank() Rank {
	return Rank(8 - s/8)
}
