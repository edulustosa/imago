package upload_test

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"testing"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/edulustosa/imago/internal/upload"
)

func TestFileSystemUpload(t *testing.T) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		t.Errorf("failed to open image: %v", err)
	}

	imgBuffer := new(bytes.Buffer)
	if err := png.Encode(imgBuffer, img); err != nil {
		t.Errorf("failed to encode image: %v", err)
	}

	ctx := context.Background()

	fsUploader := upload.NewFileSystemUploader(os.Getenv("FILESYSTEM_BASE_URL"))
	_, err = fsUploader.Upload(ctx, imgBuffer.Bytes(), "./test_data/flowers.png")
	if err != nil {
		t.Errorf("failed to upload image: %v", err)
	}

	_, err = imgio.Open("./test_data/flowers.png")
	if err != nil {
		t.Errorf("failed to open image: %v", err)
	}

	imgURL, err := fsUploader.GetImageURL(ctx, "./test_data/flowers.png")
	if err != nil {
		t.Errorf("failed to get image: %v", err)
	}

	t.Logf("image url: %s", imgURL)
}
