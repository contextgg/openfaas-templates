package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/contextgg/go-es/builder"
	"github.com/contextgg/go-es/httputils"
	"github.com/contextgg/go-sdk/hydra"
	"github.com/contextgg/go-sdk/secrets"

	"handler/function"
	// "github.com/contextgg/openfaas-templates/template/golang-es/function"
)

// Middleware used to help auth
type Middleware func(http.Handler) http.Handler

// UseHandler wraps a CommandHandler in one or more middleware.
func UseHandler(h http.Handler, middleware ...Middleware) http.Handler {
	// Apply in reverse order.
	for i := len(middleware) - 1; i >= 0; i-- {
		m := middleware[i]
		h = m(h)
	}
	return h
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}
	return duration
}
func parseInt(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil {
			return parsedVal
		}
	}
	return fallback
}

func makeStoreFactory(uri, db string, snapshot int) builder.AggregateStoreFactory {
	if len(uri) == 0 || len(db) == 0 {
		return builder.LocalStore()
	}
	return builder.Mongo(uri, db, snapshot)
}
func setupNats(build builder.ClientBuilder, natsURI, natsNS string) {
}

func main() {
	readTimeout := parseIntOrDurationValue(os.Getenv("read_timeout"), 10*time.Second)
	writeTimeout := parseIntOrDurationValue(os.Getenv("write_timeout"), 10*time.Second)

	mongodbURI := secrets.MustReadSecret("mongodb_uri", "")
	mongodbDB := secrets.MustReadSecret("mongodb_db", "")
	snapshotMin := parseInt(os.Getenv("snapshot_min"), -1)
	natsURI := secrets.MustReadSecret("nats_uri", "")
	natsNS := secrets.MustReadSecret("nats_namespace", "")

	creds := secrets.LoadBasicAuth("auth")
	hydraURL := secrets.MustReadSecret("hydra_url", "")

	middleware := []Middleware{}
	if creds != nil {
		middleware = append(middleware, secrets.AuthHandlerOptional())
	}
	if len(hydraURL) > 0 {
		middleware = append(middleware, hydra.AuthHandlerOptional(hydraURL))
	}

	storeFactory := makeStoreFactory(mongodbURI, mongodbDB, snapshotMin)
	b, err := builder.NewClientBuilder(storeFactory)
	if err != nil {
		log.Fatalf("NewClientBuilder failed %v", err)
		return
	}

	if len(natsURI) != 0 && len(natsNS) != 0 {
		b.AddPublisher(
			builder.Nats(natsURI, natsNS),
		)
	}

	function.Setup(b)

	cli, err := b.Build()
	if err != nil {
		log.Fatalf("Build failed %v", err)
		return
	}

	s := &http.Server{
		Handler:        UseHandler(httputils.CombinedHandler(cli.EventRegistry, cli.EventHandler, cli.CommandBus), middleware...),
		Addr:           fmt.Sprintf(":%d", 8082),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	log.Fatal(s.ListenAndServe())
}
