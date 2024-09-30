package main

import (
	"encoding/json"
	"fmt"
	"game/sprites"
	"image"
	"image/color"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var id = rand.Intn(200) + 2
var attackStart time.Time
var isBlockedLeft bool
var isBlockedRight bool
var isCollideLeft bool
var isCollideRight bool

type Actions struct {
	UserID           int     `json:"id"`
	MoveX            float64 `json:"moveX"`
	MoveY            float64 `json:"moveY"`
	IsAttack         bool    `json:"isAttack"`
	IsAttackInactive bool    `json:"isAttackInactive"`
	IsBlock          bool    `json:"isBlock"`
}

type Player struct {
	Img              *ebiten.Image
	X                float64
	Y                float64
	Dx               float64
	Dy               float64
	IsMe             bool
	hurtBox          image.Rectangle
	isAttack         bool
	isAttackInactive bool
	isBlock          bool
}

type Game struct {
	conn               *websocket.Conn
	player1            *Player
	player1SpriteSheet *sprites.SpriteSheet
	player2            *Player
	player2SpriteSheet *sprites.SpriteSheet
	colliders          []image.Rectangle
}

func (g *Game) handleMovement(player *Player) {
	player.Dx = 0.0
	player.Dy = 2.0
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		player.Dx = 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		player.Dx = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		player.Dy = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		player.Dy = 4
	}

	if player.isAttack && (!isBlockedLeft && !isBlockedRight) || player.isAttackInactive || player.isBlock {
		player.Dx = 0
	}

	if isBlockedRight {
		player.Dx = 4
	}
	if isBlockedLeft {
		player.Dx = -4
	}

	if isCollideRight {
		player.Dx = 0.5
	}
	if isCollideLeft {
		player.Dx = -0.5
	}

	player.X += player.Dx
	player.Y += player.Dy
	if player.X < 1 {
		player.X = 1
	}
	if player.X > 303 {
		player.X = 303
	}
	player.hurtBox = image.Rect(int(player.X), int(player.Y), int(player.X)+16, int(player.Y)+16)
	act := Actions{
		UserID:   id,
		MoveX:    player.X,
		MoveY:    player.Y,
		IsAttack: player.isAttack,
		IsBlock:  player.isBlock,
	}
	if err := g.conn.WriteJSON(act); err != nil {
		fmt.Println(err)
	}
}

func handleXCollisions(myPlayer *Player, otherPlayer *Player) {
	if myPlayer.hurtBox.Overlaps(otherPlayer.hurtBox) {
		if myPlayer.X > otherPlayer.X {
			isCollideRight = true
		} else {
			isCollideLeft = true
		}
		if myPlayer.isAttack && !otherPlayer.isBlock {
			fmt.Println("attack hit")
		}
		if myPlayer.isAttack && otherPlayer.isBlock {
			if myPlayer.X > otherPlayer.X {
				isBlockedRight = true
			} else {
				isBlockedLeft = true
			}
		}
		// if myPlayer.Dx > 0.0 && myPlayer.X < otherPlayer.X {
		// 	myPlayer.X = float64(otherPlayer.hurtBox.Min.X) - 16.0
		// } else if myPlayer.Dx < 0.0 && myPlayer.X > otherPlayer.X {
		// 	myPlayer.X = float64(otherPlayer.hurtBox.Max.X)
		// }
	} else {
		isBlockedRight = false
		isBlockedLeft = false
		isCollideRight = false
		isCollideLeft = false
	}

}

// func handleYCollisions(myPlayer *Player, otherPlayer *Player) {
// 	if myPlayer.hurtBox.Overlaps(image.Rect(
// 		int(otherPlayer.X),
// 		int(otherPlayer.Y),
// 		int(otherPlayer.X)+16,
// 		int(otherPlayer.Y)+16,
// 	)) {
// 		if myPlayer.Dy > 0.0 && myPlayer.Y < otherPlayer.Y {
// 			myPlayer.Y = float64(otherPlayer.hurtBox.Min.Y) - 16.0
// 		} else if myPlayer.Dy < 0.0 && myPlayer.Y > otherPlayer.X {
// 			myPlayer.Y = float64(otherPlayer.hurtBox.Max.Y)
// 		}
// 	}
// }

