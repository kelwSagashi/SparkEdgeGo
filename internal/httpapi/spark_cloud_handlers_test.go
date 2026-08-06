package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSparkCloudLoginReturnsToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/spark-cloud/auth/login", nil)
	rec := httptest.NewRecorder()

	Adapt((&Server{}).handleSparkCloudLogin).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["token"] == "" {
		t.Fatal("expected token")
	}
}

func TestSparkCloudEdgeRegisterReturnsMqttCredentials(t *testing.T) {
	t.Setenv("MQTT_URL", "mqtt://emqx.local:1883")
	payload := []byte(`{"name":"Edge Lab","lat":"-1.0","lng":"-2.0","tags":["lab"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/spark-cloud/edges/register", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	Adapt((&Server{}).handleSparkCloudEdgeRegister).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var body struct {
		EdgeID   string `json:"edge_id"`
		EdgeName string `json:"edge_name"`
		MQTT     struct {
			URL      string `json:"url"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"mqtt"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.EdgeID == "" || body.EdgeName != "Edge Lab" {
		t.Fatalf("unexpected edge identity: %#v", body)
	}
	if body.MQTT.URL != "mqtt://emqx.local:1883" || body.MQTT.Username == "" || body.MQTT.Password == "" {
		t.Fatalf("unexpected mqtt credentials: %#v", body.MQTT)
	}
}
