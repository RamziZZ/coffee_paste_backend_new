package store

import (
	"context"
	"database/sql"
	"errors"
)

type ExpenseStore struct {
	DB *sql.DB
}

type Expense struct {
	ID          int64  `json:"id"`
	Date        string `json:"date"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type ExpenseInput struct {
	Date        string // YYYY-MM-DD
	Name        string
	Category    string
	Amount      int64
	Description string
}

const expenseSelect = `SELECT e.id, DATE_FORMAT(e.date, '%Y-%m-%d'), e.name, e.category, e.amount, COALESCE(e.description, '')
	  FROM expenses e`

func scanExpense(r rowScanner) (Expense, error) {
	var x Expense
	err := r.Scan(&x.ID, &x.Date, &x.Name, &x.Category, &x.Amount, &x.Description)
	return x, err
}

func (s *ExpenseStore) List(ctx context.Context) ([]Expense, error) {
	rows, err := s.DB.QueryContext(ctx, expenseSelect+` ORDER BY e.date DESC, e.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Expense, 0)
	for rows.Next() {
		x, err := scanExpense(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *ExpenseStore) Get(ctx context.Context, id int64) (*Expense, error) {
	x, err := scanExpense(s.DB.QueryRowContext(ctx, expenseSelect+` WHERE e.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &x, nil
}

func (s *ExpenseStore) Create(ctx context.Context, in ExpenseInput, createdBy int64) (*Expense, error) {
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO expenses (date, name, category, amount, description, created_by)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		in.Date, in.Name, in.Category, in.Amount, in.Description, createdBy)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *ExpenseStore) Update(ctx context.Context, id int64, in ExpenseInput) (*Expense, error) {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE expenses SET date = ?, name = ?, category = ?, amount = ?, description = ? WHERE id = ?`,
		in.Date, in.Name, in.Category, in.Amount, in.Description, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(ctx, id)
}

func (s *ExpenseStore) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM expenses WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
