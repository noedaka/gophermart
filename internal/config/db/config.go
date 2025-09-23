package dbConfig

import "database/sql"

func InitDB(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY  
		);

		CREATE TABLE IF NOT EXISTS users_credentials (
			user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			login VARCHAR(128) NOT NULL UNIQUE,
			password VARCHAR(128) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS orders (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			number VARCHAR(255) NOT NULL UNIQUE,
			uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			status VARCHAR(50),
			accrual REAL DEFAULT 0
		);
	`)

	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
