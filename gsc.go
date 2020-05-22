package main

import (
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

func main() {

	AppsInfo := make(chan App, 100)
	Urls := make(chan string, 50000)
	NextUrls := make(chan string, 3000000)

	urlStore := make(map[string]bool)
	mapMutex := sync.RWMutex{}

	db := connectToServer()
	defer db.Close()
	t := time.Now()

	feedSeedurl(Urls, NextUrls, urlStore, &mapMutex)

	for i := 0; i < 150; i++ {
		go func() {
			for url := range Urls {
				getNextUrls(url, NextUrls, urlStore, &mapMutex)
			}
		}()
	}

	for i := 0; i < 4500; i++ {
		go func() {
			for url := range NextUrls {
				getAppInfo(url, AppsInfo, Urls, NextUrls)
			}
		}()
	}

	go writeToPG(AppsInfo, db)

	time.Sleep(72 * time.Hour)
	elapsed := time.Since(t)
	fmt.Printf("\nTime to scrape %d Apps is %v\n", wApps, elapsed)

}
