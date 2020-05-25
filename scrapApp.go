package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// App contains app information
type App struct {
	Name        string `json:"Name"`
	Ratings     string `json:"Ratings"`
	Adds        string `json:"Adds"`
	Publisher   string `json:"Publisher"`
	PublisherID string `json:"PublisherID"`
	Installs    string `json:"Installs"`
	Genre       string `json:"Genre"`
	URL         string `json:"Url"`
	Email       string `json:"Email"`
	Updated     string `json:"Updated"`
	Size        string `json:"Size"`
	Logo        string `json:"Logo"`
}

var rApps int
var wApps int
var naApps int
var skipped int
var urlsLeft int

func getAppInfo(url string, AppsInfo chan App, Urls chan string, NextUrls chan string) {

	resp, err := http.Get(url)
	if err != nil {
		log.Println("url error", fmt.Sprint(err))
		skipped++
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
		if i == 0 {
			temp := strings.Split(s.Find("a.hrTbp.R8zArc").AttrOr("href", "404"), "=")
			app.PublisherID = temp[len(temp)-1]
		}
		if i < 2 {
			info[i] = s.Text()
		}
	})
	doc.Find("div.hAyfc").Each(func(i int, s *goquery.Selection) {

		if s.Find("div.BgcNfc").Text() == "Updated" {
			s.Find("span.htlgb").Each(func(x int, in *goquery.Selection) {
				app.Updated = in.Text()
			})
		} else if s.Find("div.BgcNfc").Text() == "Size" {
			s.Find("span.htlgb").Each(func(x int, in *goquery.Selection) {
				app.Size = in.Text()
			})
		}

	})
	app.Publisher = info[0]
	app.Genre = info[1]
	app.Logo = doc.Find("img.T75of.sHb2Xb").AttrOr("src", "404")
	app.Ratings = doc.Find("div.BHMmbe").Text()
	app.Adds = doc.Find("div.bSIuKf").Text()
	app.Installs = doc.Find("span.EymY4b").Text()
	app.URL = url
	app.Email = doc.Find("a.hrTbp.euBY6b").Text()
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
		rApps++
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
			urlStore[next] = true
			if ok && !prs {
				select {
				case NextUrls <- "https://play.google.com" + next:
					urlsLeft++
				default:
					skipped++
				}
			}
			mapMutex.Unlock()
		})
}
