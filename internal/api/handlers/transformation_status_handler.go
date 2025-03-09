package handlers

import (
	"net/http"

	"github.com/edulustosa/imago/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

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
