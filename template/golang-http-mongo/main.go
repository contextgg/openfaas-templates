package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/contextgg/go-sdk/secrets"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"handler/function"
	// "github.com/contextgg/openfaas-templates/template/golang-http-mongo/function"
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

func getDatabase(uri, db, username, password string) (*mongo.Database, error) {
	opts := options.
		Client().
		ApplyURI(uri)

	if len(username) > 0 {
		creds := options.Credential{
			Username: username,
			Password: password,
		}
		opts = opts.SetAuth(creds)
	}

	var err error
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// test it!
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client.
		Database(db), nil
}

func main() {
	readTimeout := parseIntOrDurationValue(os.Getenv("read_timeout"), 10*time.Second)
	writeTimeout := parseIntOrDurationValue(os.Getenv("write_timeout"), 10*time.Second)

	mongodbURI := secrets.MustReadSecret("mongodb_uri", "")
	mongodbDB := secrets.MustReadSecret("mongodb_db", "")
	mongodbUsername := secrets.MustReadSecret("mongodb_username", "")
	mongodbPassword := secrets.MustReadSecret("mongodb_password", "")

	db, err := getDatabase(mongodbURI, mongodbDB, mongodbUsername, mongodbPassword)
	if err != nil {
		log.Fatalf("Could not create database: %v", err)
		return
	}

	h := function.NewHandler(db)
	s := &http.Server{
		Handler:        h,
		Addr:           fmt.Sprintf(":%d", 8082),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}
	log.Fatal(s.ListenAndServe())
}
