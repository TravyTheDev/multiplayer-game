package sprites

import "image"

type SpriteSheet struct {
	WidthInTiles  int
	HeightInTiles int
	TileSize      int
}

func NewSpriteSheet(w, h, t int) *SpriteSheet {
	return &SpriteSheet{
		WidthInTiles:  w,
		HeightInTiles: h,
		TileSize:      t,
	}
}

func (s *SpriteSheet) Rect(index int) image.Rectangle {
	x := (index % s.WidthInTiles) * s.TileSize
	y := (index / s.HeightInTiles) * s.TileSize

	return image.Rect(
		x, y, x+s.TileSize, y+s.TileSize,
	)
}
