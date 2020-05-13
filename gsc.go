package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/knq/chromedp"
	_ "github.com/lib/pq"
)

// App contains app information
type App struct {
	Name      string `json:"Name"`
	Ratings   string `json:"Ratings"`
	Adds      string `json:"Adds"`
	Publisher string `json:"Publisher"`
	Installs  string `json:"Installs"`
	Genre     string `json:"Genre"`
	URL       string `json:"Url"`
}

var rApps int
var wApps int
var naApps int
var skipped int
var urlsLeft int

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

func getAppInfo(url string, AppsInfo chan App, Urls chan string, NextUrls chan string) {

	resp, err := http.Get(url)
	if err != nil {
		log.Printf(fmt.Sprint(err))
		// time.Sleep(2000 * time.Millisecond)

		select {
		case NextUrls <- url:
		default:
			time.Sleep(2000 * time.Millisecond)
			skipped++
		}

		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf(fmt.Sprint(err))
		// time.Sleep(2000 * time.Millisecond)

		select {
		case NextUrls <- url:
		default:
			time.Sleep(2000 * time.Millisecond)
			skipped++
		}

		return
	}

	app := App{}
	var info [2]string
	app.Name = doc.Find("h1.AHFaub").Text()
	doc.Find("span.T32cc.UAO9ie").Each(func(i int, s *goquery.Selection) {
		if i < 2 {
			info[i] = s.Text()
		}
	})
	app.Publisher = info[0]
	app.Genre = info[1]
	app.Ratings = doc.Find("div.BHMmbe").Text()
	app.Adds = doc.Find("div.bSIuKf").Text()
	app.Installs = doc.Find("span.EymY4b").Text()
	app.URL = url
	rApps++
	if app.Name == "" && app.Publisher == "" {
		naApps++
		select {
		case NextUrls <- url:
		default:
			time.Sleep(1500 * time.Millisecond)
			urlsLeft--
		}

	} else {
		AppsInfo <- app
		urlsLeft--
	}

	Urls <- url

}

func writeToCSV(AppsInfo chan App, db *sql.DB) {
	for app := range AppsInfo {
		x, err := json.Marshal(app)
		if err != nil {
			panic(err)
		}

		_, err = db.Exec(sqlStatement, app.URL, x)
		if err == nil {
			wApps++
			println(rApps, wApps, naApps, skipped, urlsLeft)
		}
		// println(rApps, wApps, naApps, skipped, urlsLeft)

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

	AppsInfo := make(chan App)           //, 500)
	Urls := make(chan string, 500)       //, 1000)
	NextUrls := make(chan string, 50000) //, 20000)

	urlStore := make(map[string]bool)
	mapMutex := sync.RWMutex{}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")

	t := time.Now()

	feedSeedurl(Urls)

	for i := 0; i < 10; i++ {
		go func() {
			opts := append(chromedp.DefaultExecAllocatorOptions[:],
				chromedp.Flag("headless", true),
				chromedp.Flag("disable-gpu", false),
				chromedp.Flag("enable-automation", false),
				chromedp.Flag("disable-extensions", true),
			)

			allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
			defer cancel()

			ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
			defer cancel()
			for url := range Urls {
				getNextUrls(ctx, url, NextUrls, urlStore, &mapMutex) // go to each url to get NextUrls
				// time.Sleep(2 * time.Second)
			}
		}()
	}

	for i := 0; i < 300; i++ {
		go func() {
			for url := range NextUrls {
				getAppInfo(url, AppsInfo, Urls, NextUrls) // go to each url to get info and find more urls
				// time.Sleep(2 * time.Second)
			}
		}()

	}

	go writeToCSV(AppsInfo, db) // write the Apps in AppsInfo to a csv file

	// Finishing Tasks
	var inp string
	fmt.Scanln(&inp)

	go func() {
		fmt.Println("Next Url:", <-NextUrls)
	}()

	fmt.Scanln(&inp)

	elapsed := time.Since(t)
	fmt.Printf("\nTime to scrape %d Apps is %v\n", wApps, elapsed)

}

func getNextUrls(ctx context.Context, url string, NextUrls chan string, urlStore map[string]bool, mapMutex *sync.RWMutex) {

	var out []string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.EvaluateAsDevTools(`Array.from(document.getElementsByClassName("poRVub")).map(a => a.href);`, &out),
	); err != nil {
		log.Fatal(err)
	}
	for _, next := range out {
		mapMutex.Lock()
		_, prs := urlStore[next]
		mapMutex.Unlock()
		if !prs {
			mapMutex.Lock()

			select {
			case NextUrls <- next:
				urlStore[next] = true
				urlsLeft++
			default:
				time.Sleep(2000 * time.Millisecond)
				skipped++
			}

			mapMutex.Unlock()

		}
	}

}

func feedSeedurl(Urls chan string) {
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

	go func() {
		for i := 0; i < 37; i++ {
			Urls <- seed[i]
			// time.Sleep(10 * time.Second)

		}
	}()

	go func() {
		time.Sleep(25 * time.Minute)
		inp := "a"

		for i := 0; i < 500000; i++ {
			Urls <- fmt.Sprintf("https://play.google.com/store/search?q=%s&c=apps", inp)
			inp = biggerStr(inp)
			time.Sleep(3 * time.Second)
		}

	}()

}

func biggerStr(a string) string {
	n := len(a) - 1
	out := ""
	add := 1
	for add > 0 || n >= 0 {
		if n < 0 {
			out = "a" + out
			add = 0
		} else {
			out = string(97+(int(a[n])-int('a')+add)%26) + out
			if add == 1 && a[n] != 'z' {
				add = 0
			}
		}
		n--
	}
	return out
}
