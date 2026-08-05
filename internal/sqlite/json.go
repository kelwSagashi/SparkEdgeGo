package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type stringSliceJSON []string

func (v stringSliceJSON) Value() (driver.Value, error) {
	if v == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]string(v))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (v *stringSliceJSON) Scan(value any) error {
	if value == nil {
		*v = stringSliceJSON{}
		return nil
	}

	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type mapJSON map[string]any

func (v mapJSON) Value() (driver.Value, error) {
	if v == nil {
		return "{}", nil
	}
	data, err := json.Marshal(map[string]any(v))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (v *mapJSON) Scan(value any) error {
	if value == nil {
		*v = mapJSON{}
		return nil
	}

	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func scanJSONBytes(value any) ([]byte, error) {
	switch raw := value.(type) {
	case []byte:
		return raw, nil
	case string:
		return []byte(raw), nil
	default:
		return nil, fmt.Errorf("unsupported json storage type %T", value)
	}
}
