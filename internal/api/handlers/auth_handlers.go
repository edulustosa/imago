package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/edulustosa/imago/config"
	"github.com/edulustosa/imago/internal/api"
	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/services/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Auth struct {
	Database *pgxpool.Pool
	Env      *config.Env
}

func (a *Auth) Register(w http.ResponseWriter, r *http.Request) {
	req, problems, err := api.Decode[auth.Request](r)
	if err != nil {
		api.InvalidRequest(w, problems)
		return
	}

	userRepo := user.NewRepo(a.Database)
	authService := auth.New(userRepo)

	user, err := authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			api.SendError(w, http.StatusConflict, api.Error{Message: err.Error()})
			return
		}

		api.InternalError(w, "failed to register user", "error", err)
		return
	}

	api.Encode(w, http.StatusCreated, user)
}

type LoginResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	req, problems, err := api.Decode[auth.Request](r)
	if err != nil {
		api.InvalidRequest(w, problems)
		return
	}

	userRepo := user.NewRepo(a.Database)
	authService := auth.New(userRepo)

	user, err := authService.Login(r.Context(), &req)
	if err != nil {
		// The only error that can be returned here is ErrInvalidCredentials
		api.SendError(w, http.StatusUnauthorized, api.Error{Message: err.Error()})
		return
	}

	token, err := createJWT(user.ID, a.Env.JWTSecret)
	if err != nil {
		api.InternalError(w, "failed to create jwt", "error", err)
		return
	}

	resp := &LoginResponse{
		User:  user,
		Token: token,
	}

	api.Encode(w, http.StatusOK, resp)
}

func createJWT(userID uuid.UUID, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(secret))
}
