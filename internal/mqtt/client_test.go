package mqtt

import (
	"context"
	"encoding/json"
	"testing"
)

func TestConnectSubscribesAndPublishesOnlineStatus(t *testing.T) {
	broker := &fakeBroker{}
	client := NewClientWithBroker(broker)

	err := client.Connect(context.Background(), Config{EdgeID: "edge-1", BrokerURL: "mqtt://localhost:1883", Username: "user", Password: "pass"})
	if err != nil {
		t.Fatal(err)
	}
	if broker.config.EdgeID != "edge-1" || broker.subscribedTopic != "spark/edge-1/commands" {
		t.Fatalf("unexpected broker state %#v", broker)
	}
	if len(broker.published) != 1 || broker.published[0].Topic != "spark/edge-1/status" || string(broker.published[0].Payload) != "online" || !broker.published[0].Retain {
		t.Fatalf("unexpected published messages %#v", broker.published)
	}
}

func TestHandleCommandPublishesDoneResponse(t *testing.T) {
	broker := &fakeBroker{}
	client := NewClientWithBroker(broker)
	if err := client.Connect(context.Background(), Config{EdgeID: "edge-1", BrokerURL: "mqtt://localhost:1883"}); err != nil {
		t.Fatal(err)
	}
	broker.published = nil

	err := client.HandleCommand(context.Background(), []byte(`{"command_id":"cmd-1","type":"ping","payload":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(broker.published) != 1 || broker.published[0].Topic != ResponseTopic("edge-1") {
		t.Fatalf("unexpected response publish %#v", broker.published)
	}
	var response map[string]any
	if err := json.Unmarshal(broker.published[0].Payload, &response); err != nil {
		t.Fatal(err)
	}
	if response["command_id"] != "cmd-1" || response["status"] != "done" {
		t.Fatalf("unexpected response %#v", response)
	}
}

func TestHandleCommandIgnoresDuplicateCommandID(t *testing.T) {
	broker := &fakeBroker{}
	client := NewClientWithBroker(broker)
	if err := client.Connect(context.Background(), Config{EdgeID: "edge-1", BrokerURL: "mqtt://localhost:1883"}); err != nil {
		t.Fatal(err)
	}
	broker.published = nil
	raw := []byte(`{"command_id":"cmd-1","type":"ping","payload":{}}`)
	if err := client.HandleCommand(context.Background(), raw); err != nil {
		t.Fatal(err)
	}
	if err := client.HandleCommand(context.Background(), raw); err != nil {
		t.Fatal(err)
	}
	if len(broker.published) != 1 {
		t.Fatalf("expected one response for duplicate command, got %#v", broker.published)
	}
}

func TestTopicsMatchTypeScriptTemplates(t *testing.T) {
	edgeID := "edge-1"
	if StatusTopic(edgeID) != "spark/edge-1/status" || HeartbeatTopic(edgeID) != "spark/edge-1/heartbeat" || CommandTopic(edgeID) != "spark/edge-1/commands" {
		t.Fatal("unexpected MQTT topic template")
	}
}

type fakeBroker struct {
	config          Config
	connected       bool
	subscribedTopic string
	published       []Message
	handler         func(string, []byte)
}

func (b *fakeBroker) Connect(ctx context.Context, config Config, handler func(string, []byte)) error {
	b.config = config
	b.handler = handler
	b.connected = true
	return ctx.Err()
}

func (b *fakeBroker) Disconnect() {
	b.connected = false
}

func (b *fakeBroker) Publish(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.published = append(b.published, message)
	return nil
}

func (b *fakeBroker) Subscribe(ctx context.Context, topic string, qos byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.subscribedTopic = topic
	return nil
}

func (b *fakeBroker) IsConnected() bool {
	return b.connected
}
