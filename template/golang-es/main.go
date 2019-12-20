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
	"github.com/contextgg/go-sdk/secrets"

	"handler/function"
	// "github.com/contextgg/openfaas-templates/template/golang-es/function"
)

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

func setupMongodb(build builder.ClientBuilder, uri, db string, snapshot int) {
	if len(uri) == 0 || len(db) == 0 {
		build.SetEventStore(
			builder.LocalStore(),
		)
		return
	}

	build.SetEventStore(
		builder.Mongo(uri, db, snapshot),
	)
}
func setupNats(build builder.ClientBuilder, natsURI, natsNS string) {
	if len(natsURI) == 0 || len(natsNS) == 0 {
		return
	}

	build.AddPublisher(
		builder.Nats(natsURI, natsNS),
	)
}

func main() {
	readTimeout := parseIntOrDurationValue(os.Getenv("read_timeout"), 10*time.Second)
	writeTimeout := parseIntOrDurationValue(os.Getenv("write_timeout"), 10*time.Second)

	mongodbURI := secrets.MustReadSecret("mongodb_uri", "")
	mongodbDB := secrets.MustReadSecret("mongodb_db", "")
	snapshotMin := parseInt(os.Getenv("snapshot_min"), -1)
	natsURI := secrets.MustReadSecret("nats_uri", "")
	natsNS := secrets.MustReadSecret("nats_namespace", "")

	b := builder.NewClientBuilder()
	setupMongodb(b, mongodbURI, mongodbDB, snapshotMin)
	setupNats(b, natsURI, natsNS)
	function.Setup(b)

	cli, err := b.Build()
	if err != nil {
		log.Fatalf("Setup failed %v", err)
		return
	}

	s := &http.Server{
		Handler:        httputils.CommandHandler(cli.CommandBus),
		Addr:           fmt.Sprintf(":%d", 8082),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	log.Fatal(s.ListenAndServe())
}
