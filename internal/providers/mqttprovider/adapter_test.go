package mqttprovider

import (
	"context"
	"strings"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func TestNewRequiresBrokerURL(t *testing.T) {
	_, err := New(providers.Config{})
	if err == nil {
		t.Fatal("expected missing broker url error")
	}
	if !strings.Contains(err.Error(), "brokerUrl") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiscoverReturnsTopicResource(t *testing.T) {
	adapter, err := New(providers.Config{
		Credentials: map[string]any{"data": map[string]any{"brokerUrl": "tcp://localhost:1883"}},
		Resource:    map[string]any{"config": map[string]any{"topic": "devices/telemetry"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	resources, err := adapter.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected one resource, got %#v", resources)
	}
	if resources[0].Config["topic"] != "devices/telemetry" {
		t.Fatalf("unexpected resource config %#v", resources[0].Config)
	}
}
