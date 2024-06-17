// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
  "errors"
  "image"
  // "image/color"
  _ "image/png"
  _ "image/jpeg"
  "os"
  "log"
  "time"

  "bytes"
  "io/ioutil"

  "github.com/hajimehoshi/ebiten/v2"
  "github.com/hajimehoshi/ebiten/v2/ebitenutil"
  "github.com/hajimehoshi/ebiten/v2/audio"
  "github.com/hajimehoshi/ebiten/v2/audio/wav"
  "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
  sampleRate = 44100 // Common sample rate for audio
  frameWidth  = 128
  frameHeight = 128
  frameCount  = 4 // Total number of frames in the animation
)

var (
  audioContext *audio.Context
  soundPlayers map[string]*audio.Player
  playerSprite *ebiten.Image
  bgImage *ebiten.Image
  spriteSheet *ebiten.Image
  currentFrame int
  lastFrameChange time.Time
  frameDuration = 200 * time.Millisecond
  directionRows = map[string]int{
    "up": 0,
    "down": 131,
    "left": 261,
    "right": 390,
  } 
)

// Example placeholder for the getFrameForDirection function
// This function should return the appropriate sprite frame based on the direction
func getFrameForDirection(direction string) *ebiten.Image {
  // Determine the X position of the row for the given direction
  xPos, ok := directionRows[direction]
  if !ok {
    log.Printf("Invalid direction: %s", direction)
    return nil
  }

  // For simplicity, this example always selects the first frame in the row
  // In a real game, you might cycle through frames based on animation state
  yPos := 0 // Assuming you want the first frame; this could be dynamic

  // Calculate the source rectangle for the frame
  srcRect := image.Rect(xPos, yPos, xPos+frameWidth, yPos+frameHeight)

  // Extract and return the frame as a new *ebiten.Image
  frameImage := spriteSheet.SubImage(srcRect).(*ebiten.Image)
  return frameImage
}

// loadSpriteSheet loads the sprite sheet into memory
func loadSpriteSheet(filePath string) {
  // Read the file into memory
  data, err := ioutil.ReadFile(filePath)
  if err != nil {
    log.Fatalf("Failed to read the sprite sheet file: %v", err)
  }

  // Decode the image
  img, _, err := image.Decode(bytes.NewReader(data))
  if err != nil {
    log.Fatalf("Failed to decode the sprite sheet: %v", err)
  }

  // Convert to an Ebiten image
  spriteSheet = ebiten.NewImageFromImage(img)
}

// updateAnimation updates the current frame based on the time elapsed
func updateAnimation() {
  if time.Since(lastFrameChange) >= frameDuration {
    currentFrame = (currentFrame + 1) % frameCount
    lastFrameChange = time.Now()
  }
}

// drawFrame draws the current frame of the animation
func drawFrame(screen *ebiten.Image) {
  sx := currentFrame * frameWidth
  sy := 0 // Assuming a single row of frames in the sprite sheet
  frameImage := spriteSheet.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)
  screen.DrawImage(frameImage, nil)
}

func update(screen *ebiten.Image) error {
  updateAnimation()
  drawFrame(screen)
  return nil
}

// my audio player :)
func initAudio() {
  audioContext = audio.NewContext(sampleRate)
  soundPlayers = make(map[string]*audio.Player)

  soundFiles := map[string]string{
    "huh": "/Users/beckjonathan/projects/ebitengine-game-template/dsnoway.wav",
    "stop": "/Users/beckjonathan/projects/ebitengine-game-template/dspstop.wav",
  }

  for key, filePath := range soundFiles {
    f, err := os.Open(filePath)
    if err != nil {
      log.Fatalf("Failed to open %s: %v", filePath, err)
    }

    d, err := wav.Decode(audioContext, f)
    if err != nil {
      log.Fatal(err)
    }

    player, err := audio.NewPlayer(audioContext, d)
    if err != nil {
      log.Fatal(err)
    }

    soundPlayers[key] = player
  }
}

func init() {
  var err error
  playerSprite, _, err = ebitenutil.NewImageFromFile("/Users/beckjonathan/projects/ebitengine-game-template/docker.png")
  if err != nil {
    log.Fatal(err)
  }

  initAudio()

  // background image
  img, _, err := ebitenutil.NewImageFromFile("/Users/beckjonathan/projects/ebitengine-game-template/background.jpg")
  if err != nil {
      log.Fatalf("failed to load background image: %v", err)
  }
  bgImage = img
}

func main() {
  loadSpriteSheet("/Users/beckjonathan/projects/ebitengine-game-template/spritesheet.png")
  lastFrameChange = time.Now()
  gameWidth, gameHeight := 640, 480

  ebiten.SetWindowSize(gameWidth, gameHeight)
  ebiten.SetWindowTitle("ebitengine-game-template")

  game := &Game{
    Width:  gameWidth,
    Height: gameHeight,
    Player: &Player{Sprite: playerSprite},
  }

  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}

// Game represents the main game state
type Game struct {
  Width  int
  Height int
  Player *Player
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
  return g.Width, g.Height
}

// Update calculates game logic
func (g *Game) Update() error {

  // Pressing Q any time quits immediately
  if ebiten.IsKeyPressed(ebiten.KeyQ) {
    return errors.New("game quit by player")
  }

  // Pressing F toggles full-screen
  if inpututil.IsKeyJustPressed(ebiten.KeyF) {
    if ebiten.IsFullscreen() {
      ebiten.SetFullscreen(false)
    } else {
      ebiten.SetFullscreen(true)
    }
  }

  // Movement controls
  g.Player.Move()

  return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
  // Draw the background image
  screen.DrawImage(bgImage, nil)
  op := &ebiten.DrawImageOptions{}
  // to be able to move the player around
  op.GeoM.Translate(float64(g.Player.Coords.X), float64(g.Player.Coords.Y))
  // Draw the player sprite
  screen.DrawImage(g.Player.Sprite, op)

  updateAnimation()
  drawFrame(screen)
}

// Player is the player character in the game
type Player struct {
  Coords image.Point
  Sprite *ebiten.Image
}

// Move moves the player upwards
func (p *Player) Move() {
  moveSpeed := 5 // Pixels to move per action

  // Up
  if ebiten.IsKeyPressed(ebiten.KeyUp) {
    audioPlayer := soundPlayers["huh"]
    if(!audioPlayer.IsPlaying()) {
      audioPlayer.Rewind()
      audioPlayer.Play()
    }
    p.Sprite = getFrameForDirection("up")
    p.Coords.Y -= moveSpeed
  }
  // Down
  if ebiten.IsKeyPressed(ebiten.KeyDown) {
    audioPlayer := soundPlayers["stop"]
    if(!audioPlayer.IsPlaying()) {
      audioPlayer.Rewind()
      audioPlayer.Play()
    }
    p.Sprite = getFrameForDirection("down")
    p.Coords.Y += moveSpeed
  }

  // Left
  if ebiten.IsKeyPressed(ebiten.KeyLeft) {
    p.Sprite = getFrameForDirection("left")
    p.Coords.X -= moveSpeed
  }

  // Right
  if ebiten.IsKeyPressed(ebiten.KeyRight) {
    p.Sprite = getFrameForDirection("right")
    p.Coords.X += moveSpeed
  }
}
