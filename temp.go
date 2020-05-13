package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

// GetURLFromPage gets next urls form a page
func GetURLFromPage(inp string) []string {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var n int

	if err := chromedp.Run(ctx,
		chromedp.Navigate(inp),
		chromedp.EvaluateAsDevTools("document.getElementsByClassName('poRVub').length;", &n),
	); err != nil {
		log.Fatal(err)
	}
	var out []string
	var temp string
	for i := 0; i < 2; i++ {
		url := fmt.Sprintf("document.getElementsByClassName('poRVub')[%d].href;", i)
		if err := chromedp.Run(ctx,
			chromedp.Navigate(inp),
			chromedp.EvaluateAsDevTools(url, &temp),
		); err != nil {
			log.Fatal(err)
		}
		// println(temp)
		out = append(out, temp)
	}
	return out

}
