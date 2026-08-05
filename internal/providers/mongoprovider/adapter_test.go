package mongoprovider

import (
	"context"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func TestAdapterInsertOne(t *testing.T) {
	fake := &fakeSession{collections: map[string]*fakeCollection{"logs": {}}}
	adapter := newTestAdapter(t, "insertOne", fake)

	if err := adapter.Send(context.Background(), map[string]any{"temperature": float64(42)}); err != nil {
		t.Fatal(err)
	}
	inserted := fake.collections["logs"].inserted
	if inserted["temperature"] != float64(42) {
		t.Fatalf("unexpected insert payload %#v", inserted)
	}
	if !fake.disconnected {
		t.Fatal("expected session to disconnect after send")
	}
}

func TestAdapterUpdateOneUsesFilterAndUpdate(t *testing.T) {
	fake := &fakeSession{collections: map[string]*fakeCollection{"logs": {}}}
	adapter := newTestAdapter(t, "updateOne", fake)

	payload := map[string]any{
		"filter": map[string]any{"device": "edge-1"},
		"update": map[string]any{"$set": map[string]any{"temperature": float64(42)}},
	}
	if err := adapter.Send(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	collection := fake.collections["logs"]
	if collection.filter["device"] != "edge-1" {
		t.Fatalf("unexpected filter %#v", collection.filter)
	}
	if collection.update["$set"] == nil {
		t.Fatalf("unexpected update %#v", collection.update)
	}
}

func TestAdapterDiscoverListsCollections(t *testing.T) {
	fake := &fakeSession{collectionNames: []string{"logs", "events"}, collections: map[string]*fakeCollection{}}
	adapter := newTestAdapter(t, "insertOne", fake)

	resources, err := adapter.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 2 || resources[0].Name != "logs" || resources[1].Config["collection"] != "events" {
		t.Fatalf("unexpected resources %#v", resources)
	}
}

func TestAdapterRequiresCredentials(t *testing.T) {
	_, err := New(providers.Config{})
	if err == nil {
		t.Fatal("expected missing credentials error")
	}
}

func newTestAdapter(t *testing.T, operation string, fake *fakeSession) *Adapter {
	t.Helper()
	adapter, err := newWithConnector(providers.Config{
		Resource:    map[string]any{"config": map[string]any{"collection": "logs"}},
		Operation:   map[string]any{"config": map[string]any{"operation": operation}},
		Credentials: map[string]any{"data": map[string]any{"uri": "mongodb://localhost:27017", "database": "sparkedge"}},
	}, func(context.Context, string, string) (session, error) {
		return fake, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

type fakeSession struct {
	collections     map[string]*fakeCollection
	collectionNames []string
	disconnected    bool
}

func (s *fakeSession) Collection(name string) collection {
	if s.collections == nil {
		s.collections = map[string]*fakeCollection{}
	}
	if s.collections[name] == nil {
		s.collections[name] = &fakeCollection{}
	}
	return s.collections[name]
}

func (s *fakeSession) ListCollections(context.Context) ([]string, error) {
	return s.collectionNames, nil
}

func (s *fakeSession) Ping(context.Context) error {
	return nil
}

func (s *fakeSession) Disconnect(context.Context) error {
	s.disconnected = true
	return nil
}

type fakeCollection struct {
	inserted map[string]any
	filter   map[string]any
	update   map[string]any
}

func (c *fakeCollection) InsertOne(_ context.Context, payload map[string]any) error {
	c.inserted = payload
	return nil
}

func (c *fakeCollection) UpdateOne(_ context.Context, filter map[string]any, update map[string]any) error {
	c.filter = filter
	c.update = update
	return nil
}

func (c *fakeCollection) Find(context.Context, map[string]any) ([]map[string]any, error) {
	return []map[string]any{}, nil
}

func (c *fakeCollection) FindOne(context.Context, map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}
