package repository

import (
	"context"
	"database/sql"
	"gophermart/internal/config"
	"gophermart/internal/model"
	"time"
)

func (repo *Repository) CreateOrder(ctx context.Context, userID int64, orderNumber string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var existingUserID int64
	err = tx.QueryRowContext(ctx,
		"SELECT user_id FROM orders WHERE number = $1",
		orderNumber,
	).Scan(&existingUserID)

	if err == nil {
		if existingUserID == userID {
			return config.ErrOrderAlreadyUploadedByUser
		}
		return config.ErrOrderAlreadyUploadedByAnother
	} else if err != sql.ErrNoRows {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)",
		userID, orderNumber, "NEW",
	)

	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) GetOrdersByUserID(ctx context.Context, userID int64) ([]model.Order, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	
	rows, err := tx.QueryContext(ctx,
		`SELECT number, uploaded_at, status, accrual 
        FROM orders 
        WHERE user_id = $1 
        ORDER BY uploaded_at DESC`, userID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		var uploadedAt time.Time

		err = rows.Scan(
			&order.Number,
			&uploadedAt,
			&order.Status,
			&order.Accrual,
		)

		if err != nil {
			return nil, err
		}

		order.UploadedAt = uploadedAt.Format(time.RFC3339)
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, config.ErrNoOrders
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	
	return orders, nil
}
