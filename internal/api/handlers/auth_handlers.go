package handlers

import (
	"errors"
	"net/http"

	"github.com/edulustosa/imago/internal/api"
	"github.com/edulustosa/imago/internal/auth"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Auth struct {
	Database *pgxpool.Pool
}

func (a *Auth) Register(w http.ResponseWriter, r *http.Request) {
	req, problems, err := api.Decode[auth.Request](r)
	if err != nil {
		api.InvalidRequest(w, problems)
		return
	}

	userRepo := user.NewRepo(a.Database)
	authService := auth.New(userRepo)

	user, err := authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			api.SendError(w, http.StatusConflict, api.Error{Message: err.Error()})
			return
		}

		api.InternalError(w)
		return
	}

	api.Encode(w, http.StatusCreated, user)
}
