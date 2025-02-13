package imago_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/edulustosa/imago/internal/imago"
)

func TestImageResize(t *testing.T) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		t.Errorf("failed to get image file: %s", err.Error())
	}

	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	t.Run("resize", func(t *testing.T) {
		resizedImg := imago.Transform(img, imago.Transformations{
			Resize: imago.Resize{
				Width:  100,
				Height: 100,
			},
		})

		resizedImgWidth := resizedImg.Bounds().Dx()
		resizedImgHeight := resizedImg.Bounds().Dy()

		if resizedImgWidth != 100 || resizedImgHeight != 100 {
			t.Errorf(
				"expected image to be 100x100, got %dx%d",
				resizedImgWidth,
				resizedImgHeight,
			)
		}
	})

	t.Run("it should not resize if width and height are 0", func(t *testing.T) {
		resizedImg := imago.Transform(img, imago.Transformations{
			Resize: imago.Resize{
				Width:  0,
				Height: 0,
			},
		})

		resizedImgWidth := resizedImg.Bounds().Dx()
		resizedImgHeight := resizedImg.Bounds().Dy()

		if resizedImgWidth != imgWidth || resizedImgHeight != imgHeight {
			t.Errorf(
				"expected image to be %dx%d, got %dx%d",
				imgWidth,
				imgHeight,
				resizedImgWidth,
				resizedImgHeight,
			)
		}
	})
}

func TestImageCrop(t *testing.T) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		t.Errorf("failed to get image file: %s", err.Error())
	}

	t.Run("crop", func(t *testing.T) {
		croppedImg := imago.Transform(img, imago.Transformations{
			Crop: imago.Crop{
				Width:  100,
				Height: 100,
				X:      50,
				Y:      50,
			},
		})

		croppedImgWidth := croppedImg.Bounds().Dx()
		croppedImgHeight := croppedImg.Bounds().Dy()
		croppedImgX := croppedImg.Bounds().Min.X
		croppedImgY := croppedImg.Bounds().Min.Y

		if croppedImgWidth != 100 || croppedImgHeight != 100 {
			t.Errorf(
				"expected image to be 100x100, got %dx%d",
				croppedImgWidth,
				croppedImgHeight,
			)
		}

		if croppedImgX != 50 || croppedImgY != 50 {
			t.Errorf(
				"expected image to start at 50x50, got %dx%d",
				croppedImgX,
				croppedImgY,
			)
		}
	})

	t.Run("it should not crop if width and height are 0", func(t *testing.T) {
		croppedImg := imago.Transform(img, imago.Transformations{
			Crop: imago.Crop{
				Width:  0,
				Height: 0,
				X:      0,
				Y:      0,
			},
		})

		croppedImgWidth := croppedImg.Bounds().Dx()
		croppedImgHeight := croppedImg.Bounds().Dy()

		if croppedImgWidth != img.Bounds().Dx() || croppedImgHeight != img.Bounds().Dy() {
			t.Errorf(
				"expected image to be %dx%d, got %dx%d",
				img.Bounds().Dx(),
				img.Bounds().Dy(),
				croppedImgWidth,
				croppedImgHeight,
			)
		}
	})

}

func TestChangeImageColor(t *testing.T) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		t.Errorf("failed to get image file: %s", err.Error())
	}

	t.Run("grayscale", func(_ *testing.T) {
		grayscaleImg := imago.Transform(img, imago.Transformations{
			Filters: imago.Filters{
				Grayscale: true,
			},
		})

		_ = imgio.Save("./test_data/grayscale.jpg", grayscaleImg, imgio.JPEGEncoder(100))
	})

	t.Run("sepia", func(_ *testing.T) {
		sepiaImg := imago.Transform(img, imago.Transformations{
			Filters: imago.Filters{
				Sepia: true,
			},
		})

		_ = imgio.Save("./test_data/sepia.jpg", sepiaImg, imgio.JPEGEncoder(100))
	})
}

func TestConvertImage(t *testing.T) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		t.Errorf("failed to get image file: %s", err.Error())
	}

	t.Run("jpeg", func(t *testing.T) {
		buff := new(bytes.Buffer)
		if err := imago.Encode(buff, img, "jpeg"); err != nil {
			t.Errorf("failed to encode image: %s", err.Error())
		}

		ct := http.DetectContentType(buff.Bytes())
		assertContentType(t, "image/jpeg", ct)
	})

	t.Run("png", func(t *testing.T) {
		buff := new(bytes.Buffer)
		if err := imago.Encode(buff, img, "png"); err != nil {
			t.Errorf("failed to encode image: %s", err.Error())
		}

		ct := http.DetectContentType(buff.Bytes())
		assertContentType(t, "image/png", ct)
	})

	t.Run("gif", func(t *testing.T) {
		buff := new(bytes.Buffer)
		if err := imago.Encode(buff, img, "gif"); err != nil {
			t.Errorf("failed to encode image: %s", err.Error())
		}

		ct := http.DetectContentType(buff.Bytes())
		assertContentType(t, "image/gif", ct)
	})

	t.Run("bmp", func(t *testing.T) {
		buff := new(bytes.Buffer)
		if err := imago.Encode(buff, img, "bmp"); err != nil {
			t.Errorf("failed to encode image: %s", err.Error())
		}

		ct := http.DetectContentType(buff.Bytes())
		assertContentType(t, "image/bmp", ct)
	})

	t.Run("webp", func(t *testing.T) {
		buff := new(bytes.Buffer)
		if err := imago.Encode(buff, img, "webp"); err != nil {
			t.Errorf("failed to encode image: %s", err.Error())
		}

		ct := http.DetectContentType(buff.Bytes())
		assertContentType(t, "image/webp", ct)
	})

	t.Run("unsupported format", func(t *testing.T) {
		buff := new(bytes.Buffer)
		if err := imago.Encode(buff, img, "unsupported"); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func assertContentType(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected content type %s, got %s", expected, actual)
	}
}
