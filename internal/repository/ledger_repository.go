package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Ledger struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

type LedgerRepository struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) Create(ctx context.Context, amount float64, desc string) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO ledger (amount, description) VALUES ($1, $2)",
		amount, desc,
	)
	return err
}

func (r *LedgerRepository) List(ctx context.Context) ([]Ledger, error) {
	rows, err := r.db.Query(ctx, "SELECT id, amount, description, created_at FROM ledger ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Ledger
	for rows.Next() {
		var l Ledger
		if err := rows.Scan(&l.ID, &l.Amount, &l.Description, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (r *LedgerRepository) GetByID(ctx context.Context, id int64) (*Ledger, error) {
	row := r.db.QueryRow(ctx,
		"SELECT id, amount, description, created_at FROM ledger WHERE id=$1", id)

	var l Ledger
	if err := row.Scan(&l.ID, &l.Amount, &l.Description, &l.CreatedAt); err != nil {
		return nil, err
	}
	return &l, nil
}
