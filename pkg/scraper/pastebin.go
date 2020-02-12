package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Pastebin struct {
	Key      string
}

type PastebinSearchOptions struct {
	Terms       []string
	Limit       int
	WithContent bool
}

type PastebinResult struct {
	ScrapeUrl string `json:"scrape_url"`
	FullUrl   string `json:"full_url"`
	Date      string `json:"date"`
	Key       string `json:"key"`
	Size      string `json:"size"`
	Expire    string `json:"expire"`
	Title     string `json:"title"`
	Syntax    string `json:"syntax"`
	User      string `json:"user"`
	Raw       string `json:"raw"`
	Matched   string
}

type ScraperResponse []*PastebinResult

func getRawPasteData(key string) ([]byte, error) {
	url := fmt.Sprintf("https://pastebin.com/raw/%s", key)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("trouble requesting raw data %s: %w", url, err)
	}

	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

func filterBins(bins ScraperResponse, substrs []string) (filtered ScraperResponse) {
	for _, substr := range substrs {
		for _, bin := range bins {
			if strings.Contains(bin.Title, substr) || strings.Contains(bin.Raw, substr) {
				bin.Matched = substr
				filtered = append(filtered, bin)
			}
		}
	}

	return
}

func (pb *Pastebin) scrapeRecentPastebins(limit int) (*ScraperResponse, error) {
	url := fmt.Sprintf("https://scrape.pastebin.com/api_scraping.php?limit=%d", limit)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("trouble scraping pastebin api %s", url)
	}

	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	var scrapeResponse ScraperResponse

	err = json.Unmarshal(data, &scrapeResponse)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal response from scraper api: %s", string(data))
	}

	return &scrapeResponse, nil
}

func (pb *Pastebin) Search(options *PastebinSearchOptions) (*ScraperResponse, error) {
	response, err := pb.scrapeRecentPastebins(options.Limit)
	if err != nil {
		return nil, fmt.Errorf("error requesting recent pastebins: %v", err)
	}

	var wg sync.WaitGroup

	for _, bin := range *response {
		wg.Add(1)
		go func(bin *PastebinResult) {
			raw, _ := getRawPasteData(bin.Key)

			bin.Raw = string(raw)

			wg.Done()
		}(bin)
	}

	wg.Wait()

	bins := filterBins(*response, options.Terms)

	return &bins, nil
}

func NewPastebinScraper(key string) *Pastebin {
	return &Pastebin{
		Key:     key,
	}
}
