
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB handles PostgreSQL JSONB columns.
// Declared once here — do NOT redeclare in car.go or driver.go
type JSONB json.RawMessage


func (j JSONB) MarshalJSON() ([]byte, error) {
    if len(j) == 0 {
        return []byte("null"), nil
    }
    return json.RawMessage(j).MarshalJSON()
}

func (j *JSONB) Scan(value interface{}) error {
    if value == nil {
        *j = nil
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to scan JSONB: expected []byte, got %T", value)
    }

    var result json.RawMessage
    if err := json.Unmarshal(bytes, &result); err != nil {
        return fmt.Errorf("failed to unmarshal JSONB: %w", err)
    }

    *j = JSONB(result)
    return nil
}

func (j JSONB) Value() (driver.Value, error) {
    if len(j) == 0 {
        return nil, nil
    }
    b, err := json.RawMessage(j).MarshalJSON()
    if err != nil {
        return nil, err
    }
    return string(b), nil
}