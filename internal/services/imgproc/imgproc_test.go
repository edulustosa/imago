package imgproc_test

import (
	"bytes"
	"image"
	"testing"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/edulustosa/imago/internal/services/imgproc"
)

func TestImageTransformations(t *testing.T) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		t.Fatalf("failed to open image file: %v", err)
	}

	t.Run("resize", func(t *testing.T) {
		resizedImg := imgproc.Transform(img, &imgproc.Transformations{
			Resize: imgproc.Resize{
				Width:  100,
				Height: 100,
			},
		})

		resizedBounds := resizedImg.Bounds()
		if resizedBounds.Dx() != 100 || resizedBounds.Dy() != 100 {
			t.Errorf(
				"expected image to be 100x100, got %dx%d",
				resizedBounds.Dx(),
				resizedBounds.Dy(),
			)
		}
	})

	t.Run("crop", func(t *testing.T) {
		croppedImg := imgproc.Transform(img, &imgproc.Transformations{
			Crop: imgproc.Crop{
				Width:  100,
				Height: 100,
				X:      50,
				Y:      50,
			},
		})

		croppedBounds := croppedImg.Bounds()
		if croppedBounds.Dx() != 100 || croppedBounds.Dy() != 100 {
			t.Errorf(
				"expected image to be 100x100, got %dx%d",
				croppedBounds.Dx(),
				croppedBounds.Dy(),
			)
		}

		if croppedBounds.Min.X != 50 || croppedBounds.Min.Y != 50 {
			t.Errorf(
				"expected image to start at 50x50, got %dx%d",
				croppedBounds.Min.X,
				croppedBounds.Min.Y,
			)
		}
	})

	t.Run("rotate", func(t *testing.T) {
		rotatedImg := imgproc.Transform(img, &imgproc.Transformations{
			Rotate: 90,
		})

		originalWidth := img.Bounds().Dy()
		originalHeight := img.Bounds().Dx()
		rotatedBounds := rotatedImg.Bounds()

		if rotatedBounds.Dx() != originalHeight || rotatedBounds.Dy() != originalWidth {
			t.Errorf(
				"expected image to be %dx%d, got %dx%d",
				originalHeight,
				originalWidth,
				rotatedBounds.Dx(),
				rotatedBounds.Dy(),
			)
		}
	})

	t.Run("grayscale", func(t *testing.T) {
		grayscaleImg := imgproc.Transform(img, &imgproc.Transformations{
			Filters: imgproc.Filters{
				Grayscale: true,
			},
		})

		if !isGrayscale(grayscaleImg) {
			t.Error("expected image to be grayscale")
		}
	})

	t.Run("sepia", func(t *testing.T) {
		sepiaImg := imgproc.Transform(img, &imgproc.Transformations{
			Filters: imgproc.Filters{
				Sepia: true,
			},
		})

		if !isSepia(sepiaImg) {
			t.Error("expected image to be sepia")
		}
	})
}

func isGrayscale(img image.Image) bool {
	bounds := img.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r>>8 != g>>8 || g>>8 != b>>8 {
				return false
			}
		}
	}

	return true
}

func isSepia(img image.Image) bool {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	totalPixels := width * height
	sepiaPixels := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			if g8 == 0 || b8 == 0 {
				continue
			}

			redGreenRatio := r8 / g8
			redBlueRatio := r8 / b8

			if r8 > g8 && g8 > b8 &&
				redGreenRatio >= 1.0 && redGreenRatio <= 1.4 &&
				redBlueRatio >= 1.4 && redBlueRatio <= 2.5 {
				sepiaPixels++
			}
		}
	}

	sepiaPercentage := float64(sepiaPixels) / float64(totalPixels)
	return sepiaPercentage >= 0.7
}

func TestImageEncoding(t *testing.T) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		t.Fatalf("failed to open image file: %v", err)
	}

	testCases := []string{
		"png",
		"jpeg",
		"bmp",
		"tiff",
		"gif",
		"webp",
	}

	for _, format := range testCases {
		t.Run(format, func(t *testing.T) {
			want := format

			imgBuff := new(bytes.Buffer)
			if err := imgproc.Encode(imgBuff, img, format); err != nil {
				t.Errorf("failed to encode image to %s: %v", format, err)
			}

			_, got, err := image.Decode(imgBuff)
			if err != nil {
				t.Errorf("failed to decode image: %v", err)
			}

			if got != want {
				t.Errorf("expected image format to be %s, got %s", want, got)
			}
		})
	}

	t.Run("unsupported format", func(t *testing.T) {
		imgBuff := new(bytes.Buffer)
		err := imgproc.Encode(imgBuff, img, "unsupported")
		if err == nil {
			t.Error("expected to fail encoding image to unsupported format")
		}
		if err != imgproc.ErrUnsupportedFormat {
			t.Errorf("expected error to be %v, got %v", imgproc.ErrUnsupportedFormat, err)
		}
	})
}
