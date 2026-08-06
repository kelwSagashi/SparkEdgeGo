package runtime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templatePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

type TemplateResolver struct{}

func (resolver TemplateResolver) Resolve(input any, context map[string]any) any {
	switch value := input.(type) {
	case string:
		if isWholeTemplate(value) {
			return resolver.evaluateExpression(strings.TrimSpace(value[2:len(value)-2]), context)
		}
		return resolver.resolveString(value, context)
	case []any:
		result := make([]any, 0, len(value))
		for _, item := range value {
			result = append(result, resolver.Resolve(item, context))
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(value))
		for key, item := range value {
			result[key] = resolver.Resolve(item, context)
		}
		return result
	default:
		return input
	}
}

func (resolver TemplateResolver) ResolvePath(path string, context map[string]any) any {
	if strings.HasPrefix(strings.TrimSpace(path), "$") {
		return resolver.resolveJSONPath(path, context)
	}
	return resolver.getDeepValue(context, path)
}

func (resolver TemplateResolver) resolveString(value string, context map[string]any) string {
	return templatePattern.ReplaceAllStringFunc(value, func(match string) string {
		groups := templatePattern.FindStringSubmatch(match)
		if len(groups) != 2 {
			return match
		}
		resolved := resolver.evaluateExpression(strings.TrimSpace(groups[1]), context)
		if resolved == nil {
			return match
		}
		return stringifyTemplateValue(resolved)
	})
}

func (resolver TemplateResolver) evaluateExpression(expression string, context map[string]any) any {
	for _, part := range strings.Split(expression, "||") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if isQuoted(part) {
			return part[1 : len(part)-1]
		}
		if number, ok := parseNumberLiteral(part); ok {
			return number
		}

		value := resolver.ResolvePath(part, context)
		if isPresentTemplateValue(value) {
			return value
		}
	}
	return nil
}

func (resolver TemplateResolver) resolveJSONPath(path string, context map[string]any) any {
	normalized := strings.TrimSpace(path)
	if normalized == "$" {
		return context
	}
	if strings.HasPrefix(normalized, "$.") {
		normalized = strings.TrimPrefix(normalized, "$.")
	} else if strings.HasPrefix(normalized, "$[") {
		normalized = strings.TrimPrefix(normalized, "$")
	} else {
		return nil
	}
	parts, ok := parsePathParts(normalized)
	if !ok {
		return nil
	}
	return resolveParts(context, parts)
}

func (TemplateResolver) getDeepValue(context map[string]any, path string) any {
	parts, ok := parsePathParts(strings.TrimSpace(path))
	if !ok {
		return nil
	}
	return resolveParts(context, parts)
}

func isWholeTemplate(value string) bool {
	if !strings.HasPrefix(value, "{{") || !strings.HasSuffix(value, "}}") {
		return false
	}
	return !strings.Contains(value[2:len(value)-2], "{{")
}

func isQuoted(value string) bool {
	if len(value) < 2 {
		return false
	}
	return (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) || (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`))
}

func parseNumberLiteral(value string) (any, bool) {
	if value == "" {
		return nil, false
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, false
	}
	return number, true
}

func isPresentTemplateValue(value any) bool {
	if value == nil {
		return false
	}
	if text, ok := value.(string); ok && text == "" {
		return false
	}
	return true
}

func stringifyTemplateValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func parsePathParts(path string) ([]string, bool) {
	if path == "" {
		return nil, false
	}
	var parts []string
	for i := 0; i < len(path); {
		switch path[i] {
		case '.':
			i++
		case '[':
			end := strings.IndexByte(path[i:], ']')
			if end < 0 {
				return nil, false
			}
			token := strings.TrimSpace(path[i+1 : i+end])
			token = strings.Trim(token, `"'`)
			if token == "" {
				return nil, false
			}
			parts = append(parts, token)
			i += end + 1
		default:
			nextDot := strings.IndexByte(path[i:], '.')
			nextBracket := strings.IndexByte(path[i:], '[')
			next := len(path)
			if nextDot >= 0 {
				next = i + nextDot
			}
			if nextBracket >= 0 && i+nextBracket < next {
				next = i + nextBracket
			}
			token := strings.TrimSpace(path[i:next])
			if token == "" {
				return nil, false
			}
			parts = append(parts, token)
			i = next
		}
	}
	return parts, true
}

func resolveParts(root any, parts []string) any {
	current := root
	for _, part := range parts {
		switch typed := current.(type) {
		case map[string]any:
			current = typed[part]
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(typed) {
				return nil
			}
			current = typed[index]
		default:
			return nil
		}
		if current == nil {
			return nil
		}
	}
	return current
}
