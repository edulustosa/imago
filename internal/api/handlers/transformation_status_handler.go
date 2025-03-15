package handlers

import (
	"net/http"

	"github.com/edulustosa/imago/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

// @Summary	Get transformation status of an image transformation
// @Tags images
//
// @Param	id path string true "Callback id"
// @Produce json
//
// @Success 200 {object} queue.TransformationStatus "Status of the transformation"
// @Failure 404 {object} api.Error "Status not found"
// @Failure 500 {object} api.Error "Internal server error"
//
// @Router /images/{id}/status [get]
// @Security BearerAuth
func GetTransformationStatus(redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callbackID := chi.URLParam(r, "id")

		cmd := redisClient.Get(r.Context(), callbackID)
		if err := cmd.Err(); err != nil {
			if err == redis.Nil {
				api.SendError(w, http.StatusNotFound, api.Error{
					Message: "status not found",
				})
				return
			}

			api.InternalError(w, "failed to get status", "redis", err)
			return
		}

		statusBytes, err := cmd.Bytes()
		if err != nil {
			api.InternalError(w, "failed to decode status", "redis", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(statusBytes)
	}
}
