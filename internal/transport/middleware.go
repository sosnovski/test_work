package transport

import (
	"context"
	"time"

	"go.uber.org/zap"

	"mascot/internal/handlers"
)

type HandlerFunc func(ctx context.Context, req *ServerRequest, resp *ServerResponse)
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

func LoggingMiddleware(log *zap.Logger) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, req *ServerRequest, resp *ServerResponse) {
			start := time.Now()
			logger := log.With(
				zap.String("method", req.Method),
				zap.Any("params", req.Params),
				zap.Duration("duration", start.Sub(time.Now())),
			)

			next(ctx, req, resp)

			if resp.Error != nil {
				logger.Error(
					"handle request error",
					zap.String("error", resp.Error.Error()),
				)
			} else {
				logger.Info("handle request",
					zap.Any("result", resp.Result))
			}
		}
	}
}

func RecoverMiddleware(log *zap.Logger) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, req *ServerRequest, resp *ServerResponse) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic occurred", zap.Any("panic", err))
					resp.Id = req.Id
					resp.Error = &handlers.Error{
						Code:    handlers.ErrInternalError,
						Message: "internal server error",
					}
				}
			}()

			next(ctx, req, resp)
		}
	}
}
