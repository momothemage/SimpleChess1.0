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
	HorseDirectionOffsets = []int{-11, -19, 11, 19, -17, -7, 17, 7}
	HorseLameOffsets      = []int{-1, -9, 1, 9, -9, 1, 9, -1}
	ElephantTargets       = map[int][]int{
		2:  {18, 22},
		6:  {22, 26},
		18: {2, 38},
		22: {2, 6, 38, 42},
		26: {6, 42},
		38: {18, 22},
		42: {22, 26},
		47: {63, 67},
		51: {67, 71},
		63: {47, 83},
		67: {47, 51, 83, 87},
		71: {51, 87},
		83: {63, 67},
		87: {67, 71},
	}
	AdvisorTargets = map[int][]int{
		3: {13}, 5: {13}, 21: {13}, 23: {13}, 13: {3, 5, 21, 23},
		66: {76}, 68: {76}, 84: {76}, 86: {76}, 76: {66, 68, 84, 86},
	}
	GeneralTargets = map[int][]int{
		3: {4, 12}, 4: {3, 5, 13}, 5: {4, 14},
		12: {3, 13, 21}, 13: {4, 12, 14, 22}, 14: {5, 13, 23},
		21: {12, 22}, 22: {13, 21, 23}, 23: {14, 22},
		66: {67, 75}, 67: {66, 68, 76}, 68: {67, 77},
		75: {66, 76, 84}, 76: {67, 75, 77, 85}, 77: {68, 76, 86},
		84: {75, 85}, 85: {76, 84, 86}, 86: {77, 85},
	}
	RedGeneralPos   = []int{66, 67, 68, 75, 76, 77, 84, 85, 86}
	BlackGeneralPos = []int{3, 4, 5, 12, 13, 14, 21, 22, 23}
	BoardEdgeList   = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 17, 18, 26, 27, 35, 36, 44, 45, 53, 54, 62, 63, 71, 72, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89}
	BoardEdge       = map[int]bool{}
	NumPosToEdge    = map[int][8]int{}
)

func init() {
	for _, pos := range BoardEdgeList {
		BoardEdge[pos] = true
	}
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
