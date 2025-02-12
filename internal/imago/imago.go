package imago

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
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

type EncoderFunc func(io.Writer, image.Image) error

var encoders = map[string]EncoderFunc{
	"jpeg": toJpeg,
	"jpg":  toJpeg,
	"png":  toPng,
	"gif":  toGif,
	"bmp":  toBmp,
	"tiff": toTiff,
	"tif":  toTiff,
	"webp": toWebp,
}

func Encode(w io.Writer, img image.Image, format string) error {
	encoder, ok := encoders[format]
	if !ok {
		return fmt.Errorf("unsupported file extension: %s", format)
	}

	if err := encoder(w, img); err != nil {
		return fmt.Errorf("failed to encode %s file: %w", format, err)
	}

	return nil
}

func toJpeg(w io.Writer, img image.Image) error {
	if err := jpeg.Encode(w, img, nil); err != nil {
		return fmt.Errorf("failed to convert jpeg: %w", err)
	}

	return nil
}

func toPng(w io.Writer, img image.Image) error {
	if err := png.Encode(w, img); err != nil {
		return fmt.Errorf("failed to convert jpeg: %w", err)
	}

	return nil
}

func toWebp(w io.Writer, img image.Image) error {
	if err := webp.Encode(w, img, nil); err != nil {
		return fmt.Errorf("failed to convert jpeg: %w", err)
	}

	return nil
}

func toGif(w io.Writer, img image.Image) error {
	if err := gif.Encode(w, img, nil); err != nil {
		return fmt.Errorf("failed to convert jpeg: %w", err)
	}

	return nil
}

func toBmp(w io.Writer, img image.Image) error {
	if err := bmp.Encode(w, img); err != nil {
		return fmt.Errorf("failed to convert jpeg: %w", err)
	}

	return nil
}

func toTiff(w io.Writer, img image.Image) error {
	if err := tiff.Encode(w, img, nil); err != nil {
		return fmt.Errorf("failed to convert jpeg: %w", err)
	}

	return nil
}
