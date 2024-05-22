package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"

	"github.com/AwareRO/libaware/golang/http/handlers"
	"github.com/AwareRO/libaware/golang/metrics"
)

func RunServerWithMetrics(cfg *Config, router *httprouter.Router, collector metrics.Collector) {
	router.GET("/metrics", handlers.FromStdlib(collector.GetHttpHandler()))
	RunServer(cfg, router)
}

func RunServer(cfg *Config, handler http.Handler) {
	ctx, cancel := context.WithCancel(context.Background())

	addr := fmt.Sprintf("localhost:%d", cfg.Port)

	srv := &http.Server{
		Addr:        addr,
		Handler:     handler,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	log.Info().Str("addr", srv.Addr).Msg("Listening")

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("HTTP listen and serve")
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	<-signalChan

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(gracefullCtx); err != nil {
		log.Error().Err(err).Msg("Shutdown error")
		defer os.Exit(1)
		cancel()

		return
	} else {
		log.Info().Msg("gracefully stopped")
	}

	cancel()
}
