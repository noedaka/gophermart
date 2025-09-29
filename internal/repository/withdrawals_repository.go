package repository

import (
	"context"
	"database/sql"
	"gophermart/internal/config"
	"gophermart/internal/model"
	"time"
)

func (repo *Repository) CreateWithdrawal(ctx context.Context, userID int64, withdrawal model.Withdrawal) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var currentBalance float32
	err = tx.QueryRowContext(ctx,
		"SELECT current FROM users WHERE id = $1 FOR UPDATE",
		userID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	if currentBalance < withdrawal.Sum {
		return config.ErrNotEnoughMoney
	}

	var existingOrder string
	err = tx.QueryRowContext(ctx,
		"SELECT number FROM withdrawals WHERE number = $1",
		withdrawal.Number).Scan(&existingOrder)
	if err == nil {
		return config.ErrOrderAlreadyUploadedByUser
	} else if err != sql.ErrNoRows {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users 
         SET current = current - $1, withdrawn = withdrawn + $1 
         WHERE id = $2`,
		withdrawal.Sum, userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO withdrawals (user_id, number, sum) 
         VALUES ($1, $2, $3)`,
		userID, withdrawal.Number, withdrawal.Sum)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) GetWithdrawals(ctx context.Context, userID int64) ([]model.WithdrawalData, error) {
	rows, err := repo.db.QueryContext(ctx,
		`SELECT number, sum, processed_at 
        FROM withdrawals 
        WHERE user_id = $1 
        ORDER BY processed_at DESC`, userID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []model.WithdrawalData
	for rows.Next() {
		var withdrawal model.WithdrawalData
		var processedAt time.Time

		err = rows.Scan(
			&withdrawal.Number,
			&withdrawal.Sum,
			&processedAt,
		)

		if err != nil {
			return nil, err
		}

		withdrawal.ProcessedAt = processedAt.Format(time.RFC3339)
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, config.ErrNoOrders
	}

	return withdrawals, nil
}