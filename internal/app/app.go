package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"mascot/internal/config"
	"mascot/internal/db"
	"mascot/internal/handlers"
	"mascot/internal/repositories"
	"mascot/internal/services"
	"mascot/internal/transport"
)

type Service struct {
	logger   *zap.Logger
	closers  []Closer
	shutdown atomic.Bool
}

func NewService(logger *zap.Logger) *Service {
	return &Service{logger: logger}
}

//Start blocking method. Use goroutine
func (s *Service) Start(ctx context.Context, cfg config.Config) {
	server := transport.NewServer(transport.WithUseValidator()).UseMiddlewares(
		transport.LoggingMiddleware(s.logger),
		transport.RecoverMiddleware(s.logger),
	)

	conn, err := pgxpool.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		s.logger.Fatal("db connect", zap.Error(err))
	}

	transactor := db.NewTransactor(conn, s.logger)

	//repositories
	walletRepo := repositories.NewWallet(transactor)

	//services
	walletService := services.NewWallet(transactor, walletRepo)

	//handlers
	handler := handlers.NewHandler(walletService)

	err = server.RegisterServices(
		"getBalance", handler.GetBalance,
		"withdrawAndDeposit", handler.WithdrawAndDeposit,
		"rollbackTransaction", handler.RollbackTransaction,
	)
	if err != nil {
		s.logger.Fatal("register services", zap.Error(err))
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.SeamlessURI, server.HandleFunc())
	httpServer := http.Server{Addr: cfg.Addr, Handler: mux}

	s.AddClose(httpServer.Shutdown)

	if err := httpServer.ListenAndServe(); err != nil {
		if (s.shutdown.Load() && !errors.Is(err, http.ErrServerClosed)) || !s.shutdown.Load() {
			s.logger.Fatal("listen and serve", zap.Error(err))
		}
	}
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.shutdown.Store(true)
	for _, closer := range s.closers {
		if err := closer.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) AddClose(closer Closer) {
	s.closers = append(s.closers, closer)
}

type Closer func(context.Context) error

func (c Closer) Close(ctx context.Context) error {
	return c(ctx)
}
