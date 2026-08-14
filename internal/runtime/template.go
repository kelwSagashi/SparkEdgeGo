package runtime

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil
	}

	if condition, whenTrue, whenFalse, ok := splitTernary(expression); ok {
		if truthy(resolver.evaluateCondition(condition, context)) {
			return resolver.evaluateExpression(whenTrue, context)
		}
		return resolver.evaluateExpression(whenFalse, context)
	}

	for _, part := range splitByOperator(expression, "||") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if value, ok := resolver.resolveLiteralOrFunction(part, context); ok {
			if isPresentTemplateValue(value) {
				return value
			}
			continue
		}

		value := resolver.ResolvePath(part, context)
		if isPresentTemplateValue(value) {
			return value
		}
	}
	return nil
}

func (resolver TemplateResolver) evaluateCondition(expression string, context map[string]any) any {
	expression = strings.TrimSpace(expression)
	for _, operator := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		parts := splitByOperator(expression, operator)
		if len(parts) == 2 {
			left := resolver.evaluateExpression(parts[0], context)
			right := resolver.evaluateExpression(parts[1], context)
			return compareValues(left, right, operator)
		}
	}
	if value, ok := resolver.resolveLiteralOrFunction(expression, context); ok {
		return value
	}
	return resolver.ResolvePath(expression, context)
}

func (resolver TemplateResolver) resolveLiteralOrFunction(expression string, context map[string]any) (any, bool) {
	if isQuoted(expression) {
		return expression[1 : len(expression)-1], true
	}
	if expression == "true" {
		return true, true
	}
	if expression == "false" {
		return false, true
	}
	if expression == "null" {
		return nil, true
	}
	if number, ok := parseNumberLiteral(expression); ok {
		return number, true
	}
	if name, args, ok := parseFunctionCall(expression); ok {
		return resolver.callFunction(name, args, context), true
	}
	return nil, false
}

func (resolver TemplateResolver) callFunction(name string, args []string, context map[string]any) any {
	evaluated := make([]any, 0, len(args))
	for _, arg := range args {
		evaluated = append(evaluated, resolver.evaluateExpression(arg, context))
	}

	switch strings.ToLower(strings.TrimSpace(name)) {
	case "concat":
		var builder strings.Builder
		for _, arg := range evaluated {
			builder.WriteString(stringifyTemplateValue(arg))
		}
		return builder.String()
	case "upper":
		return strings.ToUpper(stringifyTemplateValue(firstValue(evaluated)))
	case "lower":
		return strings.ToLower(stringifyTemplateValue(firstValue(evaluated)))
	case "trim":
		return strings.TrimSpace(stringifyTemplateValue(firstValue(evaluated)))
	case "default", "coalesce":
		for _, arg := range evaluated {
			if isPresentTemplateValue(arg) {
				return arg
			}
		}
		return nil
	case "if":
		if len(args) < 3 {
			return nil
		}
		if truthy(resolver.evaluateCondition(args[0], context)) {
			return resolver.evaluateExpression(args[1], context)
		}
		return resolver.evaluateExpression(args[2], context)
	case "eq":
		return compareValues(firstValue(evaluated), secondValue(evaluated), "==")
	case "ne":
		return compareValues(firstValue(evaluated), secondValue(evaluated), "!=")
	case "gt":
		return compareValues(firstValue(evaluated), secondValue(evaluated), ">")
	case "gte":
		return compareValues(firstValue(evaluated), secondValue(evaluated), ">=")
	case "lt":
		return compareValues(firstValue(evaluated), secondValue(evaluated), "<")
	case "lte":
		return compareValues(firstValue(evaluated), secondValue(evaluated), "<=")
	case "not":
		return !truthy(firstValue(evaluated))
	case "and":
		for _, arg := range evaluated {
			if !truthy(arg) {
				return false
			}
		}
		return true
	case "or":
		for _, arg := range evaluated {
			if truthy(arg) {
				return true
			}
		}
		return false
	case "add":
		return reduceNumbers(evaluated, func(acc, current float64) float64 { return acc + current })
	case "sub":
		return reduceNumbers(evaluated, func(acc, current float64) float64 { return acc - current })
	case "mul":
		return reduceNumbers(evaluated, func(acc, current float64) float64 { return acc * current })
	case "div":
		if len(evaluated) < 2 {
			return nil
		}
		dividend, ok := toFloat(firstValue(evaluated))
		if !ok {
			return nil
		}
		for _, arg := range evaluated[1:] {
			value, ok := toFloat(arg)
			if !ok || value == 0 {
				return nil
			}
			dividend /= value
		}
		return dividend
	case "json":
		if data, err := json.Marshal(firstValue(evaluated)); err == nil {
			return string(data)
		}
		return nil
	case "string":
		return stringifyTemplateValue(firstValue(evaluated))
	case "number":
		if value, ok := toFloat(firstValue(evaluated)); ok {
			return value
		}
		return nil
	case "bool":
		return truthy(firstValue(evaluated))
	case "now":
		format := time.RFC3339
		if len(evaluated) > 0 && strings.TrimSpace(stringifyTemplateValue(evaluated[0])) != "" {
			format = stringifyTemplateValue(evaluated[0])
		}
		return time.Now().UTC().Format(format)
	case "date":
		if len(evaluated) == 0 {
			return nil
		}
		value := stringifyTemplateValue(evaluated[0])
		if value == "" {
			return nil
		}
		format := time.RFC3339
		if len(evaluated) > 1 && strings.TrimSpace(stringifyTemplateValue(evaluated[1])) != "" {
			format = stringifyTemplateValue(evaluated[1])
		}
		timestamp, err := parseTimeValue(value)
		if err != nil {
			return nil
		}
		return timestamp.UTC().Format(format)
	case "hash":
		if len(evaluated) == 0 {
			return nil
		}
		value := stringifyTemplateValue(evaluated[0])
		algorithm := "sha256"
		if len(evaluated) > 1 {
			algorithm = strings.ToLower(strings.TrimSpace(stringifyTemplateValue(evaluated[1])))
		}
		switch algorithm {
		case "sha1":
			sum := sha1.Sum([]byte(value))
			return hex.EncodeToString(sum[:])
		default:
			sum := sha256.Sum256([]byte(value))
			return hex.EncodeToString(sum[:])
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

func splitByOperator(value string, operator string) []string {
	var parts []string
	depth := 0
	inSingle := false
	inDouble := false
	start := 0

	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				depth++
			}
		case ')':
			if !inSingle && !inDouble && depth > 0 {
				depth--
			}
		}
		if !inSingle && !inDouble && depth == 0 && strings.HasPrefix(value[i:], operator) {
			parts = append(parts, strings.TrimSpace(value[start:i]))
			i += len(operator) - 1
			start = i + 1
		}
	}
	if len(parts) == 0 {
		return []string{value}
	}
	parts = append(parts, strings.TrimSpace(value[start:]))
	return parts
}

