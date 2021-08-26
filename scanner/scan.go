package scanner

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

var RecordNotFound = errors.New("record is not found")

func Scan(ctx context.Context, rows *sql.Rows, v interface{}) error {
	if rows == nil {
		return nil
	}
	defer rows.Close()

	si, err := ScanIteratorFor(v)
	if err != nil {
		return err
	}

	for rows.Next() {
		item := si.New()

		if scanErr := scanTo(context.Background(), rows, item); scanErr != nil {
			return scanErr
		}

		if err := si.Next(item); err != nil {
			return err
		}
	}

	if mustHasRecord, ok := si.(interface{ MustHasRecord() bool }); ok {
		if !mustHasRecord.MustHasRecord() {
			return RecordNotFound
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Make sure the query can be processed to completion with no errors.
	if err := rows.Close(); err != nil {
		return err
	}

	return nil
}