func (g *Game) HandleAttackAndBlock(myPlayer *Player, otherPlayer *Player) {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && !myPlayer.isBlock {
		attackStart = time.Now()
		myPlayer.isAttackInactive = true
	}
	if time.Since(attackStart) > time.Duration(83)*time.Millisecond {
		myPlayer.isAttack = true
		myPlayer.isAttackInactive = false
	}
	if time.Since(attackStart) > time.Duration(150)*time.Millisecond {
		myPlayer.isAttack = false
		myPlayer.isAttackInactive = true
	}
	if time.Since(attackStart) > time.Duration(183)*time.Millisecond {
		myPlayer.isAttack = false
		myPlayer.isAttackInactive = false
	}
	if (myPlayer.Y-otherPlayer.Y < 20 || otherPlayer.Y-myPlayer.Y < 20) && otherPlayer.isAttack {
		if myPlayer.X < otherPlayer.Y && otherPlayer.X-myPlayer.X < 18 {
			if ebiten.IsKeyPressed(ebiten.KeyLeft) {
				myPlayer.isBlock = true
			}
		}
		if myPlayer.X > otherPlayer.Y && myPlayer.X-otherPlayer.X < 18 {
			if ebiten.IsKeyPressed(ebiten.KeyRight) {
				myPlayer.isBlock = true
			}
		}
	} else {
		myPlayer.isBlock = false
	}
}

func (g *Game) Update() error {
	var myPlayer *Player
	var otherPlayer *Player
	if g.player1.IsMe {
		myPlayer = g.player1
		otherPlayer = g.player2
	} else {
		otherPlayer = g.player1
		myPlayer = g.player2
	}

	g.handleMovement(myPlayer)
	g.HandleAttackAndBlock(myPlayer, otherPlayer)
	_, a, err := g.conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("error: %v", err)
		}
	}
	var action Actions
	if err := json.Unmarshal(a, &action); err != nil {
		log.Printf("error: %v", err)
	}
	otherPlayer.X = action.MoveX
	otherPlayer.Y = action.MoveY - 2
	otherPlayer.hurtBox = image.Rect(int(otherPlayer.X), int(otherPlayer.Y), int(otherPlayer.X)+16, int(otherPlayer.Y)+16)
	otherPlayer.isAttack = action.IsAttack
	otherPlayer.isBlock = action.IsBlock

	for _, collider := range g.colliders {
		if collider.Overlaps(image.Rect(
			int(myPlayer.X),
			int(myPlayer.Y),
			int(myPlayer.X+16),
			int(myPlayer.Y+16),
		)) {
			if myPlayer.Dy > 0.0 {
				myPlayer.Y = float64(collider.Min.Y) - 16.0
			} else if myPlayer.Dy < 0.0 {
				myPlayer.Y = float64(collider.Max.Y)
			}
		}
	}
	handleXCollisions(myPlayer, otherPlayer)

	// handleYCollisions(myPlayer, otherPlayer)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{120, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}

	opts.GeoM.Translate(g.player1.X, g.player1.Y)
	screen.DrawImage(
		g.player1.Img.SubImage(
			g.player1SpriteSheet.Rect(6),
		).(*ebiten.Image),
		&opts,
	)
	opts.GeoM.Reset()

	opts2 := ebiten.DrawImageOptions{}

	opts2.GeoM.Translate(g.player2.X, g.player2.Y)
	screen.DrawImage(
		g.player2.Img.SubImage(
			g.player2SpriteSheet.Rect(0),
		).(*ebiten.Image),
		&opts2,
	)
	opts2.GeoM.Reset()

	// if g.player1.X > g.player2.X {
	// }

	// if g.player1.X < g.player2.X {
	// 	if g.player1.isAttack {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(48, 80, 64, 96),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player1.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(48, 48, 64, 64),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player1.isBlock {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(48, 16, 64, 32),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if !g.player1.isAttack && !g.player1.isBlock && !g.player1.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(48, 0, 64, 16),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// } else {
	// 	if g.player1.isAttack {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(32, 80, 48, 96),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player1.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(32, 48, 48, 64),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player1.isBlock {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(32, 16, 48, 32),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if !g.player1.isAttack && !g.player1.isBlock && !g.player1.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player1.Img.SubImage(
	// 				image.Rect(32, 0, 48, 16),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// }
	// opts.GeoM.Reset()
	// opts.GeoM.Translate(g.player2.X, g.player2.Y)

	// if g.player2.X < g.player1.X {
	// 	if g.player2.isAttack {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(48, 80, 64, 96),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player2.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(48, 48, 64, 64),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player2.isBlock {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(48, 16, 64, 32),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if !g.player2.isAttack && !g.player2.isBlock && !g.player2.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(48, 0, 64, 16),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// } else {
	// 	if g.player2.isAttack {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(32, 80, 48, 96),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player2.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(32, 48, 48, 64),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if g.player2.isBlock {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(32, 16, 48, 32),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	if !g.player2.isAttack && !g.player2.isBlock && !g.player2.isAttackInactive {
	// 		screen.DrawImage(
	// 			g.player2.Img.SubImage(
	// 				image.Rect(32, 0, 48, 16),
	// 			).(*ebiten.Image),
	// 			&opts,
	// 		)
	// 	}
	// 	opts.GeoM.Reset()
	// }
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func Connect() *websocket.Conn {
	strId := strconv.Itoa(id)
	connectStr := fmt.Sprintf("ws://localhost:8000/ws/game/" + strId)
	conn, _, err := websocket.DefaultDialer.Dial(connectStr, nil)
	if err != nil {
		log.Println("[DIAL]", err)
	}
	return conn
}

