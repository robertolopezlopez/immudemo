package dao

import (
	"context"
	"database/sql"
	"math/rand"
	"strconv"
	"time"
)

type (
	db interface {
		createTable(context.Context) error
		createIndex(context.Context) error
	}

	sqlDB struct {
		Db *sql.DB
	}

	DBase struct {
		dbWrapper db
		Db        *sql.DB
	}
)

func (d *sqlDB) createTable(ctx context.Context) error {
	_, err := d.Db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS logs(id INTEGER, ts TIMESTAMP, message VARCHAR, PRIMARY KEY id)")
	return err
}

func (d *sqlDB) createIndex(ctx context.Context) error {
	_, err := d.Db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS ON logs(ts)")
	return err
}

func (db DBase) CreateTables(ctx context.Context) error {
	if db.dbWrapper == nil {
		db.dbWrapper = &sqlDB{Db: db.Db}
	}
	if err := db.dbWrapper.createTable(ctx); err != nil {
		return err
	}
	return db.dbWrapper.createIndex(ctx)
}

func (db DBase) Log(ctx context.Context, msgs []string) error {
	tx, err := db.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ts := time.Now()
	for _, msg := range msgs {
		_, err = tx.Exec("INSERT INTO logs(id, ts, message) values (?, ?, ?)", r.Uint64(), ts, msg)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db DBase) Find(ctx context.Context, n string) ([]string, error) {
	var rows *sql.Rows
	var err error

	if n != "" {
		i, err := strconv.Atoi(n)
		if err != nil {
			return nil, err
		}
		rows, err = db.Db.QueryContext(ctx, "SELECT * FROM logs ORDER BY ts DESC LIMIT ?", i)
	} else {
		rows, err = db.Db.QueryContext(ctx, "SELECT * FROM logs ORDER BY ts")
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var id, ts, message string
	msgs := []string{}

	for {
		ok := rows.Next()
		if !ok {
			break
		}
		err = rows.Scan(&id, &ts, &message)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, message)
	}

	return msgs, nil
}

func (db DBase) Count(ctx context.Context) (int, error) {
	var c int
	err := db.Db.QueryRowContext(ctx, "SELECT COUNT(*) AS c FROM logs").Scan(&c)
	if err != nil {
		return 0, err
	}
	return c, nil
}
