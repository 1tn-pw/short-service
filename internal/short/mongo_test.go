package short

import (
	"context"
	"github.com/bugfixes/go-bugfixes/logs"
	mungo "github.com/keloran/go-config/mongo"
)

type MockMongoOperations struct {
	shouldError bool
	exists      bool
}

func (m *MockMongoOperations) GetMongoClient(ctx context.Context, cfg mungo.Mongo) error {
	if m.shouldError {
		return logs.Errorf("GetMongoClient: error")
	}

	return nil
}

func (m *MockMongoOperations) Disconnect(ctx context.Context) error {
	if m.shouldError {
		return logs.Errorf("Disconnect: error")
	}

	return nil
}

func (m *MockMongoOperations) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, logs.Errorf("InsertOne: error")
	}

	return "123", nil
}

func (m *MockMongoOperations) FindOne(ctx context.Context, filter interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, logs.Errorf("FindOne: error")
	}
	if m.exists {
		return &DocShort{
			LongURL:  "https://example.com",
			ShortURL: "https://short.example.com",
			Title:    "Example",
		}, nil
	}

	return nil, nil
}
