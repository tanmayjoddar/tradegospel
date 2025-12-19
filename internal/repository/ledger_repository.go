package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Ledger struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type LedgerRepository struct {
	db *sql.DB
}

func NewLedgerRepository(db *sql.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) Create(ctx context.Context, amount float64, desc string, actor string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		"INSERT INTO ledger (amount, description) VALUES ($1, $2)",
		amount, desc,
	)
	if err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		id = 1
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO audit_ledger (ledger_id, actor, action) VALUES ($1, $2, $3)",
		id, actor, "INSERT",
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *LedgerRepository) List(ctx context.Context) ([]Ledger, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, amount, description, created_at FROM ledger ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ledger entries: %w", err)
	}
	defer rows.Close()

	var result []Ledger
	for rows.Next() {
		var l Ledger
		if err := rows.Scan(&l.ID, &l.Amount, &l.Description, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, l)
	}
	return result, nil
}

func (r *LedgerRepository) GetByID(ctx context.Context, id int64) (*Ledger, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, amount, description, created_at FROM ledger WHERE id=$1", id)

	var l Ledger
	if err := row.Scan(&l.ID, &l.Amount, &l.Description, &l.CreatedAt); err != nil {
		return nil, fmt.Errorf("ledger entry not found: %w", err)
	}
	return &l, nil
}
