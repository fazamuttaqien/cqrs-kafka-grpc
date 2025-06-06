package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func (s *server) runHealthCheck(ctx context.Context) {
	health := healthcheck.NewHandler()

	health.AddReadinessCheck(s.config.ServiceName, healthcheck.AsyncWithContext(ctx, func() error {
		if s.config != nil {
			return nil
		}
		return errors.New("Config not loaded")
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	go func() {
		s.log.Infof("API_Gateway Kubernetes probes listening on port: %s", s.config.Probes.Port)
		if err := http.ListenAndServe(s.config.Probes.Port, health); err != nil {
			s.log.WarnMsg("ListenAndServe", err)
		}
	}()
}

func (s *server) runMetrics(cancel context.CancelFunc) {
	metricsServer := fiber.New()
	metricsServer.Use(recover.New(recover.Config{
		EnableStackTrace: false,
	}))

	go func() {
		// Convert Prometheus HTTP handler to Fiber handler
		promHandler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
		metricsServer.Get(s.config.Probes.PrometheusPath, func(c *fiber.Ctx) error {
			promHandler(c.Context())
			return nil
		})

		s.log.Infof("Metrics server is running on port: %s", s.config.Probes.PrometheusPort)
		if err := metricsServer.Listen(s.config.Probes.PrometheusPort); err != nil {
			s.log.Errorf("metricsServer.Listen: %v", err)
			cancel()
		}
	}()
}
