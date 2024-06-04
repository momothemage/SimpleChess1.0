package core

import (
	"bytes"
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

	RedGeneral   int
	BlackGeneral int

	board         [90]int // 棋盘
	colorToMove   string  // 当前行棋的颜色
	colorOpponent string  // 当前不行棋的颜色

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
	g.board = LoadPositionByPieces(fenData[0])
	g.colorToMove = fenData[1]
	g.colorOpponent = Switch(g.colorToMove)
	halfMoves := toInt(fenData[2])
	totalMoves := toInt(fenData[3]) + 1
	if g.board[goal] != 0 {
		halfMoves = 0
	}

	g.generateMoves(origin)
	for move := range g.Moves {
		if move.Target == goal {
			g.board[origin], g.board[goal] = 0, g.board[origin]
			generalPos := g.RedGeneral
			if g.colorToMove == BlackMove {
				generalPos = g.BlackGeneral
			}
			if isFacedToGeneral(g.board, generalPos, pieceColor(g.board[goal])) {
				break
			}

			opponentAllMoves := g.generateAllMoves(g.colorOpponent)
			if g.isCheckMate(getUnderAttackPos(opponentAllMoves), g.colorToMove) {
				// 送将
				break
			}

			if g.isCheckMate(getUnderAttackPos(g.generateAllMoves(g.colorToMove)), g.colorOpponent) {
				g.checkmate = true
			}

			if g.checkmate {
				for opponentMove := range opponentAllMoves {
					startPiece := g.board[opponentMove.Start]
					targetPiece := g.board[opponentMove.Target]
					g.board[opponentMove.Start], g.board[opponentMove.Target] = 0, g.board[opponentMove.Start]

					tmpGeneralPos := -1
					if pieceType(startPiece) == General {
						if g.colorOpponent == RedMove {
							tmpGeneralPos = g.RedGeneral
							g.RedGeneral = opponentMove.Target
						} else {
							tmpGeneralPos = g.BlackGeneral
							g.BlackGeneral = opponentMove.Target
						}
					}

					if g.isCheckMate(getUnderAttackPos(g.generateAllMoves(g.colorToMove)), g.colorOpponent) {
						delete(opponentAllMoves, opponentMove)
					}

					// 恢复
					g.board[opponentMove.Start], g.board[opponentMove.Target] = startPiece, targetPiece
					if pieceType(startPiece) == General {
						if g.colorOpponent == RedMove {
							g.RedGeneral = tmpGeneralPos
						} else {
							g.BlackGeneral = tmpGeneralPos
						}
					}
				}

				if len(opponentAllMoves) == 0 {
					g.gameOver = true
				}
			}

			if pieceType(g.board[goal]) == General {
				if g.colorToMove == RedMove {
					g.RedGeneral = goal
				} else {
					g.BlackGeneral = goal
				}
			}
			g.FEN = TransferBoardToPieces(g.board) + " " + switchPlayer(fenData[1]) + " " + strconv.Itoa(halfMoves) + " " + strconv.Itoa(totalMoves)
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

func (g *Game) generateMoves(start int) {
	g.Moves = make(map[Move]bool)
	piece := g.board[start]
	if isRedPiece(piece) && g.colorToMove == RedMove || isBlackPiece(piece) && g.colorToMove == BlackMove {
		switch pieceType(piece) {
		case Rook:
			g.Moves = g.generateRookMoves(start, piece)
		case Cannon:
			g.Moves = g.generateCannonMoves(start, piece)
		case Horse:
			g.Moves = g.generateHorseMoves(start, piece)
		case Elephant:
			g.Moves = g.generateElephantMoves(start, piece)
		case Advisor:
			g.Moves = g.generateAdvisorMoves(start, piece)
		case General:
			g.Moves = g.generateGeneralMoves(start, piece)
		case Soldier:
			g.Moves = g.generateSoldierMoves(start, piece)
		}
	}
}

func (g *Game) generateRookMoves(start, piece int) map[Move]bool {
	moves := make(map[Move]bool)
	currentColor := pieceColor(piece)
	for i := 0; i < 4; i++ {
		for n := 0; n < NumPosToEdge[start][i]; n++ {
			target := start + DirectionOffsets[i]*(n+1)
			targetPiece := g.board[target]
			if targetPiece == 0 {
				moves[Move{start, target}] = true
				continue
			}
			// 被友方棋子阻挡
			if pieceColor(targetPiece) == currentColor {
				break
			}
			moves[Move{start, target}] = true
			// 被敌方棋子阻挡
			if pieceColor(targetPiece) != currentColor {
				break
			}
		}
	}
	return moves
}

func (g *Game) generateCannonMoves(start, piece int) map[Move]bool {
	moves := make(map[Move]bool)
	currentColor := pieceColor(piece)
	for i := 0; i < 4; i++ {
		skipped := false
		for n := 0; n < NumPosToEdge[start][i]; n++ {
			target := start + DirectionOffsets[i]*(n+1)
			targetPiece := g.board[target]
			if !skipped {
				if targetPiece == 0 {
					moves[Move{start, target}] = true
				} else {
					skipped = true
				}
			} else {
				if targetPiece != 0 && pieceColor(targetPiece) != currentColor {
					moves[Move{start, target}] = true
					break
				}
			}
		}
	}
	return moves
}

func (g *Game) generateHorseMoves(start, piece int) map[Move]bool {
	moves := make(map[Move]bool)
	currentColor := pieceColor(piece)
	for i := 0; i < 4; i++ {
		switch NumPosToEdge[start][i+4] {
		case 0, 1:
			// 判断蹩马腿的位置是否在棋盘边缘，如果在的话说明这个目标点一定是不可达的。
			target := start + HorseDirectionOffsets[2*i]
			if target >= 0 && target <= 89 && !BoardEdge[start+HorseLameOffsets[2*i]] &&
				g.board[start+HorseLameOffsets[2*i]] == None && pieceColor(g.board[target]) != currentColor {
				moves[Move{start, target}] = true
			}
			target = start + HorseDirectionOffsets[2*i+1]
			if target >= 0 && target <= 89 && !BoardEdge[start+HorseLameOffsets[2*i+1]] &&
				g.board[start+HorseLameOffsets[2*i+1]] == None && pieceColor(g.board[target]) != currentColor {
				moves[Move{start, target}] = true
			}
		default:
			target := start + HorseDirectionOffsets[2*i]
			if g.board[start+HorseLameOffsets[2*i]] == None && pieceColor(g.board[target]) != currentColor {
				moves[Move{start, target}] = true
			}
			target = start + HorseDirectionOffsets[2*i+1]
			if g.board[start+HorseLameOffsets[2*i+1]] == None && pieceColor(g.board[target]) != currentColor {
				moves[Move{start, target}] = true
			}
		}
	}
	return moves
}

func (g *Game) generateElephantMoves(start int, piece int) map[Move]bool {
	moves := make(map[Move]bool)
	currentColor := pieceColor(piece)
	targets := ElephantTargets[start]
	for _, target := range targets {
		if g.board[(start+target)/2] == None && pieceColor(g.board[target]) != currentColor {
			moves[Move{start, target}] = true
		}
	}
	return moves
}

func (g *Game) generateAdvisorMoves(start int, piece int) map[Move]bool {
	moves := make(map[Move]bool)
	currentColor := pieceColor(piece)
	targets := AdvisorTargets[start]
	for _, target := range targets {
		if pieceColor(g.board[target]) != currentColor {
			moves[Move{start, target}] = true
		}
	}
	return moves
}

func (g *Game) generateGeneralMoves(start int, piece int) map[Move]bool {
	moves := make(map[Move]bool)
	currentColor := pieceColor(piece)
	targets := GeneralTargets[start]
	for _, target := range targets {
		if pieceColor(g.board[target]) != currentColor && !isFacedToGeneral(g.board, target, currentColor) {
			moves[Move{start, target}] = true
		}
	}
	return moves
}

func (g *Game) generateSoldierMoves(start int, piece int) map[Move]bool {
	moves := make(map[Move]bool)
	currentColor := pieceColor(piece)
	if isRedPiece(piece) {
		if pieceColor(g.board[start-9]) != currentColor {
			moves[Move{start, start - 9}] = true
		}
		if onUpperBoard(start) && pieceColor(g.board[start-1]) != currentColor {
			moves[Move{start, start - 1}] = true
		}
		if onUpperBoard(start) && pieceColor(g.board[start+1]) != currentColor {
			moves[Move{start, start + 1}] = true
		}
	} else {
		if pieceColor(g.board[start+9]) != currentColor {
			moves[Move{start, start + 9}] = true
		}
		if onLowerBoard(start) && pieceColor(g.board[start-1]) != currentColor {
			moves[Move{start, start - 1}] = true
		}
		if onLowerBoard(start) && pieceColor(g.board[start+1]) != currentColor {
			moves[Move{start, start + 1}] = true
		}
	}
	return moves
}

// messageBox 提示
func (g *Game) messageBox(screen *ebiten.Image) {
	tt, err := truetype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
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

func (g *Game) generateAllMoves(color string) map[Move]bool {
	allMoves := make(map[Move]bool)
	for pos, piece := range g.board {
		moves := make(map[Move]bool)
		if isRedPiece(piece) && color == RedMove || isBlackPiece(piece) && color == BlackMove {
			switch pieceType(piece) {
			case Rook:
				moves = g.generateRookMoves(pos, piece)
			case Cannon:
				moves = g.generateCannonMoves(pos, piece)
			case Horse:
				moves = g.generateHorseMoves(pos, piece)
			case Elephant:
				moves = g.generateElephantMoves(pos, piece)
			case Advisor:
				moves = g.generateAdvisorMoves(pos, piece)
			case General:
				moves = g.generateGeneralMoves(pos, piece)
			case Soldier:
				moves = g.generateSoldierMoves(pos, piece)
			}
		}
		for move := range moves {
			allMoves[move] = true
		}
	}
	return allMoves
}

// 获取所有可能遭受攻击的点位
func getUnderAttackPos(allMoves map[Move]bool) map[int]bool {
	underAttackPos := make(map[int]bool)
	for move := range allMoves {
		underAttackPos[move.Target] = true
	}
	return underAttackPos
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
		RedGeneral:   85,
		BlackGeneral: 4,
		selected:     -1,
		FEN:          InitialFEN,
	}
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("chess")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
