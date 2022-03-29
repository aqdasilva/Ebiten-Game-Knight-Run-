package main

import (
	_ "bytes"
	"embed"
	"github.com/blizzy78/ebitenui"
	"github.com/blizzy78/ebitenui/event"
	"github.com/blizzy78/ebitenui/image"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"math/rand"
	"time"
)

var enemeySpawn = int(ebiten.DefaultTPS * 20 * rand.Float64())

//go:embed graphics/*
var EmbeddedAssets embed.FS

type Mode int

const (
	GameWidth      = 1000
	GameHeight     = 540
	spriteSpeed    = 2
	jumpingPower   = 20
	gravity        = 1
	maxReaperCount = 10

	gameModeTitle = 0
	gameMode      = 1
	gameOvermMode = 2

	reaper01 = 0
	reaper2  = 1
	reaper3  = 2

	knightHeight = 50
	knightWidth  = 100

	reaper1Width  = 25
	reaper1Height = 50

	reaper2Width  = 25
	reaper2Height = 50

	reaper3Width  = 25
	reaper3Height = 50
)

type Sprite struct {
	xloc         int
	yloc         int
	dX           int
	dY           int
	redKnightImg *ebiten.Image
	reaper1Img   *ebiten.Image
}

//go:embed graphics/redKnightImg.png
var redKnight []byte

//go:embed graphics/reaper1.png
var reaper1 []byte

//go:embed graphics/reaper03.png
var reaper03 []byte

//go:embed graphics/knightAxe.png
var knightAxe []byte

var (
	redKnightImg *ebiten.Image
	reaper1Img   *ebiten.Image
	reaper03Img  *ebiten.Image
	knightAxeImg *ebiten.Image
	arcadeFont   font.Face

	picSlices   *ebiten.Image
	EnemiesList map[string]*ebiten.Image
	varSprite   []string
	reaperVar   []string
)

type reapers struct {
	x       int
	y       int
	kind    int
	visible bool
}

func (r *reapers) reaperMove(speed int) {
	r.x -= speed
}

func (r *reapers) reaperOffScreen() bool {
	return r.x < -50
}

type Game struct {
	player        Sprite
	score         int
	drawOps       ebiten.DrawImageOptions
	opponent      Sprite
	gameFaneto    bool
	collideKill   int
	mode          int
	count         int
	knightX       int
	knightY       int
	reapers       [maxReaperCount]*reapers
	gameOverCount int
}

func newGame() *Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) init() {
	g.score = 0
	g.knightX = 50
	g.knightY = 200
	for i := 0; i < maxReaperCount; i++ {
		g.reapers[i] = &reapers{}
	}
}

func (g *Game) Update() error {
	switch g.mode {
	case gameModeTitle:
		if g.spaceBarKeypresssed() {
			g.mode = gameMode
		}
	case gameMode:
		g.count++
		g.score = g.count / 4

	}
	processPlayerInput(g)
	return nil
}

func (g *Game) spaceBarKeypresssed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	return false
}

func (g *Game) collide() bool {
	hitKnightMinX := g.knightX + 20
	hitKnightMaxX := g.knightX + knightWidth - 15
	hitKnightMaxY := g.knightY + knightHeight - 10

	for _, r := range g.reapers {
		hitReaperMinx := r.x + 5
		hitReaperMaxX := r.x + reaper1Width - 5
		hitReaperMaxY := r.y + 5

		if r.visible {
			if hitKnightMaxX-hitReaperMinx > 0 && hitReaperMaxX-hitKnightMinX > 0 && hitKnightMaxY-hitReaperMaxY > 0 {
				return true
			}
		}
	}
	return false
}

type ButtonOpt func(b *Button)

type Button struct {
	image     *widget.ButtonImage
	TextColor *widget.ButtonTextColor

	PressedEvent *event.Event

	init *widget.MultiOnce
}
type ButtonImage struct {
	Idle     *image.NineSlice
	Pressed  *image.NineSlice
	Disabled *image.NineSlice
}
type ButtonImageImage struct {
	gameOverButton     *ebiten.Image
	startNewGameButton *ebiten.Image
}
type ButtonPressedEventArgs struct {
	Button *Button
	setX   int
	setY   int
}

type ButtonPressedHandlerFunc func(args *ButtonPressedEventArgs)

type Box struct {
	Data interface{}
	W    float64
	H    float64
	X    float64
	Y    float64
}

func (g *Game) isKeyPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	return false
}

func (g *Game) level() int {
	return g.collideKill / 10

}
func (g *Game) addScore() {
	base := 0
	switch g.collideKill {
	case 1:
		base = 20
	case 2:
		base = 40
	case 3:
		base = 60
	case 4:
		base = 80
	case 5:
		base = 100
	default:
		panic("You need to get better")
	}
	g.score += (g.level() + 1) * base
}

