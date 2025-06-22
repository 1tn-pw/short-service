package short

import (
	"context"
	mungo "github.com/keloran/go-config/database/mongo"
	"golang.org/x/net/html"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
	"go.mongodb.org/mongo-driver/bson"
)

type DocShort struct {
	LongURL     string    `bson:"long_url" json:"long_url"`
	ShortURL    string    `bson:"short_url" json:"short_url"`
	InsertDate  time.Time `bson:"insert_date" json:"insert_date"`
	Title       string    `bson:"title" json:"title"`
	Favicon     string    `bson:"favicon" json:"favicon"`
	Description string    `bson:"description" json:"description"`
}

type Short struct {
	Short string
	Long  string

	Config ConfigBuilder.Config
	CTX    context.Context

	mungo.RealMongoOperations
}

// MongoConnection holds the persistent MongoDB connection
type MongoConnection struct {
	mungo.RealMongoOperations
	Initialized bool
}

var mongoConn = &MongoConnection{}

func NewShort(ctx context.Context, cfg ConfigBuilder.Config) *Short {
	return &Short{
		Config: cfg,
		CTX:    ctx,
	}
}

// InitMongo initializes the MongoDB connection with proper pool settings
// This should be called once at service startup
func InitMongo(cfg ConfigBuilder.Config) error {
	// Modify the RawURL to include connection pool options
	originalURL := cfg.Mongo.RawURL
	if originalURL != "" && !strings.Contains(originalURL, "maxPoolSize") {
		if strings.Contains(originalURL, "?") {
			cfg.Mongo.RawURL += "&maxPoolSize=50&minPoolSize=10&maxIdleTimeMS=300000"
		} else {
			cfg.Mongo.RawURL += "?maxPoolSize=50&minPoolSize=10&maxIdleTimeMS=300000"
		}
	}

	m := &mungo.RealMongoOperations{}
	if _, err := m.GetMongoClient(cfg.Mongo); err != nil {
		return logs.Errorf("InitMongo getClient: %v", err)
	}
	if _, err := m.GetMongoDatabase(cfg.Mongo); err != nil {
		return logs.Errorf("InitMongo getDatabase: %v", err)
	}
	if _, err := m.GetMongoCollection(cfg.Mongo, "short"); err != nil {
		return logs.Errorf("InitMongo getCollection: %v", err)
	}

	mongoConn.RealMongoOperations = *m
	mongoConn.Initialized = true
	return nil
}

// CloseMongo closes the MongoDB connection
// This should be called at service shutdown
func CloseMongo(ctx context.Context) error {
	if mongoConn.Initialized && mongoConn.Client != nil {
		return mongoConn.Client.Disconnect(ctx)
	}
	return nil
}

func (s *Short) getMongo() error {
	if !mongoConn.Initialized {
		return logs.Errorf("MongoDB connection not initialized")
	}
	s.RealMongoOperations = mongoConn.RealMongoOperations
	return nil
}

func (s *Short) CreateURL(long string) (string, error) {
	short := generateShort()
	if err := s.getMongo(); err != nil {
		return "", err
	}

	// No need to disconnect - using persistent connection pool

	shortAlready, err := s.alreadyExists(long)
	if err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}

	if shortAlready != "" {
		return shortAlready, nil
	}

	dets, err := fetchDetails(long)
	if err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}

	if _, err := s.RealMongoOperations.InsertOne(s.CTX, &DocShort{
		LongURL:     long,
		ShortURL:    short,
		InsertDate:  time.Now(),
		Title:       dets.Title,
		Favicon:     dets.ShortURL,
		Description: dets.Description,
	}); err != nil {
		return "", logs.Errorf("CreateURL: %v", err)
	}

	return short, nil
}

func (s *Short) alreadyExists(long string) (string, error) {
	doc := &DocShort{}
	res := s.RealMongoOperations.FindOne(s.CTX, &bson.M{"long_url": long})
	if res.Err() != nil {
		if res.Err().Error() == "mongo: no documents in result" {
			return "", nil
		} else {
			return "", logs.Errorf("alreadyExists: %v", res.Err())
		}
	}

	if err := res.Decode(doc); err != nil {
		return "", logs.Errorf("alreadyExists: %v", err)
	}

	return doc.ShortURL, nil
}

func (s *Short) GetURL(short string) (*DocShort, error) {
	doc := &DocShort{}
	if err := s.getMongo(); err != nil {
		return nil, err
	}

	res := s.RealMongoOperations.FindOne(s.CTX, &bson.M{"short_url": short})
	if res.Err() != nil {
		return nil, logs.Errorf("GetURL: %v", res.Err())
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

func fetchDetails(url string) (*DocShort, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, logs.Errorf("fetchTitle: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = logs.Errorf("fetchTitle: %v", err)
		}
	}()
	if res.StatusCode != 200 {
		return nil, logs.Errorf("fetchTitle: %v", res.Status)
	}

	dets, err := html.Parse(res.Body)
	if err != nil {
		return nil, logs.Errorf("fetchTitle: %v", err)
	}

	title, favicon, description := extractDetails(dets)
	return &DocShort{
		Title:       title,
		ShortURL:    favicon,
		Description: description,
	}, nil
}

func extractDetails(n *html.Node) (title, favicon, description string) {
	var crawler func(*html.Node)
	crawler = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
		}
		if n.Type == html.ElementNode && n.Data == "link" {
			for _, a := range n.Attr {
				if a.Key == "rel" && (a.Val == "icon" || a.Val == "shortcut icon") {
					for _, a := range n.Attr {
						if a.Key == "href" {
							favicon = a.Val
							break
						}
					}
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "description" {
					name = a.Val
				} else if a.Key == "content" {
					content = a.Val
				}
			}
			if name == "description" {
				description = content
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			crawler(c)
		}
	}

	crawler(n)
	return
}
