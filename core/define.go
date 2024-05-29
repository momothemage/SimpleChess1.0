package core

import (
	"main/core/resource"
)

const (
	InitialFEN = "rheagaehr/9/1c5c1/s1s1s1s1s/9/9/S1S1S1S1S/1C5C1/9/RHEAGAEHR r 0 1"

	RedMove   = "r"
	BlackMove = "b"
	TotalNum  = 90
	// 棋盘宽度，棋盘高度
	ScreenWidth  = 572
	ScreenHeight = 626
	// 棋盘列数，棋盘行数
	MostInARow    = 9
	MostInAColumn = 8
	// 原点X，原点Y，棋子宽度
	OriginX    = 42
	OriginY    = 40
	PieceWidth = 54
	// 行棋方的图标X,Y
	CurrentPlayerX = 15
	CurrentPlayerY = 15
	Scale          = 0.7

	typeMask  = 0b0111
	colorMask = 0b11000
)

const (
	None     = 0
	General  = 1
	Soldier  = 2
	Horse    = 3
	Cannon   = 4
	Rook     = 5
	Elephant = 6
	Advisor  = 7

	Red   = 8
	Black = 16
)

var p2n = map[int32]int{
	'r': Black | Rook,
	'h': Black | Horse,
	'e': Black | Elephant,
	'a': Black | Advisor,
	'g': Black | General,
	'c': Black | Cannon,
	's': Black | Soldier,
	'R': Red | Rook,
	'H': Red | Horse,
	'E': Red | Elephant,
	'A': Red | Advisor,
	'G': Red | General,
	'C': Red | Cannon,
	'S': Red | Soldier,
}

var n2p = map[int]string{
	Black | Rook:     "r",
	Black | Horse:    "h",
	Black | Elephant: "e",
	Black | Advisor:  "a",
	Black | General:  "g",
	Black | Cannon:   "c",
	Black | Soldier:  "s",
	Red | Rook:       "R",
	Red | Horse:      "H",
	Red | Elephant:   "E",
	Red | Advisor:    "A",
	Red | General:    "G",
	Red | Cannon:     "C",
	Red | Soldier:    "S",
}

var p2b = map[int][]byte{
	Black | Rook:     resource.Br,
	Black | Horse:    resource.Bn,
	Black | Elephant: resource.Bb,
	Black | Advisor:  resource.Ba,
	Black | General:  resource.Bk,
	Black | Cannon:   resource.Bc,
	Black | Soldier:  resource.Bp,
	Red | Rook:       resource.Rr,
	Red | Horse:      resource.Rn,
	Red | Elephant:   resource.Rb,
	Red | Advisor:    resource.Ra,
	Red | General:    resource.Rk,
	Red | Cannon:     resource.Rc,
	Red | Soldier:    resource.Rp,
}

// 所有可能的方向
var (
	DirectionOffsets      = []int{-9, 9, -1, 1}
	HorseDirectionOffests = []int{-11, -19, 11, 19, -17, -7, 17, 7}
	HorseLameOffsets      = []int{-1, -9, 1, 9, -9, 1, 9, -1}
	NumPosToEdge          = map[int][8]int{}
)

func init() {
	for i := 0; i <= 9; i++ {
		for j := 0; j <= 8; j++ {
			index := 9*i + j
			numNorth, numSouth, numWest, numEast := i, MostInARow-i, j, MostInAColumn-j
			NumPosToEdge[index] = [8]int{numNorth, numSouth, numWest, numEast, min(numNorth, numWest), min(numSouth, numEast), min(numNorth, numEast), min(numSouth, numWest)}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
