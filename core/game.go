package core

import (
	"bytes"
	"fmt"
	"log"
	"main/core/resource"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// 此文件用于管理游戏进程
type Game struct {
	selected int //选中的格子
	FEN      string
	Moves    []Move

	board     [90]int               // 棋盘
	turn      bool                  //轮到谁走，true红方，false黑方
	lastMove  int                   //上一步棋
	flipped   bool                  //是否翻转棋盘
	gameover  bool                  //是否游戏结束
	message   string                //显示内容
	images    map[int]*ebiten.Image //图片资源
	situation *Situation            //棋局单例
}

type Situation struct {
}

// Layout 布局采用外部尺寸（例如，窗口尺寸）并返回（逻辑）屏幕尺寸，如果不使用外部尺寸，只需返回固定尺寸即可。
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 绘制棋盘
	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(resource.Boardbytes))
	if err != nil {
		log.Print(err)
		return
	}
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(img, op)
	// 绘制该行棋的阵营
	res := resource.Rk
	if strings.Split(g.FEN, " ")[1] != RedMove {
		res = resource.Bk
	}
	img, _, err = ebitenutil.NewImageFromReader(bytes.NewReader(res))
	if err != nil {
		log.Print(err)
		return
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(CurrentPlayerX, CurrentPlayerY)
	op.GeoM.Scale(Scale, Scale)
	screen.DrawImage(img, op)
	// 绘制选中的棋子
	if g.selected != -1 {
		img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(resource.Mask))
		if err != nil {
			log.Print(err)
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(getMaskPos(g.selected))
		screen.DrawImage(img, op)
	}
	// 绘制其他棋子
	g.drawByFEN(screen)
}

func (g *Game) Update() error {
	isRedMove := strings.Split(g.FEN, " ")[1] == RedMove
	board := LoadPositionFromFEN(g.FEN)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x < OriginX || y < OriginY {
			g.selected = -1
			return nil
		}
		x = (x - OriginX) / PieceWidth
		y = (y - OriginY) / PieceWidth
		if x > MostInAColumn || y > MostInARow {
			g.selected = -1
			return nil
		}
		selected := MostInARow*y + x
		if g.selected == -1 { // 未选择
			if isRedPiece(board[selected]) && isRedMove || isBlackPiece(board[selected]) && !isRedMove {
				g.selected = selected
				return nil
			}
		} else if isRedPiece(board[g.selected]) && isRedPiece(board[selected]) || isBlackPiece(board[g.selected]) && isBlackPiece(board[selected]) { // 已选红棋
			g.selected = selected
			return nil
		} else {
			g.move(g.selected, selected)
			g.selected = -1
			return nil
		}
	}
	return nil
}

func (g *Game) move(origin, goal int) {
	fenData := strings.Split(g.FEN, " ")
	board := LoadPositionByPieces(fenData[0])
	halfMoves := toInt(fenData[2])
	totalMoves := toInt(fenData[3]) + 1
	if board[goal] != 0 {
		halfMoves = 0
	}

	g.generateMoves(board, fenData[1], origin) // todo: 直接生成所有可用的move，而不是只生成当前选中的棋子
	fmt.Println(g.Moves)
	for _, move := range g.Moves {
		if move.Target == goal {
			board[origin], board[goal] = 0, board[origin]
			g.FEN = TransferBoardToPieces(board) + " " + switchPlayer(fenData[1]) + " " + strconv.Itoa(halfMoves) + " " + strconv.Itoa(totalMoves)
		}
	}
}

func (g *Game) drawByFEN(screen *ebiten.Image) {
	board := LoadPositionFromFEN(g.FEN)
	for i := 0; i <= MostInARow; i++ {
		for j := 0; j <= MostInAColumn; j++ {
			piece := board[MostInARow*i+j]
			if piece == 0 {
				continue
			}
			img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(getImage(piece)))
			if err != nil {
				log.Print(err)
				return
			}
			ebitenImage := ebiten.NewImageFromImage(img)
			g.drawChess(OriginX+PieceWidth*j, OriginY+PieceWidth*i, screen, ebitenImage)
		}
	}
}

// drawChess 绘制棋子
func (g *Game) drawChess(x, y int, screen, img *ebiten.Image) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

