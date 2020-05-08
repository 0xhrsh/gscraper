package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// App contains app information
type App struct {
	name      string
	ratings   string
	adds      string
	publisher string
	installs  string
	genre     string
	url       string
}

var rApps int
var wApps int
var naApps int
var skipped int

func writeToCSV(AppsInfo chan App, w *csv.Writer) {
	for app := range AppsInfo {
		w.Write([]string{app.name, app.publisher, app.installs, app.adds, app.genre, app.ratings, app.url})
		w.Flush()
		wApps++
		println(rApps, wApps, naApps, skipped)
	}
}

func checkError(err error) {
	if err != nil {
		log.Println(err) // Panicln(err)
	}
}

func main() {
	rApps = 0
	wApps = 0
	naApps = 0

	AppsInfo := make(chan App, 5000)
	Urls := make(chan string, 5000)
	NextUrls := make(chan string, 20000)

	urlStore := make(map[string]bool)
	mapMutex := sync.RWMutex{}

	file, _ := os.OpenFile("Data.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	var seed [37]string
	seed[0] = "https://play.google.com/store/apps/top"
	seed[1] = "https://play.google.com/store/apps"
	seed[2] = "https://play.google.com/store/apps/new"
	seed[3] = "https://play.google.com/store/apps/category/AUTO_AND_VEHICLES"
	seed[4] = "https://play.google.com/store/apps/category/BOOKS_AND_REFERENCE"
	seed[5] = "https://play.google.com/store/apps/category/BUSINESS"
	seed[6] = "https://play.google.com/store/apps/stream/baselist_featured_arcore"
	seed[7] = "https://play.google.com/store/apps/category/COMICS"
	seed[8] = "https://play.google.com/store/apps/category/COMMUNICATION"
	seed[9] = "https://play.google.com/store/apps/category/DATING"
	seed[10] = "https://play.google.com/store/apps/stream/vr_top_device_featured_category"
	seed[11] = "https://play.google.com/store/apps/category/EDUCATION"
	seed[12] = "https://play.google.com/store/apps/category/ENTERTAINMENT"
	seed[13] = "https://play.google.com/store/apps/category/EVENTS"
	seed[14] = "https://play.google.com/store/apps/category/FINANCE"
	seed[15] = "https://play.google.com/store/apps/category/FOOD_AND_DRINK"
	seed[16] = "https://play.google.com/store/apps/category/GAME"
	seed[17] = "https://play.google.com/store/apps/category/FAMILY"
	seed[18] = "https://play.google.com/store/apps/category/FAMILY?age=AGE_RANGE2"
	seed[19] = "https://play.google.com/store/apps/category/FAMILY_ACTION"
	seed[20] = "https://play.google.com/store/apps/category/FAMILY_BRAINGAMES"
	seed[21] = "https://play.google.com/store/apps/category/FAMILY_CREATE"
	seed[22] = "https://play.google.com/store/apps/category/FAMILY_EDUCATION"
	seed[23] = "https://play.google.com/store/apps/category/GAME_CASUAL"
	seed[24] = "https://play.google.com/store/apps/category/GAME_SPORTS"
	seed[25] = "https://play.google.com/store/apps/category/GAME_SIMULATION"
	seed[26] = "https://play.google.com/store/apps/category/GAME_ROLE_PLAYING"
	seed[27] = "https://play.google.com/store/apps/category/GAME_ARCADE"
	seed[28] = "https://play.google.com/store/apps/category/GAME_ADVENTURE"
	seed[29] = "https://play.google.com/store/apps/category/GAME_BOARD"
	seed[30] = "https://play.google.com/store/apps/category/GAME_CARD"
	seed[31] = "https://play.google.com/store/apps/category/GAME_CASINO"
	seed[32] = "https://play.google.com/store/apps/category/GAME_EDUCATIONAL"
	seed[33] = "https://play.google.com/store/apps/category/GAME_MUSIC"
	seed[34] = "https://play.google.com/store/apps/category/GAME_PUZZLE"
	seed[35] = "https://play.google.com/store/apps/category/GAME_RACING"
	seed[36] = "https://play.google.com/store/apps/category/BEAUTY"

	t := time.Now()
	go func() {
		for i := 0; i < 37; i++ {
			Urls <- seed[i]

		}
	}()

	for i := 0; i < 1000; i++ {
		go func() {
			for url := range Urls {
				getNextUrls(url, NextUrls, urlStore, &mapMutex) // go to each url to get NextUrls
			}
		}()
	}

	for i := 0; i < 2000; i++ {
		go func() {
			for url := range NextUrls {
				getAppInfo(url, AppsInfo, Urls, NextUrls) // go to each url to get info and find more urls
			}
		}()

	}

	go writeToCSV(AppsInfo, writer) // write the Apps in AppsInfo to a csv file

	var inp string
	fmt.Scanln(&inp)
	elapsed := time.Since(t)
	fmt.Printf("Time to scrape %d Apps is %v\n", wApps, elapsed)

}

func getNextUrls(url string, NextUrls chan string, urlStore map[string]bool, mapMutex *sync.RWMutex) {
	time.Sleep(1950 * time.Millisecond)
	resp, err := http.Get(url)
	checkError(err)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkError(err)

	doc.Find("a.poRVub").Each(
		func(i int, s *goquery.Selection) {
			next, ok := s.Attr("href")
			mapMutex.Lock()
			_, prs := urlStore[next]
			mapMutex.Unlock()
			if ok && !prs {
				mapMutex.Lock()

				select {
				case NextUrls <- "https://play.google.com" + next:
					urlStore[next] = true
				default:
					skipped++
				}

				mapMutex.Unlock()

			}
		})

}

func getAppInfo(url string, AppsInfo chan App, Urls chan string, NextUrls chan string) {
	time.Sleep(1950 * time.Millisecond)
	resp, err := http.Get(url)
	checkError(err)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkError(err)

	app := App{}
	var info [2]string
	app.name = doc.Find("h1.AHFaub").Text()
	doc.Find("span.T32cc.UAO9ie").Each(func(i int, s *goquery.Selection) {
		if i < 2 {
			info[i] = s.Text()
		}
	})
	app.publisher = info[0]
	app.genre = info[1]
	app.ratings = doc.Find("div.BHMmbe").Text()
	app.adds = doc.Find("div.bSIuKf").Text()
	app.installs = doc.Find("span.EymY4b").Text()
	app.url = url

	if app.name == "" && app.publisher == "" {
		naApps++
		time.Sleep(550 * time.Millisecond)
	} else {
		AppsInfo <- app
		time.Sleep(150 * time.Millisecond)
	}
	Urls <- url
	rApps++

}
