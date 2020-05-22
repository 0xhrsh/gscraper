package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
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

func getAppInfo(url string, AppsInfo chan App, Urls chan string, NextUrls chan string) {

	resp, err := http.Get(url)
	if err != nil {
		log.Printf(fmt.Sprint(err))
		select {
		case NextUrls <- url:
		default:
			skipped++
		}
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf(fmt.Sprint(err))
		select {
		case NextUrls <- url:
		default:
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

func getNextUrls(url string, NextUrls chan string, urlStore map[string]bool, mapMutex *sync.RWMutex) {

	resp, err := http.Get(url)
	if err != nil {
		log.Printf(fmt.Sprint(err))
		time.Sleep(2000 * time.Millisecond)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf(fmt.Sprint(err))
		time.Sleep(2000 * time.Millisecond)
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

				select {
				case NextUrls <- "https://play.google.com" + next:
					urlStore[next] = true
					urlsLeft++
				default:
					time.Sleep(2000 * time.Millisecond)
					skipped++
				}
				mapMutex.Unlock()
			}
		})
}