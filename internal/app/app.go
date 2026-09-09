package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mabarik667f/fsserver/internal/closer"
	"github.com/Mabarik667f/fsserver/pkg/config"
)

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
}

func New() *App {
	a := &App{
		diContainer: newDiContainer(),
	}

	a.initDeps()

	return a
}

func (a *App) initDeps() {
	inits := []func(){
		a.initHTTPServer,
	}

	for _, fn := range inits {
		fn()
	}
}

func (a *App) initHTTPServer() {
	a.httpServer = &http.Server{
		Addr:    config.AppConfig().ServiceHTTPConnString(),
		Handler: a.diContainer.API(),
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("server started", "addr", config.AppConfig().ServiceHTTPConnString())
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.Error("app error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("error to close http server", "err", err)
	}

	slog.Info("server turned off")

	closerCtx, closerCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closerCancel()

	if err := closer.CloseAll(closerCtx); err != nil {
		slog.Error("errors when closing resources", "err", err)
	}

	return nil
}
