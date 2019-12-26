package mongo

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	// AggregatesCollection for storing information regarding aggregates
	AggregatesCollection = "aggregates"
	// EventsCollection for storing events
	EventsCollection = "events"
)

// Create will setup a database
func Create(uri, db, username, password string, createIndexes bool) (*mongo.Database, error) {
	sublogger := log.With().
		Str("uri", uri).
		Str("db", db).
		Str("username", username).
		Bool("createIndexes", createIndexes).
		Logger()

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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		sublogger.
			Error().
			Err(err).
			Msg("Could not connect to db")
		return nil, err
	}

	// test it!
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		sublogger.
			Error().
			Err(err).
			Msg("Could not ping db")
		return nil, err
	}

	database := client.
		Database(db)

	if createIndexes {
		indexOpts := options.
			CreateIndexes().
			SetMaxTime(10 * time.Second)

		aggregatesIndex := mongo.IndexModel{
			Keys: bson.M{
				"aggregate_type": 1,
				"aggregate_id":   1,
			},
			Options: options.
				Index().
				SetUnique(true).
				SetName("aggreates.id.type"),
		}
		eventsIndex := mongo.IndexModel{
			Keys: bson.M{
				"aggregate_type": 1,
				"aggregate_id":   1,
				"version":        1,
			},
			Options: options.
				Index().
				SetUnique(true).
				SetName("events.id.type.version"),
		}

		if _, err := database.
			Collection(AggregatesCollection).
			Indexes().
			CreateOne(ctx, aggregatesIndex, indexOpts); err != nil {
			sublogger.
				Error().
				Err(err).
				Msg("Could not create aggregates index")
			return nil, err
		}
		if _, err := database.
			Collection(EventsCollection).
			Indexes().
			CreateOne(ctx, eventsIndex, indexOpts); err != nil {
			sublogger.
				Error().
				Err(err).
				Msg("Could not create events index")
			return nil, err
		}

		log.Debug().
			Msg("Indexes created successfully")
	}

	log.Debug().
		Msg("Database created successfully")
	return database, nil
}
