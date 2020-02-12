package main

import (
	"flag"
	"fmt"
	"github.com/daviddiefenderfer/pastewatch/pkg/scraper"
	"log"
	"os"
	"strings"
	"time"
)

type termsArray []string

func (arr *termsArray) String() string {
	return strings.Join(*arr, ",")
}

func (arr *termsArray) Set(value string) error {
	*arr = append(*arr, value)
	return nil
}

func arrayContains(array []string, substr string) bool {
	var contains bool

	for _, item := range array {
		if item == substr {
			contains = true
		}
	}

	return contains
}

func watchPasteBins(key string, interval int, options *scraper.PastebinSearchOptions, resultsChannel chan<-*scraper.PastebinResult) {
	pb         := scraper.NewPastebinScraper(key)
	cachedKeys := make([]string, options.Limit, options.Limit)
	ticker     := time.NewTicker(time.Duration(interval) * time.Second)

	for ; true; <-ticker.C {
		log.Println("[Pastebin] Retrieving list of bins from pastebin.com")

		results, err := pb.Search(options)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			close(resultsChannel)
			break
		}

		if len(*results) == 0 {
			log.Println("[Pastebin] No new bins matching search criteria")
			continue
		}

		for _, result := range *results {
			if !arrayContains(cachedKeys, result.Key) {
				resultsChannel <- result
				cachedKeys = append([]string{result.Key}, cachedKeys[:len(cachedKeys)-1]...)
			}
		}
	}
}

func main() {
	var term termsArray

	pastebinKey      := flag.String("pastebin-key", "", "Pastebin Developer API Key.")
	requestInterval  := flag.Int("interval", 60, "Interval to request bins at.")
	requestLimit	 := flag.Int("limit", 100, "Number of bins to retrieve per request.")

	flag.Var(&term, "term", "Terms to watch new pastebins for.")

	flag.Parse()

	resultsChannel := make(chan *scraper.PastebinResult)

	if term == nil {
		term = []string{""}
	}

	opts := &scraper.PastebinSearchOptions{
		Terms: term,
		Limit: *requestLimit,
		WithContent: true,
	}

	log.Printf("[Pastebin] Watching new bins for match on %s\n", opts.Terms)
	go watchPasteBins(*pastebinKey, *requestInterval, opts, resultsChannel)

	for result := range resultsChannel {
		log.Printf("[Pastebin] Bin matched search criteria: %s - Title: %s - URL: %s\n", result.Matched, result.Title, result.FullUrl)
	}
}
