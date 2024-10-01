package sprites

import "image"

type SpriteSheet struct {
	WidthInTiles   int
	HeightInTiles  int
	TileSizeWidth  int
	TileSizeHeight int
}

func NewSpriteSheet(w, h, tw, th int) *SpriteSheet {
	return &SpriteSheet{
		WidthInTiles:   w,
		HeightInTiles:  h,
		TileSizeWidth:  tw,
		TileSizeHeight: th,
	}
}

func (s *SpriteSheet) Rect(index int) image.Rectangle {
	x := (index % s.WidthInTiles) * s.TileSizeWidth
	y := (index / s.WidthInTiles) * s.TileSizeHeight

	return image.Rect(
		x, y, x+s.TileSizeWidth, y+s.TileSizeHeight,
	)
}
