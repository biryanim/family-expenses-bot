package repository

import (
	"context"
	"fmt"
)

func (r *Repository) ListUserIDs(ctx context.Context) ([]int64, error) {
	const sql = "SELECT telegram_id FROM users ORDER BY telegram_id"

	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return ids, nil
}
