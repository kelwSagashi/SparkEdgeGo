package runtime

import (
	"fmt"
	"strings"
)

func applyTransformScript(script string, payload map[string]any, context map[string]any) (map[string]any, error) {
	trimmed := strings.TrimSpace(script)
	if trimmed == "" {
		return payload, nil
	}

	resolver := TemplateResolver{}
	result := cloneMap(payload)
	workingContext := cloneTemplateContext(context)
	workingContext["data"] = result
	workingContext["payload"] = result

	lines := normalizedTransformLines(trimmed)
	if len(lines) == 0 {
		return result, nil
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "return "):
			return evaluateTransformReturn(strings.TrimSpace(strings.TrimPrefix(line, "return ")), result, workingContext, resolver)
		case strings.HasPrefix(line, "set "):
			assignment := strings.TrimSpace(strings.TrimPrefix(line, "set "))
			parts := splitTopLevelByOperator(assignment, '=')
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid transform set statement: %s", line)
			}
			fieldPath := normalizeTransformFieldPath(parts[0])
			if fieldPath == "" {
				return nil, fmt.Errorf("transform set statement requires a target path: %s", line)
			}
			value := parseTransformValue(parts[1], workingContext, resolver)
			result = setDeepMappedValue(result, fieldPath, value)
			workingContext["data"] = result
			workingContext["payload"] = result
			if !strings.Contains(fieldPath, ".") {
				workingContext[fieldPath] = value
			}
		case strings.HasPrefix(line, "delete "):
			targets := splitTopLevel(strings.TrimSpace(strings.TrimPrefix(line, "delete ")), ',')
			for _, target := range targets {
				fieldPath := normalizeTransformFieldPath(target)
				if fieldPath == "" {
					continue
				}
				result = deleteDeepMappedValue(result, fieldPath)
				if !strings.Contains(fieldPath, ".") {
					delete(workingContext, fieldPath)
				}
			}
			workingContext["data"] = result
			workingContext["payload"] = result
		case strings.HasPrefix(line, "merge "):
			expression := strings.TrimSpace(strings.TrimPrefix(line, "merge "))
			value := parseTransformValue(expression, workingContext, resolver)
			objectValue, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("transform merge requires an object expression: %s", line)
			}
			for key, item := range objectValue {
				result[key] = item
				workingContext[key] = item
			}
			workingContext["data"] = result
			workingContext["payload"] = result
		default:
			return nil, fmt.Errorf("unsupported transform statement: %s", line)
		}
	}

	return result, nil
}

func evaluateTransformReturn(expression string, payload map[string]any, context map[string]any, resolver TemplateResolver) (map[string]any, error) {
	expression = strings.TrimSpace(expression)
	switch expression {
	case "", "data", "payload":
		return cloneMap(payload), nil
	}

	value := parseTransformValue(expression, context, resolver)
	objectValue, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("transform return must resolve to an object")
	}
	return objectValue, nil
}

func parseTransformValue(expression string, context map[string]any, resolver TemplateResolver) any {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil
	}

	if strings.HasPrefix(expression, "{") && strings.HasSuffix(expression, "}") {
		if value, err := parseTransformObject(expression, context, resolver); err == nil {
			return value
		}
	}
	if strings.HasPrefix(expression, "[") && strings.HasSuffix(expression, "]") {
		if value, err := parseTransformArray(expression, context, resolver); err == nil {
			return value
		}
	}

	return resolver.evaluateExpression(expression, context)
}

