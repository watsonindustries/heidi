package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/watsonindustries/heidi/boetea"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoStore struct {
	client   *mongo.Client
	database *mongo.Database

	*artworkStore
}

// Create a new Mongo store with a context, URI string and db name.
func New(ctx context.Context, uri string, db string) (boetea.Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	database := client.Database(db)
	return &mongoStore{
		client:       client,
		database:     database,
		artworkStore: &artworkStore{client, database, database.Collection("artworks")},
	}, nil
}

func (m *mongoStore) Init(ctx context.Context) error {
	collections := []string{"artworks"}
	for _, col := range collections {
		err := m.database.CreateCollection(ctx, col)
		if err != nil && !errors.As(err, &mongo.CommandError{}) {
			return err
		}
	}

	return nil

}

func (m *mongoStore) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
