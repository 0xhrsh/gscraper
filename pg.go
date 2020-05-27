package main

import (
	"database/sql"
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
INSERT INTO apps (name, ratings, ads, publisher, publisherId, installs, genre, url, email, updated, size, logo)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

func writeToPG(AppsInfo chan App, db *sql.DB) {
	for app := range AppsInfo {

		_, err := db.Exec(sqlStatement, app.Name, app.Ratings, app.Ads, app.Publisher, app.PublisherID, app.Installs, app.Genre, app.URL, app.Email, app.Updated, app.Size, app.Logo)
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
