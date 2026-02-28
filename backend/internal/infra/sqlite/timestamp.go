package sqlite

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// builder is the Squirrel statement builder configured for SQLite (? placeholders).
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

// TimestampDest is a sql.Scanner that handles timestamps from SQLite (stored as
// TEXT "2006-01-02 15:04:05") and, as a safety net, time.Time from any driver.
type TimestampDest struct{ t time.Time }

func (d *TimestampDest) Scan(src interface{}) error {
	switch v := src.(type) {
	case time.Time:
		d.t = v
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return fmt.Errorf("TimestampDest: parse string %q: %w", v, err)
		}
		d.t = t.UTC()
	case []byte:
		t, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return fmt.Errorf("TimestampDest: parse bytes: %w", err)
		}
		d.t = t.UTC()
	case nil:
		d.t = time.Time{}
	default:
		return fmt.Errorf("TimestampDest: unsupported type %T", src)
	}
	return nil
}

func (d *TimestampDest) Time() time.Time { return d.t }
