package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/handler"
	"go.uber.org/zap"
)

func NewRouter(
	userHandler *handler.UserHandler,
	teamHandler *handler.TeamHandler,
	prHandler *handler.PrHandler,
	healthHandler *handler.HealthHandler,
	log *zap.Logger,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)

	router.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
		r.Get("/getReview", userHandler.GetReview)
	})

	router.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.AddTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	router.Route("/pullRequest", func(r chi.Router) {
		r.Post("create", prHandler.CreatePr)
		r.Post("/merge", prHandler.MergePr)
		r.Post("reassign", prHandler.ReassignPr)
	})

	router.Get("/health", healthHandler.HealthCheck)
	return router
}
