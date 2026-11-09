package mysqlstore

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/LeeJejune/go-school/internal/reading"
	"github.com/go-sql-driver/mysql"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Create(ctx context.Context, params reading.CreateParams) (reading.List, error) {
	if len(params.Resources) == 0 {
		return reading.List{}, errors.New("a reading list needs at least one resource")
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return reading.List{}, err
	}
	listID := hex.EncodeToString(id)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return reading.List{}, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO reading_lists (id, request_key, fingerprint, title) VALUES (?, ?, ?, ?)`,
		listID, params.RequestKey, params.Fingerprint, params.Title)
	if err != nil {
		var duplicate *mysql.MySQLError
		if errors.As(err, &duplicate) && duplicate.Number == 1062 {
			// Release the transaction before asking the pool for another connection.
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return reading.List{}, rollbackErr
			}
			return s.replay(ctx, params)
		}
		return reading.List{}, fmt.Errorf("insert list: %w", err)
	}
	for position, resource := range params.Resources {
		_, err = tx.ExecContext(ctx, `INSERT INTO reading_items
			(list_id, resource_id, title, url, position) VALUES (?, ?, ?, ?, ?)`,
			listID, resource.ID, resource.Title, resource.URL, position)
		if err != nil {
			return reading.List{}, fmt.Errorf("insert item: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		// A network error here does not prove rollback. Same-key replay is safe.
		return reading.List{}, fmt.Errorf("commit list: %w", err)
	}
	return s.Get(ctx, listID)
}

func (s *Store) replay(ctx context.Context, params reading.CreateParams) (reading.List, error) {
	var id, fingerprint string
	err := s.db.QueryRowContext(ctx, `SELECT id, fingerprint FROM reading_lists WHERE request_key = ?`, params.RequestKey).Scan(&id, &fingerprint)
	if err != nil {
		return reading.List{}, fmt.Errorf("find replay: %w", err)
	}
	if fingerprint != params.Fingerprint {
		return reading.List{}, reading.ErrConflict
	}
	return s.Get(ctx, id)
}

func (s *Store) Get(ctx context.Context, id string) (reading.List, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT l.id, l.title, l.created_at,
		i.resource_id, i.title, i.url, i.is_read
		FROM reading_lists l JOIN reading_items i ON l.id = i.list_id
		WHERE l.id = ? ORDER BY i.position`, id)
	if err != nil {
		return reading.List{}, err
	}
	defer rows.Close()
	var list reading.List
	for rows.Next() {
		var item reading.Item
		if err := rows.Scan(&list.ID, &list.Title, &list.CreatedAt,
			&item.Resource.ID, &item.Resource.Title, &item.Resource.URL, &item.IsRead); err != nil {
			return reading.List{}, err
		}
		list.Items = append(list.Items, item)
	}
	if err := rows.Err(); err != nil {
		return reading.List{}, err
	}
	if list.ID == "" {
		return reading.List{}, reading.ErrNotFound
	}
	return list, nil
}

func (s *Store) SetRead(ctx context.Context, listID, resourceID string, isRead bool) error {
	result, err := s.db.ExecContext(ctx, `UPDATE reading_items SET is_read = ? WHERE list_id = ? AND resource_id = ?`, isRead, listID, resourceID)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed > 0 {
		return err
	}
	// MySQL normally counts changed rows. Setting true twice changes zero rows,
	// which is different from a missing item.
	var exists int
	err = s.db.QueryRowContext(ctx, `SELECT 1 FROM reading_items WHERE list_id = ? AND resource_id = ?`, listID, resourceID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return reading.ErrNotFound
	}
	return err
}
