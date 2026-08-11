package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
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

type deviceOtherFieldsJSON []domain.DeviceOtherField

func (v deviceOtherFieldsJSON) Value() (driver.Value, error) {
	if v == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]domain.DeviceOtherField(v))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

type mappingCustomFieldsJSON []domain.MappingCustomField

func (v mappingCustomFieldsJSON) Value() (driver.Value, error) {
	if v == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]domain.MappingCustomField(v))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (v *mappingCustomFieldsJSON) Scan(value any) error {
	if value == nil {
		*v = mappingCustomFieldsJSON{}
		return nil
	}

	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type executionLogsJSON []domain.ExecutionLog

func (v executionLogsJSON) Value() (driver.Value, error) {
	if v == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]domain.ExecutionLog(v))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (v *executionLogsJSON) Scan(value any) error {
	if value == nil {
		*v = executionLogsJSON{}
		return nil
	}

	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type executionDestinationDetailsJSON []domain.ExecutionDestinationDetail

func (v executionDestinationDetailsJSON) Value() (driver.Value, error) {
	if v == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]domain.ExecutionDestinationDetail(v))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (v *executionDestinationDetailsJSON) Scan(value any) error {
	if value == nil {
		*v = executionDestinationDetailsJSON{}
		return nil
	}

	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type authTypeFieldsJSON []domain.AuthTypeField

func (v authTypeFieldsJSON) Value() (driver.Value, error) {
	if v == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]domain.AuthTypeField(v))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (v *authTypeFieldsJSON) Scan(value any) error {
	if value == nil {
		*v = authTypeFieldsJSON{}
		return nil
	}

	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (v *deviceOtherFieldsJSON) Scan(value any) error {
	if value == nil {
		*v = deviceOtherFieldsJSON{}
		return nil
	}

	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
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
