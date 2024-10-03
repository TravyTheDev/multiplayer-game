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
var isDash bool
var dashTimeStart time.Time

type Actions struct {
	UserID           int     `json:"id"`
	MoveX            float64 `json:"moveX"`
	MoveY            float64 `json:"moveY"`
	IsAttack         bool    `json:"isAttack"`
	IsAttackInactive bool    `json:"isAttackInactive"`
	IsBlock          bool    `json:"isBlock"`
	IsHit            bool    `json:"isHit"`
}

type Player struct {
	Img              *ebiten.Image
	X                float64
	Y                float64
	Dx               float64
	Dy               float64
	IsMe             bool
	hurtBox          image.Rectangle
	hitBox           image.Rectangle
	isAttack         bool
	isAttackInactive bool
	isBlock          bool
	isHit            bool
}

type Game struct {
	conn               *websocket.Conn
	player1            *Player
	player1SpriteSheet *sprites.SpriteSheet
	player2            *Player
	player2SpriteSheet *sprites.SpriteSheet
	colliders          []image.Rectangle
}

func handleDash(num float64) float64 {
	if isDash {
		num *= 3
	}
	return num
}

func (g *Game) handleMovement(player *Player) {
	isDash = false
	player.Dx = 0.0
	player.Dy = 2.0
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		dashTimeStart = time.Now()
		isDash = true
		if time.Since(dashTimeStart) > time.Duration(300)*time.Millisecond {
			isDash = false
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		player.Dx = 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		player.Dx = -4
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		player.Dy = -4
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		player.Dy = 6
	}

	if player.isAttack && (!isBlockedLeft && !isBlockedRight && !isDash) || player.isAttackInactive || player.isBlock {
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

	player.X += handleDash(player.Dx)
	player.Y += handleDash(player.Dy)
	if player.X < -30 {
		player.X = -30
	}
	if player.X > 470 {
		player.X = 470
	}

	act := Actions{
		UserID:           id,
		MoveX:            player.X,
		MoveY:            player.Y,
		IsAttack:         player.isAttack,
		IsAttackInactive: player.isAttackInactive,
		IsBlock:          player.isBlock,
		IsHit:            player.isHit,
	}
	if err := g.conn.WriteJSON(act); err != nil {
		fmt.Println(err)
	}
}

func handleXCollisions(myPlayer *Player, otherPlayer *Player) {
	if myPlayer.hurtBox.Overlaps(otherPlayer.hurtBox) {
		if myPlayer.X > otherPlayer.X {
			isCollideRight = true
		} else if myPlayer.X < otherPlayer.X {
			isCollideLeft = true
		}
	} else {
		isCollideRight = false
		isCollideLeft = false
	}

	if myPlayer.hitBox.Overlaps(otherPlayer.hurtBox) {
		if myPlayer.isAttack && otherPlayer.isBlock {
			if myPlayer.X > otherPlayer.X {
				isBlockedRight = true
			} else {
				isBlockedLeft = true
			}
		}
	} else {
		isBlockedRight = false
		isBlockedLeft = false
	}

	if otherPlayer.hitBox.Overlaps(myPlayer.hurtBox) {
		if otherPlayer.isAttack && !myPlayer.isBlock {
			myPlayer.isHit = true
		}
	} else {
		myPlayer.isHit = false
	}
	if myPlayer.hitBox.Overlaps(otherPlayer.hurtBox) {
		if myPlayer.isAttack && !otherPlayer.isBlock {
			otherPlayer.isHit = true
		}
	} else {
		otherPlayer.isHit = false
	}
}

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
	otherPlayer.isAttack = action.IsAttack
	otherPlayer.isAttackInactive = action.IsAttackInactive
	otherPlayer.isBlock = action.IsBlock
	otherPlayer.isHit = action.IsHit

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

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	var player1Index int
	var player2Index int
	screen.Fill(color.RGBA{120, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}
	opts2 := ebiten.DrawImageOptions{}

	if g.player1.X > g.player2.X {
		g.player1.hurtBox = image.Rect(int(g.player1.X)+93, int(g.player1.Y)+57, int(g.player1.X)+93+16, int(g.player1.Y)+57+16)
		g.player2.hurtBox = image.Rect(int(g.player2.X)+90, int(g.player2.Y)+57, int(g.player2.X)+90+16, int(g.player2.Y)+57+16)

		g.player1.hitBox = image.Rect(int(g.player1.X)+43, int(g.player1.Y)+77, int(g.player1.X)+43+16, int(g.player1.Y)+77+16)
		g.player2.hitBox = image.Rect(int(g.player2.X)+140, int(g.player2.Y)+77, int(g.player2.X)+140+16, int(g.player2.Y)+77+16)

		if g.player1.isAttackInactive {
			player1Index = 6
		}
		if g.player1.isAttack {
			player1Index = 5
		}
		if g.player1.isHit {
			player1Index = 4
		}
		if !g.player1.isAttack && !g.player1.isAttackInactive && !g.player1.isHit {
			player1Index = 7
		}

		if g.player2.isAttackInactive {
			player2Index = 1
		}
		if g.player2.isAttack {
			player2Index = 2
		}
		if g.player2.isHit {
			player2Index = 3
		}
		if !g.player2.isAttack && !g.player2.isAttackInactive && !g.player2.isHit {
			player2Index = 0
		}
	} else {
		g.player1.hurtBox = image.Rect(int(g.player1.X)+90, int(g.player1.Y)+57, int(g.player1.X)+90+16, int(g.player1.Y)+57+16)
		g.player2.hurtBox = image.Rect(int(g.player2.X)+93, int(g.player2.Y)+57, int(g.player2.X)+93+16, int(g.player2.Y)+57+16)

		g.player1.hitBox = image.Rect(int(g.player1.X)+140, int(g.player1.Y)+77, int(g.player1.X)+140+16, int(g.player1.Y)+77+16)
		g.player2.hitBox = image.Rect(int(g.player2.X)+43, int(g.player2.Y)+77, int(g.player2.X)+43+16, int(g.player2.Y)+77+16)

		if g.player1.isAttackInactive {
			player1Index = 1
		}
		if g.player1.isAttack {
			player1Index = 2
		}
		if g.player1.isHit {
			player1Index = 3
		}
		if !g.player1.isAttack && !g.player1.isAttackInactive && !g.player1.isHit {
			player1Index = 0
		}

		if g.player2.isAttackInactive {
			player2Index = 6
		}
		if g.player2.isAttack {
			player2Index = 5
		}
		if g.player2.isHit {
			player2Index = 4
		}
		if !g.player2.isAttack && !g.player2.isAttackInactive && !g.player2.isHit {
			player2Index = 7
		}
	}

	opts.GeoM.Translate(g.player1.X, g.player1.Y)
	screen.DrawImage(
		g.player1.Img.SubImage(
			g.player1SpriteSheet.Rect(player1Index),
		).(*ebiten.Image),
		&opts,
	)
	opts.GeoM.Reset()

	opts2.GeoM.Translate(g.player2.X, g.player2.Y)
	screen.DrawImage(
		g.player2.Img.SubImage(
			g.player2SpriteSheet.Rect(player2Index),
		).(*ebiten.Image),
		&opts2,
	)
	opts2.GeoM.Reset()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
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

	player1Img, _, err := ebitenutil.NewImageFromFile("assets/images/spritesheet2.png")
	if err != nil {
		log.Fatal(err)
	}
	player2Img, _, err := ebitenutil.NewImageFromFile("assets/images/spritesheet2.png")
	if err != nil {
		log.Fatal(err)
	}

	player1SpriteSheet := sprites.NewSpriteSheet(4, 2, 200, 140)
	player2SpriteSheet := sprites.NewSpriteSheet(4, 2, 200, 140)

	if err := ebiten.RunGame(&Game{
		conn: conn,
		player1: &Player{
			Img:  player1Img,
			X:    50,
			Y:    346,
			IsMe: isPlayer1,
			hurtBox: image.Rect(
				140,
				401,
				140+16,
				401+16,
			),
			hitBox: image.Rect(
				190,
				421,
				190+16,
				421+16,
			),
			isAttack:         false,
			isAttackInactive: false,
			isBlock:          false,
			isHit:            false,
		},
		player1SpriteSheet: player1SpriteSheet,
		player2: &Player{
			Img:  player2Img,
			X:    150,
			Y:    346,
			IsMe: !isPlayer1,
			hurtBox: image.Rect(
				243,
				401,
				243+16,
				401+16,
			),
			hitBox: image.Rect(
				193,
				421,
				193+16,
				421+16,
			),
			isAttack:         false,
			isAttackInactive: false,
			isBlock:          false,
			isHit:            false,
		},
		player2SpriteSheet: player2SpriteSheet,
		colliders: []image.Rectangle{
			image.Rect(-80, -1, 660, -5),
			image.Rect(-80, 440, 660, 360),
		},
	}); err != nil {
		log.Fatal(err)
	}
}
