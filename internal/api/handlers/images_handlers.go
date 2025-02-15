package handlers

import (
	"bytes"
	"errors"
	"image"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edulustosa/imago/config"
	"github.com/edulustosa/imago/internal/api"
	"github.com/edulustosa/imago/internal/domain/images"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/imago"
	"github.com/edulustosa/imago/internal/services/storage"
	"github.com/edulustosa/imago/internal/upload"
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

	file, handler, err := r.FormFile("image")
	if err != nil {
		api.SendError(w, http.StatusBadRequest, api.Error{
			Message: "failed to read file",
		})
		return
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		api.SendError(w, http.StatusBadRequest, api.Error{
			Message: "failed to decode image",
		})
		return
	}

	imgBuff := new(bytes.Buffer)
	if err := imago.Encode(imgBuff, img, format); err != nil {
		api.SendError(w, http.StatusBadRequest, api.Error{
			Message: err.Error(),
		})
		return
	}

	userRepository := user.NewRepo(h.Database)
	imageRepository := images.NewRepo(h.Database)
	s3Uploader := upload.NewS3Uploader(h.S3, h.Env.BucketName)

	imageStorage := storage.NewImageStorage(s3Uploader, userRepository, imageRepository)
	imgInfo, err := imageStorage.Upload(r.Context(), userID, imgBuff.Bytes(), storage.Metadata{
		Filename: handler.Filename,
		Format:   format,
		Alt:      r.FormValue("alt"),
	})
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "user not found",
			})
			return
		}

		api.InternalError(w, "failed to upload image", "error", err)
		return
	}

	api.Encode(w, http.StatusCreated, imgInfo)
}

type TransformRequest struct {
	Transformations imago.Transformations `json:"transformations" validate:"required"`
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

	userRepository := user.NewRepo(h.Database)
	imageRepository := images.NewRepo(h.Database)
	s3Uploader := upload.NewS3Uploader(h.S3, h.Env.BucketName)

	imageStorage := storage.NewImageStorage(s3Uploader, userRepository, imageRepository)
	imgInfo, err := imageStorage.Transform(r.Context(), userID, imageID, t.Transformations)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "user not found",
			})
			return
		}

		if errors.Is(err, storage.ErrImageNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "image not found",
			})
			return
		}

		if errors.Is(err, storage.ErrInvalidFormat) {
			api.SendError(w, http.StatusBadRequest, api.Error{
				Message: "image format not supported",
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
	imageRepository := images.NewRepo(h.Database)
	imageService := images.NewService(imageRepository, userRepository)

	imgInfo, err := imageService.GetImage(r.Context(), imageID, userID)
	if err != nil {
		if errors.Is(err, images.ErrImageNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "image not found",
			})
			return
		}

		if errors.Is(err, images.ErrUserNotFound) {
			api.SendError(w, http.StatusNotFound, api.Error{
				Message: "user not found",
			})
			return
		}
	}

	api.Encode(w, http.StatusOK, imgInfo)
}