func randNo(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func (g Game) Draw(screen *ebiten.Image) {
	g.drawOps.GeoM.Reset()
	g.drawOps.GeoM.Translate(float64(g.player.xloc), float64(g.player.yloc))
	screen.DrawImage(g.player.redKnightImg, &g.drawOps)
	g.drawOps.GeoM.Reset()
	opponent := g.opponent
	g.drawOps.GeoM.Translate(float64(opponent.xloc), float64(opponent.yloc))
	screen.DrawImage(opponent.reaper1Img, &g.drawOps)
}

func (g Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return GameWidth, GameHeight
}
func makeButtons() (GUIhandler *ebitenui.UI) {
	background := image.NewNineSliceColor(color.Gray16{})
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Top:    20,
				Bottom: 20,
			}),
			widget.GridLayoutOpts.Spacing(0, 20))),
		widget.ContainerOpts.BackgroundImage(background))

	textInfo := widget.TextOptions{}.Text("did this window work", basicfont.Face7x13, color.White)
	endGame, err := loadImageNineSlice("gameOverButton.png", 20, 0)
	if err != nil {
		log.Fatalln(err)
	}
	newGame, err := loadImageNineSlice("newGameButton.png", 20, 0)
	if err != nil {
		log.Fatalln(err)
	}
	buttonImage := &widget.ButtonImage{
		Idle:    endGame,
		Pressed: newGame,
	}
	EndButton := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text("End Game", basicfont.Face7x13, &widget.ButtonTextColor{
			Idle: color.RGBA{0xdf, 0xf4, 0xff, 0xff},
		}),
	)
	NewGameButton := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text("New Game", basicfont.Face7x13, &widget.ButtonTextColor{}),
	)

	rootContainer.AddChild(EndButton)
	rootContainer.AddChild(NewGameButton)
	rootContainer.AddChild(widget.NewText(textInfo))
	GUIhandler = &ebitenui.UI{Container: rootContainer}
	return GUIhandler

}
func main() {
	ebiten.SetWindowSize(GameWidth, GameHeight)
	ebiten.SetWindowTitle("Knight VS. Reaper")

	simpleGame := Game{score: 0}
	rand.Seed(time.Now().UnixNano())
	simpleGame.player = Sprite{
		redKnightImg: loadPNGImageFromEmbedded("redKnightImg.png"),
		xloc:         50,
		yloc:         200,
		dX:           0,
		dY:           0,
	}
	if enemeySpawn <= 0 {
		enemeySpawn = int(ebiten.DefaultTPS * 20 * rand.Float64())
	}

	simpleGame.opponent = Sprite{
		reaper1Img: loadPNGImageFromEmbedded("reaper03.png"),
		xloc:       50,
		yloc:       100,
		dX:         0,
		dY:         0,
	}

	//tried to call NewGame but was not sure how to implement simplegame and newgame together. tried many diff ways but nothing worked together

	if err := ebiten.RunGame(&simpleGame); err != nil {
		log.Fatal("Oh no! something terrible happened and the game crashed", err)
	}
}

///////////

//////// proocess images and player input&controls ////////

/////////

func loadPNGImageFromEmbedded(name string) *ebiten.Image {
	pictNames, err := EmbeddedAssets.ReadDir("graphics")
	if err != nil {
		log.Fatal("failed to read embedded dir ", pictNames, " ", err)
	}
	embeddedFile, err := EmbeddedAssets.Open("graphics/" + name)
	if err != nil {
		log.Fatal("failed to load embedded image ", embeddedFile, err)
	}
	rawImage, err := png.Decode(embeddedFile)
	if err != nil {
		log.Fatal("failed to load embedded image ", name, err)
	}
	gameImage := ebiten.NewImageFromImage(rawImage)
	return gameImage
}

func loadImageNineSlice(path string, centerWidth int, centerHeight int) (*image.NineSlice, error) {
	i := loadPNGImageFromEmbedded(path)

	w, h := i.Size()
	return image.NewNineSlice(i,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight}),
		nil
}

func processPlayerInput(theGame *Game) {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		theGame.player.dY = -spriteSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		theGame.player.dY = spriteSpeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyUp) || inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		theGame.player.dY = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		theGame.player.dX = -spriteSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		theGame.player.dX = spriteSpeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyArrowLeft) || inpututil.IsKeyJustReleased(ebiten.KeyArrowRight) {
		theGame.player.dX = 0
	}

	theGame.player.yloc += theGame.player.dY
	if theGame.player.yloc <= 0 {
		theGame.player.dY = 0
		theGame.player.yloc = 0
	} else if theGame.player.yloc+theGame.player.redKnightImg.Bounds().Size().Y > GameHeight {
		theGame.player.dY = 0
		theGame.player.yloc = GameHeight - theGame.player.redKnightImg.Bounds().Size().Y
	}
	theGame.player.xloc += theGame.player.dX
	if theGame.player.xloc <= 0 {
		theGame.player.dX = 0
		theGame.player.xloc = 0
	} else if theGame.player.xloc+theGame.player.redKnightImg.Bounds().Size().X > GameWidth {
		theGame.player.dX = 0
		theGame.player.xloc = GameWidth - theGame.player.redKnightImg.Bounds().Size().X
	}

	//sprites opponent
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		theGame.opponent.dY = -spriteSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		theGame.opponent.dY = spriteSpeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyW) || inpututil.IsKeyJustReleased(ebiten.KeyS) {
		theGame.opponent.dY = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		theGame.opponent.dX = -spriteSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		theGame.opponent.dX = spriteSpeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyA) || inpututil.IsKeyJustReleased(ebiten.KeyD) {
		theGame.opponent.dX = 0
	}

	theGame.opponent.yloc += theGame.opponent.dY
	if theGame.opponent.yloc <= 0 {
		theGame.opponent.dY = 0
		theGame.opponent.yloc = 0
	} else if theGame.opponent.yloc+theGame.opponent.reaper1Img.Bounds().Size().Y > GameHeight {
		theGame.opponent.dY = 0
		theGame.opponent.yloc = GameHeight - theGame.opponent.reaper1Img.Bounds().Size().Y
	}
	theGame.opponent.xloc += theGame.opponent.dX
	if theGame.opponent.xloc <= 0 {
		theGame.opponent.dX = 0
		theGame.opponent.xloc = 0
	} else if theGame.opponent.xloc+theGame.opponent.reaper1Img.Bounds().Size().X > GameWidth {
		theGame.opponent.dX = 0
		theGame.opponent.xloc = GameWidth - theGame.opponent.reaper1Img.Bounds().Size().X
	}
}
