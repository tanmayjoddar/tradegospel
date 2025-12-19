package repository

import (
	"context"
	"fmt"

	"github.com/prisma/prisma-client-go"
)

type Ledger struct {
	ID          int     `json:"id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

type LedgerRepository struct {
	client prisma.Client
}

func NewLedgerRepository(client prisma.Client) *LedgerRepository {
	return &LedgerRepository{client: client}
}

func (r *LedgerRepository) Create(ctx context.Context, amount float64, desc string, actor string) error {
	ledger, err := r.client.Ledger.CreateOne(
		prisma.Ledger.Amount.Set(amount),
		prisma.Ledger.Description.Set(desc),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}

	_, err = r.client.AuditLog.CreateOne(
		prisma.AuditLog.LedgerID.Set(ledger.ID),
		prisma.AuditLog.Actor.Set(actor),
		prisma.AuditLog.Action.Set("INSERT"),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (r *LedgerRepository) List(ctx context.Context) ([]Ledger, error) {
	entries, err := r.client.Ledger.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ledger entries: %w", err)
	}

	var result []Ledger
	for _, entry := range entries {
		result = append(result, Ledger{
			ID:          entry.ID,
			Amount:      entry.Amount,
			Description: entry.Description,
			CreatedAt:   entry.CreatedAt.String(),
		})
	}
	return result, nil
}

func (r *LedgerRepository) GetByID(ctx context.Context, id int64) (*Ledger, error) {
	entry, err := r.client.Ledger.FindUnique(
		prisma.Ledger.ID.Equals(int(id)),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("ledger entry not found: %w", err)
	}

	return &Ledger{
		ID:          entry.ID,
		Amount:      entry.Amount,
		Description: entry.Description,
		CreatedAt:   entry.CreatedAt.String(),
	}, nil
}
