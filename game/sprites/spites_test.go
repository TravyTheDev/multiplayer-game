package sprites_test

import (
	"game/sprites"
	"image"
	"testing"
)

func TestSprites(t *testing.T) {
	spritesheet := sprites.NewSpriteSheet(4, 2, 200, 140)
	want := image.Rect(400, 0, 600, 140)
	got := spritesheet.Rect(2)

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
