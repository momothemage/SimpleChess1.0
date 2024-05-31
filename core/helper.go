package core

import (
	"strconv"
	"strings"
)

type Move struct {
	Start, Target int
}

// rheagaehr/9/1c5c1/s1s1s1s1s/9/9/S1S1S1S1S/1C5C1/9/RHEAGAEHR r 0 1
func LoadPositionFromFEN(fen string) [90]int {
	pieces := strings.Split(fen, " ")[0]
	return LoadPositionByPieces(pieces)
}

func LoadPositionByPieces(pieces string) [90]int {
	var board = [90]int{}
	index := 0
	for _, piece := range pieces {
		switch piece {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			index += int(piece) - 48
		case '/':
			continue
		default:
			board[index] = p2n[piece]
			index++
		}
	}
	return board
}

func TransferBoardToPieces(board [90]int) string {
	blanks := 0
	res := ""
	for index, piece := range board {
		if index%9 == 0 && index != 89 && index != 0 {
			if blanks != 0 {
				res += strconv.Itoa(blanks)
			}
			res += "/"
			blanks = 0
		}
		if piece == 0 {
			blanks++
		} else {
			if blanks != 0 {
				res += strconv.Itoa(blanks)
			}
			blanks = 0
			res += n2p[piece]
		}
	}
	if blanks != 0 {
		res += strconv.Itoa(blanks)
	}
	return res
}

func getImage(piece int) []byte {
	return p2b[piece]
}

func getMaskPos(i int) (x, y float64) {
	x = float64(OriginX + i%9*PieceWidth)
	y = float64(OriginY + i/9*PieceWidth)
	return
}

func switchPlayer(player string) string {
	if player == "r" {
		return "b"
	}
	return "r"
}

func isRedPiece(piece int) bool {
	return piece != 0 && piece>>4 == 0
}

func isBlackPiece(piece int) bool {
	return piece>>4 != 0
}

func pieceColor(piece int) int {
	return piece & colorMask
}

func pieceType(piece int) int {
	return piece & typeMask
}

func isSlidingPiece(piece int) bool { // 是否是远程攻击型棋子，车/炮
	return pieceType(piece) == Rook || pieceType(piece) == Cannon
}

func onUpperBoard(pos int) bool {
	return pos >= 0 && pos <= 44
}

func onLowerBoard(pos int) bool {
	return pos >= 45 && pos <= 89
}

func getEnemyGeneralPos(board [90]int, currentColor int) int {
	posMap := make([]int, 0)
	if currentColor == Red {
		posMap = BlackGeneralPos
	} else {
		posMap = RedGeneralPos
	}
	for _, pos := range posMap {
		if pieceType(board[pos]) == General {
			return pos
		}
	}
	return -1
}

func isFacedToGeneral(board [90]int, target int, currentColor int) bool {
	offset := 9
	if currentColor == Red {
		offset = -9
	}
	target += offset
	for target >= 0 && target <= 89 {
		if pieceType(board[target]) == None {
			target += offset
			continue
		} else if pieceType(board[target]) == General {
			return true
		} else {
			return false
		}
	}
	return false
}

func toInt(str string) int {
	res, _ := strconv.Atoi(str)
	return res
}
