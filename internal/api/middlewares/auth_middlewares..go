package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/edulustosa/imago/config"
	"github.com/edulustosa/imago/internal/api"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	Env *config.Env
}

func (m *AuthMiddleware) VerifyToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := getTokenFromAuthHeader(r)
		if err != nil {
			api.SendError(w, http.StatusUnauthorized, api.Error{Message: err.Error()})
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method: %v", t.Header["alg"])
			}

			return []byte(m.Env.JWTSecret), nil
		})
		if err != nil {
			api.SendError(w, http.StatusUnauthorized, api.Error{Message: err.Error()})
			return
		}

		userID, err := parseUserIDFromToken(token)
		if err != nil {
			api.SendError(w, http.StatusUnauthorized, api.Error{Message: err.Error()})
			return
		}

		ctx := context.WithValue(r.Context(), api.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTokenFromAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return "", errors.New("invalid token format")
	}

	return token, nil
}

func parseUserIDFromToken(token *jwt.Token) (uuid.UUID, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		userID, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			return uuid.Nil, errors.New("invalid user id in token")
		}

		return userID, nil
	}

	return uuid.Nil, errors.New("invalid token claims")
}
