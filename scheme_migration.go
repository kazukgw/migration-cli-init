package migration

import (
	"database/sql"
	"time"
)

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS scheme_migrations (
		name varchar(255) NOT NULL,
		version varchar(255) NOT NULL,
		applied_at datetime NOT NULL
	)`)
	return err
}

type schemeMigration struct {
	name      string
	version   string
	appliedAt time.Time
	Migration
}

func (sm schemeMigration) save(db *sql.DB) error {
	_, err := db.Exec(
		`INSERT INTO scheme_migrations ( name, version, applied_at) VALUES (?, ?, NOW())`,
		sm.Migration.Name(),
		sm.Migration.Version(),
	)
	return err
}

func (sm schemeMigration) destroy(db *sql.DB) error {
	_, err := db.Exec(
		`DELETE FROM scheme_migrations WHERE name = ? and version = ?`,
		sm.Migration.Name(),
		sm.Migration.Version(),
	)
	return err
}

func (sm schemeMigration) appliedAtStr() string {
	if sm.appliedAt.IsZero() {
		return ""
	}
	return sm.appliedAt.Format("2006-01-02 15:04:05")
}

type schemeMigrations []schemeMigration

func loadSchemeMigrations(ms []Migration, db *sql.DB) schemeMigrations {
	if err := createTable(db); err != nil {
		panic(err.Error())
	}
	sms := schemeMigrations{}

	rows, err := db.Query("SELECT * FROM scheme_migrations ORDER BY applied_at ASC;")
	if err != nil {
		panic(err.Error())
	}
	for rows.Next() {
		var sm schemeMigration
		if err := rows.Scan(&sm.name, &sm.version, &sm.appliedAt); err != nil {
			panic(err.Error())
		}
		sms = append(sms, sm)
	}

	if len(ms) < len(sms) {
		panic("not match migrations and scheme_migrations table!")
	}
	for i, m := range ms {
		if i < len(sms) {
			sm := sms[i]
			if m.Name() != sm.name || m.Version() != sm.version {
				panic("not match migrations and scheme_migrations table!")
			}
			sm.Migration = m
			sms[i] = sm
		} else {
			sms = append(sms, schemeMigration{Migration: m})
		}
	}
	return sms
}
