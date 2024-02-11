package short

import (
	"context"
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"time"
)

type ShortDoc struct {
	LongURL  string `bson:"long_url" json:"long_url"`
	ShortURL string `bson:"short_url" json:"short_url"`
}

type Short struct {
	Short string
	Long  string

	Config ConfigBuilder.Config
	CTX    context.Context
}

func NewShort(ctx context.Context, cfg ConfigBuilder.Config) *Short {
	return &Short{
		Config: cfg,
		CTX:    ctx,
	}
}

func (s *Short) CreateURL(long string) (string, error) {
	short := generateShort()

	m := &RealMongoOperations{
		Database:   s.Config.Mongo.Database,
		Collection: s.Config.Mongo.Collections["short"],
	}
	if err := m.GetMongoClient(s.CTX, s.Config.Mongo); err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}
	defer func() {
		if err := m.Disconnect(s.CTX); err != nil {
			_ = logs.Errorf("CreateURL: %v", err)
		}
	}()

	if _, err := m.InsertOne(s.CTX, &ShortDoc{
		LongURL:  long,
		ShortURL: short,
	}); err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}

	return short, nil
}

func (s *Short) GetURL(short string) (string, error) {
	m := &RealMongoOperations{
		Database:   s.Config.Mongo.Database,
		Collection: s.Config.Mongo.Collections["short"],
	}
	if err := m.GetMongoClient(s.CTX, s.Config.Mongo); err != nil {
		return "", logs.Errorf("GetURL: %v", err)
	}
	defer func() {
		if err := m.Disconnect(s.CTX); err != nil {
			_ = logs.Errorf("GetURL: %v", err)
		}
	}()

	doc := &ShortDoc{}
	res, err := m.FindOne(s.CTX, &bson.M{"short_url": short})
	if err != nil {
		return "", logs.Errorf("GetURL: %v", err)
	}

	if err := res.Decode(doc); err != nil {
		return "", logs.Errorf("GetURL: %v", err)
	}

	return doc.LongURL, nil
}

func generateShort() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	short := make([]byte, length)
	for i := range short {
		short[i] = charset[r.Intn(len(charset))]
	}
	return string(short)
}
