package repository

import (
	"context"
	"database/sql"
	"errors"
	"gophermart/internal/model"
	"log"
)

func (repo *Repository) GetOrdersForProcessing(ctx context.Context, limit int) ([]model.Order, error) {
	rows, err := repo.db.QueryContext(ctx, `
        SELECT number, status, accrual 
        FROM orders 
        WHERE status NOT IN ('PROCESSED', 'INVALID')
        ORDER BY uploaded_at ASC
        LIMIT $1
    `, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return orders, nil
}

func (repo *Repository) UpdateOrderAccrual(ctx context.Context, order model.AccrualResponse) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
    if err := tx.Rollback(); err != nil {
        if !errors.Is(err, sql.ErrTxDone) {
            log.Printf("failed to rollback the transaction: %v", err)
        }
    }
	}()

	_, err = tx.ExecContext(ctx, `
        UPDATE orders SET status = $1, accrual = $2	
        WHERE number = $3
    `, order.Status, order.Accrual, order.Order)
	if err != nil {
		return err
	}

	if order.Status == "PROCESSED" && order.Accrual > 0 {
		_, err = tx.ExecContext(ctx, `
            UPDATE users SET current = current + $1 
            WHERE id = (SELECT user_id FROM orders WHERE number = $2)
        `, order.Accrual, order.Order)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
