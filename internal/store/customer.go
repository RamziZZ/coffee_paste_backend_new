package store

import (
	"context"
	"database/sql"
	"time"
)

type CustomerStore struct {
	DB *sql.DB
}

type Customer struct {
	ID          int64
	Name        string
	Email       string
	Phone       string
	TotalOrders int64
	TotalSpend  int64
	JoinedAt    time.Time
}

func (s *CustomerStore) List(ctx context.Context) ([]Customer, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, email, COALESCE(phone, ''), total_orders, total_spend, joined_at
		   FROM v_customers
		  ORDER BY joined_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Customer, 0)
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.TotalOrders, &c.TotalSpend, &c.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
