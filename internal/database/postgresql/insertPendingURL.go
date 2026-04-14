package postgresql

import (
	"context"
	"database/sql"
	"fmt"
)

func InsertPendingURL(ctx context.Context, db *sql.DB, url, contentType string) (bool, error) {
	res, err := db.ExecContext(ctx, `
		INSERT INTO raw_content (url, type, content, is_download)
		VALUES ($1, $2, '', FALSE)
		ON CONFLICT (url) DO NOTHING
	`, url, contentType)
	if err != nil {
		return false, fmt.Errorf("insert pending url: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return n > 0, nil
}
