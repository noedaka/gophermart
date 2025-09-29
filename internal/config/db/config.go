package dbconfig

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
            id BIGSERIAL PRIMARY KEY,
            current REAL DEFAULT 0,
            withdrawn REAL DEFAULT 0
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
        
        CREATE TABLE IF NOT EXISTS withdrawals (
            id BIGSERIAL PRIMARY KEY,
            user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            number VARCHAR(255) NOT NULL UNIQUE,
            sum REAL NOT NULL CHECK (sum >= 0),
            processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
    `)

	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	indexes := []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_id ON orders(user_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_status ON orders(status);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);",
	}

	for _, indexSQL := range indexes {
		_, err = db.Exec(indexSQL)
		if err != nil {
			return err
		}
	}

	return nil
}
