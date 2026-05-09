package infra

type Migrator interface {
	Migrate() error
}

func (d *CrawlerExtractionSQLiteStore) Migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS crawler_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id int,
			extraction TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME DEFAULT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)

	return err
}

func (d *UserSQLiteStore) Migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);	
	`)

	return err
}
