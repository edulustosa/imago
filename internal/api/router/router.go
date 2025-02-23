package router

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edulustosa/imago/config"
	"github.com/edulustosa/imago/internal/api/handlers"
	"github.com/edulustosa/imago/internal/api/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/edulustosa/imago/docs" // Swagger docs
)

type Server struct {
	Database *pgxpool.Pool
	Env      *config.Env
	S3Client *s3.Client
}

//	@title			Imago API
//	@version		1.0
//	@description	Imago is a backend system for an image processing service similar to Cloudinary.

// @host	localhost:8080
func New(srv Server) http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)

	authHandlers := &handlers.Auth{
		Database: srv.Database,
		Env:      srv.Env,
	}

	r.Post("/register", authHandlers.Register)
	r.Post("/login", authHandlers.Login)

	authMiddleware := &middlewares.AuthMiddleware{Env: srv.Env}
	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.VerifyToken)

		imagesHandler := &handlers.Images{
			Database: srv.Database,
			Env:      srv.Env,
			S3:       srv.S3Client,
		}

		r.Get("/images/{id}", imagesHandler.GetImage)
		r.Get("/images", imagesHandler.GetImages)

		r.Group(func(r chi.Router) {
			r.Use(httprate.Limit(
				10,
				1*time.Minute,
				httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
			))

			r.Post("/images", imagesHandler.Upload)
			r.Post("/images/{id}/transform", imagesHandler.Transform)
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler())

	return r
}
