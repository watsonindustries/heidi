package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/watsonindustries/heidi/boetea"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type artworkStore struct {
	client *mongo.Client
	db     *mongo.Database
	col    *mongo.Collection
}

func ArtworkStore(client *mongo.Client, database, collection string) boetea.ArtworkStore {
	db := client.Database(database)
	col := db.Collection(collection)

	return &artworkStore{
		client: client,
		db:     db,
		col:    col,
	}
}

func (a *artworkStore) Artwork(ctx context.Context, id int, url string) (*boetea.Artwork, error) {
	filter := bson.M{}
	if id != 0 {
		filter["artwork_id"] = id
	}

	if url != "" {
		filter["url"] = url
	}

	res := a.col.FindOne(ctx, filter)

	artwork := &boetea.Artwork{}
	if err := res.Decode(artwork); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, boetea.ErrArtworkNotFound
		}

		return nil, fmt.Errorf("failed to decode an artwork: %w", err)
	}

	return artwork, nil
}

func (a *artworkStore) SearchArtworks(ctx context.Context, filter boetea.ArtworkFilter, opts ...boetea.ArtworkSearchOptions) ([]*boetea.Artwork, error) {
	opt := boetea.DefaultSearchOptions()
	if len(opts) != 0 {
		opt = opts[0]
	}

	cur, err := a.col.Find(ctx, filterBSON(filter), findOptions(opt))
	if err != nil {
		return nil, err
	}

	artworks := make([]*boetea.Artwork, 0)
	err = cur.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}

	return artworks, nil
}

func findOptions(a boetea.ArtworkSearchOptions) *options.FindOptions {
	skip := a.Limit * a.Page
	sort := bson.M{a.Sort.String(): a.Order}

	return options.Find().SetLimit(a.Limit).SetSkip(skip).SetSort(sort)
}

func filterBSON(f boetea.ArtworkFilter) bson.D {
	filter := bson.D{}

	regex := func(key, value string) bson.E {
		return bson.E{Key: key, Value: bson.D{{Key: "$regex", Value: ".*" + value + ".*"}, {Key: "$options", Value: "i"}}}
	}

	regexM := func(key, value string) bson.M {
		return bson.M{key: bson.D{{Key: "$regex", Value: ".*" + value + ".*"}, {Key: "$options", Value: "i"}}}
	}

	switch {
	case len(f.IDs) != 0:
		filter = append(filter, bson.E{Key: "artwork_id", Value: bson.M{"$in": f.IDs}})
	case f.URL != "":
		filter = append(filter, bson.E{Key: "url", Value: f.URL})
	case f.Query != "":
		filter = bson.D{
			{Key: "$or", Value: []bson.M{regexM("author", f.Query), regexM("title", f.Query)}},
		}
	default:
		if f.Author != "" {
			filter = append(filter, regex("author", f.Author))
		}

		if f.Title != "" {
			filter = append(filter, regex("title", f.Title))
		}

		if f.Time != 0 {
			filter = append(filter, bson.E{Key: "created_at", Value: bson.M{"$gte": time.Now().Add(-f.Time)}})
		}
	}

	return filter
}
