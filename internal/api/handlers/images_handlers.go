package handlers

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edulustosa/imago/config"
	"github.com/edulustosa/imago/internal/api"
	"github.com/edulustosa/imago/internal/domain/img"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/services/imgproc"
	"github.com/edulustosa/imago/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Images struct {
	Database *pgxpool.Pool
	Env      *config.Env
	S3       *s3.Client
}

func (h *Images) Upload(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(api.UserIDKey).(uuid.UUID)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		api.SendError(w, http.StatusBadRequest, api.Error{
			Message: "failed to parse form",
		})
		return
	}

	imgFile, handler, err := r.FormFile("image")
	if err != nil {
		api.SendError(w, http.StatusBadRequest, api.Error{
			Message: "failed to read file",
		})
		return
	}
	defer imgFile.Close()

	imgData, err := io.ReadAll(imgFile)
	if err != nil {
		api.InternalError(w, "failed to read image", "error", err)
		return
	}

	userRepository := user.NewRepo(h.Database)
	imageRepository := img.NewRepo(h.Database)
	s3ImageStorage := storage.NewS3ImageStorage(h.S3, h.Env.BucketName)

	upload := imgproc.NewUpload(userRepository, imageRepository, s3ImageStorage)
	imgInfo, err := upload.Do(r.Context(), userID, imgData, &imgproc.ImageMetadata{
		Filename: handler.Filename,
		Format:   strings.TrimPrefix(filepath.Ext(handler.Filename), "."),
		Alt:      r.FormValue("alt"),
	})
	if err != nil {
		if errors.Is(err, imgproc.ErrUserNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "user not found",
			})
			return
		}

		if errors.Is(err, imgproc.ErrInvalidImage) {
			api.SendError(w, http.StatusBadRequest, api.Error{
				Message: err.Error(),
			})
			return
		}

		api.InternalError(w, "failed to upload image", "error", err)
		return
	}

	api.Encode(w, http.StatusCreated, imgInfo)
}

type TransformRequest struct {
	Transformations imgproc.Transformations `json:"transformations" validate:"required"`
}

func (h *Images) Transform(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(api.UserIDKey).(uuid.UUID)
	imageID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.SendError(w, http.StatusBadRequest, api.Error{
			Message: "invalid image id",
		})
		return
	}

	t, problems, err := api.Decode[TransformRequest](r)
	if err != nil {
		api.InvalidRequest(w, problems)
		return
	}

	imageRepository := img.NewRepo(h.Database)
	s3ImageStorage := storage.NewS3ImageStorage(h.S3, h.Env.BucketName)

	transformation := imgproc.NewImageTransformation(imageRepository, s3ImageStorage)
	imgInfo, err := transformation.Transform(r.Context(), imageID, userID, &t.Transformations)
	if err != nil {
		if errors.Is(err, imgproc.ErrImageNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: err.Error(),
			})
			return
		}

		if errors.Is(err, imgproc.ErrUnsupportedFormat) {
			api.SendError(w, http.StatusBadRequest, api.Error{
				Message: "unsupported image format",
			})
			return
		}

		api.InternalError(w, "failed to transform image", "error", err)
		return
	}

	api.Encode(w, http.StatusOK, imgInfo)
}

func (h *Images) GetImage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(api.UserIDKey).(uuid.UUID)
	imageID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.SendError(w, http.StatusBadRequest, api.Error{
			Message: "invalid image id",
		})
		return
	}

	userRepository := user.NewRepo(h.Database)
	imageRepository := img.NewRepo(h.Database)
	imageService := img.NewService(imageRepository, userRepository)

	imgInfo, err := imageService.GetImage(r.Context(), imageID, userID)
	if err != nil {
		if errors.Is(err, img.ErrImageNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "image not found",
			})
			return
		}

		if errors.Is(err, img.ErrUserNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "user not found",
			})
			return
		}
	}

	api.Encode(w, http.StatusOK, imgInfo)
}

func (h *Images) GetImages(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(api.UserIDKey).(uuid.UUID)
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10
	}

	userRepository := user.NewRepo(h.Database)
	imageRepository := img.NewRepo(h.Database)
	imageService := img.NewService(imageRepository, userRepository)

	imgs, err := imageService.GetImages(r.Context(), userID, page, limit)
	if err != nil {
		if errors.Is(err, img.ErrUserNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "user not found",
			})
			return
		}

		api.InternalError(w, "failed to get images", "error", err)
		return
	}

	api.Encode(w, http.StatusOK, api.Map{"images": imgs})
}
