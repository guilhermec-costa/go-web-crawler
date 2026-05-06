package infra

func (d *CrawlerExtractionSQLiteStore) Migrate() error {
	return nil
}
func (d *UserSQLiteStore) Migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);	

		CREATE TABLE IF NOT EXISTS crawler_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id int,
			extraction TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME
		);

		ALTER TABLE crawler_jobs
		ADD CONSTRAINT fk_user_id
		FOREIGN KEY (user_id)
		REFERENCES users(id);
	`)

	return err
}
