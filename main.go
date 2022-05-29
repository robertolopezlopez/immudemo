package main

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stdlib"
	"github.com/gin-gonic/gin"
)

type (
	dbase interface {
		createTables(context.Context) error

		/*
		 * log pushes the given messages to
		 */
		log(context.Context, []string) error

		find(context.Context, string) ([]string, error)

		count(context.Context) (int, error)
	}

	DBase struct {
		db *sql.DB
	}

	PostBody struct {
		Msgs []string `json:"msgs"`
	}
)

func main() {
	ctx := context.Background()

	opts := client.DefaultOptions()
	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"
	if db, ok := os.LookupEnv("DB"); ok {
		opts.Address = db
	} else {
		opts.Address = ""
	}

	dataBase := DBase{
		db: stdlib.OpenDB(opts),
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(dataBase.db)

	err := dataBase.createTables(ctx)
	if err != nil {
		log.Fatalf("creating tables: %s", err)
	}

	router := gin.Default()
	router.POST("/", func(c *gin.Context) {
		var json PostBody
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(json.Msgs) == 0 {
			c.JSON(http.StatusBadRequest, "msgs should not be empty")
			return
		}
		if err := dataBase.log(c, json.Msgs); err != nil {
			_ = c.Error(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "message(s) successfully logged",
		})
		return
	})
	router.GET("/", func(c *gin.Context) {
		msgs, err := dataBase.find(c, c.Query("n"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"logs": msgs,
		})
	})
	router.GET("/count", func(c *gin.Context) {
		count, err := dataBase.count(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

func (db *DBase) createTables(ctx context.Context) error {
	_, err := db.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS logs(id INTEGER, ts TIMESTAMP, message VARCHAR, PRIMARY KEY id)")
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS ON logs(ts)")
	return err
}

func (db *DBase) log(ctx context.Context, msgs []string) error {
	tx, err := db.db.BeginTx(ctx, nil)
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

func (db *DBase) find(ctx context.Context, n string) ([]string, error) {
	var rows *sql.Rows
	var err error

	if n != "" {
		i, err := strconv.Atoi(n)
		if err != nil {
			return nil, err
		}
		rows, err = db.db.QueryContext(ctx, "SELECT * FROM logs ORDER BY ts DESC LIMIT ?", i)
	} else {
		rows, err = db.db.QueryContext(ctx, "SELECT * FROM logs ORDER BY ts")
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

func (db *DBase) count(ctx context.Context) (int, error) {
	var c int
	err := db.db.QueryRowContext(ctx, "SELECT COUNT(*) AS c FROM logs").Scan(&c)
	if err != nil {
		return 0, err
	}
	return c, nil
}
