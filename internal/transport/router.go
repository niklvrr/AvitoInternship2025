package transport

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/handler"
	transportMiddleware "github.com/niklvrr/AvitoInternship2025/internal/transport/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func NewRouter(
	userHandler *handler.UserHandler,
	teamHandler *handler.TeamHandler,
	prHandler *handler.PrHandler,
	statsHandler *handler.StatsHandler,
	healthHandler *handler.HealthHandler,
	log *zap.Logger,
) *chi.Mux {
	router := chi.NewRouter()

	// Recovery должен быть первым для обработки паник во всех middleware
	router.Use(transportMiddleware.Recovery(log))

	// RequestID для трейсинга запросов
	router.Use(middleware.RequestID)

	// Logging для структурированного логирования всех запросов
	router.Use(transportMiddleware.Logging(log))

	// Timeout для контроля времени выполнения запросов (500ms для соблюдения SLI 300ms)
	router.Use(transportMiddleware.Timeout(500*time.Millisecond, log))

	// Metrics для сбора метрик производительности
	router.Use(transportMiddleware.Metrics)

	// Эндпоинт для Prometheus метрик
	router.Handle("/metrics", promhttp.Handler())

	router.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
		r.Get("/getReview", userHandler.GetReview)
	})

	router.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.AddTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	router.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", prHandler.CreatePr)
		r.Post("/merge", prHandler.MergePr)
		r.Post("/reassign", prHandler.ReassignPr)
	})

	router.Get("/stats", statsHandler.GetStats)

	router.Get("/health", healthHandler.HealthCheck)
	return router
}
