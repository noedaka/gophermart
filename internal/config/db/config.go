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
			password TEXT NOT NULL
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
