package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "admin"
	password = "admin"
	dbname   = "apps"
)

var sqlStatement = `
INSERT INTO apps (url, data)
VALUES ($1, $2)`

func writeToPG(AppsInfo chan App, db *sql.DB) {
	for app := range AppsInfo {
		x, err := json.Marshal(app)
		if err != nil {
			panic(err)
		}

		_, err = db.Exec(sqlStatement, app.URL, x)
		if err == nil {
			wApps++
		}
		if wApps > 0 {
			fmt.Println(rApps, wApps, naApps, skipped, urlsLeft, float32(rApps)/float32(wApps))
		} else {
			fmt.Println(rApps, wApps, naApps, skipped, urlsLeft, err)
		}

	}
}

func connectToServer() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")

	return db
}
