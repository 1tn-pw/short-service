package short

import (
	"context"
	"github.com/bugfixes/go-bugfixes/logs"
	mungo "github.com/keloran/go-config/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoOperations interface {
	GetMongoClient(ctx context.Context, cfg mungo.Mongo) error
	Disconnect(ctx context.Context) error
	InsertOne(ctx context.Context, document interface{}) (interface{}, error)
	FindOne(ctx context.Context, filter interface{}) (*mongo.SingleResult, error)
}

type RealMongoOperations struct {
	Client     *mongo.Client
	Collection string
	Database   string
}

func (r *RealMongoOperations) GetMongoClient(ctx context.Context, cfg mungo.Mongo) error {
	client, err := mungo.GetMongoClient(ctx, cfg)
	if err != nil {
		return logs.Errorf("GetMongoClient: %v", err)
	}

	r.Client = client

	return nil
}

func (r *RealMongoOperations) Disconnect(ctx context.Context) error {
	return r.Client.Disconnect(ctx)
}

func (r *RealMongoOperations) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	collection := r.Client.Database(r.Database).Collection(r.Collection)
	res, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, logs.Errorf("InsertOne: %v", err)
	}

	return res.InsertedID, nil
}

func (r *RealMongoOperations) FindOne(ctx context.Context, filter interface{}) (*mongo.SingleResult, error) {
	collection := r.Client.Database(r.Database).Collection(r.Collection)
	res := collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, logs.Errorf("FindOne: %v", res.Err())
	}

	return res, nil
}
