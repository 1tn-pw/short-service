package short

import (
	"context"
	"golang.org/x/net/html"
	"math/rand"
	"net/http"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
	"go.mongodb.org/mongo-driver/bson"
)

type DocShort struct {
	LongURL    string    `bson:"long_url" json:"long_url"`
	ShortURL   string    `bson:"short_url" json:"short_url"`
	InsertDate time.Time `bson:"insert_date" json:"insert_date"`
	Title      string    `bson:"title" json:"title"`
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

	shortAlready, err := s.alreadyExists(long)
	if err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}

	if shortAlready != "" {
		return shortAlready, nil
	}

	title, err := fetchTitle(long)
	if err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}

	if _, err := m.InsertOne(s.CTX, &DocShort{
		LongURL:    long,
		ShortURL:   short,
		InsertDate: time.Now(),
		Title:      title,
	}); err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}

	return short, nil
}

func (s *Short) alreadyExists(long string) (string, error) {
	m := &RealMongoOperations{
		Database:   s.Config.Mongo.Database,
		Collection: s.Config.Mongo.Collections["short"],
	}
	if err := m.GetMongoClient(s.CTX, s.Config.Mongo); err != nil {
		return "", logs.Errorf("alreadyExists: %v", err)
	}
	defer func() {
		if err := m.Disconnect(s.CTX); err != nil {
			_ = logs.Errorf("alreadyExists: %v", err)
		}
	}()

	doc := &DocShort{}
	res, err := m.FindOne(s.CTX, &bson.M{"long_url": long})
	if err != nil {
		if err.Error() == "FindOne: mongo: no documents in result" {
			return "", nil
		} else {
			return "", logs.Errorf("alreadyExists: %v", err)
		}
	}

	if err := res.Decode(doc); err != nil {
		return "", logs.Errorf("alreadyExists: %v", err)
	}

	return doc.ShortURL, nil
}

func (s *Short) GetURL(short string) (*DocShort, error) {
	m := &RealMongoOperations{
		Database:   s.Config.Mongo.Database,
		Collection: s.Config.Mongo.Collections["short"],
	}
	if err := m.GetMongoClient(s.CTX, s.Config.Mongo); err != nil {
		return nil, logs.Errorf("GetURL: %v", err)
	}
	defer func() {
		if err := m.Disconnect(s.CTX); err != nil {
			_ = logs.Errorf("GetURL: %v", err)
		}
	}()

	doc := &DocShort{}
	res, err := m.FindOne(s.CTX, &bson.M{"short_url": short})
	if err != nil {
		return nil, logs.Errorf("GetURL: %v", err)
	}

	if err := res.Decode(doc); err != nil {
		return nil, logs.Errorf("GetURL: %v", err)
	}

	return doc, nil
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

func fetchTitle(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", logs.Errorf("fetchTitle: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = logs.Errorf("fetchTitle: %v", err)
		}
	}()
	if res.StatusCode != 200 {
		return "", logs.Errorf("fetchTitle: %v", res.Status)
	}

	toks := html.NewTokenizer(res.Body)
	for {
		tt := toks.Next()
		if tt == html.ErrorToken {
			return "", logs.Errorf("fetchTitle: %v", toks.Err())
		}
		if tt == html.StartTagToken {
			t := toks.Token()
			if t.Data == "title" {
				toks.Next()
				return toks.Token().Data, nil
			}
		}
	}
}
