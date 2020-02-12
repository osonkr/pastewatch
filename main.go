package main

import (
	"flag"
	"fmt"
	"github.com/daviddiefenderfer/pastewatch/pkg/scraper"
	"log"
	"os"
	"strconv"
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

func defaultIntUnlessEnv(key string, defaultValue int) int {
	if value, err := strconv.Atoi(defaultUnlessEnv(key, strconv.Itoa(defaultValue))); err != nil {
		return value
	}

	return defaultValue
}

func defaultUnlessEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists != false {
		return value
	}

	return defaultValue
}

func main() {
	var term termsArray

	defaultInterval   := defaultIntUnlessEnv("REQUEST_INTERVAL", 30)
	defaultLimit      := defaultIntUnlessEnv("REQUEST_LIMIT", 100)
	defaultPastbinKey := defaultUnlessEnv("PASTEBINKEY", "")

	pastebinKey      := flag.String("pastebin-key", defaultPastbinKey, "Pastebin Developer API Key.")
	requestInterval  := flag.Int("interval", defaultInterval, "Interval to request bins at.")
	requestLimit	 := flag.Int("limit", defaultLimit, "Number of bins to retrieve per request.")

	flag.Var(&term, "term", "Terms to watch new pastebins for.")

	flag.Parse()

	if *pastebinKey == "" {
		fmt.Println("Pastebin key required")
		os.Exit(1)
	}

	if term == nil {
		term = strings.Split(defaultUnlessEnv("TERMS", ""), ",")
	}

	opts := &scraper.PastebinSearchOptions{
		Terms: term,
		Limit: *requestLimit,
		WithContent: true,
	}

	resultsChannel := make(chan *scraper.PastebinResult)

	log.Printf("[Pastebin] Watching new bins for match on %s\n", opts.Terms)
	go watchPasteBins(*pastebinKey, *requestInterval, opts, resultsChannel)

	for result := range resultsChannel {
		log.Printf("[Pastebin] Bin matched search criteria: %s - Title: %s - URL: %s\n", result.Matched, result.Title, result.FullUrl)
	}
}
