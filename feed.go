package main

import (
	"fmt"
	"sync"
)

func feedSeedurl(Urls chan string, NextUrls chan string, urlStore map[string]bool, mapMutex *sync.RWMutex) {
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

	dumpUrls := make(chan string)

	go func() {
		inp := "aaaa"
		for i := 0; i < 500000000; i++ {
			dumpUrls <- fmt.Sprintf("https://play.google.com/store/search?q=%s&c=apps", inp)
			inp = biggerStr(inp)
		}

	}()

	go func() {
		for i := 0; i < 37; i++ {
			dumpUrls <- seed[i]
		}
	}()

	go parseDumpPages(dumpUrls, NextUrls, urlStore, mapMutex)

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
