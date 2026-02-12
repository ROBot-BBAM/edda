package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// NullUUID represents a UUID that may be null.
type NullUUID struct {
	UUID  uuid.UUID
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (nu *NullUUID) Scan(value interface{}) error {
	if value == nil {
		nu.UUID, nu.Valid = uuid.Nil, false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return nu.scanBytes(v)
	case string:
		return nu.scanBytes([]byte(v))
	default:
		return fmt.Errorf("cannot scan %T into NullUUID", value)
	}
}

func (nu *NullUUID) scanBytes(src []byte) error {
	if len(src) == 0 {
		nu.UUID, nu.Valid = uuid.Nil, false
		return nil
	}

	id, err := uuid.Parse(string(src))
	if err != nil {
		return err
	}

	nu.UUID, nu.Valid = id, true
	return nil
}

// Value implements the driver.Valuer interface.
func (nu NullUUID) Value() (driver.Value, error) {
	if !nu.Valid {
		return nil, nil
	}
	return nu.UUID.String(), nil
}

// MarshalJSON implements json.Marshaler.
func (nu NullUUID) MarshalJSON() ([]byte, error) {
	if !nu.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nu.UUID.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (nu *NullUUID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nu.UUID, nu.Valid = uuid.Nil, false
		return nil
	}

	id, err := uuid.Parse(string(data[1 : len(data)-1])) // Remove quotes
	if err != nil {
		return err
	}

	nu.UUID, nu.Valid = id, true
	return nil
}
