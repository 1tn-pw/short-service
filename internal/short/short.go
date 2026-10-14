package short

import (
	"context"
	"fmt"
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

// CreateURL shortens long. A title from the caller is used as given, with its description, and the
// page isn't read; without one the page's own details are fetched.
func (s *Short) CreateURL(long, title, description string) (string, error) {
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

	dets := details(long, title, description)

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

// details are the caller's when it gave a title, so a page it already knows (or one we couldn't
// read) isn't fetched; otherwise they're the page's own.
func details(long, title, description string) *DocShort {
	if title != "" {
		return &DocShort{Title: title, Description: description}
	}
	return detailsOrNone(long)
}

// detailsOrNone is a page's details, or none: a page that can't be read (one behind a sign-in
// answers 403) still gets its short link, just without a title, icon or description.
func detailsOrNone(long string) *DocShort {
	dets, err := fetchDetails(long)
	if err != nil {
		logs.Infof("no details for %s: %v", long, err)
		return &DocShort{}
	}
	return dets
}

// detailsClient reads pages for their details; a page that never answers mustn't hold up a link.
var detailsClient = &http.Client{Timeout: 10 * time.Second}

// fetchDetails reads a page's title, icon and description. Failing is expected (private pages), so
// it's an error for the caller to shrug off, not one to report.
func fetchDetails(url string) (*DocShort, error) {
	res, err := detailsClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetchTitle: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("fetchTitle: %v", res.Status)
	}

	dets, err := html.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("fetchTitle: %w", err)
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