func parseTransformObject(expression string, context map[string]any, resolver TemplateResolver) (map[string]any, error) {
	inner := strings.TrimSpace(expression[1 : len(expression)-1])
	if inner == "" {
		return map[string]any{}, nil
	}

	result := map[string]any{}
	for _, token := range splitTopLevel(inner, ',') {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if strings.HasPrefix(token, "...") {
			spreadValue := parseTransformValue(strings.TrimSpace(strings.TrimPrefix(token, "...")), context, resolver)
			objectValue, ok := spreadValue.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("transform spread requires an object: %s", token)
			}
			for key, item := range objectValue {
				result[key] = item
			}
			continue
		}

		pair := splitTopLevelByOperator(token, ':')
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid transform object entry: %s", token)
		}
		key := normalizeTransformObjectKey(pair[0])
		if key == "" {
			return nil, fmt.Errorf("transform object entry requires a key: %s", token)
		}
		result[key] = parseTransformValue(pair[1], context, resolver)
	}
	return result, nil
}

func parseTransformArray(expression string, context map[string]any, resolver TemplateResolver) ([]any, error) {
	inner := strings.TrimSpace(expression[1 : len(expression)-1])
	if inner == "" {
		return []any{}, nil
	}

	values := make([]any, 0)
	for _, token := range splitTopLevel(inner, ',') {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if strings.HasPrefix(token, "...") {
			spreadValue := parseTransformValue(strings.TrimSpace(strings.TrimPrefix(token, "...")), context, resolver)
			arrayValue, ok := spreadValue.([]any)
			if !ok {
				return nil, fmt.Errorf("transform array spread requires a list: %s", token)
			}
			values = append(values, arrayValue...)
			continue
		}
		values = append(values, parseTransformValue(token, context, resolver))
	}
	return values, nil
}

func normalizedTransformLines(script string) []string {
	lines := strings.Split(script, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func splitTopLevel(input string, separator rune) []string {
	result := make([]string, 0)
	var current strings.Builder
	depthParen := 0
	depthBrace := 0
	depthBracket := 0
	quoted := rune(0)

	for _, char := range input {
		if quoted != 0 {
			current.WriteRune(char)
			if char == quoted {
				quoted = 0
			}
			continue
		}

		switch char {
		case '\'', '"':
			quoted = char
		case '(':
			depthParen++
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case '{':
			depthBrace++
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
		case '[':
			depthBracket++
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
		}

		if char == separator && depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(char)
	}

	result = append(result, strings.TrimSpace(current.String()))
	return result
}

func splitTopLevelByOperator(input string, separator rune) []string {
	parts := splitTopLevel(input, separator)
	if len(parts) <= 1 {
		return parts
	}
	return []string{parts[0], strings.TrimSpace(strings.Join(parts[1:], string(separator)))}
}

func normalizeTransformObjectKey(input string) string {
	trimmed := strings.TrimSpace(input)
	if isQuoted(trimmed) {
		return trimmed[1 : len(trimmed)-1]
	}
	return strings.TrimPrefix(trimmed, ".")
}

func normalizeTransformFieldPath(input string) string {
	return strings.Trim(strings.TrimSpace(input), ".")
}

func cloneTemplateContext(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = cloneAny(value)
	}
	return result
}

func cloneMap(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = cloneAny(value)
	}
	return result
}

func cloneAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, cloneAny(item))
		}
		return result
	default:
		return typed
	}
}

func setDeepMappedValue(source map[string]any, path string, value any) map[string]any {
	if path == "" {
		if typed, ok := value.(map[string]any); ok {
			return cloneMap(typed)
		}
		return source
	}
	result := cloneMap(source)
	keys := strings.Split(path, ".")
	ref := result
	for index := 0; index < len(keys)-1; index++ {
		key := keys[index]
		next, ok := ref[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			ref[key] = next
		}
		ref = next
	}
	ref[keys[len(keys)-1]] = value
	return result
}

func deleteDeepMappedValue(source map[string]any, path string) map[string]any {
	if path == "" {
		return map[string]any{}
	}
	result := cloneMap(source)
	keys := strings.Split(path, ".")
	ref := result
	for index := 0; index < len(keys)-1; index++ {
		next, ok := ref[keys[index]].(map[string]any)
		if !ok {
			return result
		}
		ref = next
	}
	delete(ref, keys[len(keys)-1])
	return result
}
