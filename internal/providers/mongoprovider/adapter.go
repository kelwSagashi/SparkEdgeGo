package mongoprovider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Adapter struct {
	uri        string
	database   string
	resource   map[string]any
	operation  map[string]any
	connect    connector
	disconnect bool
}

type connector func(context.Context, string, string) (session, error)

type session interface {
	Collection(string) collection
	ListCollections(context.Context) ([]string, error)
	Ping(context.Context) error
	Disconnect(context.Context) error
}

type collection interface {
	InsertOne(context.Context, map[string]any) error
	UpdateOne(context.Context, map[string]any, map[string]any) error
	Find(context.Context, map[string]any) ([]map[string]any, error)
	FindOne(context.Context, map[string]any) (map[string]any, error)
}

func Register(registry *providers.Registry) {
	if registry == nil {
		return
	}
	registry.Register(StrategyID, func(config providers.Config) (providers.Adapter, error) {
		return New(config)
	})
}

func New(config providers.Config) (*Adapter, error) {
	return newWithConnector(config, connectMongo)
}

func newWithConnector(config providers.Config, connect connector) (*Adapter, error) {
	credentials := mapValue(config.Credentials, "data")
	uri := stringValue(credentials, "uri")
	database := stringValue(credentials, "database", "dbName", "db_name")
	if uri == "" || database == "" {
		return nil, errors.New("mongo provider requires credentials.data.uri and credentials.data.database")
	}
	return &Adapter{
		uri:        uri,
		database:   database,
		resource:   mapValue(config.Resource, "config"),
		operation:  mapValue(config.Operation, "config"),
		connect:    connect,
		disconnect: true,
	}, nil
}

func (a *Adapter) Send(ctx context.Context, payload map[string]any) error {
	session, err := a.open(ctx)
	if err != nil {
		return err
	}
	defer a.close(ctx, session)

	collectionName := stringValue(a.resource, "collection")
	if collectionName == "" {
		return errors.New("mongo provider requires resource.config.collection")
	}
	op := stringValue(a.operation, "operation", "type")
	if op == "" {
		op = "insertOne"
	}
	coll := session.Collection(collectionName)
	switch op {
	case "insertOne":
		return coll.InsertOne(ctx, payload)
	case "updateOne":
		filter, _ := payload["filter"].(map[string]any)
		update, _ := payload["update"].(map[string]any)
		if filter == nil {
			filter = map[string]any{}
		}
		if update == nil {
			update = map[string]any{"$set": payload}
		}
		return coll.UpdateOne(ctx, filter, update)
	case "find":
		_, err := coll.Find(ctx, payload)
		return err
	case "findOne":
		_, err := coll.FindOne(ctx, payload)
		return err
	default:
		return fmt.Errorf("unsupported mongo operation %q", op)
	}
}

func (a *Adapter) Test(ctx context.Context, payload map[string]any) error {
	session, err := a.open(ctx)
	if err != nil {
		return err
	}
	defer a.close(ctx, session)
	return session.Ping(ctx)
}

func (a *Adapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	session, err := a.open(ctx)
	if err != nil {
		return nil, err
	}
	defer a.close(ctx, session)

	collections, err := session.ListCollections(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]providers.Resource, 0, len(collections))
	for _, name := range collections {
		result = append(result, providers.Resource{
			Name:   name,
			Type:   "collection",
			Config: map[string]any{"collection": name},
			Fields: []providers.Field{{Name: "collection", Type: "text"}},
		})
	}
	return result, nil
}

func (a *Adapter) open(ctx context.Context) (session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return a.connect(ctx, a.uri, a.database)
}

func (a *Adapter) close(ctx context.Context, session session) {
	if a.disconnect && session != nil {
		_ = session.Disconnect(ctx)
	}
}

type realSession struct {
	client *mongo.Client
	db     *mongo.Database
}

type realCollection struct {
	collection *mongo.Collection
}

func connectMongo(ctx context.Context, uri string, database string) (session, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	result := &realSession{client: client, db: client.Database(database)}
	if err := result.Ping(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return result, nil
}

func (s *realSession) Collection(name string) collection {
	return &realCollection{collection: s.db.Collection(name)}
}

func (s *realSession) ListCollections(ctx context.Context) ([]string, error) {
	cursor, err := s.db.ListCollections(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []struct {
		Name string `bson:"name"`
	}
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.Name)
	}
	return result, nil
}

func (s *realSession) Ping(ctx context.Context) error {
	return s.client.Ping(ctx, readpref.Primary())
}

func (s *realSession) Disconnect(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

func (c *realCollection) InsertOne(ctx context.Context, payload map[string]any) error {
	_, err := c.collection.InsertOne(ctx, bson.M(payload))
	return err
}

func (c *realCollection) UpdateOne(ctx context.Context, filter map[string]any, update map[string]any) error {
	_, err := c.collection.UpdateOne(ctx, bson.M(filter), bson.M(update), options.UpdateOne().SetUpsert(true))
	return err
}

func (c *realCollection) Find(ctx context.Context, filter map[string]any) ([]map[string]any, error) {
	cursor, err := c.collection.Find(ctx, bson.M(filter))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var result []map[string]any
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *realCollection) FindOne(ctx context.Context, filter map[string]any) (map[string]any, error) {
	var result map[string]any
	err := c.collection.FindOne(ctx, bson.M(filter)).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return map[string]any{}, nil
	}
	return result, err
}

func mapValue(source map[string]any, key string) map[string]any {
	if source == nil {
		return map[string]any{}
	}
	if value, ok := source[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func stringValue(source map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := source[key]; ok {
			switch typed := value.(type) {
			case string:
				return strings.TrimSpace(typed)
			case fmt.Stringer:
				return strings.TrimSpace(typed.String())
			}
		}
	}
	return ""
}
