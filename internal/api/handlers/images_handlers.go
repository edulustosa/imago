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
	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/img"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/queue"
	"github.com/edulustosa/imago/internal/services/imgproc"
	"github.com/edulustosa/imago/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type Images struct {
	Database    *pgxpool.Pool
	Env         *config.Env
	S3Client    *s3.Client
	RedisClient *redis.Client
	KafkaWriter *kafka.Writer
}

// @Summary	Upload an image
// @Tags		images
//
// @Accept		multipart/form-data
// @Produce		json
//
// @Param		image formData file true "Image file"
// @Param		alt formData string false "Image alt text"
//
// @Success	201	{object} models.Image
// @Failure	400	{object} api.Error "Invalid request"
// @Failure	401	{object} api.Error "Unauthorized"
// @Failure	404	{object} api.Error "User not found"
// @Failure	500	{object} api.Error "Internal server error"
//
// @Security	BearerAuth
// @Router		/images [post]
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
	s3ImageStorage := storage.NewS3ImageStorage(h.S3Client, h.Env.BucketName)

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

// @Summary	Transform an image
// @Tags		images
//
// @Accept		json
// @Produce		json
//
// @Param		id path int true "Image id"
// @Param		transformations body TransformRequest true "Image operations"
//
// @Success	200	{object} queue.TransformationStatus
// @Failure	400	{object} api.Error "Invalid parameters"
// @Failure	401	{object} api.Error "Unauthorized"
// @Failure	500	{object} api.Error "Internal server error"
//
// @Security	BearerAuth
// @Router		/images/{id}/transform [post]
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

	transformationsProducer := queue.NewTransformationProducer(h.KafkaWriter, h.RedisClient)
	processStatus, err := transformationsProducer.Enqueue(r.Context(), &queue.TransformationMessage{
		ImageID:         imageID,
		UserID:          userID,
		Transformations: &t.Transformations,
	})
	if err != nil {
		api.InternalError(w, "failed to enqueue transformation", "error", err)
		return
	}

	api.Encode(w, http.StatusAccepted, processStatus)
}

// @Summary	Get an image
// @Tags		images
//
// @Param		id path int true "Image id"
// @Produce		json
//
// @Success	200	{object} models.Image
// @Failure	400	{object} api.Error "Invalid image id"
// @Failure	401	{object} api.Error "Unauthorized"
// @Failure	404	{object} api.Error "Image or user not found"
// @Failure	500	{object} api.Error "Internal server error"
//
// @Security	BearerAuth
// @Router		/images/{id} [get]
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

type GetImagesResponse struct {
	Images []models.Image `json:"images"`
}

// @Summary	Get images
// @Tags		images
//
// @Param		page query int false "Page number"
// @Param		limit query int false "Number of images per page"
// @Produce		json
//
// @Success 200 {object} GetImagesResponse
// @Failure	401	{object} api.Error "Unauthorized"
// @Failure	404	{object} api.Error "Image or user not found"
// @Failure	500	{object} api.Error "Internal server error"
//
// @Security	BearerAuth
// @Router		/images [get]
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

	api.Encode(w, http.StatusOK, GetImagesResponse{imgs})
}
