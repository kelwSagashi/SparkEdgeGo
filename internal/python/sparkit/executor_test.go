package sparkit

import "testing"

func TestParseSparkitEnvelopeSuccess(t *testing.T) {
	envelope, ok := parseSparkitEnvelope([]byte(`{"stdout":{"ok":true,"temperature":42},"stderr":null}`))
	if !ok {
		t.Fatal("expected sparkit envelope to be detected")
	}

	stdout, ok := envelope.stdout.(map[string]any)
	if !ok {
		t.Fatalf("expected stdout object, got %#v", envelope.stdout)
	}
	if stdout["ok"] != true || stdout["temperature"] != float64(42) {
		t.Fatalf("unexpected stdout payload %#v", stdout)
	}
	if envelope.stderr != nil {
		t.Fatalf("expected nil stderr, got %#v", envelope.stderr)
	}
	if envelope.envelope["stdout"] == nil || envelope.envelope["stderr"] != nil {
		t.Fatalf("unexpected envelope %#v", envelope.envelope)
	}
}

func TestParseSparkitEnvelopeFailure(t *testing.T) {
	envelope, ok := parseSparkitEnvelope([]byte(`{"stdout":null,"stderr":{"type":"Exception","message":"Nao foi possivel conectar.","runtime_stderr":"could not open port '/dev/ttyUSB0'"}}`))
	if !ok {
		t.Fatal("expected sparkit envelope to be detected")
	}

	stderr, ok := envelope.stderr.(map[string]any)
	if !ok {
		t.Fatalf("expected structured stderr object, got %#v", envelope.stderr)
	}
	if stderr["type"] != "Exception" || stderr["message"] != "Nao foi possivel conectar." {
		t.Fatalf("unexpected stderr payload %#v", stderr)
	}

	message := structuredErrorMessage(envelope.stderr, "")
	if message != "Nao foi possivel conectar." {
		t.Fatalf("expected message to prefer structured error message, got %q", message)
	}
}

func TestStructuredErrorMessageFallsBackToRuntimeStderr(t *testing.T) {
	message := structuredErrorMessage(map[string]any{
		"runtime_stderr": "could not open port '/dev/ttyUSB0'",
	}, "")
	if message != "could not open port '/dev/ttyUSB0'" {
		t.Fatalf("expected runtime stderr fallback, got %q", message)
	}
}
