package mqtt

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/system"
)

const edgeCloudSchemaVersion = "edge-cloud.v1"

type Config struct {
	EdgeID    string
	BrokerURL string
	Username  string
	Password  string
}

type Message struct {
	Topic   string
	Payload []byte
	QOS     byte
	Retain  bool
}

type Broker interface {
	Connect(context.Context, Config, func(string, []byte)) error
	Disconnect()
	Publish(context.Context, Message) error
	Subscribe(context.Context, string, byte) error
	IsConnected() bool
}

type Command struct {
	CommandID string         `json:"command_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
}

type CommandHandler func(context.Context, map[string]any) (map[string]any, error)
type TopicHandler func(context.Context, string, []byte)
type StatsProvider func(context.Context) map[string]any

type CommandStore interface {
	FindByCommandID(context.Context, string) (domain.MqttCommand, error)
	Save(context.Context, string, string, map[string]any) (domain.MqttCommand, error)
	UpdateStatus(context.Context, string, domain.MqttCommandStatus, map[string]any, string) (domain.MqttCommand, error)
}

type QueueStore interface {
	Enqueue(context.Context, string, string) (domain.MqttQueueItem, error)
	ListPending(context.Context, int) ([]domain.MqttQueueItem, error)
	Delete(context.Context, string) error
	IncrementAttempt(context.Context, string) error
}

type Client struct {
	config        Config
	broker        Broker
	commands      CommandStore
	queue         QueueStore
	mu            sync.Mutex
	handlers      map[string]CommandHandler
	topicHandlers map[string]TopicHandler
	subscribed    map[string]struct{}
	processed     map[string]struct{}
	heartbeatStop chan struct{}
	statsStop     chan struct{}
	statsProvider StatsProvider
}

func NewClient() *Client {
	return NewClientWithBroker(&pahoBroker{})
}

func NewClientWithBroker(broker Broker) *Client {
	client := &Client{
		broker:        broker,
		handlers:      map[string]CommandHandler{},
		topicHandlers: map[string]TopicHandler{},
		subscribed:    map[string]struct{}{},
		processed:     map[string]struct{}{},
	}
	client.RegisterHandler("ping", func(context.Context, map[string]any) (map[string]any, error) {
		return map[string]any{"pong": true, "timestamp": time.Now().UTC().Format(time.RFC3339)}, nil
	})
	return client
}

func (c *Client) UseStores(commands CommandStore, queue QueueStore) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.commands = commands
	c.queue = queue
}

func (c *Client) Connect(ctx context.Context, config Config) error {
	if strings.TrimSpace(config.EdgeID) == "" || strings.TrimSpace(config.BrokerURL) == "" {
		return errors.New("mqtt requires edge_id and broker_url")
	}
	config.BrokerURL = NormalizeBrokerURL(config.BrokerURL)
	c.mu.Lock()
	c.config = config
	c.mu.Unlock()
	if err := c.broker.Connect(ctx, config, c.handleMessage); err != nil {
		return err
	}
	if err := c.SubscribeCommands(ctx); err != nil {
		return err
	}
	return c.PublishStatus(ctx, "online")
}

func (c *Client) Disconnect(ctx context.Context) error {
	c.StopHeartbeat()
	if c.IsConnected() {
		_ = c.PublishStatus(ctx, "offline")
	}
	c.broker.Disconnect()
	return nil
}

func (c *Client) IsConnected() bool {
	if c == nil || c.broker == nil {
		return false
	}
	return c.broker.IsConnected()
}

func (c *Client) RegisterHandler(commandType string, handler CommandHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[commandType] = handler
}

func (c *Client) SetStatsProvider(provider StatsProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statsProvider = provider
}

func (c *Client) SubscribeCommands(ctx context.Context) error {
	return c.broker.Subscribe(ctx, CommandTopic(c.config.EdgeID), 1)
}

func (c *Client) SyncTopicHandlers(ctx context.Context, handlers map[string]TopicHandler) error {
	c.mu.Lock()
	current := make(map[string]TopicHandler, len(handlers))
	for topic, handler := range handlers {
		current[topic] = handler
	}
	c.topicHandlers = current
	toSubscribe := make([]string, 0, len(handlers))
	for topic := range handlers {
		if _, ok := c.subscribed[topic]; !ok {
			toSubscribe = append(toSubscribe, topic)
			c.subscribed[topic] = struct{}{}
		}
	}
	c.mu.Unlock()

	for _, topic := range toSubscribe {
		if err := c.broker.Subscribe(ctx, topic, 1); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) PublishStatus(ctx context.Context, status string) error {
	return c.publish(ctx, Message{Topic: StatusTopic(c.config.EdgeID), Payload: []byte(status), QOS: 1, Retain: true})
}

func (c *Client) PublishHeartbeat(ctx context.Context) error {
	now := time.Now().UTC()
	stats := c.collectStats(ctx)
	status := "online"
	if value, ok := stats["status"].(string); ok && strings.TrimSpace(value) != "" {
		status = strings.TrimSpace(value)
	}
	_ = c.PublishStatus(ctx, "online")
	return c.PublishJSON(ctx, HeartbeatTopic(c.config.EdgeID), c.envelope("heartbeat", map[string]any{
		"status":       status,
		"ts":           now.Unix(),
		"connectivity": stats["connectivity"],
		"runtime": map[string]any{
			"uptime_seconds":   stats["uptime_seconds"],
			"goroutines":       stats["goroutines"],
			"active_instances": stats["active_instances"],
		},
		"queue_sizes":                stats["queue_sizes"],
		"oldest_pending_age_seconds": stats["oldest_pending_age_seconds"],
	}, now), false)
}

func (c *Client) PublishStats(ctx context.Context) error {
	now := time.Now().UTC()
	return c.PublishJSON(ctx, StatsTopic(c.config.EdgeID), c.envelope("stats", map[string]any{
		"status": statsStatus(c.collectStats(ctx)),
		"data":   c.collectStats(ctx),
	}, now), false)
}

func (c *Client) PublishMeta(ctx context.Context, metadata map[string]any) error {
	now := time.Now().UTC()
	payload := map[string]any{}
	for key, value := range metadata {
		payload[key] = value
	}
	delete(payload, "message_id")
	delete(payload, "schema_version")
	delete(payload, "type")
	delete(payload, "occurred_at")
	delete(payload, "timestamp")
	delete(payload, "ts")
	return c.PublishJSON(ctx, MetaTopic(c.config.EdgeID), c.envelope("meta", payload, now), false)
}

func (c *Client) PublishResponse(ctx context.Context, commandID string, status string, result map[string]any, errText string) error {
	return c.PublishJSON(ctx, ResponseTopic(c.config.EdgeID), c.envelope("command_response", map[string]any{
		"command_id": commandID,
		"status":     status,
		"result":     result,
		"error":      nullableString(errText),
	}, time.Now().UTC()), false)
}

func (c *Client) PublishLog(ctx context.Context, level string, message string) error {
	if level == "" {
		level = "info"
	}
	return c.PublishJSON(ctx, LogTopic(c.config.EdgeID), c.envelope("log", map[string]any{"level": level, "message": message}, time.Now().UTC()), false)
}

func (c *Client) PublishContext(ctx context.Context, user map[string]any) error {
	return c.PublishJSON(ctx, ContextTopic(c.config.EdgeID), c.envelope("context", map[string]any{"local_user": user}, time.Now().UTC()), false)
}

func (c *Client) PublishJSON(ctx context.Context, topic string, payload map[string]any, retain bool) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.publish(ctx, Message{Topic: topic, Payload: encoded, QOS: 1, Retain: retain})
}

func (c *Client) publish(ctx context.Context, message Message) error {
	if !c.IsConnected() {
		if c.queue != nil {
			_, err := c.queue.Enqueue(ctx, message.Topic, string(message.Payload))
			return err
		}
		return c.broker.Publish(ctx, message)
	}
	if err := c.broker.Publish(ctx, message); err != nil {
		if c.queue != nil {
			_, _ = c.queue.Enqueue(ctx, message.Topic, string(message.Payload))
		}
		return err
	}
	return nil
}

func (c *Client) StartHeartbeat(interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	c.mu.Lock()
	if c.heartbeatStop != nil {
		c.mu.Unlock()
		return
	}
	stop := make(chan struct{})
	c.heartbeatStop = stop
	c.mu.Unlock()
	go func() {
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				_ = c.PublishHeartbeat(context.Background())
				timer.Reset(c.nextHeartbeatInterval(interval))
			case <-stop:
				return
			}
		}
	}()
}

func (c *Client) StartStats(interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Minute
	}
	c.mu.Lock()
	if c.statsStop != nil {
		c.mu.Unlock()
		return
	}
	stop := make(chan struct{})
	c.statsStop = stop
	c.mu.Unlock()
	go func() {
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				_ = c.PublishStats(context.Background())
				timer.Reset(c.nextStatsInterval(interval))
			case <-stop:
				return
			}
		}
	}()
}

func (c *Client) StopHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.heartbeatStop != nil {
		close(c.heartbeatStop)
		c.heartbeatStop = nil
	}
	if c.statsStop != nil {
		close(c.statsStop)
		c.statsStop = nil
	}
}

func (c *Client) handleMessage(topic string, payload []byte) {
	if topic != CommandTopic(c.config.EdgeID) {
		c.mu.Lock()
		handlers := make(map[string]TopicHandler, len(c.topicHandlers))
		for pattern, handler := range c.topicHandlers {
			handlers[pattern] = handler
		}
		c.mu.Unlock()
		for pattern, handler := range handlers {
			if mqttTopicMatches(pattern, topic) && handler != nil {
				handler(context.Background(), topic, payload)
			}
		}
		return
	}
	_ = c.HandleCommand(context.Background(), payload)
}

func (c *Client) HandleCommand(ctx context.Context, raw []byte) error {
	var command Command
	if err := json.Unmarshal(raw, &command); err != nil {
		return err
	}
	if command.CommandID == "" || command.Type == "" {
		return errors.New("mqtt command requires command_id and type")
	}
	if c.commands != nil {
		if _, err := c.commands.FindByCommandID(ctx, command.CommandID); err == nil {
			return nil
		} else if !isNotFound(err) {
			return err
		}
		if _, err := c.commands.Save(ctx, command.CommandID, command.Type, command.Payload); err != nil {
			return err
		}
	}
	c.mu.Lock()
	if _, ok := c.processed[command.CommandID]; ok {
		c.mu.Unlock()
		return nil
	}
	c.processed[command.CommandID] = struct{}{}
	handler := c.handlers[command.Type]
	c.mu.Unlock()
	if c.commands != nil {
		_, _ = c.commands.UpdateStatus(ctx, command.CommandID, domain.MqttCommandReceived, nil, "")
	}
	_ = c.PublishResponse(ctx, command.CommandID, "received", map[string]any{"accepted": true}, "")
	if handler == nil {
		errText := fmt.Sprintf("unknown command type: %s", command.Type)
		if c.commands != nil {
			_, _ = c.commands.UpdateStatus(ctx, command.CommandID, domain.MqttCommandError, nil, errText)
		}
		_ = c.PublishResponse(ctx, command.CommandID, "error", nil, errText)
		return errors.New(errText)
	}
	if c.commands != nil {
		_, _ = c.commands.UpdateStatus(ctx, command.CommandID, domain.MqttCommandRunning, nil, "")
	}
	_ = c.PublishResponse(ctx, command.CommandID, "running", nil, "")
	handlerPayload := map[string]any{}
	for key, value := range command.Payload {
		handlerPayload[key] = value
	}
	handlerPayload["command_id"] = command.CommandID
	result, err := handler(ctx, handlerPayload)
	if err != nil {
		if c.commands != nil {
			_, _ = c.commands.UpdateStatus(ctx, command.CommandID, domain.MqttCommandError, nil, err.Error())
		}
		_ = c.PublishResponse(ctx, command.CommandID, "error", nil, err.Error())
		return err
	}
	if c.commands != nil {
		_, _ = c.commands.UpdateStatus(ctx, command.CommandID, domain.MqttCommandDone, result, "")
	}
	return c.PublishResponse(ctx, command.CommandID, "done", result, "")
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func (c *Client) Enqueue(ctx context.Context, topic string, payload []byte) error {
	if c.queue == nil {
		return errors.New("mqtt queue store is not configured")
	}
	_, err := c.queue.Enqueue(ctx, topic, string(payload))
	return err
}

func (c *Client) RetryQueue(ctx context.Context, maxAttempts int) error {
	if c.queue == nil {
		return nil
	}
	items, err := c.queue.ListPending(ctx, maxAttempts)
	if err != nil {
		return err
	}
	var failures []string
	for _, item := range items {
		err := c.broker.Publish(ctx, Message{Topic: item.Topic, Payload: []byte(item.Payload), QOS: 1})
		if err == nil {
			_ = c.queue.Delete(ctx, item.ID)
			continue
		}
		_ = c.queue.IncrementAttempt(ctx, item.ID)
		if maxAttempts > 0 && item.Attempts+1 >= maxAttempts {
			_ = c.queue.Delete(ctx, item.ID)
		}
		failures = append(failures, item.ID+": "+err.Error())
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (c *Client) collectStats(ctx context.Context) map[string]any {
	c.mu.Lock()
	provider := c.statsProvider
	c.mu.Unlock()
	if provider != nil {
		if stats := provider(ctx); stats != nil {
			return stats
		}
	}
	return system.CollectStats()
}

func statsStatus(stats map[string]any) string {
	if value, ok := stats["status"].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return "online"
}

func (c *Client) nextHeartbeatInterval(defaultInterval time.Duration) time.Duration {
	return c.intervalFromConnectivity("heartbeat_interval_seconds", defaultInterval)
}

func (c *Client) nextStatsInterval(defaultInterval time.Duration) time.Duration {
	return c.intervalFromConnectivity("stats_interval_seconds", defaultInterval)
}

func (c *Client) intervalFromConnectivity(key string, defaultInterval time.Duration) time.Duration {
	stats := c.collectStats(context.Background())
	connectivityValue, ok := stats["connectivity"].(map[string]any)
	if !ok {
		return defaultInterval
	}
	raw, ok := connectivityValue[key]
	if !ok {
		return defaultInterval
	}
	switch typed := raw.(type) {
	case int:
		if typed > 0 {
			return time.Duration(typed) * time.Second
		}
	case int64:
		if typed > 0 {
			return time.Duration(typed) * time.Second
		}
	case float64:
		if typed > 0 {
			return time.Duration(int(typed)) * time.Second
		}
	}
	return defaultInterval
}

func (c *Client) envelope(eventType string, payload map[string]any, now time.Time) map[string]any {
	result := map[string]any{
		"schema_version": edgeCloudSchemaVersion,
		"message_id":     newMessageID(),
		"type":           eventType,
		"edge_id":        c.config.EdgeID,
		"occurred_at":    now.Format(time.RFC3339),
		"timestamp":      now.Format(time.RFC3339),
	}
	for key, value := range payload {
		result[key] = value
	}
	return result
}

func newMessageID() string {
	const alphabet = "0123456789abcdef"
	result := make([]byte, 32)
	for index := range result {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return fmt.Sprintf("msg-%d", time.Now().UTC().UnixNano())
		}
		result[index] = alphabet[value.Int64()]
	}
	return string(result)
}

func StatusTopic(edgeID string) string    { return "spark/" + edgeID + "/status" }
func HeartbeatTopic(edgeID string) string { return "spark/" + edgeID + "/heartbeat" }
func CommandTopic(edgeID string) string   { return "spark/" + edgeID + "/commands" }
func ResponseTopic(edgeID string) string  { return "spark/" + edgeID + "/response" }
func MetaTopic(edgeID string) string      { return "spark/" + edgeID + "/meta" }
func LogTopic(edgeID string) string       { return "spark/" + edgeID + "/logs" }
func MetricsTopic(edgeID string) string   { return "spark/" + edgeID + "/metrics" }
func StatsTopic(edgeID string) string     { return "spark/" + edgeID + "/stats" }
func ContextTopic(edgeID string) string   { return "spark/" + edgeID + "/context" }

func NormalizeBrokerURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" {
		return trimmed
	}
	switch strings.ToLower(parsed.Scheme) {
	case "https":
		parsed.Scheme = "wss"
	case "http":
		parsed.Scheme = "ws"
	}
	if (parsed.Scheme == "ws" || parsed.Scheme == "wss") && (parsed.Path == "" || parsed.Path == "/") {
		parsed.Path = "/mqtt"
	}
	return parsed.String()
}

type pahoBroker struct {
	client paho.Client
}

func (b *pahoBroker) Connect(ctx context.Context, config Config, handler func(string, []byte)) error {
	options := paho.NewClientOptions()
	options.AddBroker(config.BrokerURL)
	options.SetClientID(config.EdgeID)
	options.SetUsername(config.Username)
	options.SetPassword(config.Password)
	options.SetCleanSession(true)
	options.SetKeepAlive(30 * time.Second)
	options.SetAutoReconnect(true)
	options.SetWill(StatusTopic(config.EdgeID), "offline", 1, true)
	options.SetDefaultPublishHandler(func(_ paho.Client, msg paho.Message) {
		handler(msg.Topic(), msg.Payload())
	})
	client := paho.NewClient(options)
	token := client.Connect()
	if err := waitToken(ctx, token); err != nil {
		return err
	}
	b.client = client
	return nil
}

func (b *pahoBroker) Disconnect() {
	if b.client != nil && b.client.IsConnected() {
		b.client.Disconnect(250)
	}
	b.client = nil
}

func (b *pahoBroker) Publish(ctx context.Context, msg Message) error {
	if b.client == nil {
		return errors.New("mqtt client is not connected")
	}
	return waitToken(ctx, b.client.Publish(msg.Topic, msg.QOS, msg.Retain, msg.Payload))
}

func (b *pahoBroker) Subscribe(ctx context.Context, topic string, qos byte) error {
	if b.client == nil {
		return errors.New("mqtt client is not connected")
	}
	return waitToken(ctx, b.client.Subscribe(topic, qos, nil))
}

func (b *pahoBroker) IsConnected() bool {
	return b.client != nil && b.client.IsConnected()
}

func waitToken(ctx context.Context, token paho.Token) error {
	done := make(chan struct{})
	go func() {
		token.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return token.Error()
	}
}

func mqttTopicMatches(pattern string, topic string) bool {
	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	for index := 0; index < len(patternParts); index++ {
		if index >= len(topicParts) {
			return patternParts[index] == "#"
		}
		switch patternParts[index] {
		case "#":
			return true
		case "+":
			continue
		default:
			if patternParts[index] != topicParts[index] {
				return false
			}
		}
	}

	return len(patternParts) == len(topicParts)
}
