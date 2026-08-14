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

func TestDisconnectPublishesOfflineStatus(t *testing.T) {
	broker := &fakeBroker{}
	client := NewClientWithBroker(broker)

	if err := client.Connect(context.Background(), Config{EdgeID: "edge-1", BrokerURL: "mqtt://localhost:1883"}); err != nil {
		t.Fatal(err)
	}
	broker.published = nil

	if err := client.Disconnect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(broker.published) != 1 || broker.published[0].Topic != StatusTopic("edge-1") || string(broker.published[0].Payload) != "offline" || !broker.published[0].Retain {
		t.Fatalf("unexpected offline publish %#v", broker.published)
	}
	if broker.connected {
		t.Fatal("expected broker to be disconnected")
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
	if len(broker.published) != 3 {
		t.Fatalf("unexpected response publish %#v", broker.published)
	}
	for _, message := range broker.published {
		if message.Topic != ResponseTopic("edge-1") {
			t.Fatalf("unexpected response topic %#v", broker.published)
		}
	}
	var received map[string]any
	if err := json.Unmarshal(broker.published[0].Payload, &received); err != nil {
		t.Fatal(err)
	}
	if received["command_id"] != "cmd-1" || received["status"] != "received" {
		t.Fatalf("unexpected received response %#v", received)
	}
	var running map[string]any
	if err := json.Unmarshal(broker.published[1].Payload, &running); err != nil {
		t.Fatal(err)
	}
	if running["status"] != "running" {
		t.Fatalf("unexpected running response %#v", running)
	}
	var done map[string]any
	if err := json.Unmarshal(broker.published[2].Payload, &done); err != nil {
		t.Fatal(err)
	}
	if done["status"] != "done" {
		t.Fatalf("unexpected done response %#v", done)
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
	if len(broker.published) != 3 {
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
	if commands.saved != 1 || commands.statuses[0] != domain.MqttCommandReceived || commands.statuses[1] != domain.MqttCommandRunning || commands.statuses[2] != domain.MqttCommandDone {
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

func TestPublishContextUsesExpectedTopicAndPayload(t *testing.T) {
	broker := &fakeBroker{connected: true}
	client := NewClientWithBroker(broker)
	client.config = Config{EdgeID: "edge-1"}

	if err := client.PublishContext(context.Background(), map[string]any{"id": "user-1", "email": "edge@example.com"}); err != nil {
		t.Fatal(err)
	}
	if len(broker.published) != 1 || broker.published[0].Topic != ContextTopic("edge-1") {
		t.Fatalf("unexpected published context %#v", broker.published)
	}
	var payload map[string]any
	if err := json.Unmarshal(broker.published[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["edge_id"] != "edge-1" {
		t.Fatalf("unexpected edge_id in context payload %#v", payload)
	}
	if payload["schema_version"] != edgeCloudSchemaVersion || payload["type"] != "context" {
		t.Fatalf("unexpected envelope in context payload %#v", payload)
	}
	localUser, ok := payload["local_user"].(map[string]any)
	if !ok || localUser["id"] != "user-1" || localUser["email"] != "edge@example.com" {
		t.Fatalf("unexpected local_user payload %#v", payload)
	}
	if _, ok := payload["timestamp"].(string); !ok {
		t.Fatalf("expected timestamp in context payload %#v", payload)
	}
}

func TestPublishHeartbeatUsesExpectedTopic(t *testing.T) {
	broker := &fakeBroker{connected: true}
	client := NewClientWithBroker(broker)
	client.config = Config{EdgeID: "edge-1"}

	if err := client.PublishHeartbeat(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(broker.published) != 1 || broker.published[0].Topic != HeartbeatTopic("edge-1") {
		t.Fatalf("unexpected heartbeat publish %#v", broker.published)
	}
	var payload map[string]any
	if err := json.Unmarshal(broker.published[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["edge_id"] != "edge-1" {
		t.Fatalf("unexpected heartbeat payload %#v", payload)
	}
	if payload["schema_version"] != edgeCloudSchemaVersion || payload["type"] != "heartbeat" || payload["status"] != "online" {
		t.Fatalf("unexpected heartbeat envelope %#v", payload)
	}
	runtimePayload, ok := payload["runtime"].(map[string]any)
	if !ok || runtimePayload["uptime_seconds"] == nil {
		t.Fatalf("expected runtime payload in heartbeat %#v", payload)
	}
}

func TestPublishStatsUsesExpectedTopicAndPayload(t *testing.T) {
	broker := &fakeBroker{connected: true}
	client := NewClientWithBroker(broker)
	client.config = Config{EdgeID: "edge-1"}

	if err := client.PublishStats(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(broker.published) != 1 || broker.published[0].Topic != StatsTopic("edge-1") {
		t.Fatalf("unexpected stats publish %#v", broker.published)
	}
	var payload map[string]any
	if err := json.Unmarshal(broker.published[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["schema_version"] != edgeCloudSchemaVersion || payload["type"] != "stats" {
		t.Fatalf("unexpected stats payload %#v", payload)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok || data["memory_mb"] == nil || data["uptime_seconds"] == nil {
		t.Fatalf("unexpected stats data %#v", payload)
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
	subscribed      []string
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
	b.subscribed = append(b.subscribed, topic)
	return nil
}

func TestSyncTopicHandlersSubscribesAndRoutesWildcardTopics(t *testing.T) {
	broker := &fakeBroker{}
	client := NewClientWithBroker(broker)
	if err := client.Connect(context.Background(), Config{EdgeID: "edge-1", BrokerURL: "mqtt://localhost:1883"}); err != nil {
		t.Fatal(err)
	}

	var receivedTopic string
	var receivedPayload string
	if err := client.SyncTopicHandlers(context.Background(), map[string]TopicHandler{
		"spark/custom/+/events": func(_ context.Context, topic string, payload []byte) {
			receivedTopic = topic
			receivedPayload = string(payload)
		},
	}); err != nil {
		t.Fatal(err)
	}

	if len(broker.subscribed) < 2 {
		t.Fatalf("expected command and custom topic subscriptions, got %#v", broker.subscribed)
	}

	broker.handler("spark/custom/device-1/events", []byte(`{"ok":true}`))
	if receivedTopic != "spark/custom/device-1/events" || receivedPayload != `{"ok":true}` {
		t.Fatalf("unexpected routed message topic=%q payload=%q", receivedTopic, receivedPayload)
	}
}

func (b *fakeBroker) IsConnected() bool {
	return b.connected
}
