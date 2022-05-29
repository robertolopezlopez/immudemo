package main

import (
	"context"
	"database/sql"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stdlib"
	"github.com/gin-gonic/gin"
	"github.com/robertolopezlopez/immudemo/authentication"
	"github.com/robertolopezlopez/immudemo/dao"
	"log"
	"net/http"
	"os"
)

type (
	dbase interface {
		/*
		 createTables creates the initial database structure in immudb
		*/
		CreateTables(context.Context) error

		/*
		 log pushes the given messages to immudb
		*/
		Log(context.Context, []string) error

		/*
		 find returns logged messages in reverse chronological order from immudb

		 optional: number of rows to fetch (string)
		*/
		Find(context.Context, string) ([]string, error)

		/*
		 count returns the number of logs stored in immudb
		*/
		Count(context.Context) (int, error)
	}

	PostBody struct {
		Msgs []string `json:"msgs"`
	}
)

var dataBase dbase

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

	dataBase = dao.DBase{
		Db: stdlib.OpenDB(opts),
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(dataBase.(dao.DBase).Db)

	err := dataBase.CreateTables(ctx)
	if err != nil {
		log.Fatalf("creating tables: %s", err)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: setupRouter(),
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.Use(authentication.HeaderAuthMiddleware())

	router.POST("/", writeLogs())

	router.GET("/", readLogs())

	router.GET("/count", countLogs())

	return router
}

func writeLogs() func(c *gin.Context) {
	return func(c *gin.Context) {
		var json PostBody
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(json.Msgs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "msgs should not be empty"})
			return
		}
		if err := dataBase.Log(c, json.Msgs); err != nil {
			_ = c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "message(s) successfully logged",
		})
		return
	}
}

func readLogs() func(c *gin.Context) {
	return func(c *gin.Context) {
		msgs, err := dataBase.Find(c, c.Query("n"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"logs": msgs,
		})
	}
}

func countLogs() func(c *gin.Context) {
	return func(c *gin.Context) {
		count, err := dataBase.Count(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	}
}
