package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"

	"function"
)

const (
	srvAddr     = ":8080"
	metricsAddr = ":8081"
	healthAddr  = ":8082"
)

type Startable interface {
	Start() error
	Shutdown()
}

type startable struct {
	h http.Handler
}

func (s *startable) Start() error {
	// Serve our handler.
	go func() {
		if err := http.ListenAndServe(srvAddr, s.h); err != nil {
			log.Panicf("error while serving: %s", err)
		}
	}()

	return nil
}
func (s *startable) Shutdown() {
}

func handler(h http.Handler) Startable {
	// Create our middleware.
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	// Wrap our main handler, we pass empty handler ID so the middleware
	// the handler label from the URL.
	h = std.Handler("", mdlw, h)

	return &startable{h: h}
}

func getStartable() Startable {
	// Create our server.
	h := function.NewService()

	switch start := h.(type) {
	case Startable:
		return start
	case http.Handler:
		return handler(start)
	default:
		panic(fmt.Sprintf("unknown service type: %T", h))
	}
}

func main() {
	// run stuff.
	startable := getStartable()

	// start it.
	if err := startable.Start(); err != nil {
		panic(err)
	}

	// Serve our metrics.
	go func() {
		if err := http.ListenAndServe(metricsAddr, promhttp.Handler()); err != nil {
			log.Panicf("error while serving metrics: %s", err)
		}
	}()

	// Serve our health.
	go func() {
		health := healthcheck.NewHandler()
		health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))

		if err := http.ListenAndServe(healthAddr, health); err != nil {
			log.Panicf("error while serving health: %s", err)
		}
	}()

	// Wait until some signal is captured.
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC

	startable.Shutdown()
}