func (g *Game) generateMoves(board [90]int, colorToMove string, start int) {
	g.Moves = make([]Move, 0)
	//for i := 0; i < TotalNum; i++ {
	piece := board[start]
	if isRedPiece(piece) && colorToMove == RedMove || isBlackPiece(piece) && colorToMove == BlackMove {
		switch pieceType(piece) {
		case Rook:
			g.generateRookMoves(board, start, piece)
		case Cannon:
			g.generateCannonMoves(board, start, piece)
		case Horse:
			g.generateHorseMoves(board, start, piece)
		case Elephant:
			g.generateElephantMoves(board, start, piece)
		case Advisor:
			g.generateAdvisorMoves(board, start, piece)
		case General:
			g.generateGeneralMoves(board, start, piece)
		}
		// todo:
	}
	//}
}

func (g *Game) generateRookMoves(board [90]int, start, piece int) {
	currentColor := pieceColor(piece)
	for i := 0; i < 4; i++ {
		for n := 0; n < NumPosToEdge[start][i]; n++ {
			target := start + DirectionOffsets[i]*(n+1)
			targetPiece := board[target]
			if targetPiece == 0 {
				g.Moves = append(g.Moves, Move{start, target})
				continue
			}
			// 被友方棋子阻挡
			if pieceColor(targetPiece) == currentColor {
				break
			}
			g.Moves = append(g.Moves, Move{start, target})
			// 被敌方棋子阻挡
			if pieceColor(targetPiece) != currentColor {
				break
			}
		}
	}
}

func (g *Game) generateCannonMoves(board [90]int, start, piece int) {
	currentColor := pieceColor(piece)
	for i := 0; i < 4; i++ {
		skipped := false
		for n := 0; n < NumPosToEdge[start][i]; n++ {
			target := start + DirectionOffsets[i]*(n+1)
			targetPiece := board[target]
			if !skipped {
				if targetPiece == 0 {
					g.Moves = append(g.Moves, Move{start, target})
				} else {
					skipped = true
				}
			} else {
				if targetPiece != 0 && pieceColor(targetPiece) != currentColor {
					g.Moves = append(g.Moves, Move{start, target})
					break
				}
			}
		}
	}
}

func (g *Game) generateHorseMoves(board [90]int, start, piece int) {
	currentColor := pieceColor(piece)
	targets := make(map[int]bool)
	for _, direction := range HorseDirectionOffsets {
		targets[start+direction] = true
	}
	for i := 0; i < 4; i++ {
		switch NumPosToEdge[start][i+4] {
		case 0, 1:
			// 判断蹩马腿的位置是否在棋盘边缘，如果在的话说明这个目标点一定是不可达的。
			target := start + HorseDirectionOffsets[2*i]
			if target >= 0 && target <= 89 && !BoardEdge[start+HorseLameOffsets[2*i]] &&
				board[start+HorseLameOffsets[2*i]] == None && pieceColor(board[target]) != currentColor {
				g.Moves = append(g.Moves, Move{start, target})
			}
			target = start + HorseDirectionOffsets[2*i+1]
			if target >= 0 && target <= 89 && !BoardEdge[start+HorseLameOffsets[2*i+1]] &&
				board[start+HorseLameOffsets[2*i+1]] == None && pieceColor(board[target]) != currentColor {
				g.Moves = append(g.Moves, Move{start, target})
			}
		default:
			target := start + HorseDirectionOffsets[2*i]
			if board[start+HorseLameOffsets[2*i]] == None && pieceColor(board[target]) != currentColor {
				g.Moves = append(g.Moves, Move{start, target})
			}
			target = start + HorseDirectionOffsets[2*i+1]
			if board[start+HorseLameOffsets[2*i+1]] == None && pieceColor(board[target]) != currentColor {
				g.Moves = append(g.Moves, Move{start, target})
			}
		}
	}
}

func (g *Game) generateElephantMoves(board [90]int, start int, piece int) {
	currentColor := pieceColor(piece)
	targets := ElephantTarget[start]
	for _, target := range targets {
		if board[(start+target)/2] == None && pieceColor(board[target]) != currentColor {
			g.Moves = append(g.Moves, Move{start, target})
		}
	}
}

func (g *Game) generateAdvisorMoves(board [90]int, start int, piece int) {
	currentColor := pieceColor(piece)
	targets := AdvisorTarget[start]
	for _, target := range targets {
		if pieceColor(board[target]) != currentColor {
			g.Moves = append(g.Moves, Move{start, target})
		}
	}
}

func (g *Game) generateGeneralMoves(board [90]int, start int, piece int) {
	currentColor := pieceColor(piece)
	targets := GeneralTarget[start]
	for _, target := range targets {
		if pieceColor(board[target]) != currentColor {
			g.Moves = append(g.Moves, Move{start, target})
		}
	}
}

// NewGame 创建象棋程序
func NewGame() {
	game := &Game{
		selected: -1,
		FEN:      InitialFEN,
	}
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("chess")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
