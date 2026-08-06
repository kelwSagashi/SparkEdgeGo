package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
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

func TestHandleCommandPersistsLifecycleInStore(t *testing.T) {
	broker := &fakeBroker{}
	commands := &fakeCommandStore{commands: map[string]domain.MqttCommand{}}
	client := NewClientWithBroker(broker)
	client.UseStores(commands, nil)
	if err := client.Connect(context.Background(), Config{EdgeID: "edge-1", BrokerURL: "mqtt://localhost:1883"}); err != nil {
		t.Fatal(err)
	}
	broker.published = nil

	if err := client.HandleCommand(context.Background(), []byte(`{"command_id":"cmd-1","type":"ping","payload":{}}`)); err != nil {
		t.Fatal(err)
	}
	if commands.saved != 1 || commands.statuses[0] != domain.MqttCommandRunning || commands.statuses[1] != domain.MqttCommandDone {
		t.Fatalf("unexpected command store %#v", commands)
	}
}

func TestRetryQueueDeletesDeliveredMessages(t *testing.T) {
	broker := &fakeBroker{connected: true}
	queue := &fakeQueueStore{items: []domain.MqttQueueItem{{ID: "queue-1", Topic: "spark/edge-1/logs", Payload: `{"ok":true}`}}}
	client := NewClientWithBroker(broker)
	client.UseStores(nil, queue)

	if err := client.RetryQueue(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if len(broker.published) != 1 || queue.deletedID != "queue-1" {
		t.Fatalf("unexpected retry state broker=%#v queue=%#v", broker, queue)
	}
}

func TestPublishJSONQueuesWhenOffline(t *testing.T) {
	broker := &fakeBroker{connected: false}
	queue := &fakeQueueStore{}
	client := NewClientWithBroker(broker)
	client.UseStores(nil, queue)
	client.config = Config{EdgeID: "edge-1"}

	if err := client.PublishLog(context.Background(), "info", "offline message"); err != nil {
		t.Fatal(err)
	}
	if len(queue.items) != 1 || queue.items[0].Topic != LogTopic("edge-1") || len(broker.published) != 0 {
		t.Fatalf("expected queued offline publish, broker=%#v queue=%#v", broker.published, queue.items)
	}
}

func TestPublishJSONQueuesOnPublishFailure(t *testing.T) {
	broker := &fakeBroker{connected: true, publishErr: errors.New("broker down")}
	queue := &fakeQueueStore{}
	client := NewClientWithBroker(broker)
	client.UseStores(nil, queue)
	client.config = Config{EdgeID: "edge-1"}

	if err := client.PublishLog(context.Background(), "info", "failed message"); err == nil {
		t.Fatal("expected publish error")
	}
	if len(queue.items) != 1 || queue.items[0].Topic != LogTopic("edge-1") {
		t.Fatalf("expected queued failed publish, queue=%#v", queue.items)
	}
}

func TestTopicsMatchTypeScriptTemplates(t *testing.T) {
	edgeID := "edge-1"
	if StatusTopic(edgeID) != "spark/edge-1/status" || HeartbeatTopic(edgeID) != "spark/edge-1/heartbeat" || CommandTopic(edgeID) != "spark/edge-1/commands" {
		t.Fatal("unexpected MQTT topic template")
	}
}

type fakeCommandStore struct {
	commands map[string]domain.MqttCommand
	saved    int
	statuses []domain.MqttCommandStatus
}

func (s *fakeCommandStore) FindByCommandID(ctx context.Context, commandID string) (domain.MqttCommand, error) {
	if command, ok := s.commands[commandID]; ok {
		return command, ctx.Err()
	}
	return domain.MqttCommand{}, errors.New("not found")
}

func (s *fakeCommandStore) Save(ctx context.Context, commandID string, commandType string, payload map[string]any) (domain.MqttCommand, error) {
	command := domain.MqttCommand{CommandID: commandID, Type: commandType, Payload: payload, Status: domain.MqttCommandPending}
	s.commands[commandID] = command
	s.saved++
	return command, ctx.Err()
}

func (s *fakeCommandStore) UpdateStatus(ctx context.Context, commandID string, status domain.MqttCommandStatus, result map[string]any, errText string) (domain.MqttCommand, error) {
	command := s.commands[commandID]
	command.Status = status
	command.Result = result
	command.Error = errText
	s.commands[commandID] = command
	s.statuses = append(s.statuses, status)
	return command, ctx.Err()
}

type fakeQueueStore struct {
	items     []domain.MqttQueueItem
	deletedID string
}

func (s *fakeQueueStore) Enqueue(ctx context.Context, topic string, payload string) (domain.MqttQueueItem, error) {
	item := domain.MqttQueueItem{ID: "queue-new", Topic: topic, Payload: payload}
	s.items = append(s.items, item)
	return item, ctx.Err()
}

func (s *fakeQueueStore) ListPending(ctx context.Context, maxAttempts int) ([]domain.MqttQueueItem, error) {
	return s.items, ctx.Err()
}

func (s *fakeQueueStore) Delete(ctx context.Context, id string) error {
	s.deletedID = id
	return ctx.Err()
}

func (s *fakeQueueStore) IncrementAttempt(ctx context.Context, id string) error {
	return ctx.Err()
}

type fakeBroker struct {
	config          Config
	connected       bool
	subscribedTopic string
	published       []Message
	handler         func(string, []byte)
	publishErr      error
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
	if b.publishErr != nil {
		return b.publishErr
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