func main() {
	conn := Connect()
	defer conn.Close()
	resp, err := http.Get("http://localhost:8000/game/isPlayer1")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var isPlayer1 bool
	if err := json.Unmarshal(bodyBytes, &isPlayer1); err != nil {
		fmt.Println(err)
	}

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")

	player1Img, _, err := ebitenutil.NewImageFromFile("assets/images/Char_3.png")
	if err != nil {
		log.Fatal(err)
	}
	player2Img, _, err := ebitenutil.NewImageFromFile("assets/images/Char_3_No_Armor.png")
	if err != nil {
		log.Fatal(err)
	}

	player1SpriteSheet := sprites.NewSpriteSheet(18, 7, 64)
	player2SpriteSheet := sprites.NewSpriteSheet(18, 7, 64)

	if err := ebiten.RunGame(&Game{
		conn: conn,
		player1: &Player{
			Img:  player1Img,
			X:    50,
			Y:    200,
			IsMe: isPlayer1,
			hurtBox: image.Rect(
				50,
				200,
				50+16,
				150+15,
			),
			isAttack:         false,
			isAttackInactive: false,
			isBlock:          true,
		},
		player1SpriteSheet: player1SpriteSheet,
		player2: &Player{
			Img:  player2Img,
			X:    150,
			Y:    200,
			IsMe: !isPlayer1,
			hurtBox: image.Rect(
				150,
				200,
				150+16,
				150+16,
			),
			isAttack:         false,
			isAttackInactive: false,
			isBlock:          false,
		},
		player2SpriteSheet: player2SpriteSheet,
		colliders: []image.Rectangle{
			image.Rect(0, 0, 320, 1),
			image.Rect(-10, 220, 330, 240),
		},
	}); err != nil {
		log.Fatal(err)
	}
}
