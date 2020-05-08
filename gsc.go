package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"sync"

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

var nApps int

func main() {
	nApps = 0

	AppsInfo := make(chan App, 2000)
	Urls := make(chan string, 2500)
	TUrls := make(chan string, 2500)
	urlStore := make(map[string]bool)
	mapMutex := sync.RWMutex{}

	seed := "/store/apps/details?id=com.boxingclub.realpunch.free"

	mapMutex.Lock()
	urlStore[seed] = true
	mapMutex.Unlock()

	TUrls <- "https://play.google.com" + seed

	file, _ := os.OpenFile("Data.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 0; i < 5000; i++ {
		go getAppInfo(Urls, TUrls, AppsInfo)
	}

	for i := 0; i < 2500; i++ {
		go getUrls(Urls, TUrls, urlStore, &mapMutex)
	}

	for i := 0; i < 5000; i++ {
		go writeToCSV(AppsInfo, writer)
	}

	var first string
	fmt.Scanln(&first)

}

func writeToCSV(AppsInfo chan App, w *csv.Writer) {
	for app := range AppsInfo {
		w.Write([]string{app.name, app.publisher, app.installs, app.adds, app.genre, app.ratings, app.url})
		w.Flush()
	}
}

func getUrls(Urls chan string, TUrls chan string, urlStore map[string]bool, mapMutex *sync.RWMutex) {

	for url := range Urls {
		resp, err := http.Get(url)
		if err != nil {
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return
		}

		doc.Find("a.poRVub").Each(
			func(i int, s *goquery.Selection) {
				next, ok := s.Attr("href")
				mapMutex.Lock()
				_, prs := urlStore[next]
				mapMutex.Unlock()
				if ok && !prs {
					mapMutex.Lock()
					urlStore[next] = true
					mapMutex.Unlock()

					TUrls <- "https://play.google.com" + next
				}
			})
	}
	return
}

func getAppInfo(Urls chan string, TUrls chan string, AppsInfo chan App) {

	for url := range TUrls {

		Urls <- url
		resp, err := http.Get(url)
		if err != nil {
			// log.Panicln(err)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			// log.Panicln(err)
			return
		}

		app := App{}
		var info [2]string

		app.name = doc.Find("h1.AHFaub").Text()
		doc.Find("span.T32cc.UAO9ie").Each(func(i int, s *goquery.Selection) {
			if i > 1 {
				return
			}
			info[i] = s.Text()
		})
		app.publisher = info[0]
		app.genre = info[1]
		app.ratings = doc.Find("div.BHMmbe").Text()
		app.adds = doc.Find("div.bSIuKf").Text()
		app.installs = doc.Find("span.EymY4b").Text()
		app.url = url
		nApps++
		println(nApps)
		AppsInfo <- app

	}
	return
}
