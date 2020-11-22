package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/alyrot/menuToText/pkg/parser"
)

/*
invDateError is returned when MenuCache deems a date to far in the future or to far in the past
*/
var invDateError = errors.New("date in invalid range")

/*
roundToDay helper function that truncates time from t
*/
func roundToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

type MenuCacher interface {
	GetMenu(date time.Time) ([]*parser.Dish, error)
	Refresh() error
}

/*
MenuCache is a cached Data Model for parser.Dish values served on a day
*/
type MenuCache struct {
	lock         sync.RWMutex
	dateToDishes map[time.Time][]*parser.Dish
	download     Downloader
	parse        parser.UKSHParserI
	errorLog     *log.Logger
	infoLog      *log.Logger
}

/*
NewMenuCache creates and fills a new MenuCache.
*/
func NewMenuCache(errorLog, infoLog *log.Logger) (*MenuCache, error) {
	mc := &MenuCache{
		lock:         sync.RWMutex{},
		dateToDishes: nil,
		download:     &realDownloader{},
		parse:        &parser.UKSHParser{},
		errorLog:     errorLog,
		infoLog:      infoLog,
	}
	if err := mc.Refresh(); err != nil {
		return nil, fmt.Errorf("failed to initialize cache with data: %v", err)
	}
	return mc, nil
}

/*
Refresh, fetches the current menuHandler, parses it and completely rebuilds the cache. Calling function may not
hold mc.lock.RLock() as we call mc.lock.Lock()
*/
func (mc *MenuCache) Refresh() error {
	mc.infoLog.Println("Executing MenuCache.Refresh")
	mc.lock.Lock()
	defer mc.lock.Unlock()
	pdfs, err := extractPDFsFromMenuSite(mc.download)
	if err != nil {
		return err
	}

	//clear cache
	mc.dateToDishes = make(map[time.Time][]*parser.Dish)

	//rebuild cache
	for i := range pdfs {
		dishes, err := mc.parse.PDFToDishes(pdfs[i])
		if err != nil {
			return err
		}

		for j := range dishes {
			list, ok := mc.dateToDishes[roundToDay(dishes[j].Date)]
			if !ok {
				mc.dateToDishes[roundToDay(dishes[j].Date)] = []*parser.Dish{dishes[j]}
			} else {
				list = append(list, dishes[j])
				mc.dateToDishes[roundToDay(dishes[j].Date)] = list
			}
		}
	}
	return nil
}

/*
GetMenu, returns the dishes for date if they have been published yet
*/
func (mc *MenuCache) GetMenu(date time.Time) ([]*parser.Dish, error) {
	date = roundToDay(date)

	mc.lock.RLock()
	dishes, ok := mc.dateToDishes[date]
	mc.lock.RUnlock()
	if !ok {
		//only refresh if date is in valid range
		if date.Before(roundToDay(time.Now().In(time.Local))) {
			return nil, fmt.Errorf("GetMenu: %w: %v is in the past", invDateError, date)
		}
		if date.After(roundToDay(time.Now()).Add(7 * 24 * time.Hour)) {
			return nil, fmt.Errorf("GetMenu: %w: %v is more than 7 days in the future", invDateError, date)
		}
		//Refresh cache; if still not there return error
		if err := mc.Refresh(); err != nil {
			return nil, fmt.Errorf("GetMenu: failed to refresh: %v", err)
		}
		mc.lock.RLock()
		dishes, ok := mc.dateToDishes[date]
		mc.lock.RUnlock()
		if !ok {
			return nil, fmt.Errorf("GetMenu: failed to get dishes for %v", date)
		}
		return dishes, nil
	}
	return dishes, nil
}

/*
Abstraction of http Get requests for testing
*/
type Downloader interface {
	Get(url string) ([]byte, error)
}

type realDownloader struct{}

/*
Get, does a http get request to url, consumes the body and returns it
*/
func (d *realDownloader) Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get for %v failed: %v", url, err)
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %v", err)
	}
	return respBytes, nil
}

/*
extractLinks, extracts the UKSH menuHandler download links from site
*/
func extractLinks(site []byte) ([]string, error) {
	//missing slash at the end is important
	const hostPrefix = "https://www.uksh.de"
	hrefRegexp := regexp.MustCompile(`<a.+>Speiseplan Bistro.+<\/a>`)
	linkFromHref := regexp.MustCompile(`\/.+\.pdf`)

	menuHREFs := hrefRegexp.FindAll(site, -1)
	if menuHREFs == nil {
		return nil, fmt.Errorf("exractLinks: menuHandler href regexp found no matches\n")
	}

	links := make([]string, 0, len(menuHREFs))
	for i := range menuHREFs {
		linkBytes := linkFromHref.Find(menuHREFs[i])
		if linkBytes == nil {
			return nil, fmt.Errorf("extractLinks: failed to extract link from partial match %v\n", menuHREFs[i])
		}
		//note that hostPrefix has no trailing slash and linkFromHref assures we start with a slash
		links = append(links, hostPrefix+string(linkBytes))
	}

	return links, nil
}

/*
extractPDFsFromMenuSite, downloads the lunch menuHandler PDFs from the UKSH website
*/
func extractPDFsFromMenuSite(d Downloader) ([][]byte, error) {
	const menuBaseURL = "https://www.uksh.de/servicesternnord/Unser+Speisenangebot/Speisepl%C3%A4ne+L%C3%BCbeck/UKSH_Bistro+L%C3%BCbeck-p-346.html"
	site, err := d.Get(menuBaseURL)
	if err != nil {
		return nil, fmt.Errorf("extractPDFsFromMenuSite: failed to fetch site: %v", err)
	}
	links, err := extractLinks(site)
	if err != nil {
		return nil, fmt.Errorf("extractPDFsFromMenuSite: failed to extract links: %v", err)
	}

	if l := len(links); l > 2 {
		return nil, fmt.Errorf("extractPDFsFromMenuSite: expected 2 links, got %v", l)
	}

	pdfs := make([][]byte, 0, len(links))
	for i := range links {
		tmp, err := d.Get(links[i])
		if err != nil {
			return nil, fmt.Errorf("extractPDFsFromMenuSite: failed to fetch pdf: %v", err)
		}
		pdfs = append(pdfs, tmp)
	}

	return pdfs, nil
}
