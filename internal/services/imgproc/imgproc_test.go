package imgproc_test

import (
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
		t.Skip("could not be implemented")
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