func splitTernary(value string) (string, string, string, bool) {
	depth := 0
	inSingle := false
	inDouble := false
	question := -1
	colon := -1

	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				depth++
			}
		case ')':
			if !inSingle && !inDouble && depth > 0 {
				depth--
			}
		case '?':
			if !inSingle && !inDouble && depth == 0 && question == -1 {
				question = i
			}
		case ':':
			if !inSingle && !inDouble && depth == 0 && question != -1 {
				colon = i
				return strings.TrimSpace(value[:question]), strings.TrimSpace(value[question+1 : colon]), strings.TrimSpace(value[colon+1:]), true
			}
		}
	}
	return "", "", "", false
}

func parseFunctionCall(expression string) (string, []string, bool) {
	open := strings.IndexByte(expression, '(')
	if open <= 0 || !strings.HasSuffix(strings.TrimSpace(expression), ")") {
		return "", nil, false
	}
	name := strings.TrimSpace(expression[:open])
	argsText := strings.TrimSpace(expression[open+1 : len(expression)-1])
	if name == "" {
		return "", nil, false
	}
	if argsText == "" {
		return name, []string{}, true
	}
	return name, splitArguments(argsText), true
}

func splitArguments(value string) []string {
	var parts []string
	depth := 0
	inSingle := false
	inDouble := false
	start := 0

	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				depth++
			}
		case ')':
			if !inSingle && !inDouble && depth > 0 {
				depth--
			}
		case ',':
			if !inSingle && !inDouble && depth == 0 {
				parts = append(parts, strings.TrimSpace(value[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(value[start:]))
	return parts
}

func firstValue(values []any) any {
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func secondValue(values []any) any {
	if len(values) < 2 {
		return nil
	}
	return values[1]
}

func truthy(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case bool:
		return typed
	case string:
		return strings.TrimSpace(typed) != "" && typed != "false" && typed != "0"
	case int:
		return typed != 0
	case int64:
		return typed != 0
	case float64:
		return typed != 0
	case []any:
		return len(typed) > 0
	case map[string]any:
		return len(typed) > 0
	default:
		return true
	}
}

func compareValues(left any, right any, operator string) bool {
	if leftNumber, ok := toFloat(left); ok {
		if rightNumber, ok := toFloat(right); ok {
			switch operator {
			case "==":
				return leftNumber == rightNumber
			case "!=":
				return leftNumber != rightNumber
			case ">":
				return leftNumber > rightNumber
			case ">=":
				return leftNumber >= rightNumber
			case "<":
				return leftNumber < rightNumber
			case "<=":
				return leftNumber <= rightNumber
			}
		}
	}
	leftText := stringifyTemplateValue(left)
	rightText := stringifyTemplateValue(right)
	switch operator {
	case "==":
		return leftText == rightText
	case "!=":
		return leftText != rightText
	case ">":
		return leftText > rightText
	case ">=":
		return leftText >= rightText
	case "<":
		return leftText < rightText
	case "<=":
		return leftText <= rightText
	default:
		return false
	}
}

func toFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func reduceNumbers(values []any, operation func(float64, float64) float64) any {
	if len(values) == 0 {
		return nil
	}
	acc, ok := toFloat(values[0])
	if !ok {
		return nil
	}
	for _, value := range values[1:] {
		number, ok := toFloat(value)
		if !ok {
			return nil
		}
		acc = operation(acc, number)
	}
	return acc
}

func parseTimeValue(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time value: %s", value)
}
