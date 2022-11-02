package main

import (
	"context"

	"function"

	"github.com/contextcloud/graceful"
	"github.com/contextcloud/graceful/config"
	"github.com/contextcloud/graceful/srv"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.NewConfig(ctx)
	if err != nil {
		panic(err)
	}

	handler, err := function.NewHandler(ctx, cfg)
	if err != nil {
		panic(err)
	}

	startable, err := srv.NewStartable(cfg.SrvAddr, handler)
	if err != nil {
		panic(err)
	}

	tracer, err := srv.NewTracer(ctx, cfg)
	if err != nil {
		panic(err)
	}

	multi := srv.NewMulti(
		tracer,
		srv.NewMetricsServer(cfg.MetricsAddr),
		srv.NewHealth(cfg.HealthAddr),
		startable,
	)

	// graceful?
	graceful.Run(ctx, multi)
	cancel()

	<-ctx.Done()
}
