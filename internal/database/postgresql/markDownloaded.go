package postgresql

import (
	"context"
	"database/sql"
	"fmt"
)

func MarkDownloaded(ctx context.Context, db *sql.DB, url, content string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE raw_content
		SET content = $2, is_download = TRUE, updated_at = now()
		WHERE url = $1
	`, url, content)
	if err != nil {
		return fmt.Errorf("mark downloaded: %w", err)
	}
	return nil
}
