package mqttprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

type Adapter struct {
	credentials map[string]any
	resource    map[string]any
	operation   map[string]any
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
	credentials := mapValue(config.Credentials, "data")
	brokerURL := stringValue(credentials, "brokerUrl", "broker_url", "url")
	if strings.TrimSpace(brokerURL) == "" {
		return nil, errors.New("mqtt provider requires credentials.data.brokerUrl")
	}

	return &Adapter{
		credentials: credentials,
		resource:    mapValue(config.Resource, "config"),
		operation:   mapValue(config.Operation, "config"),
	}, nil
}

func (a *Adapter) Send(ctx context.Context, payload map[string]any) error {
	topic := stringValue(a.resource, "topic")
	if topic == "" {
		return errors.New("mqtt provider requires resource.config.topic")
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client, err := a.connect(ctx)
	if err != nil {
		return err
	}
	defer a.disconnect(client)

	qos := byte(intValue(a.operation, 0, "qos"))
	retained := boolValue(a.operation, false, "retained")
	token := client.Publish(topic, qos, retained, encoded)
	return waitToken(ctx, token, 15*time.Second, "mqtt publish timeout")
}

func (a *Adapter) Test(ctx context.Context, payload map[string]any) error {
	client, err := a.connect(ctx)
	if err != nil {
		return err
	}
	defer a.disconnect(client)
	return nil
}

func (a *Adapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	return []providers.Resource{
		{
			Name:   "MQTT topic",
			Type:   "topic",
			Config: a.resource,
			Fields: []providers.Field{
				{Name: "topic", Type: "string"},
			},
		},
	}, ctx.Err()
}

func (a *Adapter) connect(ctx context.Context) (mqtt.Client, error) {
	options := mqtt.NewClientOptions()
	options.AddBroker(stringValue(a.credentials, "brokerUrl", "broker_url", "url"))
	options.SetClientID(defaultString(
		stringValue(a.credentials, "clientId", "client_id"),
		fmt.Sprintf("sparkedge-%d", time.Now().UnixNano()),
	))
	if username := stringValue(a.credentials, "username"); username != "" {
		options.SetUsername(username)
	}
	if password := stringValue(a.credentials, "password"); password != "" {
		options.SetPassword(password)
	}
	options.SetConnectTimeout(15 * time.Second)
	options.SetWriteTimeout(15 * time.Second)
	options.SetAutoReconnect(false)
	options.SetCleanSession(true)

	client := mqtt.NewClient(options)
	token := client.Connect()
	if err := waitToken(ctx, token, 15*time.Second, "mqtt connect timeout"); err != nil {
		return nil, err
	}
	return client, nil
}

func (a *Adapter) disconnect(client mqtt.Client) {
	if client == nil || !client.IsConnected() {
		return
	}
	client.Disconnect(250)
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

func boolValue(source map[string]any, fallback bool, keys ...string) bool {
	for _, key := range keys {
		if value, ok := source[key]; ok {
			if typed, ok := value.(bool); ok {
				return typed
			}
		}
	}
	return fallback
}

func intValue(source map[string]any, fallback int, keys ...string) int {
	for _, key := range keys {
		if value, ok := source[key]; ok {
			switch typed := value.(type) {
			case int:
				return typed
			case int32:
				return int(typed)
			case int64:
				return int(typed)
			case float64:
				return int(typed)
			}
		}
	}
	return fallback
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func waitToken(ctx context.Context, token mqtt.Token, timeout time.Duration, timeoutMessage string) error {
	done := make(chan error, 1)
	go func() {
		token.Wait()
		done <- token.Error()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	case <-time.After(timeout):
		return errors.New(timeoutMessage)
	}
}
