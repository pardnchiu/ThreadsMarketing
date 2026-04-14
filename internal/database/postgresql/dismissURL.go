package postgresql

import (
	"context"
	"database/sql"
	"fmt"
)

func DismissURL(ctx context.Context, db *sql.DB, url string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE raw_content
		SET dismiss = TRUE, updated_at = now()
		WHERE url = $1
	`, url)
	if err != nil {
		return fmt.Errorf("dismiss url: %w", err)
	}
	return nil
}
