package lifecycle

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const defaultShutdownTimeout = 5 * time.Second

// GracefulShutdown manages the graceful shutdown process for the HTTP server
func GracefulShutdown(server *http.Server, logger *zap.Logger) chan bool {
	done := make(chan bool, 1)

	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		<-ctx.Done()

		logger.Info("initiating graceful shutdown",
			zap.String("signal", ctx.Err().Error()),
		)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server forced to shutdown",
				zap.Error(err),
				zap.Duration("timeout", defaultShutdownTimeout),
			)
		}

		logger.Info("server exited gracefully")
		done <- true
	}()

	return done
}
