package core

import (
	"bytes"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image/color"
	"log"
	"main/core/resource"
	"strconv"
	"strings"

	"github.com/golang/freetype/truetype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Game 此类用于管理游戏进程
type Game struct {
	selected int //选中的格子
	FEN      string
	Moves    map[Move]bool

	RedGeneral   int // todo: 帅的位置，需要被更新
	BlackGeneral int // todo: 将的位置，需要被更新

	board     [90]int               // 棋盘
	checkmate bool                  // 是否被将军
	lastMove  int                   //上一步棋
	flipped   bool                  //是否翻转棋盘
	gameOver  bool                  //是否游戏结束
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
	if g.gameOver {
		g.messageBox(screen)
	}
}

func (g *Game) Update() error {
	isRedMove := strings.Split(g.FEN, " ")[1] == RedMove
	board := LoadPositionFromFEN(g.FEN)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.gameOver {
			g.gameOver = false
			g.FEN = InitialFEN
			board = LoadPositionFromFEN(g.FEN)
			g.selected = -1
		}

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
	colorToMove := fenData[1]
	halfMoves := toInt(fenData[2])
	totalMoves := toInt(fenData[3]) + 1
	if board[goal] != 0 {
		halfMoves = 0
	}

	g.generateMoves(board, colorToMove, origin)
	// fmt.Println(g.Moves)
	for move := range g.Moves {
		if move.Target == goal {
			if g.checkmate {
				// todo: 计算所有敌方棋子的攻击范围，并判断此步是不是解除了将军/送将
				// todo: 如果不能解，则此步不合法，需要此处break
			}
			if g.isCheckMate(g.getUnderAttackPos(board, Switch(colorToMove)), Switch(colorToMove)) {
				g.checkmate = true
			}
			if g.checkmate && len(g.generateAllMoves(board, Switch(colorToMove))) == 0 {
				g.gameOver = true
			}
			board[origin], board[goal] = 0, board[origin]
			g.FEN = TransferBoardToPieces(board) + " " + switchPlayer(fenData[1]) + " " + strconv.Itoa(halfMoves) + " " + strconv.Itoa(totalMoves)
			break
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
	g.Moves = make(map[Move]bool)
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
		case Soldier:
			g.generateSoldierMoves(board, start, piece)
		}
	}
}

func (g *Game) generateRookMoves(board [90]int, start, piece int) {
	currentColor := pieceColor(piece)
	for i := 0; i < 4; i++ {
		for n := 0; n < NumPosToEdge[start][i]; n++ {
			target := start + DirectionOffsets[i]*(n+1)
			targetPiece := board[target]
			if targetPiece == 0 {
				g.Moves[Move{start, target}] = true
				continue
			}
			// 被友方棋子阻挡
			if pieceColor(targetPiece) == currentColor {
				break
			}
			g.Moves[Move{start, target}] = true
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
					g.Moves[Move{start, target}] = true
				} else {
					skipped = true
				}
			} else {
				if targetPiece != 0 && pieceColor(targetPiece) != currentColor {
					g.Moves[Move{start, target}] = true
					break
				}
			}
		}
	}
}

func (g *Game) generateHorseMoves(board [90]int, start, piece int) {
	currentColor := pieceColor(piece)
	for i := 0; i < 4; i++ {
		switch NumPosToEdge[start][i+4] {
		case 0, 1:
			// 判断蹩马腿的位置是否在棋盘边缘，如果在的话说明这个目标点一定是不可达的。
			target := start + HorseDirectionOffsets[2*i]
			if target >= 0 && target <= 89 && !BoardEdge[start+HorseLameOffsets[2*i]] &&
				board[start+HorseLameOffsets[2*i]] == None && pieceColor(board[target]) != currentColor {
				g.Moves[Move{start, target}] = true
			}
			target = start + HorseDirectionOffsets[2*i+1]
			if target >= 0 && target <= 89 && !BoardEdge[start+HorseLameOffsets[2*i+1]] &&
				board[start+HorseLameOffsets[2*i+1]] == None && pieceColor(board[target]) != currentColor {
				g.Moves[Move{start, target}] = true
			}
		default:
			target := start + HorseDirectionOffsets[2*i]
			if board[start+HorseLameOffsets[2*i]] == None && pieceColor(board[target]) != currentColor {
				g.Moves[Move{start, target}] = true
			}
			target = start + HorseDirectionOffsets[2*i+1]
			if board[start+HorseLameOffsets[2*i+1]] == None && pieceColor(board[target]) != currentColor {
				g.Moves[Move{start, target}] = true
			}
		}
	}
}

func (g *Game) generateElephantMoves(board [90]int, start int, piece int) {
	currentColor := pieceColor(piece)
	targets := ElephantTargets[start]
	for _, target := range targets {
		if board[(start+target)/2] == None && pieceColor(board[target]) != currentColor {
			g.Moves[Move{start, target}] = true
		}
	}
}

func (g *Game) generateAdvisorMoves(board [90]int, start int, piece int) {
	currentColor := pieceColor(piece)
	targets := AdvisorTargets[start]
	for _, target := range targets {
		if pieceColor(board[target]) != currentColor {
			g.Moves[Move{start, target}] = true
		}
	}
}

func (g *Game) generateGeneralMoves(board [90]int, start int, piece int) {
	currentColor := pieceColor(piece)
	targets := GeneralTargets[start]
	for _, target := range targets {
		if pieceColor(board[target]) != currentColor && !isFacedToGeneral(board, target, currentColor) {
			g.Moves[Move{start, target}] = true
		}
	}
}

func (g *Game) generateSoldierMoves(board [90]int, start int, piece int) {
	currentColor := pieceColor(piece)
	if isRedPiece(piece) {
		if pieceColor(board[start-9]) != currentColor {
			g.Moves[Move{start, start - 9}] = true
		} else if onUpperBoard(start) && pieceColor(board[start-1]) != currentColor {
			g.Moves[Move{start, start - 1}] = true
		} else if onUpperBoard(start) && pieceColor(board[start+1]) != currentColor {
			g.Moves[Move{start, start + 1}] = true
		}
	} else {
		if pieceColor(board[start+9]) != currentColor {
			g.Moves[Move{start, start + 9}] = true
		} else if onLowerBoard(start) && pieceColor(board[start-1]) != currentColor {
			g.Moves[Move{start, start - 1}] = true
		} else if onLowerBoard(start) && pieceColor(board[start+1]) != currentColor {
			g.Moves[Move{start, start + 1}] = true
		}
	}
}

// messageBox 提示
func (g *Game) messageBox(screen *ebiten.Image) {
	tt, err := truetype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		fmt.Print(err)
		return
	}
	arcadeFont := truetype.NewFace(tt, &truetype.Options{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	text.Draw(screen, "Click mouse to restart", arcadeFont, 150, 320, color.Black)
	return
}

func (g *Game) generateAllMoves(board [90]int, colorToMove string) []Move {
	// todo: 直接生成所有可用的move，而不是只生成当前选中的棋子
	return nil
}

func (g *Game) getUnderAttackPos(board [90]int, colorToMove string) map[int]bool {
	// todo: 获取所有可能遭受攻击的点位
	return nil
}

func (g *Game) isCheckMate(underAttackPos map[int]bool, color string) bool {
	if color == RedMove {
		return underAttackPos[g.RedGeneral]
	}
	return underAttackPos[g.BlackGeneral]
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
