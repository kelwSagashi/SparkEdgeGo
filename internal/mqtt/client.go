package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

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

type Client struct {
	config        Config
	broker        Broker
	mu            sync.Mutex
	handlers      map[string]CommandHandler
	processed     map[string]struct{}
	heartbeatStop chan struct{}
}

func NewClient() *Client {
	return NewClientWithBroker(&pahoBroker{})
}

func NewClientWithBroker(broker Broker) *Client {
	client := &Client{
		broker:    broker,
		handlers:  map[string]CommandHandler{},
		processed: map[string]struct{}{},
	}
	client.RegisterHandler("ping", func(context.Context, map[string]any) (map[string]any, error) {
		return map[string]any{"pong": true, "timestamp": time.Now().UTC().Format(time.RFC3339)}, nil
	})
	return client
}

func (c *Client) Connect(ctx context.Context, config Config) error {
	if strings.TrimSpace(config.EdgeID) == "" || strings.TrimSpace(config.BrokerURL) == "" {
		return errors.New("mqtt requires edge_id and broker_url")
	}
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

func (c *Client) SubscribeCommands(ctx context.Context) error {
	return c.broker.Subscribe(ctx, CommandTopic(c.config.EdgeID), 1)
}

func (c *Client) PublishStatus(ctx context.Context, status string) error {
	return c.broker.Publish(ctx, Message{Topic: StatusTopic(c.config.EdgeID), Payload: []byte(status), QOS: 1, Retain: true})
}

func (c *Client) PublishHeartbeat(ctx context.Context) error {
	return c.PublishJSON(ctx, HeartbeatTopic(c.config.EdgeID), map[string]any{"edge_id": c.config.EdgeID, "ts": time.Now().Unix()}, false)
}

func (c *Client) PublishResponse(ctx context.Context, commandID string, status string, result map[string]any, errText string) error {
	return c.PublishJSON(ctx, ResponseTopic(c.config.EdgeID), map[string]any{
		"command_id": commandID,
		"status":     status,
		"result":     result,
		"error":      nullableString(errText),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}, false)
}

func (c *Client) PublishLog(ctx context.Context, level string, message string) error {
	if level == "" {
		level = "info"
	}
	return c.PublishJSON(ctx, LogTopic(c.config.EdgeID), map[string]any{"level": level, "message": message, "timestamp": time.Now().UTC().Format(time.RFC3339)}, false)
}

func (c *Client) PublishContext(ctx context.Context, user map[string]any) error {
	return c.PublishJSON(ctx, ContextTopic(c.config.EdgeID), map[string]any{"edge_id": c.config.EdgeID, "local_user": user, "timestamp": time.Now().UTC().Format(time.RFC3339)}, false)
}

func (c *Client) PublishJSON(ctx context.Context, topic string, payload map[string]any, retain bool) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.broker.Publish(ctx, Message{Topic: topic, Payload: encoded, QOS: 1, Retain: retain})
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
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = c.PublishHeartbeat(context.Background())
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
}

func (c *Client) handleMessage(topic string, payload []byte) {
	if topic != CommandTopic(c.config.EdgeID) {
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
	c.mu.Lock()
	if _, ok := c.processed[command.CommandID]; ok {
		c.mu.Unlock()
		return nil
	}
	c.processed[command.CommandID] = struct{}{}
	handler := c.handlers[command.Type]
	c.mu.Unlock()
	if handler == nil {
		errText := fmt.Sprintf("unknown command type: %s", command.Type)
		_ = c.PublishResponse(ctx, command.CommandID, "error", nil, errText)
		return errors.New(errText)
	}
	result, err := handler(ctx, command.Payload)
	if err != nil {
		_ = c.PublishResponse(ctx, command.CommandID, "error", nil, err.Error())
		return err
	}
	return c.PublishResponse(ctx, command.CommandID, "done", result, "")
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
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
