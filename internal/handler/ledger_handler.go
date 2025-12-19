package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ledger-go-system/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerHandler struct {
	repo *repository.LedgerRepository
}

func NewLedgerHandler(db *pgxpool.Pool) *LedgerHandler {
	return &LedgerHandler{
		repo: repository.NewLedgerRepository(db),
	}
}

func (h *LedgerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), body.Amount, body.Description); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *LedgerHandler) List(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(data)
}

func (h *LedgerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	data, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(data)
}
