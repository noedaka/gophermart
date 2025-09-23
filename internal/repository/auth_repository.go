package repository

import (
	"context"
	"database/sql"
	"errors"
	"gophermart/internal/auth"
	"gophermart/internal/model"
)

func (repo *Repository) GetUserByCredentials(ctx context.Context, userCredentials model.UserCredentials) (int64, error) {
	var userFromDB model.UserCredentials
	err := repo.db.QueryRowContext(ctx,
		"SELECT user_id, password FROM users_credentials WHERE login = $1", userCredentials.Login,
	).Scan(&userFromDB.ID, &userFromDB.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, nil
		}

		return -1, err
	}

	isPasswordCorrect := auth.CheckPasswordHash(userCredentials.Password, userFromDB.Password)
	if isPasswordCorrect {
		return userFromDB.ID, nil
	}

	return -1, nil
}

func (repo *Repository) CreateUserWithCredentials(ctx context.Context, userCredentials model.UserCredentials) (int64, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	isFree, err := repo.isLoginFree(ctx, userCredentials.Login)
	if err != nil {
		return -1, err
	}

	if !isFree {
		return -1, nil
	}

	userCredentials.ID, err = repo.createNewUser(ctx, tx)
	if err != nil {
		return -1, err
	}

	hashedPassword, err := auth.HashPassword(userCredentials.Password)
	if err != nil {
		return -1, err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO users_credentials(user_id, login, password) VALUES ($1, $2, $3)",
		userCredentials.ID, userCredentials.Login, hashedPassword)
	if err != nil {
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}

	return userCredentials.ID, nil
}

func (repo *Repository) createNewUser(ctx context.Context, tx *sql.Tx) (int64, error) {
	var userID int64
	err := tx.QueryRowContext(ctx, "INSERT INTO users DEFAULT VALUES RETURNING id").Scan(&userID)
	if err != nil {
		return -1, err
	}

	return userID, nil
}

func (repo *Repository) isLoginFree(ctx context.Context, login string) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users_credentials WHERE login = $1", login,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 1 {
		return false, nil
	} else {
		return true, nil
	}
}