package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/knq/chromedp"
)

func parseDumpPages(dumpUrls chan string, NextUrls chan string, urlStore map[string]bool, mapMutex *sync.RWMutex) {

	for i := 0; i < 5; i++ {
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

			for url := range dumpUrls {
				var out []string
				var n int
				var res []byte
				if err := chromedp.Run(ctx,
					chromedp.Navigate(url),
					chromedp.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`, &res),
					chromedp.Sleep(2500*time.Millisecond),
					chromedp.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`, &res),
					chromedp.Sleep(2500*time.Millisecond),
					chromedp.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`, &res),
					chromedp.Sleep(2500*time.Millisecond),
					chromedp.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`, &res),
					chromedp.Sleep(2500*time.Millisecond),
					chromedp.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`, &res),
					chromedp.Sleep(2000*time.Millisecond),
					chromedp.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`, &res),
					chromedp.Sleep(1500*time.Millisecond),
					chromedp.EvaluateAsDevTools(`document.getElementsByClassName("poRVub").length;`, &n),
					chromedp.EvaluateAsDevTools(`Array.from(document.getElementsByClassName("poRVub")).map(a => a.href);`, &out),
				); err != nil {
					log.Fatal(err)
				}
				fmt.Println("====>", n)
				for _, next := range out {
					mapMutex.Lock()
					_, prs := urlStore[next]
					urlStore[next] = true
					if !prs {
						select {
						case NextUrls <- next:
							urlsLeft++
						default:
							skipped++
						}
					}
					mapMutex.Unlock()

				}
			}
		}()
	}
}
