package imago

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
)

type Transformations struct {
	Resize  Resize  `json:"resize"`
	Crop    Crop    `json:"crop"`
	Rotate  float64 `json:"rotate"`
	Format  string  `json:"format"`
	Filters Filters `json:"filters"`
}

type Resize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Crop struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

type Filters struct {
	Grayscale bool `json:"grayscale"`
	Sepia     bool `json:"sepia"`
}

func Transform(img image.Image, t Transformations) image.Image {
	if t.Resize.Width > 0 || t.Resize.Height > 0 {
		img = transform.Resize(img, t.Resize.Width, t.Resize.Height, transform.Linear)
	}

	if t.Crop.Width > 0 || t.Crop.Height > 0 {
		img = transform.Crop(
			img,
			image.Rect(t.Crop.X, t.Crop.Y, t.Crop.X+t.Crop.Width, t.Crop.Y+t.Crop.Height),
		)
	}

	img = transform.Rotate(img, t.Rotate, nil)
	if t.Filters.Grayscale {
		img = effect.Grayscale(img)
	}

	if t.Filters.Sepia {
		img = effect.Sepia(img)
	}

	return img
}
