package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fjrevoredo/openclaudio/internal/auth"
	"github.com/fjrevoredo/openclaudio/internal/config"
	webapp "github.com/fjrevoredo/openclaudio/internal/web"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hash-password":
			password, err := auth.ReadPasswordArgOrStdin(os.Args[2:])
			if err != nil {
				log.Fatalf("read password: %v", err)
			}
			hash, err := auth.HashPassword(password)
			if err != nil {
				log.Fatalf("hash password: %v", err)
			}
			fmt.Println(hash)
			return
		default:
			log.Fatalf("unknown command: %s", os.Args[1])
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server, err := webapp.New(cfg)
	if err != nil {
		log.Fatalf("build server: %v", err)
	}

	httpServer := &http.Server{
		Addr:              cfg.ListenAddr(),
		Handler:           server,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("openclaudio listening on http://%s", cfg.ListenAddr())
	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}
}
