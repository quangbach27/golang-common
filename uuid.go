package common

import (
	"database/sql/driver"
	"fmt"
	"uuid"
)

type UUID [16]byte

func NewUUIDv7() UUID {
	return UUID(uuid.NewV7())
}

func MustUUIDFromString(s string) UUID {
	return UUID(uuid.MustParse(s))
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) MarshalText() ([]byte, error) {
	return uuid.UUID(u).MarshalText()
}

func (u *UUID) UnmarshalText(data []byte) error {
	var parsed uuid.UUID
	if err := parsed.UnmarshalText(data); err != nil {
		return err
	}

	*u = UUID(parsed)
	return nil
}

func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

func (u *UUID) Scan(src any) error {
	var s string
	switch v := src.(type) {
	case nil:
		return nil
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("Scan: unable to scan type %T into UUID", src)
	}

	parsed, err := uuid.Parse(s)
	if err != nil {
		return fmt.Errorf("Scan: %w", err)
	}

	*u = UUID(parsed)
	return nil
}

func (u UUID) IsZero() bool {
	return u == UUID(uuid.Nil())
}
