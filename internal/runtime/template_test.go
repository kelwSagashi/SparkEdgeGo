package runtime

import "testing"

func TestTemplateResolverPreservesTypeForWholeTemplate(t *testing.T) {
	resolver := TemplateResolver{}
	context := map[string]any{
		"script": map[string]any{
			"temperature": float64(42.5),
			"active":      true,
		},
	}

	if got := resolver.Resolve("{{script.temperature}}", context); got != float64(42.5) {
		t.Fatalf("expected numeric value, got %#v", got)
	}
	if got := resolver.Resolve("{{script.active}}", context); got != true {
		t.Fatalf("expected boolean value, got %#v", got)
	}
}

func TestTemplateResolverInterpolatesStringsAndKeepsMissingTemplates(t *testing.T) {
	resolver := TemplateResolver{}
	context := map[string]any{"device": map[string]any{"name": "Pump 1"}}

	got := resolver.Resolve("Device {{device.name}} / {{device.missing}}", context)
	if got != "Device Pump 1 / {{device.missing}}" {
		t.Fatalf("unexpected interpolation: %#v", got)
	}
}

func TestTemplateResolverSupportsFallbackExpressions(t *testing.T) {
	resolver := TemplateResolver{}
	context := map[string]any{"script": map[string]any{"empty": "", "value": "online"}}

	if got := resolver.Resolve("{{script.empty || script.value || 'offline'}}", context); got != "online" {
		t.Fatalf("expected fallback to script.value, got %#v", got)
	}
	if got := resolver.Resolve("{{script.missing || 'offline'}}", context); got != "offline" {
		t.Fatalf("expected string fallback, got %#v", got)
	}
	if got := resolver.Resolve("{{script.missing || 123}}", context); got != float64(123) {
		t.Fatalf("expected numeric fallback, got %#v", got)
	}
}

func TestTemplateResolverSupportsSimpleJSONPath(t *testing.T) {
	resolver := TemplateResolver{}
	context := map[string]any{
		"script": map[string]any{
			"items": []any{
				map[string]any{"value": "first"},
				map[string]any{"value": "second"},
			},
		},
	}

	if got := resolver.Resolve("{{$.script.items[1].value}}", context); got != "second" {
		t.Fatalf("expected JSONPath value, got %#v", got)
	}
	if got := resolver.ResolvePath("$.script.items[0]['value']", context); got != "first" {
		t.Fatalf("expected bracket JSONPath value, got %#v", got)
	}
}

func TestTemplateResolverRecursivelyResolvesObjectsAndArrays(t *testing.T) {
	resolver := TemplateResolver{}
	context := map[string]any{"output": map[string]any{"voltage": float64(220)}}
	input := map[string]any{
		"nested": map[string]any{"voltage": "{{output.voltage}}"},
		"labels": []any{"v={{output.voltage}}"},
	}

	got := resolver.Resolve(input, context).(map[string]any)
	if got["nested"].(map[string]any)["voltage"] != float64(220) {
		t.Fatalf("unexpected nested value: %#v", got)
	}
	if got["labels"].([]any)[0] != "v=220" {
		t.Fatalf("unexpected array value: %#v", got)
	}
}

func TestTemplateResolverSupportsFunctionsAndConditionals(t *testing.T) {
	resolver := TemplateResolver{}
	context := map[string]any{
		"device": map[string]any{"name": "edge-alpha"},
		"script": map[string]any{"temperature": float64(42)},
	}

	if got := resolver.Resolve("{{concat('device-', upper(device.name))}}", context); got != "device-EDGE-ALPHA" {
		t.Fatalf("unexpected concat/upper result: %#v", got)
	}

	if got := resolver.Resolve("{{if(script.temperature > 40, 'hot', 'ok')}}", context); got != "hot" {
		t.Fatalf("unexpected conditional result: %#v", got)
	}

	if got := resolver.Resolve("{{add(script.temperature, 8)}}", context); got != float64(50) {
		t.Fatalf("unexpected add result: %#v", got)
	}
}
